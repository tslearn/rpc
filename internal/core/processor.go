package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	rootName               = "#"
	freeGroups             = 512
	processorStatusClosed  = 0
	processorStatusRunning = 1
)

var (
	nodeNameRegex   = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	actionNameRegex = regexp.MustCompile(
		`^([_a-zA-Z][_0-9a-zA-Z]*)|(\$onMount)|(\$onUnmount)|(\$onUpdateConfig)$`,
	)
)

type rpcActionNode struct {
	path       string
	meta       *rpcActionMeta
	service    *rpcServiceNode
	cacheFN    ActionCacheFunc
	reflectFn  reflect.Value
	callString string
	argTypes   []reflect.Type
	indicator  *base.PerformanceIndicator
}

type rpcServiceNode struct {
	path    string
	addMeta *ServiceMeta
	depth   uint16
	isMount bool
	data    Map
	sync.Mutex
}

func (p *rpcServiceNode) GetConfig(key string) (Any, bool) {
	v, ok := p.data[key]
	return v, ok
}

func (p *rpcServiceNode) SetConfig(key string, value Any) bool {
	if p.data != nil {
		p.data[key] = value
		return true
	} else {
		return false
	}
}

// Processor ...
type Processor struct {
	status            int32
	actionsMap        map[string]*rpcActionNode
	servicesMap       map[string]*rpcServiceNode
	maxNodeDepth      uint16
	maxCallDepth      uint16
	threads           []*rpcThread
	systemThread      *rpcThread
	freeCHArray       []chan *rpcThread
	readThreadPos     uint64
	writeThreadPos    uint64
	panicSubscription *base.PanicSubscription
	fnError           func(err *base.Error)
	closeCH           chan string
}

// NewProcessor ...
func NewProcessor(
	numOfThreads int,
	maxNodeDepth int16,
	maxCallDepth int16,
	threadBufferSize uint32,
	fnCache ActionCache,
	closeTimeout time.Duration,
	mountServices []*ServiceMeta,
	onReturnStream func(stream *Stream),
) (*Processor, *base.Error) {
	if onReturnStream == nil {
		return nil, errors.ErrProcessorOnReturnStreamIsNil
	}

	if numOfThreads <= 0 {
		return nil, errors.ErrNumOfThreadsIsWrong
	} else if maxNodeDepth <= 0 {
		return nil, errors.ErrMaxNodeDepthIsWrong
	} else if maxCallDepth <= 0 {
		return nil, errors.ErrProcessorMaxCallDepthIsWrong
	} else {
		fnError := func(err *base.Error) {
			defer func() {
				_ = recover()
			}()
			stream := NewStream()
			stream.WriteUint64(err.GetCode())
			stream.WriteString(err.GetMessage())
			onReturnStream(stream)
		}

		size := ((numOfThreads + freeGroups - 1) / freeGroups) * freeGroups
		ret := &Processor{
			status:         processorStatusRunning,
			actionsMap:     make(map[string]*rpcActionNode),
			servicesMap:    make(map[string]*rpcServiceNode),
			maxNodeDepth:   uint16(maxNodeDepth),
			maxCallDepth:   uint16(maxCallDepth),
			threads:        make([]*rpcThread, size),
			freeCHArray:    nil,
			readThreadPos:  0,
			writeThreadPos: 0,
			fnError:        fnError,
			closeCH:        make(chan string),
		}

		// subscribe panic
		ret.panicSubscription = base.SubscribePanic(fnError)

		// init system thread
		ret.systemThread = newThread(
			ret,
			closeTimeout,
			2048,
			func(stream *Stream) {},
			func(thread *rpcThread) {},
		)

		// mount nodes
		ret.servicesMap[rootName] = &rpcServiceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
			data:    Map{},
		}

		for _, meta := range mountServices {
			if err := ret.mountNode(rootName, meta, fnCache); err != nil {
				return nil, err
			}
		}

		// start config update
		go func() {
			counter := uint64(0)
			for atomic.LoadInt32(&ret.status) == processorStatusRunning {
				time.Sleep(100 * time.Millisecond)
				counter++
				if counter%40 == 0 {
					ret.onUpdateConfig()
				}
			}
			ret.closeCH <- ""
		}()

		// start threads
		freeCHArray := make([]chan *rpcThread, freeGroups)
		for i := 0; i < freeGroups; i++ {
			freeCHArray[i] = make(chan *rpcThread, size/freeGroups)
		}
		ret.freeCHArray = freeCHArray

		for i := 0; i < size; i++ {
			thread := newThread(
				ret,
				closeTimeout,
				threadBufferSize,
				onReturnStream,
				func(thread *rpcThread) {
					defer func() {
						_ = recover()
					}()
					freeCHArray[atomic.AddUint64(
						&ret.writeThreadPos,
						1,
					)%freeGroups] <- thread
				},
			)
			ret.threads[i] = thread
			ret.freeCHArray[i%freeGroups] <- thread
		}

		return ret, nil
	}
}

// Close ...
func (p *Processor) Close() bool {
	if atomic.CompareAndSwapInt32(
		&p.status,
		processorStatusRunning,
		processorStatusClosed,
	) {
		// wait for config update thread finish
		<-p.closeCH

		// close worker threads
		for i := 0; i < len(p.threads); i++ {
			go func(idx int) {
				if p.threads[idx].Close() {
					p.closeCH <- ""
				} else {
					p.closeCH <- p.threads[idx].GetExecActionDebug()
				}
			}(i)
		}

		// wait all rpcThread close
		errMap := make(map[string]int)
		for i := 0; i < len(p.threads); i++ {
			if errString := <-p.closeCH; errString != "" {
				if v, ok := errMap[errString]; ok {
					errMap[errString] = v + 1
				} else {
					errMap[errString] = 1
				}
			}
		}

		errList := make([]string, 0)
		for k, v := range errMap {
			if v > 1 {
				errList = append(errList, fmt.Sprintf("%s (%d goroutines)", k, v))
			} else {
				errList = append(errList, fmt.Sprintf("%s (%d goroutine)", k, v))
			}
		}

		if len(errList) > 0 {
			p.fnError(
				errors.ErrActionCloseTimeout.AddDebug(base.ConcatString(
					"the following actions can not close: \n\t",
					strings.Join(errList, "\n\t"),
				)),
			)
		}

		for _, freeCH := range p.freeCHArray {
			close(freeCH)
		}

		if p.panicSubscription != nil {
			p.panicSubscription.Close()
			p.panicSubscription = nil
		}

		return len(errList) == 0
	}

	return false
}

// PutStream ...
func (p *Processor) PutStream(stream *Stream) (ret bool) {
	defer func() {
		if v := recover(); v != nil {
			ret = false
		}
	}()

	if thread := <-p.freeCHArray[atomic.AddUint64(
		&p.readThreadPos,
		1,
	)%freeGroups]; thread != nil {
		success := thread.PutStream(stream)
		if !success {
			p.freeCHArray[atomic.AddUint64(
				&p.writeThreadPos,
				1,
			)%freeGroups] <- thread
		}
		return success
	}

	return false
}

// BuildCache ...
func (p *Processor) BuildCache(pkgName string, path string) *base.Error {
	retMap := make(map[string]bool)
	for _, action := range p.actionsMap {
		if fnTypeString, err := getFuncKind(action.reflectFn); err == nil {
			retMap[fnTypeString] = true
		}
	}

	fnKinds := make([]string, 0)
	for key := range retMap {
		fnKinds = append(fnKinds, key)
	}

	return buildFuncCache(pkgName, path, fnKinds)
}

func (p *Processor) onUpdateConfig() {
	for key := range p.servicesMap {
		p.invokeSystemAction("onUpdateConfig", key)
	}
}

func (p *Processor) invokeSystemAction(name string, path string) {
	unmountPath := path + ":$" + name
	if _, ok := p.actionsMap[unmountPath]; ok {
		stream, _ := MakeRequestStream(true, 0, unmountPath, "")
		defer func() {
			stream.Release()
		}()
		p.systemThread.Eval(stream, func(_ *Stream) {})
	}
}

func (p *Processor) mountNode(
	parentServiceNodePath string,
	nodeMeta *ServiceMeta,
	fnCache ActionCache,
) *base.Error {
	if nodeMeta == nil {
		return errors.ErrProcessorNodeMetaIsNil
	} else if !nodeNameRegex.MatchString(nodeMeta.name) {
		return errors.ErrServiceName.
			AddDebug(fmt.Sprintf("service name %s is illegal", nodeMeta.name)).
			AddDebug(nodeMeta.fileLine)
	} else if nodeMeta.service == nil {
		return errors.ErrServiceIsNil.AddDebug(nodeMeta.fileLine)
	} else {
		parentNode := p.servicesMap[parentServiceNodePath]
		servicePath := parentServiceNodePath + "." + nodeMeta.name
		if parentNode.depth+1 > p.maxNodeDepth { // depth overflows
			return errors.ErrServiceOverflow.
				AddDebug(fmt.Sprintf(
					"service path %s is overflow (max depth: %d, current depth:%d)",
					servicePath,
					p.maxNodeDepth, parentNode.depth+1,
				)).
				AddDebug(nodeMeta.fileLine)
		} else if item, ok := p.servicesMap[servicePath]; ok { // path is occupied
			return errors.ErrServiceName.
				AddDebug(fmt.Sprintf("duplicated service name %s", nodeMeta.name)).
				AddDebug(fmt.Sprintf(
					"current:\n%s\nconflict:\n%s",
					base.AddPrefixPerLine(nodeMeta.fileLine, "\t"),
					base.AddPrefixPerLine(item.addMeta.fileLine, "\t"),
				))
		} else {
			node := &rpcServiceNode{
				path:    servicePath,
				addMeta: nodeMeta,
				depth:   parentNode.depth + 1,
				data:    Map{},
				isMount: false,
			}

			for k, v := range nodeMeta.data {
				node.data[k] = v
			}

			// mount the node
			p.servicesMap[servicePath] = node

			// mount the actions
			for _, actionMeta := range nodeMeta.service.actions {
				if err := p.mountAction(node, actionMeta, fnCache); err != nil {
					p.unmount(servicePath)
					return err
				}
			}

			// invoke onUpdateConfig
			p.invokeSystemAction("onUpdateConfig", servicePath)

			// invoke onMount
			p.invokeSystemAction("onMount", servicePath)
			node.isMount = true

			// mount children
			for _, v := range nodeMeta.service.children {
				err := p.mountNode(servicePath, v, fnCache)
				if err != nil {
					p.unmount(servicePath)
					return err
				}
			}

			return nil
		}
	}
}

func (p *Processor) unmount(path string) {
	// clean sub nodes
	subPath := path + "."
	if _, ok := p.servicesMap[subPath]; ok {
		p.unmount(subPath + ".")
	}

	// clean node or sibling nodes
	for key, v := range p.servicesMap {
		if strings.HasPrefix(key, path) {
			if v.isMount {
				p.invokeSystemAction("onUnmount", key)
				v.isMount = false
			}
			delete(p.servicesMap, key)
		}
	}

	// clean actions
	for key := range p.actionsMap {
		if strings.HasPrefix(key, path) {
			delete(p.actionsMap, key)
		}
	}
}

func (p *Processor) mountAction(
	serviceNode *rpcServiceNode,
	meta *rpcActionMeta,
	fnCache ActionCache,
) *base.Error {
	if meta == nil {
		return errors.ErrProcessorActionMetaIsNil
	} else if !actionNameRegex.MatchString(meta.name) {
		return errors.ErrActionName.
			AddDebug(fmt.Sprintf("action name %s is illegal", meta.name)).
			AddDebug(meta.fileLine)
	} else if meta.handler == nil {
		return errors.ErrActionHandler.
			AddDebug("handler is nil").
			AddDebug(meta.fileLine)
	} else if fn := reflect.ValueOf(meta.handler); fn.Kind() != reflect.Func {
		return errors.ErrActionHandler.AddDebug(fmt.Sprintf(
			"handler must be func(rt %s, ...) %s",
			convertTypeToString(runtimeType),
			convertTypeToString(returnType),
		)).AddDebug(meta.fileLine)
	} else if fnTypeString, err := getFuncKind(fn); err != nil {
		// Check action handler is right
		return err.AddDebug(meta.fileLine)
	} else {
		actionPath := serviceNode.path + ":" + meta.name
		if item, ok := p.actionsMap[actionPath]; ok {
			// check the action path is not occupied
			return errors.ErrActionName.
				AddDebug(base.ConcatString("duplicated action name ", meta.name)).
				AddDebug(fmt.Sprintf(
					"current:\n%s\nconflict:\n%s",
					base.AddPrefixPerLine(meta.fileLine, "\t"),
					base.AddPrefixPerLine(item.meta.fileLine, "\t"),
				))
		}

		// mount the action
		numOfArgs := fn.Type().NumIn()
		argTypes := make([]reflect.Type, numOfArgs)
		argStrings := make([]string, numOfArgs)
		for i := 0; i < numOfArgs; i++ {
			argTypes[i] = fn.Type().In(i)
			argStrings[i] = convertTypeToString(argTypes[i])
		}

		actionNode := &rpcActionNode{
			path:      actionPath,
			meta:      meta,
			service:   serviceNode,
			cacheFN:   nil,
			reflectFn: fn,
			callString: fmt.Sprintf(
				"%s(%s) %s",
				actionPath,
				strings.Join(argStrings, ", "),
				convertTypeToString(returnType),
			),
			argTypes:  argTypes,
			indicator: base.NewPerformanceIndicator(),
		}

		if fnCache != nil {
			actionNode.cacheFN = fnCache.Get(fnTypeString)
		}

		p.actionsMap[actionPath] = actionNode
		return nil
	}
}

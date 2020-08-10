package internal

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

const rootName = "#"
const freeGroups = 1024

var (
	nodeNameRegex  = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	replyNameRegex = regexp.MustCompile(`^[_a-zA-Z][_0-9a-zA-Z]*$`)
)

type rpcReplyNode struct {
	path       string
	meta       *rpcReplyMeta
	cacheFN    ReplyCacheFunc
	reflectFn  reflect.Value
	callString string
	argTypes   []reflect.Type
	indicator  *rpcPerformanceIndicator
}

type rpcServiceNode struct {
	path    string
	addMeta *ServiceMeta
	depth   uint
}

// Processor ...
type Processor struct {
	isDebug           bool
	repliesMap        map[string]*rpcReplyNode
	servicesMap       map[string]*rpcServiceNode
	maxNodeDepth      uint64
	maxCallDepth      uint64
	threads           []*rpcThread
	freeCHArray       []chan *rpcThread
	readThreadPos     uint64
	writeThreadPos    uint64
	panicSubscription *rpcPanicSubscription
	fnError           func(err Error)
	Lock
}

// NewProcessor ...
func NewProcessor(
	isDebug bool,
	numOfThreads int,
	maxNodeDepth int,
	maxCallDepth int,
	fnCache ReplyCache,
	closeTimeout time.Duration,
	mountServices []*ServiceMeta,
	onReturnStream func(stream *Stream),
) *Processor {
	if onReturnStream == nil {
		return nil
	}

	fnError := func(err Error) {
		defer func() {
			recover()
		}()
		stream := NewStream()
		stream.WriteUint64(uint64(err.GetKind()))
		stream.WriteString(err.GetMessage())
		stream.WriteString(err.GetDebug())
		onReturnStream(stream)
	}

	if numOfThreads <= 0 {
		fnError(
			NewKernelPanic("numOfThreads is wrong").
				AddDebug(string(debug.Stack())),
		)
		return nil
	} else if maxNodeDepth <= 0 {
		fnError(
			NewKernelPanic("maxNodeDepth is wrong").
				AddDebug(string(debug.Stack())),
		)
		return nil
	} else if maxCallDepth <= 0 {
		fnError(
			NewKernelPanic("maxCallDepth is wrong").
				AddDebug(string(debug.Stack())),
		)
		return nil
	} else {
		size := ((numOfThreads + freeGroups - 1) / freeGroups) * freeGroups
		ret := &Processor{
			isDebug:        isDebug,
			repliesMap:     make(map[string]*rpcReplyNode),
			servicesMap:    make(map[string]*rpcServiceNode),
			maxNodeDepth:   uint64(maxNodeDepth),
			maxCallDepth:   uint64(maxCallDepth),
			threads:        make([]*rpcThread, size, size),
			freeCHArray:    nil,
			readThreadPos:  0,
			writeThreadPos: 0,
			fnError:        fnError,
		}

		// mount nodes
		ret.servicesMap[rootName] = &rpcServiceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
		}

		for _, meta := range mountServices {
			if err := ret.mountNode(rootName, meta, fnCache); err != nil {
				ret.Panic(err)
				return nil
			}
		}

		// subscribe panic
		ret.panicSubscription = subscribePanic(fnError)

		// start threads
		freeCHArray := make([]chan *rpcThread, freeGroups, freeGroups)
		for i := 0; i < freeGroups; i++ {
			freeCHArray[i] = make(chan *rpcThread, size/freeGroups)
		}
		ret.freeCHArray = freeCHArray

		for i := 0; i < size; i++ {
			thread := newThread(
				ret,
				closeTimeout,
				onReturnStream,
				func(thread *rpcThread) {
					freeCHArray[atomic.AddUint64(
						&ret.writeThreadPos,
						1,
					)%freeGroups] <- thread
				},
			)
			ret.threads[i] = thread
			ret.freeCHArray[i%freeGroups] <- thread
		}
		return ret
	}
}

// Close ...
func (p *Processor) Close() bool {
	return p.CallWithLock(func() interface{} {
		if p.panicSubscription == nil {
			return false
		}

		closeCH := make(chan string)

		for i := 0; i < len(p.threads); i++ {
			go func(idx int) {
				if p.threads[idx].Close() {
					closeCH <- ""
				} else {
					closeCH <- p.threads[idx].GetExecReplyDebug()
				}
			}(i)
		}

		// wait all rpcThread close
		errMap := make(map[string]int)
		for i := 0; i < len(p.threads); i++ {
			if errString := <-closeCH; errString != "" {
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
				NewReplyPanic(ConcatString(
					"the following replies can not close: \n\t",
					strings.Join(errList, "\n\t"),
				)),
			)
		}

		for _, freeCH := range p.freeCHArray {
			close(freeCH)
		}

		p.panicSubscription.Close()
		p.panicSubscription = nil
		return len(errList) == 0
	}).(bool)
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

// Panic ...
func (p *Processor) Panic(err Error) {
	p.fnError(err)
}

// BuildCache ...
func (p *Processor) BuildCache(pkgName string, path string) Error {
	retMap := make(map[string]bool)
	for _, reply := range p.repliesMap {
		if fnTypeString, err := getFuncKind(reply.reflectFn); err == nil {
			retMap[fnTypeString] = true
		}
	}

	fnKinds := make([]string, 0)
	for key := range retMap {
		fnKinds = append(fnKinds, key)
	}

	return buildFuncCache(pkgName, path, fnKinds)
}

func (p *Processor) mountNode(
	parentServiceNodePath string,
	nodeMeta *ServiceMeta,
	fnCache ReplyCache,
) Error {
	if nodeMeta == nil {
		// check nodeMeta is not nil
		return NewKernelPanic("nodeMeta is nil").
			AddDebug(string(debug.Stack()))
	} else if !nodeNameRegex.MatchString(nodeMeta.name) {
		// check nodeMeta.name is valid
		return NewRuntimePanic(
			fmt.Sprintf("service name %s is illegal", nodeMeta.name),
		).AddDebug(nodeMeta.fileLine)
	} else if nodeMeta.service == nil {
		// check nodeMeta.service is not nil
		return NewRuntimePanic("service is nil").AddDebug(nodeMeta.fileLine)
	} else {
		parentNode := p.servicesMap[parentServiceNodePath]
		servicePath := parentServiceNodePath + "." + nodeMeta.name
		if uint64(parentNode.depth+1) > p.maxNodeDepth {
			// check max node depth overflow
			return NewRuntimePanic(fmt.Sprintf(
				"service path %s is too long",
				servicePath,
			)).AddDebug(nodeMeta.fileLine)
		} else if item, ok := p.servicesMap[servicePath]; ok {
			// check the mount path is not occupied
			return NewRuntimePanic(fmt.Sprintf(
				"duplicated service name %s",
				nodeMeta.name,
			)).AddDebug(fmt.Sprintf(
				"current:\n%s\nconflict:\n%s",
				AddPrefixPerLine(nodeMeta.fileLine, "\t"),
				AddPrefixPerLine(item.addMeta.fileLine, "\t"),
			))
		} else {
			node := &rpcServiceNode{
				path:    servicePath,
				addMeta: nodeMeta,
				depth:   parentNode.depth + 1,
			}

			// mount the node
			p.servicesMap[servicePath] = node

			// mount the replies
			for _, replyMeta := range nodeMeta.service.replies {
				err := p.mountReply(node, replyMeta, fnCache)
				if err != nil {
					delete(p.servicesMap, servicePath)
					return err
				}
			}

			// mount children
			for _, v := range nodeMeta.service.children {
				err := p.mountNode(node.path, v, fnCache)
				if err != nil {
					delete(p.servicesMap, servicePath)
					return err
				}
			}

			return nil
		}
	}
}

func (p *Processor) mountReply(
	serviceNode *rpcServiceNode,
	meta *rpcReplyMeta,
	fnCache ReplyCache,
) Error {
	if meta == nil {
		// check the rpcReplyMeta is nil
		return NewKernelPanic("meta is nil").
			AddDebug(string(debug.Stack()))
	} else if !replyNameRegex.MatchString(meta.name) {
		// check the name
		return NewRuntimePanic(
			fmt.Sprintf("reply name %s is illegal", meta.name),
		).AddDebug(meta.fileLine)
	} else if meta.handler == nil {
		// check the reply handler is nil
		return NewRuntimePanic("handler is nil").AddDebug(meta.fileLine)
	} else if fn := reflect.ValueOf(meta.handler); fn.Kind() != reflect.Func {
		// Check reply handler is Func
		return NewRuntimePanic(fmt.Sprintf(
			"handler must be func(ctx %s, ...) %s",
			convertTypeToString(contextType),
			convertTypeToString(returnType),
		)).AddDebug(meta.fileLine)
	} else if fnTypeString, err := getFuncKind(fn); err != nil {
		// Check reply handler is right
		return NewRuntimePanic(err.Error()).AddDebug(meta.fileLine)
	} else {
		replyPath := serviceNode.path + ":" + meta.name
		if item, ok := p.repliesMap[replyPath]; ok {
			// check the reply path is not occupied
			return NewRuntimePanic(fmt.Sprintf(
				"reply name %s is duplicated",
				meta.name,
			)).AddDebug(fmt.Sprintf(
				"current:\n%s\nconflict:\n%s",
				AddPrefixPerLine(meta.fileLine, "\t"),
				AddPrefixPerLine(item.meta.fileLine, "\t"),
			))
		}

		// mount the replyRecord
		numOfArgs := fn.Type().NumIn()
		argTypes := make([]reflect.Type, numOfArgs, numOfArgs)
		argStrings := make([]string, numOfArgs, numOfArgs)
		for i := 0; i < numOfArgs; i++ {
			argTypes[i] = fn.Type().In(i)
			argStrings[i] = convertTypeToString(argTypes[i])
		}

		replyNode := &rpcReplyNode{
			path:      replyPath,
			meta:      meta,
			cacheFN:   nil,
			reflectFn: fn,
			callString: fmt.Sprintf(
				"%s(%s) %s",
				replyPath,
				strings.Join(argStrings, ", "),
				convertTypeToString(returnType),
			),
			argTypes:  argTypes,
			indicator: newPerformanceIndicator(),
		}

		if fnCache != nil {
			replyNode.cacheFN = fnCache.Get(fnTypeString)
		}

		p.repliesMap[replyPath] = replyNode
		return nil
	}
}

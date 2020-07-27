package internal

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
	"sync/atomic"
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
	addMeta *rpcChildMeta
	depth   uint
}

// Processor ...
type Processor struct {
	isDebug            bool
	fnCache            ReplyCache
	repliesMap         map[string]*rpcReplyNode
	servicesMap        map[string]*rpcServiceNode
	maxNodeDepth       uint64
	maxCallDepth       uint64
	threads            []*rpcThread
	freeThreadsCHGroup []chan *rpcThread
	readThreadPos      uint64
	writeThreadPos     uint64
	panicSubscription  *rpcPanicSubscription
	fnError            func(err Error)
	Lock
}

// NewProcessor ...
func NewProcessor(
	isDebug bool,
	numOfThreads uint,
	maxNodeDepth uint,
	maxCallDepth uint,
	fnCache ReplyCache,
	mountServices []*rpcChildMeta,
	onReturnStream func(stream *Stream),
) *Processor {
	if onReturnStream == nil {
		return nil
	}

	fnError := func(err Error) {
		defer func() {
			// ignore this error.
			// it has tried its best to report panic.
			recover()
		}()
		stream := NewStream()
		stream.WriteUint64(uint64(err.GetKind()))
		stream.WriteString(err.GetMessage())
		stream.WriteString(err.GetDebug())
		onReturnStream(stream)
	}

	if numOfThreads == 0 {
		fnError(
			NewKernelPanic("rpc: numOfThreads is zero").
				AddDebug(string(debug.Stack())),
		)
		return nil
	} else if maxNodeDepth == 0 {
		fnError(
			NewKernelPanic("rpc: maxNodeDepth is zero").
				AddDebug(string(debug.Stack())),
		)
		return nil
	} else if maxCallDepth == 0 {
		fnError(
			NewKernelPanic("rpc: maxCallDepth is zero").
				AddDebug(string(debug.Stack())),
		)
		return nil
	} else {
		size := int(((numOfThreads + freeGroups - 1) / freeGroups) * freeGroups)
		ret := &Processor{
			isDebug:            isDebug,
			fnCache:            fnCache,
			repliesMap:         make(map[string]*rpcReplyNode),
			servicesMap:        make(map[string]*rpcServiceNode),
			maxNodeDepth:       uint64(maxNodeDepth),
			maxCallDepth:       uint64(maxCallDepth),
			threads:            make([]*rpcThread, size, size),
			freeThreadsCHGroup: nil,
			readThreadPos:      0,
			writeThreadPos:     0,
			fnError:            fnError,
		}

		// mount nodes
		ret.servicesMap[rootName] = &rpcServiceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
		}
		for _, meta := range mountServices {
			if err := ret.mountNode(rootName, meta); err != nil {
				ret.Panic(err)
				return nil
			}
		}

		// subscribe panic
		ret.panicSubscription = subscribePanic(fnError)

		// start threads
		freeThreadsCHGroup := make([]chan *rpcThread, freeGroups, freeGroups)
		for i := 0; i < freeGroups; i++ {
			freeThreadsCHGroup[i] = make(chan *rpcThread, size/freeGroups)
		}
		ret.freeThreadsCHGroup = freeThreadsCHGroup

		for i := 0; i < size; i++ {
			thread := newThread(ret, onReturnStream, func(thread *rpcThread) {
				freeThreadsCHGroup[atomic.AddUint64(
					&ret.writeThreadPos,
					1,
				)%freeGroups] <- thread
			})
			ret.threads[i] = thread
			ret.freeThreadsCHGroup[i%freeGroups] <- thread
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

		closeCH := make(chan string, len(p.threads))

		for i := 0; i < len(p.threads); i++ {
			go func(idx int) {

				if p.threads[idx].Close() {
					closeCH <- ""
				} else if node := p.threads[idx].GetReplyNode(); node != nil {
					closeCH <- p.threads[idx].GetExecReplyFileLine()
				} else {
					closeCH <- ""
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
				errList = append(errList, fmt.Sprintf("%s (%d routines)", k, v))
			} else {
				errList = append(errList, fmt.Sprintf("%s (%d routine)", k, v))
			}
		}

		if len(errList) > 0 {
			p.fnError(
				NewReplyPanic(ConcatString(
					"rpc: the following replies can not close after 20 seconds: \n\t",
					strings.Join(errList, "\n\t"),
				)),
			)
		}

		for _, freeCH := range p.freeThreadsCHGroup {
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

	if thread := <-p.freeThreadsCHGroup[atomic.AddUint64(
		&p.readThreadPos,
		1,
	)%freeGroups]; thread != nil {
		success := thread.PutStream(stream)
		if !success {
			p.freeThreadsCHGroup[atomic.AddUint64(
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
		if fnTypeString, ok := getFuncKind(reply.meta.handler); ok {
			retMap[fnTypeString] = true
		}
	}

	fnKinds := make([]string, 0)
	for key := range retMap {
		fnKinds = append(fnKinds, key)
	}

	if err := buildFuncCache(pkgName, path, fnKinds); err != nil {
		return NewRuntimePanic(err.Error()).AddDebug(string(debug.Stack()))
	}

	return nil
}

func (p *Processor) mountNode(
	parentServiceNodePath string,
	nodeMeta *rpcChildMeta,
) Error {
	if nodeMeta == nil {
		// check nodeMeta is not nil
		return NewKernelPanic("rpc: nodeMeta is nil").
			AddDebug(string(debug.Stack()))
	} else if !nodeNameRegex.MatchString(nodeMeta.name) {
		// check nodeMeta.name is valid
		return NewRuntimePanic(
			fmt.Sprintf("rpc: service name %s is illegal", nodeMeta.name),
		).AddDebug(nodeMeta.fileLine)
	} else if nodeMeta.service == nil {
		// check nodeMeta.service is not nil
		return NewRuntimePanic("rpc: service is nil").AddDebug(nodeMeta.fileLine)
	} else if parentNode, ok := p.servicesMap[parentServiceNodePath]; !ok {
		// check if parent node is exist
		return NewKernelPanic(
			"rpc: can not find parent service",
		).AddDebug(string(debug.Stack()))
	} else {
		servicePath := parentServiceNodePath + "." + nodeMeta.name
		if uint64(parentNode.depth+1) > p.maxNodeDepth {
			// check max node depth overflow
			return NewRuntimePanic(fmt.Sprintf(
				"rpc: service path %s is too long",
				servicePath,
			)).AddDebug(nodeMeta.fileLine)
		} else if item, ok := p.servicesMap[servicePath]; ok {
			// check the mount path is not occupied
			return NewRuntimePanic(fmt.Sprintf(
				"rpc: duplicated service name %s",
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
				err := p.mountReply(node, replyMeta)
				if err != nil {
					delete(p.servicesMap, servicePath)
					return err
				}
			}

			// mount children
			for _, v := range nodeMeta.service.children {
				err := p.mountNode(node.path, v)
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
) Error {
	if serviceNode == nil {
		// check the node is nil
		return NewKernelPanic("rpc: service node is nil").
			AddDebug(string(debug.Stack()))
	} else if meta == nil {
		// check the rpcReplyMeta is nil
		return NewKernelPanic("rpc: meta is nil").
			AddDebug(string(debug.Stack()))
	} else if !replyNameRegex.MatchString(meta.name) {
		// check the name
		return NewRuntimePanic(
			fmt.Sprintf("rpc: reply name %s is illegal", meta.name),
		).AddDebug(meta.fileLine)
	} else if meta.handler == nil {
		// check the reply handler is nil
		return NewRuntimePanic("rpc: reply handler is nil").AddDebug(meta.fileLine)
	} else if fn := reflect.ValueOf(meta.handler); fn.Kind() != reflect.Func {
		// Check reply handler is Func
		return NewRuntimePanic(fmt.Sprintf(
			"rpc: reply handler must be func(ctx %s, ...) %s",
			convertTypeToString(contextType),
			convertTypeToString(returnType),
		)).AddDebug(meta.fileLine)
	} else if argsErrorPos := getArgumentsErrorPosition(fn); argsErrorPos == 0 {
		// Check reply handler arguments types
		return NewRuntimePanic(fmt.Sprintf(
			"rpc: reply handler 1st argument type must be %s",
			convertTypeToString(contextType),
		)).AddDebug(meta.fileLine)
	} else if argsErrorPos > 0 {
		// Check reply handler arguments types
		return NewRuntimePanic(fmt.Sprintf(
			"rpc: reply handler %s argument type %s is not supported",
			ConvertOrdinalToString(1+uint(argsErrorPos)),
			fmt.Sprintf("%s", fn.Type().In(argsErrorPos)),
		)).AddDebug(meta.fileLine)
	} else if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		// Check return type
		return NewRuntimePanic(
			fmt.Sprintf(
				"rpc: reply handler return type must be %s",
				convertTypeToString(returnType),
			)).AddDebug(meta.fileLine)
	} else {
		replyPath := serviceNode.path + ":" + meta.name
		if item, ok := p.repliesMap[replyPath]; ok {
			// check the reply path is not occupied
			return NewRuntimePanic(fmt.Sprintf(
				"rpc: reply name %s is duplicated",
				meta.name,
			)).AddDebug(fmt.Sprintf(
				"current:\n%s\nconflict:\n%s",
				AddPrefixPerLine(meta.fileLine, "\t"),
				AddPrefixPerLine(item.meta.fileLine, "\t"),
			))
		}

		// mount the replyRecord
		argTypes := make([]reflect.Type, fn.Type().NumIn(), fn.Type().NumIn())
		argStrings := make([]string, fn.Type().NumIn(), fn.Type().NumIn())
		for i := 0; i < len(argTypes); i++ {
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

		if fnTypeString, ok := getFuncKind(meta.handler); ok && p.fnCache != nil {
			replyNode.cacheFN = p.fnCache.Get(fnTypeString)
		}

		p.repliesMap[replyPath] = replyNode
		return nil
	}
}

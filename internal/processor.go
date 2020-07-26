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
	replyMeta  *rpcReplyMeta
	cacheFN    ReplyCacheFunc
	reflectFn  reflect.Value
	callString string
	argTypes   []reflect.Type
	indicator  *rpcPerformanceIndicator
}

func (p *rpcReplyNode) GetPath() string {
	return p.path
}

func (p *rpcReplyNode) GetDebug() string {
	if p.replyMeta != nil {
		return ConcatString(p.path, " ", p.replyMeta.fileLine)
	}
	return p.path
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
	Lock
}

// NewProcessor ...
func NewProcessor(
	isDebug bool,
	numOfThreads uint,
	maxNodeDepth uint,
	maxCallDepth uint,
	fnCache ReplyCache,
) (*Processor, Error) {
	if numOfThreads == 0 {
		return nil, NewKernelError("rpc: numOfThreads is zero").
			AddDebug(string(debug.Stack()))
	} else if maxNodeDepth == 0 {
		return nil, NewKernelError("rpc: maxNodeDepth is zero").
			AddDebug(string(debug.Stack()))
	} else if maxCallDepth == 0 {
		return nil, NewKernelError("rpc: maxCallDepth is zero").
			AddDebug(string(debug.Stack()))
	} else {
		size := ((numOfThreads + freeGroups - 1) / freeGroups) * freeGroups
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
		}
		// mount root node
		ret.servicesMap[rootName] = &rpcServiceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
		}
		return ret, nil
	}
}

func (p *Processor) IsDebug() bool {
	return p.isDebug
}

func (p *Processor) Start(
	onReturnStream func(stream *Stream),
) Error {
	return ConvertToError(p.CallWithLock(func() interface{} {
		if onReturnStream == nil {
			return NewKernelError(ErrStringUnexpectedNil).
				AddDebug(string(debug.Stack()))
		} else if p.freeThreadsCHGroup != nil {
			return NewKernelError(ErrStringUnexpectedNil).
				AddDebug(string(debug.Stack()))
		} else {
			p.panicSubscription = SubscribePanic(func(err Error) {
				defer func() {
					// ignore this error.
					// it has tried its best to report panic.
					recover()
				}()
				stream := NewStream()
				stream.SetStreamKind(StreamKindResponsePanic)
				stream.WriteUint64(uint64(err.GetKind()))
				stream.WriteString(err.GetMessage())
				stream.WriteString(err.GetDebug())
				onReturnStream(stream)
				stream.Release()
			})

			freeThreadsCHGroup := make([]chan *rpcThread, freeGroups, freeGroups)
			for i := 0; i < freeGroups; i++ {
				freeThreadsCHGroup[i] = make(chan *rpcThread, len(p.threads)/freeGroups)
			}
			p.freeThreadsCHGroup = freeThreadsCHGroup

			for i := 0; i < len(p.threads); i++ {
				thread := newThread(p, onReturnStream, func(thread *rpcThread) {
					freeThreadsCHGroup[atomic.AddUint64(
						&p.writeThreadPos,
						1,
					)%freeGroups] <- thread
				})
				p.threads[i] = thread
				p.freeThreadsCHGroup[i%freeGroups] <- thread
			}
			return Error(nil)
		}
	}))
}

func (p *Processor) Stop() Error {
	return ConvertToError(p.CallWithLock(func() interface{} {
		if p.freeThreadsCHGroup == nil {
			return NewKernelError(ErrStringUnexpectedNil).
				AddDebug(string(debug.Stack()))
		} else {
			closeCH := make(chan string, len(p.threads))

			for i := 0; i < len(p.threads); i++ {
				go func(idx int) {
					if !p.threads[idx].Stop() && p.threads[idx].execReplyNode != nil {
						closeCH <- p.threads[idx].execReplyNode.GetDebug()
					} else {
						closeCH <- ""
					}
					p.threads[idx] = nil
				}(i)
			}

			// wait all rpcThread stop
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
					errList = append(errList, fmt.Sprintf(
						"%s (%d routines)",
						k,
						v,
					))
				} else {
					errList = append(errList, fmt.Sprintf(
						"%s (%d routine)",
						k,
						v,
					))
				}
			}

			p.freeThreadsCHGroup = nil
			p.panicSubscription.Close()
			p.panicSubscription = nil

			if len(errList) > 0 {
				// this is because reply is still running
				return NewReplyPanic(ConcatString(
					"rpc: the following routine can not stop after 20 seconds: \n\t",
					strings.Join(errList, "\n\t"),
				))
			} else {
				return Error(nil)
			}
		}
	}))
}

func (p *Processor) PutStream(stream *Stream) bool {
	if freeThreadsCHGroup := p.freeThreadsCHGroup; freeThreadsCHGroup == nil {
		return false
	} else if thread := <-freeThreadsCHGroup[atomic.AddUint64(
		&p.readThreadPos,
		1,
	)%freeGroups]; thread == nil {
		return false
	} else {
		return thread.PutStream(stream)
	}
}

// BuildCache ...
func (p *Processor) BuildCache(pkgName string, path string) Error {
	retMap := make(map[string]bool)
	for _, reply := range p.repliesMap {
		if fnTypeString, ok := getFuncKind(reply.replyMeta.handler); ok {
			retMap[fnTypeString] = true
		}
	}

	fnKinds := make([]string, 0)
	for key := range retMap {
		fnKinds = append(fnKinds, key)
	}

	if err := buildFuncCache(pkgName, path, fnKinds); err != nil {
		return NewRuntimeError(err.Error()).AddDebug(string(debug.Stack()))
	}

	return nil
}

// AddChildService ...
func (p *Processor) AddService(
	name string,
	service *Service,
	debug string,
) Error {
	if service == nil {
		return NewRuntimeError("rpc: service is nil").AddDebug(debug)
	}

	return p.mountNode(rootName, &rpcChildMeta{
		name:     name,
		service:  service,
		fileLine: debug,
	})
}

func (p *Processor) mountNode(
	parentServiceNodePath string,
	nodeMeta *rpcChildMeta,
) Error {
	// check nodeMeta is not nil
	if nodeMeta == nil {
		return NewKernelError(ErrStringUnexpectedNil).
			AddDebug(string(debug.Stack()))
	}

	// check nodeMeta.name is valid
	if !nodeNameRegex.MatchString(nodeMeta.name) {
		return NewBaseError(
			fmt.Sprintf("rpc: service name \"%s\" is illegal", nodeMeta.name),
		).AddDebug(nodeMeta.fileLine)
	}

	// check nodeMeta.service is not nil
	if nodeMeta.service == nil {
		return NewBaseError("rpc: service is nil").AddDebug(nodeMeta.fileLine)
	}

	// check max node depth overflow
	parentNode, ok := p.servicesMap[parentServiceNodePath]
	if !ok {
		return NewBaseError(
			"rpc: mountNode: parentNode is nil",
		).AddDebug(nodeMeta.fileLine)
	}
	servicePath := parentServiceNodePath + "." + nodeMeta.name
	if uint64(parentNode.depth+1) > p.maxNodeDepth {
		return NewBaseError(fmt.Sprintf(
			"Service path depth %s is too long, it must be less or equal than %d",
			servicePath,
			p.maxNodeDepth,
		)).AddDebug(nodeMeta.fileLine)
	}

	// check the mount path is not occupied
	if item, ok := p.servicesMap[servicePath]; ok {
		return NewBaseError(fmt.Sprintf(
			"Service name \"%s\" is duplicated",
			nodeMeta.name,
		)).AddDebug(fmt.Sprintf(
			"Current:\n%s\nConflict:\n%s",
			AddPrefixPerLine(nodeMeta.fileLine, "\t"),
			AddPrefixPerLine(item.addMeta.fileLine, "\t"),
		))
	}

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

func (p *Processor) mountReply(
	serviceNode *rpcServiceNode,
	replyMeta *rpcReplyMeta,
) Error {
	// check the node is nil
	if serviceNode == nil {
		return NewBaseError("rpc: mountReply: node is nil")
	}

	// check the rpcReplyMeta is nil
	if replyMeta == nil {
		return NewBaseError("rpc: mountReply: rpcReplyMeta is nil")
	}

	// check the name
	if !replyNameRegex.MatchString(replyMeta.name) {
		return NewBaseError(
			fmt.Sprintf("Reply name %s is illegal", replyMeta.name),
		).AddDebug(replyMeta.fileLine)
	}

	// check the reply path is not occupied
	replyPath := serviceNode.path + ":" + replyMeta.name
	if item, ok := p.repliesMap[replyPath]; ok {
		return NewBaseError(fmt.Sprintf(
			"Reply name %s is duplicated",
			replyMeta.name,
		)).AddDebug(fmt.Sprintf(
			"Current:\n%s\nConflict:\n%s",
			AddPrefixPerLine(replyMeta.fileLine, "\t"),
			AddPrefixPerLine(item.replyMeta.fileLine, "\t"),
		))
	}

	// check the reply handler is nil
	if replyMeta.handler == nil {
		return NewBaseError("Reply handler is nil").AddDebug(replyMeta.fileLine)
	}

	// Check reply handler is Func
	fn := reflect.ValueOf(replyMeta.handler)
	if fn.Kind() != reflect.Func {
		return NewBaseError(fmt.Sprintf(
			"Reply handler must be func(ctx %s, ...) %s",
			convertTypeToString(contextType),
			convertTypeToString(returnType),
		)).AddDebug(replyMeta.fileLine)
	}

	// Check reply handler arguments types
	argumentsErrorPos := getArgumentsErrorPosition(fn)
	if argumentsErrorPos == 0 {
		return NewBaseError(fmt.Sprintf(
			"Reply handler 1st argument type must be %s",
			convertTypeToString(contextType),
		)).AddDebug(replyMeta.fileLine)
	} else if argumentsErrorPos > 0 {
		return NewBaseError(fmt.Sprintf(
			"Reply handler %s argument type <%s> not supported",
			ConvertOrdinalToString(1+uint(argumentsErrorPos)),
			fmt.Sprintf("%s", fn.Type().In(argumentsErrorPos)),
		)).AddDebug(replyMeta.fileLine)
	}

	// Check return type
	if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return NewBaseError(
			fmt.Sprintf(
				"Reply handler return type must be %s",
				convertTypeToString(returnType),
			)).AddDebug(replyMeta.fileLine)
	}

	// mount the replyRecord
	argTypes := make([]reflect.Type, fn.Type().NumIn(), fn.Type().NumIn())
	argStrings := make([]string, fn.Type().NumIn(), fn.Type().NumIn())
	for i := 0; i < len(argTypes); i++ {
		argTypes[i] = fn.Type().In(i)
		argStrings[i] = convertTypeToString(argTypes[i])
	}
	argString := strings.Join(argStrings, ", ")

	cacheFN := ReplyCacheFunc(nil)
	if fnTypeString, ok := getFuncKind(replyMeta.handler); ok && p.fnCache != nil {
		cacheFN = p.fnCache.Get(fnTypeString)
	}

	p.repliesMap[replyPath] = &rpcReplyNode{
		path:      replyPath,
		replyMeta: replyMeta,
		cacheFN:   cacheFN,
		reflectFn: fn,
		callString: fmt.Sprintf(
			"%s(%s) %s",
			replyPath,
			argString,
			convertTypeToString(returnType),
		),
		argTypes:  argTypes,
		indicator: newPerformanceIndicator(),
	}

	return nil
}

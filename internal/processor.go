package internal

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
)

const rootName = "$"
const freeGroups = 1024

var (
	nodeNameRegex  = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	replyNameRegex = regexp.MustCompile(`^[_a-zA-Z][_0-9a-zA-Z]*$`)
)

type replyNode struct {
	replyMeta   *rpcReplyMeta
	cacheFN     RPCReplyCacheFunc
	reflectFn   reflect.Value
	callString  string
	debugString string
	argTypes    []reflect.Type
	indicator   *performanceIndicator
}

type serviceNode struct {
	path    string
	addMeta *rpcAddChildMeta
	depth   uint
}

// Processor ...
type Processor struct {
	isDebug            bool
	fnCache            RPCReplyCache
	repliesMap         map[string]*replyNode
	servicesMap        map[string]*serviceNode
	maxNodeDepth       uint64
	maxCallDepth       uint64
	threads            []*rpcThread
	freeThreadsCHGroup []chan *rpcThread
	readThreadPos      uint64
	writeThreadPos     uint64
	Lock
}

// NewProcessor ...
func NewProcessor(
	isDebug bool,
	numOfThreads uint,
	maxNodeDepth uint,
	maxCallDepth uint,
	fnCache RPCReplyCache,
) *Processor {
	if numOfThreads == 0 {
		return nil
	} else if maxNodeDepth == 0 {
		return nil
	} else if maxCallDepth == 0 {
		return nil
	} else {
		size := ((numOfThreads + freeGroups - 1) / freeGroups) * freeGroups
		ret := &Processor{
			isDebug:            isDebug,
			fnCache:            fnCache,
			repliesMap:         make(map[string]*replyNode),
			servicesMap:        make(map[string]*serviceNode),
			maxNodeDepth:       uint64(maxNodeDepth),
			maxCallDepth:       uint64(maxCallDepth),
			threads:            make([]*rpcThread, size, size),
			freeThreadsCHGroup: nil,
			readThreadPos:      0,
			writeThreadPos:     0,
		}
		// mount root node
		ret.servicesMap[rootName] = &serviceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
		}
		return ret
	}
}

func (p *Processor) Start(
	onEvalFinish func(stream *RPCStream, success bool),
	onPanic func(v interface{}, debug string),
) Error {
	return ConvertToError(p.CallWithLock(func() interface{} {
		size := len(p.threads)

		if p.freeThreadsCHGroup != nil {
			return NewError("Processor: Start: it has already benn started")
		} else {
			freeThreadsCHGroup := make(
				[]chan *rpcThread,
				freeGroups,
				freeGroups,
			)
			for i := 0; i < freeGroups; i++ {
				freeThreadsCHGroup[i] = make(
					chan *rpcThread,
					size/freeGroups,
				)
			}
			p.freeThreadsCHGroup = freeThreadsCHGroup

			for i := 0; i < size; i++ {
				thread := newThread(
					p,
					func(thread *rpcThread, stream *RPCStream, success bool) {
						onEvalFinish(stream, success)

						defer func() {
							// do not panic when freeThreadsCH was closed,
							recover()
						}()
						freeThreadsCHGroup[atomic.AddUint64(
							&p.writeThreadPos,
							1,
						)%freeGroups] <- thread
					},
					onPanic,
				)
				p.threads[i] = thread
				p.freeThreadsCHGroup[i%freeGroups] <- thread
			}
			return nil
		}
	}))
}

func (p *Processor) PutStream(stream *RPCStream) bool {
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

func (p *Processor) Stop() Error {
	return ConvertToError(p.CallWithLock(func() interface{} {
		if p.freeThreadsCHGroup == nil {
			return NewError("Processor: Start: it has already benn stopped")
		} else {
			for i := 0; i < freeGroups; i++ {
				close(p.freeThreadsCHGroup[i])
			}
			p.freeThreadsCHGroup = nil
			numOfThreads := len(p.threads)
			closeCH := make(chan string, numOfThreads)

			for i := 0; i < numOfThreads; i++ {
				go func(idx int) {
					if !p.threads[idx].Stop() && p.threads[idx].execReplyNode != nil {
						closeCH <- p.threads[idx].execReplyNode.debugString
					} else {
						closeCH <- ""
					}
					p.threads[idx] = nil
				}(i)
			}

			// wait all thread stop
			errMap := make(map[string]int)
			for i := 0; i < numOfThreads; i++ {
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

			if len(errList) > 0 {
				return NewError(ConcatString(
					"Processor: Stop: The following routine still running: \n\t",
					strings.Join(errList, "\n\t"),
				))
			} else {
				return nil
			}
		}
	}))
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

	return buildFuncCache(pkgName, path, fnKinds)
}

// AddChild ...
func (p *Processor) AddService(
	name string,
	service *Service,
	debug string,
) Error {
	if service == nil {
		return NewErrorByDebug(
			"Service is nil",
			debug,
		)
	}

	return p.mountNode(rootName, &rpcAddChildMeta{
		name:    name,
		service: service,
		debug:   debug,
	})
}

func (p *Processor) mountNode(
	parentServiceNodePath string,
	nodeMeta *rpcAddChildMeta,
) Error {
	// check nodeMeta is not nil
	if nodeMeta == nil {
		return NewError("rpc: mountNode: nodeMeta is nil")
	}

	// check nodeMeta.name is valid
	if !nodeNameRegex.MatchString(nodeMeta.name) {
		return NewErrorByDebug(
			fmt.Sprintf("Service name \"%s\" is illegal", nodeMeta.name),
			nodeMeta.debug,
		)
	}

	// check nodeMeta.service is not nil
	if nodeMeta.service == nil {
		return NewErrorByDebug(
			"Service is nil",
			nodeMeta.debug,
		)
	}

	// check max node depth overflow
	parentNode, ok := p.servicesMap[parentServiceNodePath]
	if !ok {
		return NewErrorByDebug(
			"rpc: mountNode: parentNode is nil",
			nodeMeta.debug,
		)
	}
	servicePath := parentServiceNodePath + "." + nodeMeta.name
	if uint64(parentNode.depth+1) > p.maxNodeDepth {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Service path depth %s is too long, it must be less or equal than %d",
				servicePath,
				p.maxNodeDepth,
			),
			nodeMeta.debug,
		)
	}

	// check the mount path is not occupied
	if item, ok := p.servicesMap[servicePath]; ok {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Service name \"%s\" is duplicated",
				nodeMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				AddPrefixPerLine(nodeMeta.debug, "\t"),
				AddPrefixPerLine(item.addMeta.debug, "\t"),
			),
		)
	}

	node := &serviceNode{
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
	serviceNode *serviceNode,
	replyMeta *rpcReplyMeta,
) Error {
	// check the node is nil
	if serviceNode == nil {
		return NewError("rpc: mountReply: node is nil")
	}

	// check the replyMeta is nil
	if replyMeta == nil {
		return NewError("rpc: mountReply: replyMeta is nil")
	}

	// check the name
	if !replyNameRegex.MatchString(replyMeta.name) {
		return NewErrorByDebug(
			fmt.Sprintf("Reply name %s is illegal", replyMeta.name),
			replyMeta.debug,
		)
	}

	// check the reply path is not occupied
	replyPath := serviceNode.path + ":" + replyMeta.name
	if item, ok := p.repliesMap[replyPath]; ok {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Reply name %s is duplicated",
				replyMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				AddPrefixPerLine(replyMeta.debug, "\t"),
				AddPrefixPerLine(item.replyMeta.debug, "\t"),
			),
		)
	}

	// check the reply handler is nil
	if replyMeta.handler == nil {
		return NewErrorByDebug(
			"Reply handler is nil",
			replyMeta.debug,
		)
	}

	// Check reply handler is Func
	fn := reflect.ValueOf(replyMeta.handler)
	if fn.Kind() != reflect.Func {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Reply handler must be func(ctx %s, ...) %s",
				convertTypeToString(contextType),
				convertTypeToString(returnType),
			),
			replyMeta.debug,
		)
	}

	// Check reply handler arguments types
	argumentsErrorPos := getArgumentsErrorPosition(fn)
	if argumentsErrorPos == 0 {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Reply handler 1st argument type must be %s",
				convertTypeToString(contextType),
			),
			replyMeta.debug,
		)
	} else if argumentsErrorPos > 0 {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Reply handler %s argument type <%s> not supported",
				ConvertOrdinalToString(1+uint(argumentsErrorPos)),
				fmt.Sprintf("%s", fn.Type().In(argumentsErrorPos)),
			),
			replyMeta.debug,
		)
	}

	// Check return type
	if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return NewErrorByDebug(
			fmt.Sprintf(
				"Reply handler return type must be %s",
				convertTypeToString(returnType),
			),
			replyMeta.debug,
		)
	}

	// mount the replyRecord
	fileLine := ""
	debugArr := FindLinesByPrefix(replyMeta.debug, "-01")
	if len(debugArr) > 0 {
		arr := strings.Split(debugArr[0], " ")
		if len(arr) == 3 {
			fileLine = arr[2]
		}
	}

	argTypes := make([]reflect.Type, fn.Type().NumIn(), fn.Type().NumIn())
	argStrings := make([]string, fn.Type().NumIn(), fn.Type().NumIn())
	for i := 0; i < len(argTypes); i++ {
		argTypes[i] = fn.Type().In(i)
		argStrings[i] = convertTypeToString(argTypes[i])
	}
	argString := strings.Join(argStrings, ", ")

	cacheFN := RPCReplyCacheFunc(nil)
	if fnTypeString, ok := getFuncKind(replyMeta.handler); ok && p.fnCache != nil {
		cacheFN = p.fnCache.Get(fnTypeString)
	}

	p.repliesMap[replyPath] = &replyNode{
		replyMeta: replyMeta,
		cacheFN:   cacheFN,
		reflectFn: fn,
		callString: fmt.Sprintf(
			"%s(%s) %s",
			replyPath,
			argString,
			convertTypeToString(returnType),
		),
		debugString: fmt.Sprintf("%s %s", replyPath, fileLine),
		argTypes:    argTypes,
		indicator:   newPerformanceIndicator(),
	}

	return nil
}

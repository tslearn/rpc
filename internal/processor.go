package internal

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
)

const rootName = "$"
const freeGroupSize = 1024

var (
	nodeNameRegex = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	echoNameRegex = regexp.MustCompile(`^[_a-zA-Z][_0-9a-zA-Z]*$`)
)

type rpcReplyNode struct {
	serviceNode *rpcServiceNode
	path        string
	echoMeta    *rpcEchoMeta
	cacheFN     RPCReplyCacheFunc
	reflectFn   reflect.Value
	callString  string
	debugString string
	argTypes    []reflect.Type
	indicator   *RPCPerformanceIndicator
}

type rpcServiceNode struct {
	path    string
	addMeta *rpcNodeMeta
	depth   uint
}

// RPCProcessor ...
type RPCProcessor struct {
	isDebug            bool
	fnCache            RPCReplyCache
	onEvalFinish       func(stream *RPCStream, success bool)
	echosMap           map[string]*rpcReplyNode
	nodesMap           map[string]*rpcServiceNode
	maxNodeDepth       uint64
	maxCallDepth       uint64
	threads            []*rpcThread
	freeThreadsCHGroup []chan *rpcThread
	readThreadPos      uint64
	writeThreadPos     uint64
	RPCLock
}

// NewRPCProcessor ...
func NewRPCProcessor(
	isDebug bool,
	numOfThreads uint,
	maxNodeDepth uint,
	maxCallDepth uint,
	onEvalFinish func(stream *RPCStream, success bool),
	fnCache RPCReplyCache,
) *RPCProcessor {
	if numOfThreads == 0 {
		return nil
	} else if maxNodeDepth == 0 {
		return nil
	} else if maxCallDepth == 0 {
		return nil
	} else if onEvalFinish == nil {
		return nil
	} else {
		size := ((numOfThreads + freeGroupSize - 1) / freeGroupSize) * freeGroupSize
		ret := &RPCProcessor{
			isDebug:            isDebug,
			fnCache:            fnCache,
			onEvalFinish:       onEvalFinish,
			echosMap:           make(map[string]*rpcReplyNode),
			nodesMap:           make(map[string]*rpcServiceNode),
			maxNodeDepth:       uint64(maxNodeDepth),
			maxCallDepth:       uint64(maxCallDepth),
			threads:            make([]*rpcThread, size, size),
			freeThreadsCHGroup: nil,
			readThreadPos:      0,
			writeThreadPos:     0,
		}
		// mount root node
		ret.nodesMap[rootName] = &rpcServiceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
		}
		return ret
	}
}

func (p *RPCProcessor) Start() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		size := len(p.threads)

		if p.freeThreadsCHGroup != nil {
			return NewRPCError("RPCProcessor: Start: it has already benn started")
		} else {
			freeThreadsCHGroup := make(
				[]chan *rpcThread,
				freeGroupSize,
				freeGroupSize,
			)
			for i := 0; i < freeGroupSize; i++ {
				freeThreadsCHGroup[i] = make(
					chan *rpcThread,
					size/freeGroupSize,
				)
			}
			p.freeThreadsCHGroup = freeThreadsCHGroup

			for i := 0; i < size; i++ {
				thread := newThread(
					p,
					func(thread *rpcThread, stream *RPCStream, success bool) {
						p.onEvalFinish(stream, success)
						freeThreadsCHGroup[atomic.AddUint64(
							&p.writeThreadPos,
							1,
						)%freeGroupSize] <- thread
					},
				)
				p.threads[i] = thread
				p.freeThreadsCHGroup[i%freeGroupSize] <- thread
			}
			return nil
		}
	}))
}

func (p *RPCProcessor) PutStream(stream *RPCStream) bool {
	if freeThreadsCHGroup := p.freeThreadsCHGroup; freeThreadsCHGroup != nil {
		thread := <-freeThreadsCHGroup[atomic.AddUint64(
			&p.readThreadPos,
			1,
		)%freeGroupSize]
		if thread != nil {
			return thread.PutStream(stream)
		} else {
			return false
		}
	} else {
		return false
	}
}

func (p *RPCProcessor) Stop() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		if p.freeThreadsCHGroup == nil {
			return NewRPCError("RPCProcessor: Start: it has already benn stopped")
		} else {
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
				return NewRPCError(ConcatString(
					"RPCProcessor: Stop: The following routine still running: \n\t",
					strings.Join(errList, "\n\t"),
				))
			} else {
				return nil
			}
		}
	}))
}

// BuildCache ...
func (p *RPCProcessor) BuildCache(pkgName string, path string) RPCError {
	retMap := make(map[string]bool)
	for _, echo := range p.echosMap {
		if fnTypeString, ok := getFuncKind(echo.echoMeta.handler); ok {
			retMap[fnTypeString] = true
		}
	}

	fnKinds := make([]string, 0)
	for key := range retMap {
		fnKinds = append(fnKinds, key)
	}

	return buildFuncCache(pkgName, path, fnKinds)
}

// AddService ...
func (p *RPCProcessor) AddService(
	name string,
	service Service,
	debug string,
) RPCError {
	serviceMeta, ok := service.(*rpcService)
	if !ok {
		return NewRPCErrorByDebug(
			"Service is nil",
			debug,
		)
	}

	return p.mountNode(rootName, &rpcNodeMeta{
		name:        name,
		serviceMeta: serviceMeta,
		debug:       debug,
	})
}

func (p *RPCProcessor) mountNode(
	parentServiceNodePath string,
	nodeMeta *rpcNodeMeta,
) RPCError {
	// check nodeMeta is not nil
	if nodeMeta == nil {
		return NewRPCError("rpc: mountNode: nodeMeta is nil")
	}

	// check nodeMeta.name is valid
	if !nodeNameRegex.MatchString(nodeMeta.name) {
		return NewRPCErrorByDebug(
			fmt.Sprintf("Service name \"%s\" is illegal", nodeMeta.name),
			nodeMeta.debug,
		)
	}

	// check nodeMeta.serviceMeta is not nil
	if nodeMeta.serviceMeta == nil {
		return NewRPCErrorByDebug(
			"Service is nil",
			nodeMeta.debug,
		)
	}

	// check max node depth overflow
	parentNode, ok := p.nodesMap[parentServiceNodePath]
	if !ok {
		return NewRPCErrorByDebug(
			"rpc: mountNode: parentNode is nil",
			nodeMeta.debug,
		)
	}
	servicePath := parentServiceNodePath + "." + nodeMeta.name
	if uint64(parentNode.depth+1) > p.maxNodeDepth {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Service path depth %s is too long, it must be less or equal than %d",
				servicePath,
				p.maxNodeDepth,
			),
			nodeMeta.debug,
		)
	}

	// check the mount path is not occupied
	if item, ok := p.nodesMap[servicePath]; ok {
		return NewRPCErrorByDebug(
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

	node := &rpcServiceNode{
		path:    servicePath,
		addMeta: nodeMeta,
		depth:   parentNode.depth + 1,
	}

	// mount the node
	p.nodesMap[servicePath] = node

	// mount the replies
	for _, echoMeta := range nodeMeta.serviceMeta.replies {
		err := p.mountEcho(node, echoMeta)
		if err != nil {
			delete(p.nodesMap, servicePath)
			return err
		}
	}

	// mount children
	for _, v := range nodeMeta.serviceMeta.children {
		err := p.mountNode(node.path, v)
		if err != nil {
			delete(p.nodesMap, servicePath)
			return err
		}
	}

	return nil
}

func (p *RPCProcessor) mountEcho(
	serviceNode *rpcServiceNode,
	echoMeta *rpcEchoMeta,
) RPCError {
	// check the node is nil
	if serviceNode == nil {
		return NewRPCError("rpc: mountEcho: node is nil")
	}

	// check the echoMeta is nil
	if echoMeta == nil {
		return NewRPCError("rpc: mountEcho: echoMeta is nil")
	}

	// check the name
	if !echoNameRegex.MatchString(echoMeta.name) {
		return NewRPCErrorByDebug(
			fmt.Sprintf("Reply name %s is illegal", echoMeta.name),
			echoMeta.debug,
		)
	}

	// check the echo path is not occupied
	echoPath := serviceNode.path + ":" + echoMeta.name
	if item, ok := p.echosMap[echoPath]; ok {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Reply name %s is duplicated",
				echoMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				AddPrefixPerLine(echoMeta.debug, "\t"),
				AddPrefixPerLine(item.echoMeta.debug, "\t"),
			),
		)
	}

	// check the echo handler is nil
	if echoMeta.handler == nil {
		return NewRPCErrorByDebug(
			"Reply handler is nil",
			echoMeta.debug,
		)
	}

	// Check echo handler is Func
	fn := reflect.ValueOf(echoMeta.handler)
	if fn.Kind() != reflect.Func {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Reply handler must be func(ctx %s, ...) %s",
				convertTypeToString(contextType),
				convertTypeToString(returnType),
			),
			echoMeta.debug,
		)
	}

	// Check echo handler arguments types
	argumentsErrorPos := getArgumentsErrorPosition(fn)
	if argumentsErrorPos == 0 {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Reply handler 1st argument type must be %s",
				convertTypeToString(contextType),
			),
			echoMeta.debug,
		)
	} else if argumentsErrorPos > 0 {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Reply handler %s argument type <%s> not supported",
				ConvertOrdinalToString(1+uint(argumentsErrorPos)),
				fmt.Sprintf("%s", fn.Type().In(argumentsErrorPos)),
			),
			echoMeta.debug,
		)
	}

	// Check return type
	if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Reply handler return type must be %s",
				convertTypeToString(returnType),
			),
			echoMeta.debug,
		)
	}

	// mount the echoRecord
	fileLine := ""
	debugArr := FindLinesByPrefix(echoMeta.debug, "-01")
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
	if fnTypeString, ok := getFuncKind(echoMeta.handler); ok && p.fnCache != nil {
		cacheFN = p.fnCache.Get(fnTypeString)
	}

	p.echosMap[echoPath] = &rpcReplyNode{
		serviceNode: serviceNode,
		path:        echoPath,
		echoMeta:    echoMeta,
		cacheFN:     cacheFN,
		reflectFn:   fn,
		callString: fmt.Sprintf(
			"%s(%s) %s",
			echoPath,
			argString,
			convertTypeToString(returnType),
		),
		debugString: fmt.Sprintf("%s %s", echoPath, fileLine),
		argTypes:    argTypes,
		indicator:   NewRPCPerformanceIndicator(),
	}

	return nil
}

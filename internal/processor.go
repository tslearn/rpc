package internal

import (
	"fmt"
	"github.com/tslearn/rpcc/util"
	"math/rand"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

const (
	rootName                 = "$"
	numOfThreadPerThreadPool = 8
	numOfThreadPoolPerCore   = 2
	numOfMinThreadPool       = 2
	numOfMaxThreadPool       = 2
	//numOfThreadPerThreadPool = 8192
	//numOfThreadPoolPerCore   = 2
	//numOfMinThreadPool       = 2
	//numOfMaxThreadPool       = 64
)

var (
	nodeNameRegex = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	echoNameRegex = regexp.MustCompile(`^[_a-zA-Z][_0-9a-zA-Z]*$`)
)

func getFuncKind(fn interface{}) (string, bool) {
	if fn == nil {
		return "", false
	} else if reflectFn := reflect.ValueOf(fn); reflectFn.Kind() != reflect.Func {
		return "", false
	} else if reflectFn.Type().NumIn() < 1 ||
		reflectFn.Type().In(0) != reflect.ValueOf(nilContext).Type() {
		return "", false
	} else if reflectFn.Type().NumOut() != 1 ||
		reflectFn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return "", false
	} else {
		sb := util.NewStringBuilder()
		defer sb.Release()

		for i := 1; i < reflectFn.Type().NumIn(); i++ {
			switch reflectFn.Type().In(i) {
			case bytesType:
				sb.AppendByte('X')
			case arrayType:
				sb.AppendByte('A')
			case mapType:
				sb.AppendByte('M')
			case int64Type:
				sb.AppendByte('I')
			case uint64Type:
				sb.AppendByte('U')
			case boolType:
				sb.AppendByte('B')
			case float64Type:
				sb.AppendByte('F')
			case stringType:
				sb.AppendByte('S')
			default:
				return "", false
			}
		}

		return sb.String(), true
	}
}

func convertTypeToString(reflectType reflect.Type) string {
	switch reflectType {
	case nil:
		return "<nil>"
	case contextType:
		return "rpc.Context"
	case returnType:
		return "rpc.Return"
	case bytesType:
		return "rpc.Bytes"
	case arrayType:
		return "rpc.Array"
	case mapType:
		return "rpc.Map"
	case boolType:
		return "rpc.Bool"
	case int64Type:
		return "rpc.Int"
	case uint64Type:
		return "rpc.Uint"
	case float64Type:
		return "rpc.Float"
	case stringType:
		return "rpc.String"
	default:
		return reflectType.String()
	}
}

func getArgumentsErrorPosition(fn reflect.Value) int {
	if fn.Type().NumIn() < 1 {
		return 0
	} else if fn.Type().In(0) != reflect.ValueOf(nilContext).Type() {
		return 0
	} else {
		for i := 1; i < fn.Type().NumIn(); i++ {
			switch fn.Type().In(i) {
			case bytesType:
				continue
			case arrayType:
				continue
			case mapType:
				continue
			case int64Type:
				continue
			case uint64Type:
				continue
			case boolType:
				continue
			case float64Type:
				continue
			case stringType:
				continue
			default:
				return i
			}
		}
		return -1
	}
}

type rpcEchoNode struct {
	serviceNode *rpcServiceNode
	path        string
	echoMeta    *rpcEchoMeta
	cacheFN     RPCCacheFunc
	reflectFn   reflect.Value
	callString  string
	debugString string
	argTypes    []reflect.Type
	indicator   *PerformanceIndicator
}

type rpcServiceNode struct {
	path    string
	addMeta *rpcNodeMeta
	depth   uint
}

// RPCProcessor ...
type RPCProcessor struct {
	isRunning    bool
	fnCache      RPCCache
	callback     fnProcessorCallback
	echosMap     map[string]*rpcEchoNode
	nodesMap     map[string]*rpcServiceNode
	threadPools  []*rpcThreadPool
	maxNodeDepth uint64
	maxCallDepth uint64
	util.AutoLock
}

var fnGetRuntimeNumberOfCPU = func() int {
	return runtime.NumCPU()
}

// NewRPCProcessor ...
func NewRPCProcessor(
	maxNodeDepth uint,
	maxCallDepth uint,
	callback fnProcessorCallback,
	fnCache RPCCache,
) *RPCProcessor {
	numOfThreadPool := uint32(fnGetRuntimeNumberOfCPU() * numOfThreadPoolPerCore)
	if numOfThreadPool < numOfMinThreadPool {
		numOfThreadPool = numOfMinThreadPool
	}
	if numOfThreadPool > numOfMaxThreadPool {
		numOfThreadPool = numOfMaxThreadPool
	}

	ret := &RPCProcessor{
		isRunning:    false,
		fnCache:      fnCache,
		callback:     callback,
		echosMap:     make(map[string]*rpcEchoNode),
		nodesMap:     make(map[string]*rpcServiceNode),
		threadPools:  make([]*rpcThreadPool, numOfThreadPool, numOfThreadPool),
		maxNodeDepth: uint64(maxNodeDepth),
		maxCallDepth: uint64(maxCallDepth),
	}

	// mount root node
	ret.nodesMap[rootName] = &rpcServiceNode{
		path:    rootName,
		addMeta: nil,
		depth:   0,
	}

	return ret
}

// Start ...
func (p *RPCProcessor) Start() bool {
	return p.CallWithLock(func() interface{} {
		if !p.isRunning {
			p.isRunning = true
			for i := 0; i < len(p.threadPools); i++ {
				p.threadPools[i] = newThreadPool(p)
			}
			return true
		}

		return false
	}).(bool)
}

// Stop ...
func (p *RPCProcessor) Stop() RPCError {
	return ConvertToRPCError(p.CallWithLock(func() interface{} {
		if !p.isRunning {
			return NewRPCError("RPCProcessor: Stop: it has already benn stopped")
		} else {
			p.isRunning = false
			numOfThreadPools := len(p.threadPools)
			closeCH := make(chan []string, numOfThreadPools)
			for i := 0; i < numOfThreadPools; i++ {
				go func(idx int) {
					if ok, errList := p.threadPools[idx].stop(); !ok {
						closeCH <- errList
					} else {
						closeCH <- nil
					}
				}(i)
			}

			// wait all thread stop
			errMap := make(map[string]int)
			for i := 0; i < numOfThreadPools; i++ {
				if errList := <-closeCH; errList != nil {
					for _, errString := range errList {
						if v, ok := errMap[errString]; ok {
							errMap[errString] = v + 1
						} else {
							errMap[errString] = 1
						}
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
				return NewRPCError(util.ConcatString(
					"RPCProcessor: Stop: The following routine still running: \n\t",
					strings.Join(errList, "\n\t"),
				))
			} else {
				return nil
			}
		}
	}))
}

// PutStream ...
func (p *RPCProcessor) PutStream(stream *RPCStream) bool {
	// PutStream stream in a random thread pool
	threadPool := p.threadPools[int(rand.Int31())%len(p.threadPools)]
	if threadPool != nil {
		if thread := threadPool.allocThread(); thread != nil {
			thread.put(stream)
			return true
		}
		return false
	}

	return false
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
				util.AddPrefixPerLine(nodeMeta.debug, "\t"),
				util.AddPrefixPerLine(item.addMeta.debug, "\t"),
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

	// mount the echos
	for _, echoMeta := range nodeMeta.serviceMeta.echos {
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
			fmt.Sprintf("Echo name %s is illegal", echoMeta.name),
			echoMeta.debug,
		)
	}

	// check the echo path is not occupied
	echoPath := serviceNode.path + ":" + echoMeta.name
	if item, ok := p.echosMap[echoPath]; ok {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Echo name %s is duplicated",
				echoMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				util.AddPrefixPerLine(echoMeta.debug, "\t"),
				util.AddPrefixPerLine(item.echoMeta.debug, "\t"),
			),
		)
	}

	// check the echo handler is nil
	if echoMeta.handler == nil {
		return NewRPCErrorByDebug(
			"Echo handler is nil",
			echoMeta.debug,
		)
	}

	// Check echo handler is Func
	fn := reflect.ValueOf(echoMeta.handler)
	if fn.Kind() != reflect.Func {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Echo handler must be func(ctx %s, ...) %s",
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
				"Echo handler 1st argument type must be %s",
				convertTypeToString(contextType),
			),
			echoMeta.debug,
		)
	} else if argumentsErrorPos > 0 {
		return NewRPCErrorByDebug(
			fmt.Sprintf(
				"Echo handler %s argument type <%s> not supported",
				util.ConvertOrdinalToString(1+uint(argumentsErrorPos)),
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
				"Echo handler return type must be %s",
				convertTypeToString(returnType),
			),
			echoMeta.debug,
		)
	}

	// mount the echoRecord
	fileLine := ""
	debugArr := util.FindLinesByPrefix(echoMeta.debug, "-01")
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

	cacheFN := RPCCacheFunc(nil)
	if fnTypeString, ok := getFuncKind(echoMeta.handler); ok && p.fnCache != nil {
		cacheFN = p.fnCache.Get(fnTypeString)
	}

	p.echosMap[echoPath] = &rpcEchoNode{
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
		indicator:   NewPerformanceIndicator(),
	}

	return nil
}

package common

import (
	"fmt"
	"github.com/rpccloud/rpc/common"
	"math/rand"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RootName            = "$"
	NumOfThreadPerBlock = 16
	NumOfBlockPerSlot   = 512
	NumOfThreadGCSwipe  = 8192
	NumOfSlotPerCore    = 4
	NumOfMinSlot        = 4
	NumOfMaxSlot        = 128
)

var (
	NodeNameRegex = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	EchoNameRegex = regexp.MustCompile(`^[_a-zA-Z][_0-9a-zA-Z]*$`)
)

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcThread                                       ///////////////
////////////////////////////////////////////////////////////////////////////////

//func (p *executor) startRun() {
//	atomic.StoreInt64(&p.execNS, common.TimeNowNS())
//}
//
//func (p *executor) stopRun() {
//	p.execDepth = 0
//	p.execRecord = nil
//	atomic.StoreInt64(&p.execNS, 0)
//	p.execArgs = p.execArgs[:0]
//}

func consume(processor *rpcProcessor, stream *common.RPCStream) {

}

type rpcThread struct {
	processor      *rpcProcessor
	ch             chan *common.RPCStream
	execNS         int64
	execStream     *common.RPCStream
	execDepth      uint64
	execFrom       string
	execEchoNode   *rpcEchoNode
	execArgs       []reflect.Value
	execVarInt64   int64
	execVarUint64  uint64
	execVarFloat64 float64
	execVarBool    bool
	execSuccessful bool
}

func newThread(processor *rpcProcessor) *rpcThread {
	return &rpcThread{
		processor:  processor,
		ch:         make(chan *common.RPCStream),
		execArgs:   make([]reflect.Value, 0, 16),
		execStream: common.NewRPCStream(),
		execNS:     0,
	}
}

func (p *rpcThread) getRunDuration() time.Duration {
	timeNS := atomic.LoadInt64(&p.execNS)
	if timeNS == 0 {
		return 0
	}
	return common.TimeSpanFrom(timeNS)
}

func (p *rpcThread) toRun() bool {
	return atomic.CompareAndSwapInt64(&p.execNS, 0, common.TimeNowNS())
}

func (p *rpcThread) toSweep() {
	p.execDepth = 0
	p.execEchoNode = nil
	p.execArgs = p.execArgs[:0]
	atomic.StoreInt64(&p.execNS, -1)
}

func (p *rpcThread) toFree() bool {
	return atomic.CompareAndSwapInt64(&p.execNS, -1, 0)
}

func (p *rpcThread) start() {
	go func() {
		for node := <-p.ch; node != nil; node = <-p.ch {
			consume(p.processor, node)
			p.toSweep()
		}
	}()
}

func (p *rpcThread) put(stream *common.RPCStream) {
	p.ch <- stream
}

func (p *rpcThread) stop() {
	close(p.ch)
}

func (p *rpcThread) eval(inStream *common.RPCStream) Return {
	// create context
	ctx := serverContext{thread: p}

	// if the header is error, we can not find the method to return
	headerBytes, ok := inStream.ReadUnsafeBytes()
	if !ok || len(headerBytes) != 12 {
		p.processor.logger.Error("rpc data format error")
		return nilReturn
	}
	p.execStream.WriteBytes(headerBytes)

	// read echo path
	echoPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return ctx.writeError("rpc data format error", "")
	}
	if p.execEchoNode, ok = p.processor.echosMap[echoPath]; !ok {
		return ctx.Errorf("rpc echo path %s is not mounted", echoPath)
	}

	// read depth
	if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return ctx.writeError("rpc data format error", "")
	}
	if p.execDepth > p.processor.maxCallDepth {
		return ctx.Errorf(
			"rpc current call depth (%d) is overflow. limit(%d)",
			p.execDepth,
			p.processor.maxCallDepth,
		)
	}

	// read from
	if p.execFrom, ok = inStream.ReadUnsafeString(); !ok {
		return ctx.writeError("rpc data format error", "")
	}

	// build callArgs
	argStartPos := inStream.GetReadPos()

	defer func() {
		if err := recover(); err != nil {
			ctx.writeError(
				fmt.Sprintf(
					"rpc call %s runtime error: %s",
					p.execEchoNode.callString,
					err,
				),
				"",
			)
		}
		p.execEchoNode.indicator.count(
			p.getRunDuration(),
			p.execFrom,
			p.execSuccessful,
		)
	}()

	fnCache := p.execEchoNode.fnCache
	if fnCache != nil {
		ok = fnCache(&ctx, inStream, p.execEchoNode.echoMeta.handler)
	} else {
		p.execArgs = append(p.execArgs, reflect.ValueOf(ctx))
		for i := 1; i < len(p.execEchoNode.argTypes); i++ {
			var rv reflect.Value

			switch p.execEchoNode.argTypes[i].Kind() {
			case reflect.Int64:
				p.execVarInt64, ok = inStream.ReadInt64()
				rv = reflect.ValueOf(p.execVarInt64)
				break
			case reflect.Uint64:
				p.execVarUint64, ok = inStream.ReadUint64()
				rv = reflect.ValueOf(p.execVarUint64)
				break
			case reflect.Float64:
				p.execVarFloat64, ok = inStream.ReadFloat64()
				rv = reflect.ValueOf(p.execVarFloat64)
				break
			case reflect.Bool:
				p.execVarBool, ok = inStream.ReadBool()
				rv = reflect.ValueOf(p.execVarBool)
				break
			default:
				if p.execEchoNode.argTypes[i] == reflect.ValueOf(rpcBytes).Type() {
					bVar := inStream.ReadRPCBytes()
					ok = bVar.OK()
					rv = reflect.ValueOf(bVar)
				} else if p.execEchoNode.argTypes[i] == reflect.ValueOf(rpcString).Type() {
					sVar := inStream.ReadRPCString()
					ok = sVar.OK()
					rv = reflect.ValueOf(sVar)
				} else if p.execEchoNode.argTypes[i] == reflect.ValueOf(rpcArray).Type() {
					aVar, ok := inStream.ReadRPCArray()
					rv = reflect.ValueOf(aVar)
				} else if p.execEchoNode.argTypes[i] == reflect.ValueOf(rpcMap).Type() {
					mVar, ok := inStream.ReadRPCMap()
					rv = reflect.ValueOf(mVar)
				} else {
					ok = false
				}
			}

			if !ok {
				break
			}

			p.execArgs = append(p.execArgs, rv)
		}
	}

	if ok {
		ok = inStream.IsReadFinish()
	}
	if !ok {
		inStream.SetReadPos(argStartPos)
		remoteArgsType := make([]string, 0, 0)
		for inStream.CanReadNext() {
			val, ok := inStream.Read()
			if !ok {
				return ctx.writeError("rpc data format error", "")
			}

			if val == nil {
				remoteArgsType = append(remoteArgsType, "nil")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(readTypeBytes).Type() {
				remoteArgsType = append(remoteArgsType, "RPCBytes")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(readTypeString).Type() {
				remoteArgsType = append(remoteArgsType, "RPCString")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(readTypeArray).Type() {
				remoteArgsType = append(remoteArgsType, "RPCArray")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(readTypeMap).Type() {
				remoteArgsType = append(remoteArgsType, "RPCMap")
			} else {
				remoteArgsType = append(remoteArgsType, reflect.ValueOf(val).Type().String())
			}
		}

		return ctx.writeError(
			fmt.Sprintf(
				"rpc echo arguments not match\nCalled: %s(%s) Return\nRequired: %s",
				echoPath,
				strings.Join(remoteArgsType, ", "),
				p.execEchoNode.callString,
			),
			"",
		)
	}

	if fnCache == nil {
		p.execEchoNode.reflectFn.Call(p.execArgs)
	}

	return nilReturn
}

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcThreadSlot                                   ///////////////
////////////////////////////////////////////////////////////////////////////////
type rpcThreadArray [NumOfThreadPerBlock]*rpcThread
type rpcThreadSlot struct {
	isRunning  bool
	threads    []*rpcThread
	cacheArray chan *rpcThreadArray
	emptyArray chan *rpcThreadArray
	sync.Mutex
}

func newThreadSlot() *rpcThreadSlot {
	size := uint32(NumOfBlockPerSlot * NumOfThreadPerBlock)
	return &rpcThreadSlot{
		isRunning: false,
		threads:   make([]*rpcThread, size, size),
	}
}

func (p *rpcThreadSlot) start(processor *rpcProcessor) bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.cacheArray = make(chan *rpcThreadArray, NumOfBlockPerSlot)
		p.emptyArray = make(chan *rpcThreadArray, NumOfBlockPerSlot)
		for i := 0; i < NumOfBlockPerSlot; i++ {
			threadArray := &rpcThreadArray{}
			for n := 0; n < NumOfThreadPerBlock; n++ {
				thread := newThread(processor)
				p.threads[i*NumOfThreadPerBlock+n] = thread
				threadArray[n] = thread
				thread.start()
			}
			p.cacheArray <- threadArray
		}
		go p.gc()
		p.isRunning = true
		return true
	} else {
		return false
	}
}

func (p *rpcThreadSlot) stop() bool {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		for i := 0; i < len(p.threads); i++ {
			p.threads[i].stop()
			p.threads[i] = nil
		}
		close(p.cacheArray)
		close(p.emptyArray)
		p.isRunning = false
		return true
	} else {
		return false
	}
}

func (p *rpcThreadSlot) gc() {
	gIndex := 0
	threadArray := <-p.emptyArray
	arrIndex := 0
	totalThreads := len(p.threads)
	for {
		for i := 0; i < NumOfThreadGCSwipe; i++ {
			gIndex = (gIndex + 1) % totalThreads
			if p.threads[gIndex].execNS == -1 {
				if p.threads[gIndex].toFree() {
					threadArray[arrIndex] = p.threads[gIndex]
					arrIndex++
					if arrIndex == NumOfThreadPerBlock {
						p.cacheArray <- threadArray
						threadArray = <-p.emptyArray
						arrIndex = 0
						if threadArray == nil {
							return
						}
					}
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (p *rpcThreadSlot) put(stream *common.RPCStream) {
	arrayOfThreads := <-p.cacheArray

	for i := 0; i < NumOfThreadPerBlock; i++ {
		thread := arrayOfThreads[i]
		thread.toRun()
		thread.put(stream)
		arrayOfThreads[i] = nil
	}

	p.emptyArray <- arrayOfThreads
}

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcProcessor                                    ///////////////
////////////////////////////////////////////////////////////////////////////////
type rpcProcessor struct {
	isRunning    bool
	logger       *common.Logger
	echosMap     map[string]*rpcEchoNode
	servicesMap  map[string]*rpcServiceNode
	slots        []*rpcThreadSlot
	maxNodeDepth uint64
	maxCallDepth uint64
	sync.Mutex
}

func newProcessor(
	logger *common.Logger,
	maxNodeDepth uint,
	maxCallDepth uint,
) *rpcProcessor {
	numOfSlots := uint32(runtime.NumCPU() * NumOfSlotPerCore)
	if numOfSlots < NumOfMinSlot {
		numOfSlots = NumOfMinSlot
	}
	if numOfSlots > NumOfMaxSlot {
		numOfSlots = NumOfMaxSlot
	}

	ret := &rpcProcessor{
		isRunning:    false,
		logger:       logger,
		echosMap:     make(map[string]*rpcEchoNode),
		servicesMap:  make(map[string]*rpcServiceNode),
		slots:        make([]*rpcThreadSlot, numOfSlots, numOfSlots),
		maxNodeDepth: uint64(maxNodeDepth),
		maxCallDepth: uint64(maxCallDepth),
	}

	// mount root node
	ret.servicesMap[RootName] = &rpcServiceNode{
		path:    RootName,
		addMeta: nil,
		depth:   0,
	}

	return ret
}

func (p *rpcProcessor) start() bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		for i := 0; i < len(p.slots); i++ {
			p.slots[i] = newThreadSlot()
			p.slots[i].start(p)
		}
		p.isRunning = true
		return true
	} else {
		return false
	}
}

func (p *rpcProcessor) stop() bool {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		for i := 0; i < len(p.slots); i++ {
			p.slots[i].stop()
			p.slots[i] = nil
		}
		p.isRunning = false
		return true
	} else {
		return false
	}
}

func (p *rpcProcessor) put(
	stream *common.RPCStream,
	goroutineFixedRand *rand.Rand,
) {
	// get a random uint32
	randInt := 0
	if goroutineFixedRand == nil {
		randInt = int(common.GetRandUint32())
	} else {
		randInt = goroutineFixedRand.Int()
	}
	// put stream in a random slot
	p.slots[randInt%len(p.slots)].put(stream)
}

func (p *rpcProcessor) AddService(
	name string,
	serviceMeta *rpcServiceMeta,
) *rpcProcessor {
	err := p.mountService(RootName, &rpcAddMeta{
		name:        name,
		serviceMeta: serviceMeta,
		debug:       common.GetStackString(1),
	})

	if err != nil {
		p.logger.Error(err)
	}

	return p
}

func (p *rpcProcessor) mountService(
	parentServiceNodePath string,
	addMeta *rpcAddMeta,
) common.RPCError {
	// check addMeta is not nil
	if addMeta == nil {
		return common.NewRPCError("mountService addMeta is nil")
	}

	// check addMeta.serviceMeta is not nil
	if addMeta.serviceMeta == nil {
		return common.NewRPCErrorWithDebug(
			"mountService service is nil",
			addMeta.debug,
		)
	}

	// check addMeta.name is valid
	if !NodeNameRegex.MatchString(addMeta.name) {
		return common.NewRPCErrorWithDebug(
			fmt.Sprintf("mountService name %s is illegal", addMeta.name),
			addMeta.debug,
		)
	}

	// check max node depth overflow
	parentNode, ok := p.servicesMap[parentServiceNodePath]
	if !ok {
		return common.NewRPCError(
			"mountService parentNode is nil",
		)
	}
	servicePath := parentServiceNodePath + "." + addMeta.name
	if uint64(parentNode.depth+1) > p.maxNodeDepth {
		return common.NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Service path %s is too long, it must be less or equal than %d",
				servicePath,
				p.maxNodeDepth,
			),
			addMeta.debug,
		)
	}

	// check the mount path is not occupied
	if item, ok := p.servicesMap[servicePath]; ok {
		return common.NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Add name %s is duplicated",
				addMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				common.AddPrefixPerLine(addMeta.debug, "\t"),
				common.AddPrefixPerLine(item.addMeta.debug, "\t"),
			),
		)
	}

	node := &rpcServiceNode{
		path:    servicePath,
		addMeta: addMeta,
		depth:   parentNode.depth + 1,
	}

	// mount the node
	p.servicesMap[servicePath] = node

	// mount the echos
	for _, echoMeta := range addMeta.serviceMeta.echos {
		err := p.mountEcho(node, echoMeta)
		if err != nil {
			return err
		}
	}

	// mount children
	for _, v := range addMeta.serviceMeta.children {
		err := p.mountService(node.path, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *rpcProcessor) mountEcho(
	serviceNode *rpcServiceNode,
	echoMeta *rpcEchoMeta,
) common.RPCError {
	// check the node is nil
	if serviceNode == nil {
		return common.NewRPCError("mountEcho node is nil")
	}

	// check the echoMeta is nil
	if echoMeta == nil {
		return common.NewRPCError("mountEcho scheme is nil")
	}

	// check the name
	if !EchoNameRegex.MatchString(echoMeta.name) {
		return common.NewRPCErrorWithDebug(
			fmt.Sprintf("Echo name %s is illegal", echoMeta.name),
			echoMeta.debug,
		)
	}

	// check the echo path is not occupied
	echoPath := serviceNode.path + ":" + echoMeta.name
	if item, ok := p.echosMap[echoPath]; ok {
		return common.NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Echo name %s is duplicated",
				echoMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				common.AddPrefixPerLine(echoMeta.debug, "\t"),
				common.AddPrefixPerLine(item.echoMeta.debug, "\t"),
			),
		)
	}

	// check the echo handler is nil
	if echoMeta.handler == nil {
		return common.NewRPCErrorWithDebug(
			"Echo handler is nil",
			echoMeta.debug,
		)
	}

	// Check echo handler is Func
	fn := reflect.ValueOf(echoMeta.handler)
	if fn.Kind() != reflect.Func {
		return common.NewRPCErrorWithDebug(
			"Echo handler must be func(ctx Context, ...) Return",
			echoMeta.debug,
		)
	}

	// Check echo handler arguments types
	argumentsErrorPos := getArgumentsErrorPosition(fn)
	if argumentsErrorPos == 0 {
		return common.NewRPCErrorWithDebug(
			"Echo handler 1st argument type must be Context",
			echoMeta.debug,
		)
	} else if argumentsErrorPos > 0 {
		return common.NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Echo handler %s argument type <%s> not supported",
				common.ConvertOrdinalToString(1+uint(argumentsErrorPos)),
				fmt.Sprintf("%s", fn.Type().In(argumentsErrorPos)),
			),
			echoMeta.debug,
		)
	}

	// Check return type
	if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(pReturn).Type() {
		return common.NewRPCErrorWithDebug(
			"Echo handler return type must be Return",
			echoMeta.debug,
		)
	}

	// mount the echoRecord
	fileLine := ""
	debugArr := common.FindLinesByPrefix(echoMeta.debug, "-01")
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

		if argTypes[i] == reflect.ValueOf(rpcBytes).Type() {
			argStrings[i] = "RPCBytes"
		} else if argTypes[i] == reflect.ValueOf(rpcString).Type() {
			argStrings[i] = "RPCString"
		} else if argTypes[i] == reflect.ValueOf(rpcArray).Type() {
			argStrings[i] = "RPCArray"
		} else if argTypes[i] == reflect.ValueOf(rpcMap).Type() {
			argStrings[i] = "RPCMap"
		} else {
			argStrings[i] = argTypes[i].String()
		}
	}
	argString := strings.Join(argStrings[1:], ", ")

	p.echosMap[echoPath] = &rpcEchoNode{
		serviceNode: serviceNode,
		path:        echoPath,
		echoMeta:    echoMeta,
		fnCache:     getFCache(echoMeta.handler),
		reflectFn:   fn,
		callString:  fmt.Sprintf("%s(%s)", echoPath, argString),
		debugString: fmt.Sprintf("%s %s", echoPath, fileLine),
		argTypes:    argTypes,
		indicator:   newIndicator(),
	}

	p.logger.Infof(
		"rpc: mounted %s %s",
		p.echosMap[echoPath].callString,
		fileLine,
	)

	return nil
}

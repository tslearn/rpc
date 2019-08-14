package core

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RootName           = "$"
	NumOfThreadPerSlot = 8192
	NumOfThreadGCSwipe = 4096
	NumOfSlotPerCore   = 4
	NumOfMinSlot       = 4
	NumOfMaxSlot       = 128
)

var (
	NodeNameRegex = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	EchoNameRegex = regexp.MustCompile(`^[_a-zA-Z][_0-9a-zA-Z]*$`)
)

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcThread                                       ///////////////
////////////////////////////////////////////////////////////////////////////////
type rpcThread struct {
	processor      *rpcProcessor
	ch             chan *rpcStream
	execNS         int64
	execStream     *rpcStream
	execDepth      uint64
	execEchoNode   *rpcEchoNode
	execArgs       []reflect.Value
	execSuccessful bool

	execInnerContext *rpcInnerContext
}

func newThread(processor *rpcProcessor) *rpcThread {
	ret := rpcThread{
		processor:  processor,
		ch:         make(chan *rpcStream),
		execArgs:   make([]reflect.Value, 0, 16),
		execStream: NewRPCStream(),
		execNS:     0,
	}

	ret.execInnerContext = &rpcInnerContext{
		stream:       nil,
		serverThread: &ret,
		clientThread: nil,
	}

	return &ret
}

func (p *rpcThread) getRunDuration() time.Duration {
	timeNS := atomic.LoadInt64(&p.execNS)
	if timeNS == 0 {
		return 0
	}
	return TimeSpanFrom(timeNS)
}

func (p *rpcThread) toRun() bool {
	return atomic.CompareAndSwapInt64(&p.execNS, 0, TimeNowNS())
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
		for stream := <-p.ch; stream != nil; stream = <-p.ch {
			p.eval(stream)
			p.toSweep()
		}
	}()
}

func (p *rpcThread) stop() {
	close(p.ch)
}

func (p *rpcThread) put(stream *rpcStream) {
	p.ch <- stream
}

func (p *rpcThread) eval(inStream *rpcStream) *rpcReturn {
	processor := p.processor
	// create context
	p.execInnerContext.stream = inStream
	ctx := &rpcContext{inner: p.execInnerContext}

	// if the header is error, we can not find the method to return
	headerBytes, ok := inStream.ReadUnsafeBytes()
	if !ok || len(headerBytes) != 16 {
		processor.logger.Error("rpc data format error")
		return nilReturn
	}
	p.execStream.WriteBytes(headerBytes)

	// read echo path
	echoPath, ok := inStream.ReadUnsafeString()
	if !ok {
		return ctx.writeError("rpc data format error", "")
	}
	if p.execEchoNode, ok = processor.echosMap[echoPath]; !ok {
		return ctx.Errorf("rpc echo path %s is not mounted", echoPath)
	}

	// read depth
	if p.execDepth, ok = inStream.ReadUint64(); !ok {
		return ctx.writeError("rpc data format error", "")
	}
	if p.execDepth > processor.maxCallDepth {
		return ctx.Errorf(
			"rpc current call depth (%d) is overflow. limit(%d)",
			p.execDepth,
			processor.maxCallDepth,
		)
	}

	// read from
	from, ok := inStream.ReadUnsafeString()
	if !ok {
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
				GetStackString(1),
			)
		}
		p.execEchoNode.indicator.count(
			p.getRunDuration(),
			from,
			p.execSuccessful,
		)
	}()

	fnCache := p.execEchoNode.fnCache
	if fnCache != nil {
		ok = fnCache(ctx, inStream, p.execEchoNode.echoMeta.handler)
	} else {
		p.execArgs = append(p.execArgs, reflect.ValueOf(ctx))
		for i := 1; i < len(p.execEchoNode.argTypes); i++ {
			var rv reflect.Value

			switch p.execEchoNode.argTypes[i].Kind() {
			case reflect.Int64:
				iVar, success := inStream.ReadInt64()
				if !success {
					ok = false
				} else {
					rv = reflect.ValueOf(iVar)
				}
				break
			case reflect.Uint64:
				uVar, success := inStream.ReadUint64()
				if !success {
					ok = false
				} else {
					rv = reflect.ValueOf(uVar)
				}
				break
			case reflect.Float64:
				fVar, success := inStream.ReadFloat64()
				if !success {
					ok = false
				} else {
					rv = reflect.ValueOf(fVar)
				}
				break
			case reflect.Bool:
				bVar, success := inStream.ReadBool()
				if !success {
					ok = false
				} else {
					rv = reflect.ValueOf(bVar)
				}
				break
			case reflect.String:
				sVar, success := inStream.ReadString()
				if !success {
					ok = false
				} else {
					rv = reflect.ValueOf(sVar)
				}
				break
			default:
				if p.execEchoNode.argTypes[i] == reflect.ValueOf(emptyBytes).Type() {
					xVar, success := inStream.ReadBytes()
					if !success {
						ok = false
					} else {
						rv = reflect.ValueOf(xVar)
					}
				} else if p.execEchoNode.argTypes[i] == reflect.ValueOf(nilRPCArray).Type() {
					aVar, success := inStream.ReadRPCArray(ctx)
					if !success {
						ok = false
					} else {
						rv = reflect.ValueOf(aVar)
					}
				} else if p.execEchoNode.argTypes[i] == reflect.ValueOf(nilRPCMap).Type() {
					mVar, success := inStream.ReadRPCMap(ctx)
					if !success {
						ok = false
					} else {
						rv = reflect.ValueOf(mVar)
					}
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
			val, ok := inStream.Read(ctx)
			if !ok {
				return ctx.writeError("rpc data format error", "")
			}

			if val == nil {
				remoteArgsType = append(remoteArgsType, "nil")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(emptyBytes).Type() {
				remoteArgsType = append(remoteArgsType, "[]byte")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(nilRPCArray).Type() {
				remoteArgsType = append(remoteArgsType, "RPCArray")
			} else if reflect.ValueOf(val).Type() == reflect.ValueOf(nilRPCMap).Type() {
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

	processor.callback(p.execStream, p.execSuccessful)
	inStream.Reset()
	p.execStream = inStream

	return nilReturn
}

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcThreadSlot                                   ///////////////
////////////////////////////////////////////////////////////////////////////////
type rpcThreadSlot struct {
	isRunning   bool
	gcFinish    chan bool
	threads     []*rpcThread
	freeThreads chan *rpcThread
	sync.Mutex
}

func newThreadSlot() *rpcThreadSlot {
	return &rpcThreadSlot{
		isRunning: false,
		threads:   make([]*rpcThread, NumOfThreadPerSlot, NumOfThreadPerSlot),
	}
}

func (p *rpcThreadSlot) start(processor *rpcProcessor) bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.gcFinish = make(chan bool)
		p.freeThreads = make(chan *rpcThread, NumOfThreadPerSlot)
		for i := 0; i < NumOfThreadPerSlot; i++ {
			thread := newThread(processor)
			thread.start()
			p.threads[i] = thread
			p.freeThreads <- thread
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
		close(p.freeThreads)
		p.isRunning = false
		<-p.gcFinish
		for i := 0; i < len(p.threads); i++ {
			p.threads[i].stop()
			p.threads[i] = nil
		}
		return true
	} else {
		return false
	}
}

func (p *rpcThreadSlot) gc() {
	gIndex := 0
	totalThreads := len(p.threads)
	for p.isRunning {
		for i := 0; i < NumOfThreadGCSwipe; i++ {
			gIndex = (gIndex + 1) % totalThreads
			thread := p.threads[gIndex]
			if thread.execNS == -1 {
				if thread.toFree() {
					p.freeThreads <- thread
				}
			}
		}
		time.Sleep(16 * time.Millisecond)
	}
	p.gcFinish <- true
}

func (p *rpcThreadSlot) put(stream *rpcStream) {
	thread := <-p.freeThreads
	thread.toRun()
	thread.put(stream)
}

////////////////////////////////////////////////////////////////////////////////
///////////////  rpcProcessor                                    ///////////////
////////////////////////////////////////////////////////////////////////////////

type rpcProcessor struct {
	isRunning    bool
	logger       *Logger
	callback     fnProcessorCallback
	echosMap     map[string]*rpcEchoNode
	servicesMap  map[string]*rpcServiceNode
	slots        []*rpcThreadSlot
	maxNodeDepth uint64
	maxCallDepth uint64
	sync.Mutex
}

func newProcessor(
	logger *Logger,
	maxNodeDepth uint,
	maxCallDepth uint,
	callback fnProcessorCallback,
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
		callback:     callback,
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

func (p *rpcProcessor) put(stream *rpcStream) bool {
	// put stream in a random slot
	slot := p.slots[int(GetRandUint32())%len(p.slots)]
	if slot != nil {
		slot.put(stream)
		return true
	} else {
		return false
	}
}

func (p *rpcProcessor) AddService(
	name string,
	serviceMeta *rpcServiceMeta,
) *rpcProcessor {
	err := p.mountService(RootName, &rpcAddMeta{
		name:        name,
		serviceMeta: serviceMeta,
		debug:       GetStackString(1),
	})

	if err != nil {
		p.logger.Error(err)
	}

	return p
}

func (p *rpcProcessor) mountService(
	parentServiceNodePath string,
	addMeta *rpcAddMeta,
) RPCError {
	// check addMeta is not nil
	if addMeta == nil {
		return NewRPCError("mountService addMeta is nil")
	}

	// check addMeta.serviceMeta is not nil
	if addMeta.serviceMeta == nil {
		return NewRPCErrorWithDebug(
			"mountService service is nil",
			addMeta.debug,
		)
	}

	// check addMeta.name is valid
	if !NodeNameRegex.MatchString(addMeta.name) {
		return NewRPCErrorWithDebug(
			fmt.Sprintf("mountService name %s is illegal", addMeta.name),
			addMeta.debug,
		)
	}

	// check max node depth overflow
	parentNode, ok := p.servicesMap[parentServiceNodePath]
	if !ok {
		return NewRPCError(
			"mountService parentNode is nil",
		)
	}
	servicePath := parentServiceNodePath + "." + addMeta.name
	if uint64(parentNode.depth+1) > p.maxNodeDepth {
		return NewRPCErrorWithDebug(
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
		return NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Add name %s is duplicated",
				addMeta.name,
			),
			fmt.Sprintf(
				"Current:\n%s\nConflict:\n%s",
				AddPrefixPerLine(addMeta.debug, "\t"),
				AddPrefixPerLine(item.addMeta.debug, "\t"),
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
) RPCError {
	// check the node is nil
	if serviceNode == nil {
		return NewRPCError("mountEcho node is nil")
	}

	// check the echoMeta is nil
	if echoMeta == nil {
		return NewRPCError("mountEcho scheme is nil")
	}

	// check the name
	if !EchoNameRegex.MatchString(echoMeta.name) {
		return NewRPCErrorWithDebug(
			fmt.Sprintf("Echo name %s is illegal", echoMeta.name),
			echoMeta.debug,
		)
	}

	// check the echo path is not occupied
	echoPath := serviceNode.path + ":" + echoMeta.name
	if item, ok := p.echosMap[echoPath]; ok {
		return NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Echo name %s is duplicated",
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
		return NewRPCErrorWithDebug(
			"Echo handler is nil",
			echoMeta.debug,
		)
	}

	// Check echo handler is Func
	fn := reflect.ValueOf(echoMeta.handler)
	if fn.Kind() != reflect.Func {
		return NewRPCErrorWithDebug(
			"Echo handler must be func(ctx Context, ...) Return",
			echoMeta.debug,
		)
	}

	// Check echo handler arguments types
	argumentsErrorPos := getArgumentsErrorPosition(fn)
	if argumentsErrorPos == 0 {
		return NewRPCErrorWithDebug(
			"Echo handler 1st argument type must be Context",
			echoMeta.debug,
		)
	} else if argumentsErrorPos > 0 {
		return NewRPCErrorWithDebug(
			fmt.Sprintf(
				"Echo handler %s argument type <%s> not supported",
				ConvertOrdinalToString(1+uint(argumentsErrorPos)),
				fmt.Sprintf("%s", fn.Type().In(argumentsErrorPos)),
			),
			echoMeta.debug,
		)
	}

	// Check return type
	if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(nilReturn).Type() {
		return NewRPCErrorWithDebug(
			"Echo handler return type must be Return",
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

		if argTypes[i] == reflect.ValueOf(emptyBytes).Type() {
			argStrings[i] = "[]byte"
		} else if argTypes[i] == reflect.ValueOf(nilRPCArray).Type() {
			argStrings[i] = "RPCArray"
		} else if argTypes[i] == reflect.ValueOf(nilRPCMap).Type() {
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

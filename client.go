package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"net/url"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const sendItemStatusRunning = int32(1)
const sendItemStatusFinish = int32(2)

type sendItem struct {
	id         uint64
	status     int32
	startTime  time.Time
	sendTime   time.Time
	timeout    time.Duration
	returnCH   chan *Stream
	sendStream *Stream
	next       *sendItem
}

var sendItemCache = &sync.Pool{
	New: func() interface{} {
		return &sendItem{
			sendStream: core.NewStream(),
		}
	},
}

func newSendItem() *sendItem {
	ret := sendItemCache.Get().(*sendItem)
	ret.id = 0
	ret.status = sendItemStatusRunning
	ret.startTime = time.Time{}
	ret.sendTime = time.Time{}
	ret.timeout = 0
	ret.returnCH = make(chan *Stream, 1)
	ret.next = nil
	return ret
}

func (p *sendItem) Return(stream *Stream) bool {
	if stream == nil {
		return false
	} else if !atomic.CompareAndSwapInt32(
		&p.status,
		sendItemStatusRunning,
		sendItemStatusFinish,
	) {
		stream.Release()
		return false
	} else {
		p.returnCH <- stream
		return true
	}
}

func (p *sendItem) Timeout() bool {
	if atomic.CompareAndSwapInt32(
		&p.status,
		sendItemStatusRunning,
		sendItemStatusFinish,
	) {
		// return timeout stream
		stream := core.NewStream()
		stream.SetCallbackID(p.sendStream.GetCallbackID())
		stream.WriteUint64(uint64(core.ErrorKindReply))
		stream.WriteString(core.ErrStringTimeout)
		stream.WriteString("")
		p.returnCH <- stream
		return true
	}

	return false
}

func (p *sendItem) Release() {
	close(p.returnCH)
	p.returnCH = nil
	p.sendStream.Reset()
	sendItemCache.Put(p)
}

// Client ...
type Client struct {
	sessionString        string
	conn                 core.IStreamConn
	logWriter            LogWriter
	preSendHead          *sendItem
	preSendTail          *sendItem
	sendMap              map[uint64]*sendItem
	systemSeed           uint64
	readTimeout          time.Duration
	writeTimeout         time.Duration
	readLimit            int64
	currCallbackID       uint64
	maxCallbackID        uint64
	callbackSize         int64
	lastControlSendTime  time.Time
	lastTimeoutCheckTime time.Time
	statusManager        core.StatusManager
	mutex                *sync.Mutex
}

// Dial ...
func Dial(connectString string) (*Client, Error) {
	urlInfo, err := url.Parse(connectString)

	if err != nil {
		return nil, core.NewRuntimePanic(err.Error())
	}

	switch urlInfo.Scheme {
	case "ws":
		return newClient(core.NewWebSocketClientAdapter(connectString)), nil
	default:
		return nil,
			core.NewRuntimePanic(fmt.Sprintf("unknown scheme %s", urlInfo.Scheme))
	}
}

func newClient(adapter core.IClientAdapter) *Client {
	ret := &Client{
		sessionString:        "",
		conn:                 nil,
		logWriter:            NewStdoutLogWriter(),
		preSendHead:          nil,
		preSendTail:          nil,
		sendMap:              make(map[uint64]*sendItem),
		systemSeed:           0,
		readTimeout:          0,
		writeTimeout:         0,
		readLimit:            0,
		currCallbackID:       0,
		maxCallbackID:        0,
		callbackSize:         0,
		lastControlSendTime:  time.Now().Add(-10 * time.Second),
		lastTimeoutCheckTime: time.Now().Add(-10 * time.Second),
		statusManager:        core.StatusManager{},
		mutex:                &sync.Mutex{},
	}

	ret.statusManager.SetRunning(nil)

	go func() {
		for ret.statusManager.IsRunning() {
			adapter.Open(ret.onConnRun, ret.onError)
		}
		ret.statusManager.SetClosed(nil)
	}()

	go func() {
		for ret.statusManager.IsRunning() {
			now := base.TimeNow()
			ret.tryToTimeout(now)
			ret.tryToDeliverControlMessage(now)
			for ret.tryToDeliverPreSendMessage() {
				// loop until failed
			}
			time.Sleep(100 * time.Millisecond)
		}

		ret.statusManager.SetClosed(func() {
			adapter.Close(ret.onError)
		})
	}()

	return ret
}

// Close ...
func (p *Client) Close() bool {
	waitCH := chan bool(nil)

	if !p.statusManager.SetClosing(func(ch chan bool) {
		waitCH = ch
	}) {
		p.onError(core.NewRuntimePanic(
			"it is not running",
		).AddDebug(string(debug.Stack())))
		return false
	}

	select {
	case <-waitCH:
		return true
	case <-time.After(20 * time.Second):
		p.onError(core.NewRuntimePanic(
			"can not close within 20 seconds",
		).AddDebug(string(debug.Stack())))
		return false
	}
}

func (p *Client) initConn(conn core.IStreamConn) Error {
	// get the sequence
	p.mutex.Lock()
	p.systemSeed++
	sequence := p.systemSeed
	p.mutex.Unlock()

	sendStream := core.NewStream()
	backStream := (*core.Stream)(nil)

	defer func() {
		sendStream.Release()
		if backStream != nil {
			backStream.Release()
		}
	}()

	sendStream.SetCallbackID(0)
	sendStream.WriteInt64(controlStreamKindInit)
	sendStream.WriteUint64(sequence)
	sendStream.WriteString(p.sessionString)

	if conn == nil {
		return core.NewKernelPanic(
			"Client: initConn: conn is nil",
		).AddDebug(string(debug.Stack()))
	} else if err := conn.WriteStream(sendStream, 3*time.Second); err != nil {
		return err
	} else if backStream, err = conn.ReadStream(3*time.Second, 0); err != nil {
		return err
	} else if backStream.GetCallbackID() != 0 {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if kind, ok := backStream.ReadInt64(); !ok ||
		kind != controlStreamKindInitBack {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if seq, ok := backStream.ReadUint64(); !ok || seq != sequence {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if sessionString, ok := backStream.ReadString(); !ok ||
		len(sessionString) < 34 {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if readTimeoutMS, ok := backStream.ReadInt64(); !ok ||
		readTimeoutMS <= 0 {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if writeTimeoutMS, ok := backStream.ReadInt64(); !ok ||
		writeTimeoutMS <= 0 {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if readLimit, ok := backStream.ReadInt64(); !ok ||
		readLimit <= 0 {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if callBackSize, ok := backStream.ReadInt64(); !ok ||
		callBackSize <= 0 {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else if !backStream.IsReadFinish() {
		return core.NewProtocolError(core.ErrStringBadStream)
	} else {
		p.sessionString = sessionString
		p.readTimeout = time.Duration(readTimeoutMS) * time.Millisecond
		p.writeTimeout = time.Duration(writeTimeoutMS) * time.Millisecond
		p.readLimit = readLimit
		p.callbackSize = callBackSize
		return nil
	}
}

func (p *Client) onConnRun(conn core.IStreamConn) {
	// init conn
	if err := p.initConn(conn); err != nil {
		p.onError(err)
		return
	}

	// set the conn
	p.mutex.Lock()
	p.conn = conn
	p.mutex.Unlock()

	err := Error(nil)

	// clear the conn when finish
	defer func() {
		if err != nil {
			p.onError(err)
		}

		p.mutex.Lock()
		p.conn = nil
		p.mutex.Unlock()

		if err := conn.Close(); err != nil {
			p.onError(err)
		}
	}()

	// receive messages
	for p.statusManager.IsRunning() {
		if stream, e := conn.ReadStream(p.readTimeout, p.readLimit); e != nil {
			if e != core.ErrTransportStreamConnIsClosed {
				err = e
			}
			return
		} else if callbackID := stream.GetCallbackID(); callbackID > 0 {
			p.mutex.Lock()
			if v, ok := p.sendMap[callbackID]; ok {
				if v.Return(stream) {
					delete(p.sendMap, callbackID)
				}
			}
			p.mutex.Unlock()
		} else if kind, ok := stream.ReadInt64(); !ok ||
			kind != controlStreamKindRequestIdsBack {
			err = core.NewProtocolError(core.ErrStringBadStream)
			return
		} else if maxCallbackID, ok := stream.ReadUint64(); !ok {
			err = core.NewProtocolError(core.ErrStringBadStream)
			return
		} else if !stream.IsReadFinish() {
			err = core.NewProtocolError(core.ErrStringBadStream)
			return
		} else {
			p.mutex.Lock()
			if maxCallbackID > p.maxCallbackID {
				p.maxCallbackID = maxCallbackID
			}
			p.mutex.Unlock()
		}
	}
}

func (p *Client) getHeartbeatDuration() time.Duration {
	return time.Duration(float64(p.readTimeout/time.Millisecond) * 0.8)
}

func (p *Client) tryToDeliverControlMessage(now time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	deltaTime := now.Sub(p.lastControlSendTime)
	if p.conn == nil {
		return
	} else if deltaTime < 1000*time.Millisecond {
		return
	} else if deltaTime < p.getHeartbeatDuration() &&
		int64(len(p.sendMap)) > p.callbackSize/2 {
		return
	} else {
		p.lastControlSendTime = now
		p.systemSeed++

		sendStream := core.NewStream()
		sendStream.SetCallbackID(0)
		sendStream.WriteInt64(controlStreamKindRequestIds)
		sendStream.WriteUint64(p.systemSeed)
		sendStream.WriteUint64(p.currCallbackID)

		for key := range p.sendMap {
			sendStream.WriteUint64(key)
		}

		if err := p.conn.WriteStream(
			sendStream,
			p.writeTimeout,
		); err != nil {
			p.onError(err)
		}

		sendStream.Release()
	}
}

func (p *Client) tryToTimeout(now time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if now.Sub(p.lastTimeoutCheckTime) > 800*time.Millisecond {
		p.lastTimeoutCheckTime = now

		// sweep pre send list
		preValidItem := (*sendItem)(nil)
		item := p.preSendHead
		for item != nil {
			if now.Sub(item.startTime) > item.timeout && item.Timeout() {
				nextItem := item.next

				if preValidItem == nil {
					p.preSendHead = nextItem
				} else {
					preValidItem.next = nextItem
				}

				if item == p.preSendTail {
					p.preSendTail = preValidItem
					if p.preSendTail != nil {
						p.preSendTail.next = nil
					}
				}
				item = nextItem
			} else {
				preValidItem = item
				item = item.next
			}
		}

		// sweep send map
		for key, value := range p.sendMap {
			if now.Sub(value.startTime) > value.timeout && value.Timeout() {
				delete(p.sendMap, key)
			}
		}
	}
}

func (p *Client) tryToDeliverPreSendMessage() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.statusManager.IsRunning() { // not running
		return false
	} else if p.conn == nil { // not connected
		return false
	} else if p.preSendHead == nil { // preSend queue is empty
		return false
	} else if p.currCallbackID >= p.maxCallbackID { // id is not available
		return false
	} else {
		// get and set the send item
		item := p.preSendHead
		if item == p.preSendTail {
			p.preSendHead = nil
			p.preSendTail = nil
		} else {
			p.preSendHead = p.preSendHead.next
		}
		p.currCallbackID++
		item.id = p.currCallbackID
		item.next = nil
		item.sendStream.SetCallbackID(item.id)

		// set to sendMap
		p.sendMap[item.id] = item

		// try to send
		if err := p.conn.WriteStream(
			item.sendStream,
			p.writeTimeout,
		); err != nil {
			p.onError(err)
			return false
		}

		item.sendTime = base.TimeNow()
		return true
	}
}

// SendMessage ...
func (p *Client) SendMessage(
	timeout time.Duration,
	target string,
	args ...interface{},
) (interface{}, Error) {
	item := newSendItem()
	defer item.Release()

	item.startTime = base.TimeNow()
	item.timeout = timeout

	// set depth
	item.sendStream.SetDepth(0)
	// write target
	item.sendStream.WriteString(target)
	// write from
	item.sendStream.WriteString("@")
	// write args
	for i := 0; i < len(args); i++ {
		if item.sendStream.Write(args[i]) != core.StreamWriteOK {
			return nil, core.NewRuntimePanic(
				"Client: send: args not supported",
			)
		}
	}

	// add item to the list tail
	p.mutex.Lock()
	if p.preSendTail == nil {
		p.preSendHead = item
		p.preSendTail = item
	} else {
		p.preSendTail.next = item
		p.preSendTail = item
	}
	p.mutex.Unlock()

	// wait for response
	if stream := <-item.returnCH; stream == nil {
		return nil, core.NewKernelPanic("stream is nil").
			AddDebug(string(debug.Stack()))
	} else if errKind, ok := stream.ReadUint64(); !ok {
		return nil, core.NewProtocolError(core.ErrStringBadStream)
	} else if errKind == uint64(core.ErrorKindNone) {
		if ret, ok := stream.Read(); ok {
			return ret, nil
		}
		return nil, core.NewProtocolError(core.ErrStringBadStream)
	} else if message, ok := stream.ReadString(); !ok {
		return nil, core.NewProtocolError(core.ErrStringBadStream)
	} else if dbg, ok := stream.ReadString(); !ok {
		return nil, core.NewProtocolError(core.ErrStringBadStream)
	} else if !stream.IsReadFinish() {
		return nil, core.NewProtocolError(core.ErrStringBadStream)
	} else {
		return nil, core.NewError(core.ErrorKind(errKind), message, dbg)
	}
}

func (p *Client) onError(err Error) {
	fmt.Println("client", err)
}

// End ***** Client ***** //

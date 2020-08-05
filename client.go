package rpc

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
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
			sendStream: internal.NewStream(),
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
		stream := internal.NewStream()
		stream.SetCallbackID(p.sendStream.GetCallbackID())
		stream.WriteUint64(uint64(internal.ErrorKindReply))
		stream.WriteString(internal.ErrStringTimeout)
		stream.WriteString("")
		p.returnCH <- stream
		return true
	} else {
		return false
	}
}

func (p *sendItem) Release() {
	close(p.returnCH)
	p.returnCH = nil
	p.sendStream.Reset()
	sendItemCache.Put(p)
}

// End ***** sendItem ***** //

// Begin ***** Client ***** //
type Client struct {
	sessionString        string
	conn                 internal.IStreamConn
	logWriter            LogWriter
	preSendHead          *sendItem
	preSendTail          *sendItem
	sendMap              map[uint64]*sendItem
	systemSeed           uint64
	readTimeout          time.Duration
	writeTimeout         time.Duration
	readLimit            int64
	currCallbackId       uint64
	maxCallbackId        uint64
	callbackSize         int64
	lastControlSendTime  time.Time
	lastTimeoutCheckTime time.Time
	internal.Lock
	internal.StatusManager
}

// Dial ...
func Dial(connectString string) (*Client, Error) {
	if u, e := url.Parse(connectString); e != nil {
		return nil, internal.NewRuntimePanic(e.Error())
	} else {
		switch u.Scheme {
		case "ws":
			return newClient(internal.NewWebSocketClientAdapter(connectString)), nil
		default:
			return nil,
				internal.NewRuntimePanic(fmt.Sprintf("unknown scheme %s", u.Scheme))
		}
	}
}

func newClient(endPoint internal.IAdapter) *Client {
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
		currCallbackId:       0,
		maxCallbackId:        0,
		callbackSize:         0,
		lastControlSendTime:  time.Now().Add(-10 * time.Second),
		lastTimeoutCheckTime: time.Now().Add(-10 * time.Second),
	}

	ret.SetRunning(nil)

	go func() {
		for ret.IsRunning() {
			endPoint.Open(ret.onConnRun, ret.onError)
		}
		ret.SetClosed(nil)
	}()

	go func() {
		for ret.IsRunning() {
			now := internal.TimeNow()
			ret.tryToTimeout(now)
			ret.tryToDeliverControlMessage(now)
			for ret.tryToDeliverPreSendMessage() {
				// loop until failed
			}
			time.Sleep(100 * time.Millisecond)
		}

		ret.SetClosed(func() {
			endPoint.Close(ret.onError)
		})
	}()

	return ret
}

func (p *Client) Close() bool {
	waitCH := chan bool(nil)

	if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
	}) {
		p.onError(internal.NewRuntimePanic(
			"it is not running",
		).AddDebug(string(debug.Stack())))
		return false
	} else {
		select {
		case <-waitCH:
			return true
		case <-time.After(20 * time.Second):
			p.onError(internal.NewRuntimePanic(
				"can not close within 20 seconds",
			).AddDebug(string(debug.Stack())))
			return false
		}
	}
}

func (p *Client) initConn(conn internal.IStreamConn) Error {
	// get the sequence
	sequence := p.CallWithLock(func() interface{} {
		p.systemSeed++
		ret := p.systemSeed
		return ret
	}).(uint64)

	sendStream := internal.NewStream()
	backStream := (*internal.Stream)(nil)

	defer func() {
		sendStream.Release()
		if backStream != nil {
			backStream.Release()
		}
	}()

	sendStream.SetCallbackID(0)
	sendStream.SetSequence(sequence)
	sendStream.WriteInt64(controlStreamKindInit)
	sendStream.WriteString(p.sessionString)

	if conn == nil {
		return internal.NewKernelPanic(
			"Client: initConn: conn is nil",
		).AddDebug(string(debug.Stack()))
	} else if err := conn.WriteStream(sendStream, 3*time.Second); err != nil {
		return err
	} else if backStream, err = conn.ReadStream(3*time.Second, 0); err != nil {
		return err
	} else if backStream.GetCallbackID() != 0 {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if backStream.GetSequence() != sequence {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if kind, ok := backStream.ReadInt64(); !ok ||
		kind != controlStreamKindInitBack {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if sessionString, ok := backStream.ReadString(); !ok ||
		len(sessionString) < 34 {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if readTimeoutMS, ok := backStream.ReadInt64(); !ok ||
		readTimeoutMS <= 0 {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if writeTimeoutMS, ok := backStream.ReadInt64(); !ok ||
		writeTimeoutMS <= 0 {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if readLimit, ok := backStream.ReadInt64(); !ok ||
		readLimit <= 0 {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if callBackSize, ok := backStream.ReadInt64(); !ok ||
		callBackSize <= 0 {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else if !backStream.IsReadFinish() {
		return internal.NewProtocolError(internal.ErrStringBadStream)
	} else {
		p.sessionString = sessionString
		p.readTimeout = time.Duration(readTimeoutMS) * time.Millisecond
		p.writeTimeout = time.Duration(writeTimeoutMS) * time.Millisecond
		p.readLimit = readLimit
		p.callbackSize = callBackSize
		return nil
	}
}

func (p *Client) onConnRun(conn internal.IStreamConn) {
	// init conn
	if err := p.initConn(conn); err != nil {
		p.onError(err)
		return
	}

	// set the conn
	p.DoWithLock(func() {
		p.conn = conn
	})

	err := Error(nil)

	// clear the conn when finish
	defer func() {
		if err != nil {
			p.onError(err)
		}
		p.DoWithLock(func() {
			p.conn = nil
		})
	}()

	// receive messages
	for p.IsRunning() {
		if stream, e := conn.ReadStream(p.readTimeout, p.readLimit); e != nil {
			if e != internal.ErrTransportStreamConnIsClosed {
				err = e
			}
			return
		} else if callbackID := stream.GetCallbackID(); callbackID > 0 {
			p.DoWithLock(func() {
				if v, ok := p.sendMap[callbackID]; ok {
					if v.Return(stream) {
						delete(p.sendMap, callbackID)
					}
				}
			})
		} else if kind, ok := stream.ReadInt64(); !ok ||
			kind != controlStreamKindRequestIdsBack {
			err = internal.NewProtocolError(internal.ErrStringBadStream)
			return
		} else if maxCallbackId, ok := stream.ReadUint64(); !ok {
			err = internal.NewProtocolError(internal.ErrStringBadStream)
			return
		} else if !stream.IsReadFinish() {
			err = internal.NewProtocolError(internal.ErrStringBadStream)
			return
		} else {
			p.DoWithLock(func() {
				if maxCallbackId > p.maxCallbackId {
					p.maxCallbackId = maxCallbackId
				}
			})
		}
	}
}

func (p *Client) getHeartbeatDuration() time.Duration {
	return time.Duration(float64(p.readTimeout/time.Millisecond) * 0.8)
}

func (p *Client) tryToDeliverControlMessage(now time.Time) {
	p.DoWithLock(func() {
		deltaTime := now.Sub(p.lastControlSendTime)
		if p.conn == nil {
			return
		} else if deltaTime < 500*time.Millisecond {
			return
		} else if deltaTime < p.getHeartbeatDuration() &&
			int64(len(p.sendMap)) > p.callbackSize/2 {
			return
		} else if deltaTime < p.getHeartbeatDuration() &&
			int64(p.maxCallbackId-p.currCallbackId) > p.callbackSize/2 {
			return
		} else {
			p.lastControlSendTime = now
			p.systemSeed++

			sendStream := internal.NewStream()
			sendStream.SetCallbackID(0)
			sendStream.SetSequence(p.systemSeed)
			sendStream.WriteInt64(controlStreamKindRequestIds)
			sendStream.WriteUint64(p.currCallbackId)

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
	})
}

func (p *Client) tryToTimeout(now time.Time) {
	p.DoWithLock(func() {
		if now.Sub(p.lastTimeoutCheckTime) < 600*time.Millisecond {
			return
		} else {
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
	})
}

func (p *Client) tryToDeliverPreSendMessage() bool {
	return p.CallWithLock(func() interface{} {
		if !p.IsRunning() { // not running
			return false
		} else if p.conn == nil { // not connected
			return false
		} else if p.preSendHead == nil { // preSend queue is empty
			return false
		} else if p.currCallbackId >= p.maxCallbackId { // id is not available
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
			p.currCallbackId++
			item.id = p.currCallbackId
			item.next = nil
			item.sendStream.SetCallbackID(item.id)
			item.sendStream.SetSequence(0)

			// set to sendMap
			p.sendMap[item.id] = item

			// try to send
			if err := p.conn.WriteStream(
				item.sendStream,
				p.writeTimeout,
			); err != nil {
				p.onError(err)
				return false
			} else {
				item.sendTime = internal.TimeNow()
				return true
			}
		}
	}).(bool)
}

// sendMessage ...
func (p *Client) sendMessage(
	timeout time.Duration,
	target string,
	args ...interface{},
) (interface{}, Error) {
	item := newSendItem()
	defer item.Release()

	item.startTime = internal.TimeNow()
	item.timeout = timeout

	// write target
	item.sendStream.WriteString(target)
	// write depth
	item.sendStream.WriteUint64(0)
	// write from
	item.sendStream.WriteString("@")
	// write args
	for i := 0; i < len(args); i++ {
		if item.sendStream.Write(args[i]) != internal.StreamWriteOK {
			return nil, internal.NewRuntimePanic(
				"Client: send: args not supported",
			)
		}
	}

	// add item to the list tail
	p.DoWithLock(func() {
		if p.preSendTail == nil {
			p.preSendHead = item
			p.preSendTail = item
		} else {
			p.preSendTail.next = item
			p.preSendTail = item
		}
	})

	// wait for response
	if stream := <-item.returnCH; stream == nil {
		return nil, internal.NewKernelPanic("stream is nil").
			AddDebug(string(debug.Stack()))
	} else if errKind, ok := stream.ReadUint64(); !ok {
		return nil, internal.NewProtocolError(internal.ErrStringBadStream)
	} else if errKind == uint64(internal.ErrorKindNone) {
		if ret, ok := stream.Read(); !ok {
			return nil, internal.NewProtocolError(internal.ErrStringBadStream)
		} else {
			return ret, nil
		}
	} else if message, ok := stream.ReadString(); !ok {
		return nil, internal.NewProtocolError(internal.ErrStringBadStream)
	} else if debug, ok := stream.ReadString(); !ok {
		return nil, internal.NewProtocolError(internal.ErrStringBadStream)
	} else if !stream.IsReadFinish() {
		return nil, internal.NewProtocolError(internal.ErrStringBadStream)
	} else {
		return nil, internal.NewError(internal.ErrorKind(errKind), message, debug)
	}
}

//func (p *sessionServer) onError(err Error) {
//  p.onSessionError(0, err)
//}
//
//func (p *sessionServer) onSessionError(sessionID uint64, err Error) {
//  fmt.Println("client", sessionID, err)
//}

func (p *Client) onError(err Error) {
	fmt.Println("client", err)
}

// End ***** Client ***** //

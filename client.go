package rpcc

import (
	"github.com/tslearn/rpcc/internal"
	"sync"
	"sync/atomic"
	"time"
)

// Begin ***** sendItem ***** //
const sendItemStatusRunning = int32(1)
const sendItemStatusFinish = int32(2)

type sendItem struct {
	id        uint64
	status    int32
	startMS   int64
	sendMS    int64
	timeoutMS int64
	finishCH  chan bool
	stream    Stream
	next      *sendItem
}

var sendItemCache = &sync.Pool{
	New: func() interface{} {
		return &sendItem{
			id:        0,
			status:    sendItemStatusRunning,
			startMS:   0,
			sendMS:    0,
			timeoutMS: 0,
			finishCH:  nil,
			stream:    internal.NewRPCStream(),
			next:      nil,
		}
	},
}

func newSendItem() *sendItem {
	ret := sendItemCache.Get().(*sendItem)
	ret.id = 0
	ret.status = sendItemStatusRunning
	ret.startMS = 0
	ret.sendMS = 0
	ret.timeoutMS = 0
	ret.finishCH = make(chan bool, 1)
	ret.next = nil
	return ret
}

func (p *sendItem) Return(stream Stream) bool {
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
		p.stream.Release()
		p.stream = stream
		p.finishCH <- true
		return true
	}
}

func (p *sendItem) Timeout() bool {
	if atomic.CompareAndSwapInt32(
		&p.status,
		sendItemStatusRunning,
		sendItemStatusFinish,
	) {
		p.finishCH <- false
		return true
	} else {
		return false
	}
}

func (p *sendItem) Release() {
	close(p.finishCH)
	p.finishCH = nil
	p.stream.Reset()
	sendItemCache.Put(p)
}

// End ***** sendItem ***** //

// Begin ***** Client ***** //
type Client struct {
	isOpen             bool
	closeCH            chan bool
	sessionString      string
	conn               IStreamConnection
	logger             *internal.Logger
	endPoint           IAdapter
	preSendHead        *sendItem
	preSendTail        *sendItem
	sendMap            map[uint64]*sendItem
	systemSeed         uint64
	readTimeout        time.Duration
	writeTimeout       time.Duration
	readLimit          int64
	writeLimit         int64
	currCallbackId     uint64
	maxCallbackId      uint64
	callbackSize       int64
	lastControlSendMS  int64
	lastTimeoutCheckMS int64
	internal.Lock
}

func NewClient(endPoint IAdapter) *Client {
	return &Client{
		isOpen:             false,
		closeCH:            nil,
		sessionString:      "",
		conn:               nil,
		logger:             internal.NewLogger(nil),
		endPoint:           endPoint,
		preSendHead:        nil,
		preSendTail:        nil,
		sendMap:            make(map[uint64]*sendItem),
		systemSeed:         0,
		readTimeout:        0,
		writeTimeout:       0,
		readLimit:          0,
		writeLimit:         0,
		currCallbackId:     0,
		maxCallbackId:      0,
		callbackSize:       0,
		lastControlSendMS:  0,
		lastTimeoutCheckMS: 0,
	}
}

func (p *Client) Open() Error {
	return internal.ConvertToError(p.CallWithLock(func() interface{} {
		if p.isOpen {
			return internal.NewError(
				"Client: Start: it has already been opened",
			)
		} else {
			closeCH := make(chan bool, 1)
			p.closeCH = closeCH
			p.isOpen = true

			go func() {
				defer func() {
					p.DoWithLock(func() {
						p.closeCH = nil
						p.isOpen = false
						closeCH <- true
					})
				}()

				// make sure the endPoint is not running
				if p.endPoint.IsRunning() {
					p.endPoint.Close(p.onError)
				}

				for p.IsRunning() {
					if !p.endPoint.IsRunning() {
						p.endPoint.Open(p.onConnRun, p.onError)
					}

					now := internal.TimeNowMS()
					p.tryToTimeout(now)
					p.tryToDeliverControlMessage(now)
					for p.tryToDeliverPreSendMessage() {
						// loop until failed
					}

					time.Sleep(100 * time.Millisecond)
				}

				// close the endPoint if it is running
				if p.endPoint.IsRunning() {
					p.endPoint.Close(p.onError)
				}
			}()

			return nil
		}
	}))
}

func (p *Client) Close() Error {
	err := Error(nil)

	closeCH := p.CallWithLock(func() interface{} {
		if p.closeCH == nil {
			err = internal.NewError(
				"Client: Stop: has not been opened",
			)
			return nil
		} else {
			ret := p.closeCH
			p.closeCH = nil
			return ret
		}
	})

	if closeCH == nil {
		return err
	} else {
		select {
		case <-closeCH.(chan bool):
			return nil
		case <-time.After(10 * time.Second):
			return internal.NewError(
				"Client: Stop: can not close in 10 seconds",
			)
		}
	}
}

func (p *Client) IsRunning() bool {
	return p.CallWithLock(func() interface{} {
		return p.closeCH != nil
	}).(bool)
}

func (p *Client) initConn(conn IStreamConnection) Error {
	// get the sequence
	sequence := p.CallWithLock(func() interface{} {
		p.systemSeed++
		ret := p.systemSeed
		return ret
	}).(uint64)

	sendStream := internal.NewRPCStream()
	backStream := (*internal.RPCStream)(nil)

	defer func() {
		sendStream.Release()
		if backStream != nil {
			backStream.Release()
		}
	}()

	sendStream.SetCallbackID(0)
	sendStream.SetSequence(sequence)
	sendStream.WriteInt64(SystemStreamKindInit)
	sendStream.WriteString(p.sessionString)

	if conn == nil {
		return internal.NewError(
			"Client: initConn: conn is nil",
		)
	} else if err := conn.WriteStream(
		sendStream,
		configWriteTimeout,
		configServerReadLimit,
	); err != nil {
		return err
	} else if backStream, err = conn.ReadStream(
		configReadTimeout,
		configServerWriteLimit,
	); err != nil {
		return err
	} else if backStream.GetCallbackID() != 0 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if backStream.GetSequence() != sequence {
		return internal.NewError(
			"Client: initConn: sequence omit",
		)
	} else if kind, ok := backStream.ReadInt64(); !ok ||
		kind != SystemStreamKindInitBack {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if sessionString, ok := backStream.ReadString(); !ok ||
		len(sessionString) < 34 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if readTimeoutMS, ok := backStream.ReadInt64(); !ok ||
		readTimeoutMS <= 0 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if writeTimeoutMS, ok := backStream.ReadInt64(); !ok ||
		writeTimeoutMS <= 0 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if readLimit, ok := backStream.ReadInt64(); !ok ||
		readLimit <= 0 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if writeLimit, ok := backStream.ReadInt64(); !ok ||
		writeLimit <= 0 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if callBackSize, ok := backStream.ReadInt64(); !ok ||
		callBackSize <= 0 {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else if !backStream.IsReadFinish() {
		return internal.NewError(
			"Client: initConn: stream format error",
		)
	} else {
		p.sessionString = sessionString
		p.readTimeout = time.Duration(readTimeoutMS) * time.Millisecond
		p.writeTimeout = time.Duration(writeTimeoutMS) * time.Millisecond
		p.readLimit = readLimit
		p.writeLimit = writeLimit
		p.callbackSize = callBackSize
		return nil
	}
}

func (p *Client) onConnRun(conn IStreamConnection) {
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
			err = e
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
			kind != SystemStreamKindRequestIdsBack {
			err = internal.NewError(
				"Client: initConn: stream format error",
			)
			return
		} else if maxCallbackId, ok := stream.ReadUint64(); !ok {
			err = internal.NewError(
				"Client: initConn: stream format error",
			)
			return
		} else if !stream.IsReadFinish() {
			err = internal.NewError(
				"Client: initConn: stream format error",
			)
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

func (p *Client) getHeartbeatMS() int64 {
	return int64(float64(p.readTimeout/time.Millisecond) * 0.8)
}

func (p *Client) tryToDeliverControlMessage(nowMS int64) {
	p.DoWithLock(func() {
		delta := nowMS - p.lastControlSendMS
		if p.conn == nil {
			return
		} else if delta < 500 {
			return
		} else if delta < p.getHeartbeatMS() &&
			int64(len(p.sendMap)) > p.callbackSize/2 {
			return
		} else if delta < p.getHeartbeatMS() &&
			int64(p.maxCallbackId-p.currCallbackId) > p.callbackSize/2 {
			return
		} else {
			p.lastControlSendMS = nowMS
			p.systemSeed++

			sendStream := internal.NewRPCStream()
			sendStream.SetCallbackID(0)
			sendStream.SetSequence(p.systemSeed)
			sendStream.WriteInt64(SystemStreamKindRequestIds)
			sendStream.WriteUint64(p.currCallbackId)

			for key := range p.sendMap {
				sendStream.WriteUint64(key)
			}

			if err := p.conn.WriteStream(
				sendStream,
				p.writeTimeout,
				p.writeLimit,
			); err != nil {
				p.onError(err)
			}

			sendStream.Release()
		}
	})
}

func (p *Client) tryToTimeout(nowMS int64) {
	p.DoWithLock(func() {
		if nowMS-p.lastTimeoutCheckMS < 600 {
			return
		} else {
			p.lastTimeoutCheckMS = nowMS

			// sweep pre send list
			preValidItem := (*sendItem)(nil)
			item := p.preSendHead
			for item != nil {
				if nowMS-item.startMS > item.timeoutMS && item.Timeout() {
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
				if nowMS-value.startMS > value.timeoutMS && value.Timeout() {
					delete(p.sendMap, key)
				}
			}
		}
	})
}

func (p *Client) tryToDeliverPreSendMessage() bool {
	return p.CallWithLock(func() interface{} {
		if p.closeCH == nil { // close
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
			item.stream.SetCallbackID(item.id)
			item.stream.SetSequence(0)

			// set to sendMap
			p.sendMap[item.id] = item

			// try to send
			if err := p.conn.WriteStream(
				item.stream,
				p.writeTimeout,
				p.writeLimit,
			); err != nil {
				p.onError(err)
				return false
			} else {
				item.sendMS = internal.TimeNowMS()
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

	item.startMS = internal.TimeNowMS()
	item.timeoutMS = int64(timeout / time.Millisecond)

	// write target
	item.stream.WriteString(target)
	// write depth
	item.stream.WriteUint64(0)
	// write from
	item.stream.WriteString("@")
	// write args
	for i := 0; i < len(args); i++ {
		if item.stream.Write(args[i]) != internal.RPCStreamWriteOK {
			return nil, internal.NewError(
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
	if ok := <-item.finishCH; !ok {
		return nil, internal.NewError("timeout")
	} else if success, ok := item.stream.ReadBool(); !ok {
		return nil, internal.NewError("data format error")
	} else if success {
		if ret, ok := item.stream.Read(); !ok {
			return nil, internal.NewError("data format error")
		} else {
			return ret, nil
		}
	} else {
		if message, ok := item.stream.ReadString(); !ok {
			return nil, internal.NewError("data format error")
		} else if debug, ok := item.stream.ReadString(); !ok {
			return nil, internal.NewError("data format error")
		} else {
			return nil, internal.NewErrorByDebug(message, debug)
		}
	}
}

func (p *Client) onError(err Error) {
	p.logger.Error(err.Error())
}

// End ***** Client ***** //

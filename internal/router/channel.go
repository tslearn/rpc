package router

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	channelActionInit      = 1
	channelActionContinue  = 2
	channelActionReset     = 3
	channelActionDataBlock = 4
)

type ConnectMeta struct {
	addr      string
	tlsConfig *tls.Config
	id        *base.GlobalID
}

func connReadBytes(conn net.Conn, timeout time.Duration, b []byte) (int, *base.Error) {
	pos := 0
	length := 2

	if len(b) < 2 {
		return -1, base.ErrRouterConnProtocol
	}

	if e := conn.SetReadDeadline(base.TimeNow().Add(timeout)); e != nil {
		return -1, base.ErrRouterConnRead.AddDebug(e.Error())
	}

	for pos < length {
		if n, e := conn.Read(b[pos:]); e != nil {
			return -1, base.ErrRouterConnRead.AddDebug(e.Error())
		} else {
			pos += n

			if length == 2 && pos >= 2 {
				length = int(binary.LittleEndian.Uint16(b))
				if length > len(b) {
					return -1, base.ErrRouterConnProtocol
				}
			}
		}
	}

	return length, nil
}

func connWriteBytes(conn net.Conn, timeout time.Duration, b []byte) *base.Error {
	pos := 0

	if len(b) < 2 || int(binary.LittleEndian.Uint16(b)) != len(b) {
		return base.ErrRouterConnProtocol
	}

	if e := conn.SetWriteDeadline(base.TimeNow().Add(timeout)); e != nil {
		return base.ErrRouterConnWrite.AddDebug(e.Error())
	}

	for pos < len(b) {
		if n, e := conn.Write(b[pos:]); e != nil {
			return base.ErrRouterConnWrite.AddDebug(e.Error())
		} else {
			pos += n
		}
	}

	return nil
}

type Channel struct {
	needReset              uint8
	conn                   net.Conn
	streamCH               chan *rpc.Stream
	streamHub              rpc.IStreamHub
	sendPrepareSequence    uint64
	sendSequence           uint64
	sendConfirmSequence    uint64
	sendBuffers            [numOfCacheBuffer][bufferSize]byte
	receiveSequence        uint64
	receiveBuffer          [bufferSize]byte
	receiveStreamGenerator *rpc.StreamGenerator
	closeCH                chan bool
	orcManager             *base.ORCManager
	sync.Mutex
}

func NewChannel(
	index uint16,
	connMeta *ConnectMeta,
	streamCH chan *rpc.Stream,
	streamHub rpc.IStreamHub,
) *Channel {
	ret := &Channel{
		needReset:              0,
		conn:                   nil,
		streamCH:               streamCH,
		streamHub:              streamHub,
		sendPrepareSequence:    0,
		sendSequence:           0,
		sendConfirmSequence:    0,
		receiveSequence:        0,
		receiveStreamGenerator: rpc.NewStreamGenerator(streamHub),
		closeCH:                make(chan bool),
		orcManager:             base.NewORCManager(),
	}

	ret.orcManager.Open(func() bool {
		return true
	})

	if connMeta != nil {
		go func() {
			ret.orcManager.Run(func(isRunning func() bool) bool {
				for isRunning() {
					startNS := base.TimeNow().UnixNano()
					if err := ret.runMasterThread(index, connMeta); err != nil {
						streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(err))
					}
					base.WaitAtLeastDurationWhenRunning(
						startNS,
						isRunning,
						time.Second,
					)
				}

				return true
			})
		}()
	}

	return ret
}

func (p *Channel) runMasterThread(index uint16, connMeta *ConnectMeta) *base.Error {
	var conn net.Conn
	var e error
	buffer := make([]byte, 40)

	// dail
	if connMeta.tlsConfig == nil {
		conn, e = net.Dial("tcp", connMeta.addr)
	} else {
		conn, e = tls.Dial("tcp", connMeta.addr, connMeta.tlsConfig)
	}

	// deal dail error
	if e != nil {
		return base.ErrRouterConnDial.AddDebug(e.Error())
	}

	// make init frame
	binary.LittleEndian.PutUint16(buffer, 40)
	binary.LittleEndian.PutUint16(buffer[2:], channelActionInit)
	binary.LittleEndian.PutUint16(buffer[4:], index)
	binary.LittleEndian.PutUint64(buffer[6:], connMeta.id.GetID())
	buffer[14] = p.needReset
	binary.LittleEndian.PutUint64(buffer[16:], p.sendConfirmSequence)
	binary.LittleEndian.PutUint64(buffer[24:], p.sendSequence)
	binary.LittleEndian.PutUint64(buffer[32:], p.receiveSequence)

	// send init frame
	if err := connWriteBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	// read response frame
	if _, err := connReadBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	switch binary.LittleEndian.Uint16(buffer[2:]) {
	case channelActionContinue:

	case channelActionReset:
		p.needReset = 0
		p.sendPrepareSequence = 0
		p.sendSequence = 0
		p.sendConfirmSequence = 0
		p.receiveSequence = 0
	default:
		p.needReset = 1
		return base.ErrRouterConnProtocol
	}

	p.RunWithConn(conn)

	return nil
}

func (p *Channel) runSlaveThread(conn net.Conn, buffer [40]byte) *base.Error {
	rNeedReset := buffer[14]
	rSendConfirmSequence := binary.LittleEndian.Uint64(buffer[16:])
	rSendSequence := binary.LittleEndian.Uint64(buffer[24:])
	rReceiveSequence := binary.LittleEndian.Uint64(buffer[32:])

	checkConfirmSequenceAndReceiveSequence :=
		rSendConfirmSequence <= p.receiveSequence &&
			rSendConfirmSequence+numOfCacheBuffer >= p.receiveSequence &&
			p.sendConfirmSequence <= rReceiveSequence &&
			p.sendConfirmSequence+numOfCacheBuffer >= rReceiveSequence

	checkSendSequenceAndReceiveSequence :=
		rSendSequence >= p.receiveSequence &&
			rSendSequence <= p.receiveSequence+numOfCacheBuffer &&
			p.sendSequence >= rReceiveSequence &&
			p.sendSequence <= rReceiveSequence+numOfCacheBuffer

	if p.needReset == 0 && rNeedReset == 0 &&
		checkConfirmSequenceAndReceiveSequence &&
		checkSendSequenceAndReceiveSequence {
		p.sendSequence = rReceiveSequence
		binary.LittleEndian.PutUint16(buffer[2:], channelActionContinue)
	} else {
		p.needReset = 0
		p.sendPrepareSequence = 0
		p.sendSequence = 0
		p.sendConfirmSequence = 0
		p.receiveSequence = 0
		binary.LittleEndian.PutUint16(buffer[2:], channelActionReset)
	}

	// send response frame
	if err := connWriteBytes(conn, time.Second, buffer[:]); err != nil {
		return err
	}

	// run
	p.RunWithConn(conn)

	return nil
}

func (p *Channel) RunWithConn(conn net.Conn) {
	running := uint32(1)

	isRunning := func() bool {
		isChannelRunning := p.orcManager.GetRunningFn()
		return isChannelRunning() && atomic.LoadUint32(&running) == 1
	}

	p.conn = conn
	waitCH := make(chan bool)

	go func() {
		p.runMakeFrame(isRunning)
		atomic.StoreUint32(&running, 0)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		_ = p.runRead(conn)
		atomic.StoreUint32(&running, 0)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		_ = p.runWrite(conn, isRunning)
		atomic.StoreUint32(&running, 0)
		_ = conn.Close()
		waitCH <- true
	}()

	<-waitCH
	<-waitCH
	<-waitCH
	p.conn = nil
}

//func (p *Channel) updateReceiveSequence(sequence uint64) *base.Error {
//    p.Lock()
//    defer p.Unlock()
//
//    if p.receiveSequence+1 == sequence {
//        p.receiveSequence = sequence
//        return nil
//    } else {
//        return base.ErrRouterConnProtocol
//    }
//}

//func (p *Channel) lockSendPrepareSequence() uint64 {
//    p.Lock()
//    defer p.Unlock()
//
//    if p.sendPrepareSequence-p.sendConfirmSequence < numOfCacheBuffer {
//        return p.sendPrepareSequence
//    }
//
//    return 0
//}
//
//func (p *Channel) unlockSendPrepareSequence() {
//    p.Lock()
//    defer p.Unlock()
//
//    p.sendPrepareSequence += 1
//}

//func (p *Channel) updateSendSuccessSequence(sequence uint64) *base.Error {
//    p.Lock()
//    defer p.Unlock()
//
//    for i := p.sendSuccessSequence + 1; i < sequence; i++ {
//        if binary.LittleEndian.Uint64(p.sendBuffers[i%numOfCacheBuffer][:]) != i {
//            p.sendCurrentSequence = 0
//            p.sendSuccessSequence = 0
//            p.receiveSequence = 0
//            return base.ErrRouterConnProtocol
//        }
//    }
//
//    p.sendSuccessSequence = sequence
//
//    return nil
//}

func (p *Channel) canPrepare() bool {
	sendPrepare := atomic.LoadUint64(&p.sendPrepareSequence)
	sendConfirm := atomic.LoadUint64(&p.sendConfirmSequence)
	return sendPrepare-sendConfirm < numOfCacheBuffer
}

func (p *Channel) runMakeFrame(isRunning func() bool) {
	stream := (*rpc.Stream)(nil)
	streamPos := 0

	for isRunning() {
		// get frame id
		for isRunning() && !p.canPrepare() {
			time.Sleep(30 * time.Millisecond)
		}

		if !isRunning() {
			return
		}

		frameID := atomic.AddUint64(&p.sendPrepareSequence, 1)

		// init data frame
		frameBuffer := p.sendBuffers[frameID%numOfCacheBuffer][:]
		binary.LittleEndian.PutUint16(frameBuffer[2:], channelActionDataBlock)
		binary.LittleEndian.PutUint64(frameBuffer[4:], frameID)

		// gat first stream
		if stream == nil {
			stream = <-p.streamCH
		}

		// write stream to buffer
		bufferPos := 12
		for stream != nil && bufferSize-bufferPos >= 512 {
			peekBuf, finish := stream.PeekBufferSlice(
				streamPos, bufferSize-bufferPos,
			)
			copyLen := copy(frameBuffer[bufferPos:], peekBuf)
			streamPos += copyLen
			bufferPos += copyLen

			if finish {
				streamPos = 0
				select {
				case stream = <-p.streamCH:
				default:
					stream = nil
				}
			}
		}

		binary.LittleEndian.PutUint16(frameBuffer, 12)
	}
}

func (p *Channel) runRead(conn net.Conn) *base.Error {
	for {
		if n, err := connReadBytes(
			conn, 3*time.Second, p.receiveBuffer[:],
		); err != nil {
			return err
		} else if n < 12 {
			return base.ErrRouterConnProtocol
		} else if binary.LittleEndian.Uint16(p.receiveBuffer[2:]) != channelActionDataBlock {
			return base.ErrRouterConnProtocol
		} else if err := p.updateReceiveSequence(binary.LittleEndian.Uint64(p.receiveBuffer[4:])); err != nil {
			return err
		} else {
			if err := p.receiveStreamGenerator.OnBytes(p.receiveBuffer[12:]); err != nil {
				return err
			}
		}
	}
}

func (p *Channel) runWrite(conn net.Conn, isRunning func() bool) *base.Error {
	stream := (*rpc.Stream)(nil)
	streamPos := 0

	for {

		stream := <-p.streamCH
		writePos := 0

		for {
			peekBuf, finish := stream.PeekBufferSlice(writePos, 1024)

			peekLen := len(peekBuf)

			if peekLen <= 0 {
				return errors.New("error stream")
			}

			start := 0
			for start < peekLen {
				if n, e := conn.Write(peekBuf[start:]); e != nil {
					return e
				} else {
					start += n
				}
			}

			writePos += peekLen

			if finish {
				break
			}
		}
	}
}

func (p *Channel) Close() {
	p.orcManager.Close(func() bool {
		return true
	}, func() {

	})
}

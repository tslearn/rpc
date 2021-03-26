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
	channelActionInitRequest       = 1
	channelActionInitResponseOK    = 2
	channelActionInitResponseError = 3
	channelDataBlock               = 4

	runningStatusNone        = 0
	runningStatusRunning     = 1
	runningStatusWaitForExit = 2
)

type ConnectMeta struct {
	addr      string
	tlsConfig *tls.Config
	id        *base.GlobalID
}

type Channel struct {
	isMaster               bool
	conn                   net.Conn
	streamCH               chan *rpc.Stream
	streamHub              rpc.IStreamHub
	sendCurrentSequence    uint64
	sendSuccessSequence    uint64
	sendBuffers            [numOfCacheBuffer][bufferSize]byte
	receiveSequence        uint64
	receiveBuffer          [bufferSize]byte
	receiveStreamGenerator *rpc.StreamGenerator
	closeCH                chan bool
	orcManager             *base.ORCManager
	makeFrameRunningStatus uint32
	sync.Mutex
}

func connReadBytes(conn net.Conn, timeout time.Duration, b []byte) (int, *base.Error) {
	pos := 0
	length := 2

	if e := conn.SetReadDeadline(base.TimeNow().Add(timeout)); e != nil {
		return -1, base.ErrRouterConnRead.AddDebug(e.Error())
	}

	for pos < length {
		if n, e := conn.Read(b); e != nil {
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

func NewChannel(
	index uint16,
	connectMeta *ConnectMeta,
	streamCH chan *rpc.Stream,
	streamHub rpc.IStreamHub,
) *Channel {
	ret := &Channel{
		isMaster:               connectMeta != nil,
		conn:                   nil,
		streamCH:               streamCH,
		streamHub:              streamHub,
		sendCurrentSequence:    0,
		sendSuccessSequence:    0,
		receiveSequence:        0,
		receiveStreamGenerator: rpc.NewStreamGenerator(streamHub),
		closeCH:                make(chan bool),
		orcManager:             base.NewORCManager(),
		makeFrameRunningStatus: runningStatusNone,
	}

	ret.orcManager.Open(func() bool {
		return true
	})

	if connectMeta != nil {
		go func() {
			ret.orcManager.Run(func(isRunning func() bool) bool {
				for isRunning() {
					var conn net.Conn
					var e error

					if connectMeta.tlsConfig == nil {
						conn, e = net.Dial("tcp", connectMeta.addr)
					} else {
						conn, e = tls.Dial("tcp", connectMeta.addr, connectMeta.tlsConfig)
					}

					if e != nil {
						streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
							base.ErrRouterConnDial.AddDebug(e.Error()),
						))
						continue
					}

					if err := ret.initMasterConn(
						conn, index, connectMeta.id.GetID(),
					); err != nil {
						streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(err))
						_ = conn.Close()
						continue
					} else {
						ret.RunWithConn(conn)
					}
				}
				return true
			})
		}()
	}

	return ret
}

func (p *Channel) initSlaveConn(
	conn net.Conn,
	remoteSendSequence uint64,
	remoteReceiveSequence uint64,
) *base.Error {
	needToSync := false
	buffer := make([]byte, 32)
	binary.LittleEndian.PutUint16(buffer, 32)
	binary.LittleEndian.PutUint64(buffer[16:], p.sendCurrentSequence)
	binary.LittleEndian.PutUint64(buffer[24:], p.receiveSequence)

	if p.receiveSequence > remoteSendSequence ||
		p.receiveSequence <= remoteSendSequence-numOfCacheBuffer ||
		remoteReceiveSequence > p.sendCurrentSequence ||
		remoteReceiveSequence <= p.sendCurrentSequence-numOfCacheBuffer {
		p.sendCurrentSequence = 0
		p.receiveSequence = 0
		binary.LittleEndian.PutUint16(buffer[2:], channelActionInitResponseError)
	} else {
		binary.LittleEndian.PutUint16(buffer[2:], channelActionInitResponseOK)
		needToSync = true
	}

	// send buffer
	if err := connWriteBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	// send unreceived buffer to remote
	if needToSync {
		for i := remoteReceiveSequence + 1; i <= p.sendCurrentSequence; i++ {
			if err := connWriteBytes(
				conn,
				time.Second,
				p.sendBuffers[i%numOfCacheBuffer][:],
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Channel) initMasterConn(
	conn net.Conn,
	index uint16,
	slotID uint64,
) *base.Error {
	buffer := make([]byte, 32)

	// send channelActionInitRequest
	binary.LittleEndian.PutUint16(buffer, 32)
	binary.LittleEndian.PutUint16(buffer[2:], channelActionInitRequest)
	binary.LittleEndian.PutUint16(buffer[4:], index)
	binary.LittleEndian.PutUint64(buffer[8:], slotID)
	binary.LittleEndian.PutUint64(buffer[16:], p.sendCurrentSequence)
	binary.LittleEndian.PutUint64(buffer[24:], p.receiveSequence)
	if err := connWriteBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	// receive channelActionInitResponse
	if _, err := connReadBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	if binary.LittleEndian.Uint16(buffer[2:]) != channelActionInitResponseOK {
		p.sendCurrentSequence = 0
		p.sendSuccessSequence = 0
		p.receiveSequence = 0
		return base.ErrRouterConnProtocol
	}

	remoteSendSequence := binary.LittleEndian.Uint64(buffer[16:])
	remoteReceiveSequence := binary.LittleEndian.Uint64(buffer[24:])
	if p.receiveSequence > remoteSendSequence ||
		p.receiveSequence <= remoteSendSequence-numOfCacheBuffer ||
		remoteReceiveSequence > p.sendCurrentSequence ||
		remoteReceiveSequence <= p.sendCurrentSequence-numOfCacheBuffer {
		p.sendCurrentSequence = 0
		p.sendSuccessSequence = 0
		p.receiveSequence = 0
		return base.ErrRouterConnProtocol
	}

	// send unreceived buffer to remote
	for i := remoteReceiveSequence + 1; i <= p.sendCurrentSequence; i++ {
		if err := connWriteBytes(
			conn,
			time.Second,
			p.sendBuffers[i%numOfCacheBuffer][:],
		); err != nil {
			return err
		}
	}

	return nil
}

func (p *Channel) setConn(conn net.Conn) {
	p.Lock()
	defer p.Unlock()

	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn = conn
}

func (p *Channel) getConn() net.Conn {
	p.Lock()
	defer p.Unlock()

	return p.conn
}

func (p *Channel) RunWithConn(conn net.Conn) bool {
	running := uint32(1)
	isRunning := func() bool {
		return atomic.LoadUint32(&running) == 1
	}

	p.setConn(conn)
	waitCH := make(chan bool)

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
	p.setConn(nil)
	return true
}

func (p *Channel) updateReceiveSequence(sequence uint64) *base.Error {
	p.Lock()
	defer p.Unlock()

	if p.receiveSequence+1 == sequence {
		p.receiveSequence = sequence
		return nil
	} else {
		return base.ErrRouterConnProtocol
	}
}

func (p *Channel) requireSendCurrentSequence() uint64 {
	p.Lock()
	defer p.Unlock()

	if p.sendCurrentSequence-p.sendSuccessSequence < numOfCacheBuffer {
		p.sendCurrentSequence += 1
		return p.sendCurrentSequence
	}

	return 0
}

func (p *Channel) updateSendSuccessSequence(sequence uint64) *base.Error {
	p.Lock()
	defer p.Unlock()

	for i := p.sendSuccessSequence + 1; i < sequence; i++ {
		if binary.LittleEndian.Uint64(p.sendBuffers[i%numOfCacheBuffer][:]) != i {
			p.sendCurrentSequence = 0
			p.sendSuccessSequence = 0
			p.receiveSequence = 0
			return base.ErrRouterConnProtocol
		}
	}

	p.sendSuccessSequence = sequence

	return nil
}

func (p *Channel) runRead(conn net.Conn) *base.Error {
	for {
		if n, err := connReadBytes(conn, 3*time.Second, p.receiveBuffer[:]); err != nil {
			return err
		} else if n < 12 {
			return base.ErrRouterConnProtocol
		} else if binary.LittleEndian.Uint16(p.receiveBuffer[2:]) != channelDataBlock {
			return base.ErrRouterConnProtocol
		} else if err := p.updateReceiveSequence(
			binary.LittleEndian.Uint64(p.receiveBuffer[4:]),
		); err != nil {
			return err
		} else {
			if err := p.receiveStreamGenerator.OnBytes(p.receiveBuffer[12:]); err != nil {
				return err
			}
		}
	}
}

func (p *Channel) waitMakeFrameFinish() {
	for atomic.LoadUint32(&p.makeFrameRunningStatus) != runningStatusNone {
		time.Sleep(10 * time.Millisecond)
	}
}

func (p *Channel) runMakeFrame() *base.Error {
	isRunning := func() bool {
		return atomic.LoadUint32(&p.makeFrameRunningStatus) == runningStatusRunning
	}

	if atomic.CompareAndSwapUint32(
		&p.makeFrameRunningStatus,
		runningStatusNone,
		runningStatusRunning,
	) {
		defer func() {
			atomic.StoreUint32(&p.makeFrameRunningStatus, runningStatusNone)
		}()

		stream := (*rpc.Stream)(nil)
		streamPos := 0

		for isRunning() {
			startTime := base.TimeNow()
			//deadLine := startTime.Add(2 * time.Second)

			// get frame id
			frameID := p.requireSendCurrentSequence()
			for isRunning() && frameID == 0 {
				time.Sleep(30 * time.Millisecond)
				frameID = p.requireSendCurrentSequence()
			}

			// init data frame
			frameBuffer := p.sendBuffers[frameID%numOfCacheBuffer][:]
			binary.LittleEndian.PutUint16(frameBuffer, 20)
			binary.LittleEndian.PutUint16(frameBuffer[2:], channelDataBlock)
			binary.LittleEndian.PutUint64(frameBuffer[4:], frameID)

			// gat first stream
			if stream == nil {
				for {
					select {
					case stream = <-p.streamCH:
					case <-time.After(100 * time.Millisecond):

					}
				}
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
		if conn := p.getConn(); conn != nil {
			_ = conn.Close()
		}
		return true
	}, func() {

	})
}

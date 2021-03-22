package router

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"sync"
	"time"
)

const (
	channelActionInitRequest       = 1
	channelActionInitResponseOK    = 2
	channelActionInitResponseError = 4
	channelActionInitOK            = 3
	channelDataBlock               = 5
)

type ConnectMeta struct {
	addr      string
	tlsConfig *tls.Config
	id        *base.GlobalID
}

type Channel struct {
	isMaster        bool
	conn            net.Conn
	streamCH        chan *rpc.Stream
	streamHub       rpc.IStreamHub
	sendSequence    uint64
	sendBuffers     [numOfCacheBuffer][bufferSize]byte
	sendStream      *rpc.Stream
	sendStreamPos   int
	receiveSequence uint64
	receiveBuffer   [bufferSize]byte
	closeCH         chan bool
	orcManager      *base.ORCManager
	sync.Mutex
}

func connReadBytes(conn net.Conn, b []byte) *base.Error {
	pos := 0
	for pos < len(b) {
		if e := conn.SetReadDeadline(base.TimeNow().Add(time.Second)); e != nil {
			return base.ErrRouterConnRead.AddDebug(e.Error())
		} else if n, e := conn.Read(b); e != nil {
			return base.ErrRouterConnRead.AddDebug(e.Error())
		} else {
			pos += n
		}
	}
	return nil
}

func connWriteBytes(conn net.Conn, b []byte) *base.Error {
	pos := 0
	for pos < len(b) {
		if e := conn.SetWriteDeadline(base.TimeNow().Add(time.Second)); e != nil {
			return base.ErrRouterConnWrite.AddDebug(e.Error())
		} else if n, e := conn.Write(b[pos:]); e != nil {
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
		isMaster:        connectMeta != nil,
		conn:            nil,
		streamCH:        streamCH,
		streamHub:       streamHub,
		sendSequence:    0,
		sendStream:      nil,
		sendStreamPos:   0,
		receiveSequence: 0,
		closeCH:         make(chan bool),
		orcManager:      base.NewORCManager(),
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
	binary.LittleEndian.PutUint64(buffer[16:], p.sendSequence)
	binary.LittleEndian.PutUint64(buffer[24:], p.receiveSequence)

	if p.receiveSequence > remoteSendSequence ||
		p.receiveSequence <= remoteSendSequence-numOfCacheBuffer ||
		remoteReceiveSequence > p.sendSequence ||
		remoteReceiveSequence <= p.sendSequence-numOfCacheBuffer {
		p.sendSequence = 0
		p.receiveSequence = 0
		binary.LittleEndian.PutUint16(buffer, channelActionInitResponseError)
	} else {
		binary.LittleEndian.PutUint16(buffer, channelActionInitResponseOK)
		needToSync = true
	}

	// send buffer
	if err := connWriteBytes(conn, buffer); err != nil {
		return err
	}

	// send unreceived buffer to remote
	if needToSync {
		for i := remoteReceiveSequence + 1; i <= p.sendSequence; i++ {
			if err := connWriteBytes(
				conn,
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
	binary.LittleEndian.PutUint16(buffer, channelActionInitRequest)
	binary.LittleEndian.PutUint16(buffer[2:], index)
	binary.LittleEndian.PutUint64(buffer[8:], slotID)
	binary.LittleEndian.PutUint64(buffer[16:], p.sendSequence)
	binary.LittleEndian.PutUint64(buffer[24:], p.receiveSequence)
	if err := connWriteBytes(conn, buffer); err != nil {
		return err
	}

	// receive channelActionInitResponse
	if err := connReadBytes(conn, buffer); err != nil {
		return err
	}

	if binary.LittleEndian.Uint16(buffer) != channelActionInitResponseOK {
		p.sendSequence = 0
		p.receiveSequence = 0
		return base.ErrRouterConnProtocol
	}

	remoteSendSequence := binary.LittleEndian.Uint64(buffer[16:])
	remoteReceiveSequence := binary.LittleEndian.Uint64(buffer[24:])
	if p.receiveSequence > remoteSendSequence ||
		p.receiveSequence <= remoteSendSequence-numOfCacheBuffer ||
		remoteReceiveSequence > p.sendSequence ||
		remoteReceiveSequence <= p.sendSequence-numOfCacheBuffer {
		p.sendSequence = 0
		p.receiveSequence = 0
		return base.ErrRouterConnProtocol
	}

	// send ok to remote
	binary.LittleEndian.PutUint16(buffer, channelActionInitOK)
	if err := connWriteBytes(conn, buffer); err != nil {
		return err
	}

	// send unreceived buffer to remote
	for i := remoteReceiveSequence + 1; i <= p.sendSequence; i++ {
		if err := connWriteBytes(
			conn,
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
	p.setConn(conn)
	waitCH := make(chan bool)

	go func() {
		_ = p.runRead(conn)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		_ = p.runWrite(conn)
		_ = conn.Close()
		waitCH <- true
	}()

	<-waitCH
	<-waitCH
	p.setConn(nil)
	return true
}

func (p *Channel) runRead(conn net.Conn) error {
	buffer := make([]byte, 8192)
	bufferStart := 0
	bufferPos := 0

	stream := (*rpc.Stream)(nil)

	for {
		if stream == nil {
			// read the head
			for bufferPos-bufferStart < rpc.StreamHeadSize {
				if n, e := conn.Read(buffer[bufferPos:]); e != nil {
					return e
				} else {
					bufferPos += n
				}
			}

			// create the stream
			stream = rpc.NewStream()

			// put header to stream
			stream.PutBytesTo(
				buffer[bufferStart:bufferStart+rpc.StreamHeadSize],
				0,
			)
			bufferStart += rpc.StreamHeadSize
		} else if remainBytes := bufferPos - bufferStart; remainBytes > 0 {
			streamLength := int(stream.GetLength())
			readMaxBytes := base.MinInt(
				remainBytes,
				streamLength-stream.GetWritePos(),
			)

			if readMaxBytes >= 0 {
				stream.PutBytes(buffer[bufferStart : bufferStart+readMaxBytes])
				if stream.GetWritePos() == streamLength {
					if stream.CheckStream() {
						p.streamHub.OnReceiveStream(stream)
					} else {
						return errors.New("stream error")
					}
				}
			} else {
				return errors.New("stream error")
			}
		}

		// move the buffer
		if len(buffer)-bufferStart <= rpc.StreamHeadSize {
			copy(buffer[0:], buffer[bufferStart:])
			bufferPos -= bufferStart
			bufferStart = 0
		}
	}
}

func (p *Channel) runWrite(conn net.Conn) error {
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

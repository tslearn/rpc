package tcp

import (
	"encoding/binary"
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"time"
)

const bufferSize = 1024

type tcpStreamConn struct {
	readBuf  []byte
	writeBuf []byte
	readPos  int
	conn     net.Conn
	u32Buf   [4]byte
}

func NewTCPStreamConn(
	conn net.Conn,
) *tcpStreamConn {
	if conn == nil {
		return nil
	}

	ret := &tcpStreamConn{
		readBuf:  make([]byte, bufferSize),
		writeBuf: make([]byte, bufferSize),
		conn:     conn,
	}

	return ret
}

func (p *tcpStreamConn) ReadStream(
	timeout time.Duration,
	readLimit int64,
) (*core.Stream, *base.Error) {
	if e := p.conn.SetReadDeadline(
		base.TimeNow().Add(timeout),
	); e != nil {
		return nil, errors.ErrTCPStreamConnReadStream.AddDebug("SetReadDeadline error")
	}

	for p.readPos < 4 {
		n, e := p.conn.Read(p.readBuf[p.readPos:])
		if e != nil {
			return nil, errors.ErrTCPStreamConnReadStream.AddDebug(e.Error())
		}
		p.readPos += n
	}

	streamLen := int(binary.LittleEndian.Uint32(p.readBuf))

	if streamLen >= int(readLimit) {
		return nil, errors.ErrTCPStreamConnReadStream.AddDebug("readLimit exceeds")
	}

	if streamLen < core.StreamHeadSize {
		return nil, errors.ErrTCPStreamConnReadStream.AddDebug("stream error")
	}

	for p.readPos < core.StreamHeadSize+4 {
		n, e := p.conn.Read(p.readBuf[p.readPos:])
		if e != nil {
			return nil, errors.ErrTCPStreamConnReadStream.AddDebug(e.Error())
		}
		p.readPos += n
	}

	ret := core.NewStream()

	if p.readPos >= streamLen+4 {
		ret.PutBytesTo(p.readBuf[4:streamLen+4], 0)
		p.readPos = copy(p.readBuf, p.readBuf[streamLen+4:p.readPos])
	} else {
		ret.PutBytesTo(p.readBuf[4:p.readPos], 0)
		streamLen -= p.readPos
		p.readPos = 0

		for streamLen > 0 {
			n, e := p.conn.Read(p.readBuf[p.readPos:])
			if e != nil {
				ret.Release()
				return nil, errors.ErrTCPStreamConnReadStream.AddDebug(e.Error())
			}
			p.readPos = n

			if streamLen >= p.readPos {
				ret.PutBytes(p.readBuf[:p.readPos])
				streamLen -= p.readPos
				p.readPos = 0
			} else {
				ret.PutBytes(p.readBuf[:streamLen])
				streamLen = 0
				p.readPos = copy(p.readBuf, p.readBuf[streamLen:p.readPos])
			}
		}
	}

	return ret, nil
}

func (p *tcpStreamConn) WriteStream(
	stream *core.Stream,
	timeout time.Duration,
) *base.Error {
	if e := p.conn.SetWriteDeadline(
		base.TimeNow().Add(timeout),
	); e != nil {
		return errors.ErrTCPStreamConnWriteStream.AddDebug("SetReadDeadline error")
	}

	buffer := stream.GetBufferUnsafe()

	if len(buffer) < bufferSize-4 { // fast write
		binary.LittleEndian.PutUint32(p.writeBuf, uint32(len(buffer)))
		copy(p.writeBuf[4:], buffer)
		pos := 0
		for pos < len(buffer)+4 {
			n, e := p.conn.Write(p.writeBuf[pos : len(buffer)+4])
			if e != nil {
				return errors.ErrTCPStreamConnWriteStream.AddDebug(e.Error())
			}
			pos += n
		}
		return nil
	}

	binary.LittleEndian.PutUint32(p.u32Buf[0:], uint32(len(buffer)))
	pos := 0
	for pos < 4 {
		n, e := p.conn.Write(p.u32Buf[pos:])
		if e != nil {
			return errors.ErrTCPStreamConnWriteStream.AddDebug(e.Error())
		}
		pos += n
	}

	pos = 0
	for pos < len(buffer) {
		n, e := p.conn.Write(stream.GetBufferUnsafe())
		if e != nil {
			return errors.ErrTCPStreamConnWriteStream.AddDebug(e.Error())
		}
		pos += n
	}

	return nil
}

func (p *tcpStreamConn) Close() *base.Error {
	if e := p.conn.Close(); e != nil {
		return errors.ErrTCPStreamConnClose.AddDebug("SetReadDeadline error")
	} else {
		return nil
	}
}

type tcpServerAdapter struct {
	addr   string
	server net.Listener
	base.StatusManager
}

// NewTCPServerAdapter ...
func NewTCPServerAdapter(addr string) internal.IServerAdapter {
	return &tcpServerAdapter{
		addr:   addr,
		server: nil,
	}
}

// Open ...
func (p *tcpServerAdapter) Open(
	onConnRun func(internal.IStreamConn, net.Addr),
	onError func(uint64, *base.Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		panic("onConnRun is nil")
	} else {
		server, e := net.Listen("tcp", p.addr)
		if e != nil {
			onError(0, errors.ErrWebsocketServerAdapterWSServerListenAndServe.
				AddDebug(e.Error()))
			return
		}

		if !p.SetRunning(func() {
			p.server = server
		}) {
			onError(0, errors.ErrWebsocketServerAdapterAlreadyRunning)
		} else {
			for {
				// Listen for an incoming connection.
				conn, err := p.server.Accept()
				if err == nil {
					conn.(*net.TCPConn).SetNoDelay(false)
					go onConnRun(NewTCPStreamConn(conn), conn.RemoteAddr())
				} else {
					onError(0, errors.ErrWebsocketServerAdapterWSServerListenAndServe.
						AddDebug(err.Error()))
					break
				}
			}

			p.SetClosing(nil)
			p.SetClosed(func() {
				p.server = nil
			})
		}
	}
}

// Close ...
func (p *tcpServerAdapter) Close(onError func(uint64, *base.Error)) {
	waitCH := chan bool(nil)
	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		if e := p.server.Close(); e != nil {
			onError(
				0,
				errors.ErrWebsocketServerAdapterWSServerClose.
					AddDebug(e.Error()),
			)
		}
	}) {
		onError(0, errors.ErrWebsocketServerAdapterNotRunning)
	} else {
		select {
		case <-waitCH:
		case <-time.After(3 * time.Second):
			onError(0, errors.ErrWebsocketServerAdapterCloseTimeout)
		}
	}
}

type tcpClientAdapter struct {
	conn          internal.IStreamConn
	connectString string
	base.StatusManager
}

// NewTCPClientAdapter ...
func NewTCPClientAdapter(connectString string) internal.IClientAdapter {
	return &tcpClientAdapter{
		conn:          nil,
		connectString: connectString,
	}
}

func (p *tcpClientAdapter) Open(
	onConnRun func(internal.IStreamConn),
	onError func(*base.Error),
) {
	if onError == nil {
		panic("onError is nil")
	} else if onConnRun == nil {
		panic("onConnRun is nil")
	} else if conn, err := net.Dial("tcp", p.connectString); err != nil {
		onError(errors.ErrWebsocketClientAdapterDial.AddDebug(err.Error()))
	} else {
		conn.(*net.TCPConn).SetNoDelay(false)
		streamConn := NewTCPStreamConn(conn)
		if !p.SetRunning(func() {
			p.conn = streamConn
		}) {
			_ = conn.Close()
			onError(errors.ErrWebsocketClientAdapterAlreadyRunning)
		} else {
			onConnRun(streamConn)
			p.SetClosing(nil)
			p.SetClosed(func() {
				p.conn = nil
			})
		}
	}
}

func (p *tcpClientAdapter) Close(onError func(*base.Error)) {
	waitCH := chan bool(nil)

	if onError == nil {
		panic("onError is nil")
	} else if !p.SetClosing(func(ch chan bool) {
		waitCH = ch
		if e := p.conn.Close(); e != nil {
			onError(e)
		}
	}) {
		onError(errors.ErrWebsocketClientAdapterNotRunning)
	} else {
		select {
		case <-waitCH:
		case <-time.After(3 * time.Second):
			onError(errors.ErrWebsocketClientAdapterCloseTimeout)
		}
	}
}

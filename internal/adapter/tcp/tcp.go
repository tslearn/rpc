package tcp

import (
	"encoding/binary"
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"io"
	"net"
	"time"
)

const bufferSize = 1024

type tcpStreamConn struct {
	readBuf []byte
	readPos int
	conn    net.Conn
}

func newTCPStreamConn(
	conn net.Conn,
) *tcpStreamConn {
	if conn == nil {
		return nil
	}

	ret := &tcpStreamConn{
		readBuf: make([]byte, bufferSize),
		conn:    conn,
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

	if p.readPos >= streamLen+8 {
		ret.PutBytesTo(p.readBuf[8:streamLen+8], 0)
		p.readPos = copy(p.readBuf, p.readBuf[streamLen+8:p.readPos-streamLen-8])
	} else {
		ret.PutBytesTo(p.readBuf[8:p.readPos], 0)
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
	pos := 0

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

//
//var (
//  wsUpgradeManager = websocket.Upgrader{
//    ReadBufferSize:    1024,
//    WriteBufferSize:   1024,
//    EnableCompression: false,
//    CheckOrigin: func(r *http.Request) bool {
//      return true
//    },
//  }
//)

type tcpServerAdapter struct {
	addr   string
	server net.Listener
	base.StatusManager
}

// NewWebsocketServerAdapter ...
func NewWebsocketServerAdapter(addr string) internal.IServerAdapter {
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
					onConnRun(newTCPStreamConn(conn), conn.RemoteAddr())
				} else if err == io.EOF {
					break
				} else {
					onError(0, errors.ErrWebsocketServerAdapterWSServerListenAndServe.
						AddDebug(e.Error()))
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

// NewWebsocketClientAdapter ...
func NewWebsocketClientAdapter(connectString string) internal.IClientAdapter {
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
		streamConn := newTCPStreamConn(conn)
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

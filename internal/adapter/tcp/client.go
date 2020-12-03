// +build linux freebsd dragonfly darwin

package tcp

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const tcpClientAdapterRunning = uint32(1)
const tcpClientAdapterClosing = uint32(2)
const tcpClientAdapterClosed = uint32(0)

type tcpClientAdapter struct {
	status  uint32
	closeCH chan bool
	addr    string
	conn    net.Conn
	sync.Mutex
}

// NewTCPClientAdapter ...
func NewTCPClientAdapter(addr string) adapter.IAdapter {
	return &tcpClientAdapter{
		status:  tcpClientAdapterClosed,
		closeCH: make(chan bool),
		addr:    addr,
	}
}

// Open ...
func (p *tcpClientAdapter) Open(receiver adapter.XReceiver) {
	if p.conn != nil {
		receiver.OnEventConnError(nil, errors.ErrTCPClientAdapterAlreadyRunning)
	} else if conn, e := net.Dial("tcp", p.addr); e != nil {
		receiver.OnEventConnError(
			nil,
			errors.ErrTCPClientAdapterDail.AddDebug(e.Error()),
		)
	} else if fd, e := adapter.GetFD(conn); e != nil {
		receiver.OnEventConnError(
			nil,
			errors.ErrTCPClientAdapterDail.AddDebug(e.Error()),
		)
		_ = conn.Close()
	} else {
		atomic.StoreUint32(&p.status, tcpClientAdapterRunning)
		p.conn = conn
		manager := adapter.NewLoopManager(1, receiver)
		manager.AllocChannel().AddConn(
			adapter.NewEventConn(adapter.NewReceiverHook(
				receiver,
				nil,
				func(eventConn *adapter.EventConn) {
					if eventConn != nil && eventConn.GetFD() == fd {
						go func() {
							p.Close(receiver)
						}()
					}
				},
				nil,
				nil,
			), conn, fd),
		)

		for atomic.LoadUint32(&p.status) == tcpClientAdapterRunning {
			time.Sleep(50 * time.Millisecond)
		}

		manager.Close()
		p.conn = nil
		atomic.StoreUint32(&p.status, tcpClientAdapterClosed)
		p.closeCH <- true
	}
}

// Close ...
func (p *tcpClientAdapter) Close(receiver adapter.XReceiver) {
	if atomic.CompareAndSwapUint32(
		&p.status,
		tcpClientAdapterRunning,
		tcpClientAdapterClosing,
	) {
		if e := p.conn.Close(); e != nil {
			receiver.OnEventConnError(
				nil,
				errors.ErrTCPClientAdapterClose.AddDebug(e.Error()),
			)
		}

		<-p.closeCH
	} else {
		receiver.OnEventConnError(nil, errors.ErrTCPClientAdapterNotRunning)
	}
}

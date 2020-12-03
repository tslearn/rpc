// +build linux freebsd dragonfly darwin

package tcp

import (
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/x/adapter"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
)

const tcpServerAdapterRunning = uint32(1)
const tcpServerAdapterClosing = uint32(2)
const tcpServerAdapterClosed = uint32(0)

type tcpServerAdapter struct {
	status  uint32
	closeCH chan bool
	addr    string
	server  net.Listener
	sync.Mutex
}

// NewTCPServerAdapter ...
func NewTCPServerAdapter(addr string) adapter.IAdapter {
	return &tcpServerAdapter{
		status:  tcpServerAdapterClosed,
		closeCH: make(chan bool),
		addr:    addr,
		server:  nil,
	}
}

// Open ...
func (p *tcpServerAdapter) Open(receiver adapter.XReceiver) {
	cpus := runtime.NumCPU()
	if p.server != nil {
		receiver.OnEventConnError(nil, errors.ErrTCPServerAdapterAlreadyRunning)
	} else if server, e := net.Listen("tcp", p.addr); e != nil {
		receiver.OnEventConnError(
			nil,
			errors.ErrTCPServerAdapterListen.AddDebug(e.Error()),
		)
	} else {
		atomic.StoreUint32(&p.status, tcpServerAdapterRunning)
		p.server = server
		manager := adapter.NewLoopManager(cpus, receiver)
		for atomic.LoadUint32(&p.status) == tcpServerAdapterRunning {
			if conn, e := p.server.Accept(); e != nil {
				receiver.OnEventConnError(
					nil,
					errors.ErrTCPServerAdapterAccept.AddDebug(e.Error()),
				)
			} else if fd, e := adapter.GetFD(conn); e != nil {
				receiver.OnEventConnError(
					nil,
					errors.ErrTCPServerAdapterAccept.AddDebug(e.Error()),
				)
			} else {
				manager.AllocChannel().AddConn(
					adapter.NewEventConn(receiver, conn, fd),
				)
			}
		}
		manager.Close()
		p.server = nil
		atomic.StoreUint32(&p.status, tcpServerAdapterClosed)
		p.closeCH <- true
	}
}

// Close ...
func (p *tcpServerAdapter) Close(receiver adapter.XReceiver) {
	if atomic.CompareAndSwapUint32(
		&p.status,
		tcpServerAdapterRunning,
		tcpServerAdapterClosing,
	) {
		<-p.closeCH
	} else {
		receiver.OnEventConnError(nil, errors.ErrTCPServerAdapterNotRunning)
	}
}

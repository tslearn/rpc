package adapter

import (
	"net"

	"github.com/rpccloud/rpc/internal/errors"
)

// // ISyncServer ...
// type ISyncServer interface {
// 	Open() error
// 	Run() error
// 	Close() error
// }

// // SyncServerTCP ...
// type SyncServerTCP struct {
// 	network  string
// 	addr     string
// 	listener net.Listener
// }

// // Open ...
// func (p *SyncServerTCP) Open() (e error) {
// 	p.listener, e = net.Listen(p.network, p.addr)
// 	return
// }

// // Run ...
// func (p *SyncServerTCP) Run() error {
// 	for {
// 		conn, err := p.listener.Accept()

// 		if err == nil {
// 			go onConnRun(NewTCPStreamConn(conn), conn.RemoteAddr())
// 		} else {

// 			break
// 		}
// 	}
// }

// SyncServerAdapter ...
type SyncServerAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int
	receiver IReceiver
}

// NewSyncServerAdapter ...
func NewSyncServerAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *RunnableService {
	return NewRunnableService(&SyncClientAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
		receiver: receiver,
	})
}

// OnRun ...
func (p *SyncServerAdapter) OnRun(service *RunnableService) {
	switch p.network {
	case "tcp":
		p.runAsTCPServer(service)
	}
}

func (p *SyncServerAdapter) runAsTCPServer(service *RunnableService) {
	listener, e := net.Listen(p.network, p.addr)

	if e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
		return
	}

	// for service.IsRunning() {
	// 	conn, e := listener.Accept()

	// 	if e != nil {
	// 		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
	// 	} else {
	// 		//go onConnRun(NewTCPStreamConn(conn), conn.RemoteAddr())
	// 	}
	// }

	if e = listener.Close(); e != nil {
		p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(e.Error()))
	}
}

// OnClose ...
func (p *SyncServerAdapter) OnClose(_ *RunnableService) {

}

// Close ...
func (p *SyncServerAdapter) Close() {
	// do nothing
}

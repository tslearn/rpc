package async

import (
	"github.com/rpccloud/rpc/internal/adapter"
)

//
//manager := xadapter.NewManager(
//p.network,
//p.addr, func(err *base.Error) {
//  receiver.OnEventConnError(nil, err)
//},
//fnConnect,
//runtime.NumCPU(),
//)

type TCPServerAdapter struct {
	network  string
	addr     string
	rBufSize int
	wBufSize int

	manager *common.Manager
}

func NewAsyncTCPServerAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
) *adapter.RunnableService {
	return adapter.NewRunnableService(&TCPServerAdapter{
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
	})
}

func (p *TCPServerAdapter) OnOpen() bool {

}

func (p *TCPServerAdapter) OnRun(service *RunnableService) bool {

}

func (p *TCPServerAdapter) OnClose() {

}

package adapter

import "github.com/rpccloud/rpc/internal/base"

// -----------------------------------------------------------------------------
// asyncTCPServerService
// -----------------------------------------------------------------------------
type asyncTCPServerService struct {
	adapter    *Adapter
	ln         *XListener
	orcManager *base.ORCManager
}

func (p *asyncTCPServerService) Open() bool {
	return p.orcManager.Open(func() bool {
		adapter := p.adapter

		p.ln = NewXListener(
			adapter.network,
			adapter.addr,
			func(conn IConn) {
				conn.SetNext(NewStreamConn(conn, adapter.receiver))
				runIConnOnServer(conn)
			},
			func(err *base.Error) {
				adapter.receiver.OnConnError(nil, err)
			},
			adapter.rBufSize,
			adapter.wBufSize,
		)

		if p.ln == nil {
			return false
		}

		return p.ln.Open()
	})
}

func (p *asyncTCPServerService) Run() bool {
	return true
}

func (p *asyncTCPServerService) Close() bool {
	return p.orcManager.Close(func() bool {
		p.ln.Close()
		return true
	}, func() {
		p.ln = nil
	})
}

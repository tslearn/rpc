package adapter

import (
    "context"
    "crypto/tls"
    "fmt"
    "net"
    "net/url"
    "time"

    "github.com/rpccloud/rpc/internal/adapter/common"
    "github.com/rpccloud/rpc/internal/errors"

    "github.com/gobwas/ws"
    "github.com/rpccloud/rpc/internal/base"
)

// ClientTCP ...
type ClientTCP struct {
    adapter    *ClientAdapter
    conn       *common.NetConn
    orcManager *base.ORCManager
}

// NewClientTCP ...
func NewClientTCP(adapter *ClientAdapter) base.IORCService {
    return &ClientTCP{
        adapter:    adapter,
        conn:       nil,
        orcManager: base.NewORCManager(),
    }
}

// Open ...
func (p *ClientTCP) Open() bool {
    return p.orcManager.Open(func() bool {
        var e error
        var conn net.Conn

        adapter := p.adapter

        if adapter.tlsConfig == nil {
            conn, e = net.Dial(adapter.network, adapter.addr)
        } else {
            conn, e = tls.Dial(adapter.network, adapter.addr, adapter.tlsConfig)
        }

        if e != nil {
            adapter.receiver.OnConnError(
                nil,
                errors.ErrTemp.AddDebug(e.Error()),
            )
            return false
        }

        p.conn = adapter.CreateNetConn(conn)
        return true
    })
}

// Run ...
func (p *ClientTCP) Run() bool {
    return p.orcManager.Run(func(isRunning func() bool) {
        p.conn.OnOpen()
        for {
            if ok := p.conn.OnReadReady(); !ok {
                break
            }
        }
        p.conn.OnClose()
    })
}

// Close ...
func (p *ClientTCP) Close() bool {
    return p.orcManager.Close(func() {
        p.conn.Close()
    }, func() {
        p.conn = nil
    })
}

// ClientWebsocket ...
type ClientWebsocket struct {
    adapter    *ClientAdapter
    conn       *common.NetConn
    orcManager *base.ORCManager
}

// NewClientWebsocket ...
func NewClientWebsocket(adapter *ClientAdapter) base.IORCService {
    return &ClientWebsocket{
        adapter:    adapter,
        conn:       nil,
        orcManager: base.NewORCManager(),
    }
}

// Open ...
func (p *ClientWebsocket) Open() bool {
    return p.orcManager.Open(func() bool {
        adapter := p.adapter
        dialer := &ws.Dialer{TLSConfig: adapter.tlsConfig}
        u := url.URL{Scheme: adapter.network, Host: adapter.addr, Path: "/"}
        conn, _, _, e := dialer.Dial(context.Background(), u.String())

        if e != nil {
            adapter.receiver.OnConnError(
                nil,
                errors.ErrTemp.AddDebug(e.Error()),
            )
            return false
        }

        p.conn = adapter.CreateNetConn(conn)

        return true
    })
}

// Run ...
func (p *ClientWebsocket) Run() bool {
    return p.orcManager.Run(func(isRunning func() bool) {
        p.conn.OnOpen()
        for {
            if ok := p.conn.OnReadReady(); !ok {
                break
            }
        }
        p.conn.OnClose()
    })
}

// Close ...
func (p *ClientWebsocket) Close() bool {
    return p.orcManager.Close(func() {
        p.conn.Close()
    }, func() {
        p.conn = nil
    })
}

// ClientAdapter ...
type ClientAdapter struct {
    network    string
    addr       string
    tlsConfig  *tls.Config
    rBufSize   int
    wBufSize   int
    receiver   common.IReceiver
    client     base.IORCService
    orcManager *base.ORCManager
}

// NewClientAdapter ...
func NewClientAdapter(
    network string,
    addr string,
    tlsConfig *tls.Config,
    rBufSize int,
    wBufSize int,
    receiver common.IReceiver,
) *ClientAdapter {
    return &ClientAdapter{
        network:    network,
        addr:       addr,
        tlsConfig:  tlsConfig,
        rBufSize:   rBufSize,
        wBufSize:   wBufSize,
        receiver:   receiver,
        client:     nil,
        orcManager: base.NewORCManager(),
    }
}

func (p *ClientAdapter) CreateNetConn(conn net.Conn) *common.NetConn {
    ret := common.NewNetConn(false, conn, p.rBufSize, p.wBufSize)
    ret.SetNext(common.NewStreamConn(ret, p.receiver))
    return ret
}

// Open ...
func (p *ClientAdapter) Open() bool {
    return p.orcManager.Open(func() bool {
        switch p.network {
        case "tcp4":
            fallthrough
        case "tcp6":
            fallthrough
        case "tcp":
            p.client = NewClientTCP(p)
            return true
        case "ws":
            fallthrough
        case "wss":
            p.client = NewClientWebsocket(p)
            return true
        default:
            p.receiver.OnConnError(nil, errors.ErrTemp.AddDebug(
                fmt.Sprintf("unsupported protocol %s", p.network),
            ))
            return false
        }
    })
}

// Run ...
func (p *ClientAdapter) Run() bool {
    return p.orcManager.Run(func(isRunning func() bool) {
        for isRunning() {
            start := base.TimeNow()
            p.client.Open()
            p.client.Run()
            p.client.Close()

            if isRunning() {
                if delta := base.TimeNow().Sub(start); delta < time.Second {
                    time.Sleep(time.Second - delta)
                }
            }
        }
    })
}

// Close ...
func (p *ClientAdapter) Close() bool {
    return p.orcManager.Close(func() {
        p.client.Close()
    }, func() {
        p.client = nil
    })
}

// +build linux freebsd dragonfly darwin

package tcp

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/xadapter"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const tcpServerAdapterLoading = uint32(1)
const tcpServerAdapterRunning = uint32(2)
const tcpServerAdapterClosing = uint32(3)
const tcpServerAdapterClosed = uint32(0)

type tcpServerAdapter struct {
	status   uint32
	closeCH  chan bool
	network  string
	addr     string
	rBufSize int
	wBufSize int
	sync.Mutex
}

// NewTCPServerAdapter ...
func NewTCPServerAdapter(
	network string,
	addr string,
	rBufSize int,
	wBufSize int,
) adapter.IAdapter {
	return &tcpServerAdapter{
		status:   tcpServerAdapterClosed,
		closeCH:  make(chan bool),
		network:  network,
		addr:     addr,
		rBufSize: rBufSize,
		wBufSize: wBufSize,
	}
}

// Open ...
func (p *tcpServerAdapter) Open(receiver adapter.XReceiver) {
	fnConnect := func(
		fd int, lAddr net.Addr, rAddr net.Addr, channel *xadapter.Channel,
	) *xadapter.ChannelConn {
		channelConn := xadapter.NewChannelConn(
			channel, fd, lAddr, rAddr, p.rBufSize, p.wBufSize,
		)
		channelConn.SetNext(xadapter.NewStreamConn(channelConn))
		return channelConn
	}

	if manager := xadapter.NewManager(
		p.network,
		p.addr, func(err *base.Error) {
			receiver.OnEventConnError(nil, err)
		},
		fnConnect,
		runtime.NumCPU(),
	); manager == nil {
		// do nothing
	} else if atomic.CompareAndSwapUint32(
		&p.status,
		tcpServerAdapterClosed,
		tcpServerAdapterLoading,
	) {
		atomic.StoreUint32(&p.status, tcpServerAdapterRunning)
		for atomic.LoadUint32(&p.status) == tcpServerAdapterRunning {
			time.Sleep(50 * time.Millisecond)
		}
		manager.Close()
		atomic.StoreUint32(&p.status, tcpServerAdapterClosed)
		p.closeCH <- true
	} else {
		manager.Close()
		receiver.OnEventConnError(nil, errors.ErrTCPServerAdapterAlreadyRunning)
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

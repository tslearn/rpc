// +build linux darwin

package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
	"sync/atomic"
)

// TCPListener ...
type TCPListener struct {
	status   int32
	lnFD     int
	lnAddr   net.Addr
	poller   *Poller
	onAccept func(fd int, localAddr net.Addr, remoteAddr net.Addr)
	onError  func(err *base.Error)
}

// NewTCPListener ...
func NewTCPListener(
	network string,
	addr string,
	onAccept func(fd int, localAddr net.Addr, remoteAddr net.Addr),
	onError func(err *base.Error),
) *TCPListener {
	ret := &TCPListener{
		status:   tcpListenerStatusRunning,
		onAccept: onAccept,
		onError:  onError,
	}

	if lnFD, lnAddr, e := listen(network, addr); e != nil {
		onError(errors.ErrTCPListener.AddDebug(e.Error()))
		return nil
	} else if poller := NewPoller(
		onError,
		func() {},
		func(fd int) {
			if fd == ret.lnFD {
				if connFD, sa, e := unix.Accept(ret.lnFD); e != nil {
					if e != unix.EAGAIN {
						ret.onError(errors.ErrTCPListener.AddDebug(e.Error()))
					}
				} else {
					if remoteAddr := sockAddrToTCPAddr(sa); remoteAddr != nil {
						onAccept(connFD, ret.lnAddr, remoteAddr)
					} else {
						_ = unix.Close(connFD)
					}
				}
			}
		},
		func(fd int) {},
		func(fd int) {},
	); poller == nil {
		_ = unix.Close(lnFD)
		return nil
	} else if e := poller.RegisterFD(lnFD); e != nil {
		_ = unix.Close(lnFD)
		poller.Close()
		return nil
	} else {
		ret.lnFD = lnFD
		ret.lnAddr = lnAddr
		ret.poller = poller
		return ret
	}
}

// Close ...
func (p *TCPListener) Close() {
	if atomic.CompareAndSwapInt32(
		&p.status,
		tcpListenerStatusRunning,
		tcpListenerStatusClosed,
	) {
		_ = unix.Close(p.lnFD)
		p.poller.Close()
	}
}

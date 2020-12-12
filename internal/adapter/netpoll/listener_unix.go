// +build linux freebsd dragonfly darwin

package netpoll

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
	"sync/atomic"
)

const tcpListenerStatusRunning = 1
const tcpListenerStatusClosed = 0

type TCPListener struct {
	status   int32
	lnFD     int
	lnAddr   net.Addr
	poller   *Poller
	onAccept func(fd int, localAddr net.Addr, remoteAddr net.Addr)
	onError  func(err *base.Error)
}

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

	if lnFD, lnAddr, e := TCPSocket(network, addr); e != nil {
		onError(errors.ErrTCPListener.AddDebug(e.Error()))
		return nil
	} else if poller := NewPoller(
		onError,
		func() {},
		func() {},
		ret.OnRead,
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

func (p *TCPListener) OnRead(fd int) {
	if fd == p.lnFD {
		if connFD, sa, e := unix.Accept(p.lnFD); e != nil {
			if e != unix.EAGAIN {
				p.onError(errors.ErrTCPListener.AddDebug(e.Error()))
			}
		} else {
			if remoteAddr := sockAddrToTCPAddr(sa); remoteAddr != nil {
				p.onAccept(connFD, p.lnAddr, remoteAddr)
			} else {
				_ = unix.Close(connFD)
			}
		}
	}
}

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

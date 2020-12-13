// +build linux darwin freebsd dragonfly

package netpoll

import (
	"net"
	"sync/atomic"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
)

const tcpListenerStatusRunning = 1
const tcpListenerStatusClosed = 0

func listen(network string, addr string) (int, net.Addr, error) {
	if sockAddr, family, netAddr, err := getTCPSockAddr(
		network, addr,
	); err != nil {
		return 0, nil, err
	} else if fd, err := sysSocket(
		family, unix.SOCK_STREAM, unix.IPPROTO_TCP,
	); err != nil {
		return 0, nil, err
	} else if err := unix.SetsockoptInt(
		fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1,
	); err != nil {
		_ = unix.Close(fd)
		return 0, nil, err
	} else if err := unix.Bind(fd, sockAddr); err != nil {
		_ = unix.Close(fd)
		return 0, nil, err
	} else if err := unix.Listen(fd, 128); err != nil {
		_ = unix.Close(fd)
		return 0, nil, err
	} else {
		return fd, netAddr, nil
	}
}

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

// OnRead ...
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

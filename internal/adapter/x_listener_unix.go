// +build linux darwin

package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
)

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
	network    string
	addr       string
	onAccept   func(fd int, localAddr net.Addr, remoteAddr net.Addr)
	onError    func(err *base.Error)
	lnFD       int
	lnAddr     net.Addr
	poller     *Poller
	orcManager *base.ORCManager
}

// NewTCPListener ...
func NewTCPListener(
	network string,
	addr string,
	onAccept func(fd int, localAddr net.Addr, remoteAddr net.Addr),
	onError func(err *base.Error),
) *TCPListener {
	return &TCPListener{
		network:    network,
		addr:       addr,
		onAccept:   onAccept,
		onError:    onError,
		lnFD:       0,
		lnAddr:     nil,
		poller:     nil,
		orcManager: base.NewORCManager(),
	}
}

func (p *TCPListener) Open() bool {
	return p.orcManager.Open(func() bool {
		if lnFD, lnAddr, e := listen(p.network, p.addr); e != nil {
			p.onError(errors.ErrTCPListener.AddDebug(e.Error()))
			return false
		} else if poller := NewPoller(
			p.onError,
			func() {},
			func(fd int) {
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
			},
			func(fd int) {},
			func(fd int) {},
		); poller == nil {
			_ = unix.Close(lnFD)
			return false
		} else if e := poller.RegisterFD(lnFD); e != nil {
			_ = unix.Close(lnFD)
			poller.Close()
			return false
		} else {
			p.lnFD = lnFD
			p.lnAddr = lnAddr
			p.poller = poller
			return true
		}
	})
}

func (p *TCPListener) Run() bool {
	return true
}

// Close ...
func (p *TCPListener) Close() bool {
	return p.orcManager.Close(func() {
		if e := unix.Close(p.lnFD); e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
		}
		p.poller.Close()
	}, func() {
		p.lnFD = 0
		p.lnAddr = nil
		p.poller = nil
	})
}

// +build linux darwin

package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
	"runtime"
)

func getChannelSize() int {
	return int(runtime.NumCgoCall())
}

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

// XListener ...
type XListener struct {
	network    string
	addr       string
	onAccept   func(conn IConn)
	onError    func(err *base.Error)
	rBufSize   int
	wBufSize   int
	lnFD       int
	lnAddr     net.Addr
	mainPoller *Poller
	channels   []*Channel
	curChannel *Channel
	curRemains uint64
	orcManager *base.ORCManager
}

// NewXListener ...
func NewXListener(
	network string,
	addr string,
	onAccept func(conn IConn),
	onError func(err *base.Error),
	rBufSize int,
	wBufSize int,
) *XListener {
	channelSize := getChannelSize()

	ret := &XListener{
		network:    network,
		addr:       addr,
		onAccept:   onAccept,
		onError:    onError,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		lnFD:       0,
		lnAddr:     nil,
		mainPoller: nil,
		channels:   make([]*Channel, channelSize),
		curChannel: nil,
		curRemains: 0,
		orcManager: base.NewORCManager(),
	}

	for i := 0; i < channelSize; i++ {
		channel := NewChannel(onError)

		if channel != nil {
			ret.channels[i] = channel
		} else {
			// clean up and return nil
			for j := 0; j < i; j++ {
				ret.channels[j].Close()
				ret.channels[j] = nil
			}
			return nil
		}
	}

	return ret
}

func (p *XListener) allocChannel() *Channel {
	if p.curRemains <= 0 {
		maxConn := int64(-1)
		for i := 0; i < len(p.channels); i++ {
			if connCount := p.channels[i].GetActiveConnCount(); connCount > maxConn {
				p.curChannel = p.channels[i]
				maxConn = connCount
			}
		}
		p.curRemains = 256
	}

	p.curRemains--
	return p.curChannel
}

func (p *XListener) Open() bool {
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
							p.onAccept(NewXConn(
								p.allocChannel(),
								connFD,
								p.lnAddr,
								remoteAddr,
								p.rBufSize,
								p.wBufSize,
							))
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
			p.mainPoller = poller
			return true
		}
	})
}

func (p *XListener) Run() bool {
	return true
}

// Close ...
func (p *XListener) Close() bool {
	return p.orcManager.Close(func() bool {
		if e := unix.Close(p.lnFD); e != nil {
			p.onError(errors.ErrTemp.AddDebug(e.Error()))
		}
		p.mainPoller.Close()

		// cl
		waitCH := make(chan bool)
		channelSize := len(p.channels)

		for i := 0; i < channelSize; i++ {
			go func(idx int) {
				p.channels[idx].Close()
				waitCH <- true
			}(i)
		}

		for i := 0; i < channelSize; i++ {
			<-waitCH
		}
		return true
	}, func() {
		p.lnFD = 0
		p.lnAddr = nil
		p.mainPoller = nil
	})
}

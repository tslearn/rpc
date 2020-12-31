// +build linux darwin

package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

func getTCPSockAddr(
	network string,
	addr string,
) (unix.Sockaddr, int, *net.TCPAddr, *base.Error) {
	if addr, e := net.ResolveTCPAddr(network, addr); e != nil {
		return nil, unix.AF_UNSPEC, nil, errors.ErrTemp.AddDebug(e.Error())
	} else if addr.IP.To4() != nil || network == "tcp4" {
		sa4 := &unix.SockaddrInet4{Port: addr.Port}

		if addr.IP != nil {
			if len(addr.IP) == 16 {
				copy(sa4.Addr[:], addr.IP[12:16])
			} else {
				copy(sa4.Addr[:], addr.IP)
			}
		}
		return sa4, unix.AF_INET, addr, nil
	} else if addr.IP.To16() != nil || network == "tcp6" {
		sa6 := &unix.SockaddrInet6{Port: addr.Port}

		if addr.IP != nil {
			copy(sa6.Addr[:], addr.IP)
		}

		if addr.Zone != "" {
			if netInterface, e := net.InterfaceByName(addr.Zone); e == nil {
				sa6.ZoneId = uint32(netInterface.Index)
			} else {
				return nil, unix.AF_UNSPEC, nil, errors.ErrTemp.AddDebug(e.Error())
			}
		}

		return sa6, unix.AF_INET6, addr, nil
	} else if network == "tcp" {
		return &unix.SockaddrInet4{Port: addr.Port}, unix.AF_INET, addr, nil
	} else {
		return nil, unix.AF_UNSPEC, nil, errors.ErrTemp.AddDebug("tcp: get proto error")
	}
}

func ip6ZoneToString(zone int) string {
	if zone == 0 {
		return ""
	}
	if ifi, err := net.InterfaceByIndex(zone); err == nil {
		return ifi.Name
	}
	return strconv.FormatUint(uint64(zone), 10)
}

func sockAddrToIPAndZone(sa unix.Sockaddr) (net.IP, string) {
	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		ip := make([]byte, 16)
		// V4InV6Prefix
		ip[10] = 0xff
		ip[11] = 0xff
		copy(ip[12:16], sa.Addr[:])
		return ip, ""
	case *unix.SockaddrInet6:
		ip := make([]byte, 16)
		copy(ip, sa.Addr[:])
		return ip, ip6ZoneToString(int(sa.ZoneId))
	}
	return nil, ""
}

func sockAddrToTCPAddr(sa unix.Sockaddr) *net.TCPAddr {
	ip, zone := sockAddrToIPAndZone(sa)

	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		return &net.TCPAddr{IP: ip, Port: sa.Port}
	case *unix.SockaddrInet6:
		return &net.TCPAddr{IP: ip, Port: sa.Port, Zone: zone}
	}
	return nil
}

//func sockAddrToUDPAddr(sa unix.Sockaddr) *net.UDPAddr {
//	ip, zone := sockAddrToIPAndZone(sa)
//
//	switch sa := sa.(type) {
//	case *unix.SockaddrInet4:
//		return &net.UDPAddr{IP: ip, Port: sa.Port}
//	case *unix.SockaddrInet6:
//		return &net.UDPAddr{IP: ip, Port: sa.Port, Zone: zone}
//	}
//	return nil
//}

// Channel ...
type Channel struct {
	onError         func(err *base.Error)
	activeConnCount int64
	poller          *Poller
	connMap         map[int]IConn
	addCH           chan IConn
	sync.Mutex
}

// NewChannel ...
func NewChannel(onError func(err *base.Error)) *Channel {
	ret := &Channel{
		onError:         onError,
		activeConnCount: 0,
		poller:          nil,
		connMap:         make(map[int]IConn),
		addCH:           make(chan IConn, 4096),
	}

	ret.poller = NewPoller(
		ret.onError,
		ret.onTrigger,
		ret.onFDRead,
		ret.onFDWrite,
		ret.onFDClose,
	)

	if ret.poller == nil {
		return nil
	}

	return ret
}

// Close ...
func (p *Channel) Close() {
	p.Lock()
	defer p.Unlock()

	p.poller.Close()
}

// AddConn ...
func (p *Channel) AddConn(conn IConn) {
	p.addCH <- conn
	_ = p.poller.Trigger()
}

func (p *Channel) onTrigger() {
	for {
		select {
		case conn := <-p.addCH:
			if e := p.poller.RegisterFD(conn.GetFD()); e != nil {
				p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
			} else {
				p.connMap[conn.GetFD()] = conn
				conn.OnOpen()
			}
		default:
			return
		}
	}
}

func (p *Channel) onFDRead(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnReadReady()
	}
}

func (p *Channel) onFDWrite(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		conn.OnWriteReady()
	}
}

func (p *Channel) onFDClose(fd int) {
	if conn, ok := p.connMap[fd]; ok {
		if e := unix.Close(fd); e != nil {
			conn.OnError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		} else {
			delete(p.connMap, fd)
			conn.OnClose()
		}
	}
}

// CloseFD ...
func (p *Channel) CloseFD(fd int) error {
	if e := p.poller.UnregisteredFD(fd); e != nil {
		return e
	}

	if e := unix.Close(fd); e != nil {
		return e
	}

	return nil
}

// SetWriteFD ...
func (p *Channel) SetWriteFD(fd int, isWatch bool) error {
	if isWatch {
		return p.poller.AddWrite(fd)
	}

	return p.poller.DelWrite(fd)
}

// GetActiveConnCount ...
func (p *Channel) GetActiveConnCount() int64 {
	return atomic.LoadInt64(&p.activeConnCount)
}

// Manager ...
type Manager struct {
	channels   []*Channel
	onError    func(err *base.Error)
	curChannel *Channel
	curRemains uint64
}

// NewManager ...
func NewManager(
	onError func(err *base.Error),
	size int,
) *Manager {
	if size < 1 {
		size = 1
	}

	ret := &Manager{
		channels:   make([]*Channel, size),
		onError:    onError,
		curChannel: nil,
		curRemains: 0,
	}

	for i := 0; i < size; i++ {
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

// Close ...
func (p *Manager) Close() {
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
}

// AllocChannel ...
func (p *Manager) AllocChannel() *Channel {
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

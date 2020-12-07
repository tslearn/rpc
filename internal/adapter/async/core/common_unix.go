// +build linux freebsd dragonfly darwin

package xadapter

import (
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
)

func getTCPSockAddr(
	network string,
	addr string,
) (unix.Sockaddr, int, *net.TCPAddr, error) {
	if addr, err := net.ResolveTCPAddr(network, addr); err != nil {
		return nil, unix.AF_UNSPEC, nil, err
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
			if netInterface, err := net.InterfaceByName(addr.Zone); err != nil {
				return nil, unix.AF_UNSPEC, nil, err
			} else {
				sa6.ZoneId = uint32(netInterface.Index)
			}
		}

		return sa6, unix.AF_INET6, addr, nil
	} else if network == "tcp" {
		return &unix.SockaddrInet4{Port: addr.Port}, unix.AF_INET, addr, nil
	} else {
		return nil, unix.AF_UNSPEC, nil, errors.New("tcp: get proto error")
	}
}

func TCPSocket(network string, addr string) (int, net.Addr, error) {
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

func sockAddrToUDPAddr(sa unix.Sockaddr) *net.UDPAddr {
	ip, zone := sockAddrToIPAndZone(sa)
	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		return &net.UDPAddr{IP: ip, Port: sa.Port}
	case *unix.SockaddrInet6:
		return &net.UDPAddr{IP: ip, Port: sa.Port, Zone: zone}
	}
	return nil
}

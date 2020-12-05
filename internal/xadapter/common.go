package xadapter

import (
	"golang.org/x/sys/unix"
	"net"
	"strconv"
)

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

// +build linux freebsd dragonfly darwin

package xadapter

import (
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"strconv"
)

func getTCPSockaddr(proto, addr string) (sa unix.Sockaddr, family int, tcpAddr *net.TCPAddr, err error) {
	var tcpVersion string

	tcpAddr, err = net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return
	}

	tcpVersion, err = determineTCPProto(proto, tcpAddr)
	if err != nil {
		return
	}

	switch tcpVersion {
	case "tcp":
		sa, family = &unix.SockaddrInet4{Port: tcpAddr.Port}, unix.AF_INET
	case "tcp4":
		sa4 := &unix.SockaddrInet4{Port: tcpAddr.Port}

		if tcpAddr.IP != nil {
			if len(tcpAddr.IP) == 16 {
				copy(sa4.Addr[:], tcpAddr.IP[12:16]) // copy last 4 bytes of slice to array
			} else {
				copy(sa4.Addr[:], tcpAddr.IP) // copy all bytes of slice to array
			}
		}

		sa, family = sa4, unix.AF_INET
	case "tcp6":
		sa6 := &unix.SockaddrInet6{Port: tcpAddr.Port}

		if tcpAddr.IP != nil {
			copy(sa6.Addr[:], tcpAddr.IP) // copy all bytes of slice to array
		}

		if tcpAddr.Zone != "" {
			var iface *net.Interface
			iface, err = net.InterfaceByName(tcpAddr.Zone)
			if err != nil {
				return
			}

			sa6.ZoneId = uint32(iface.Index)
		}

		sa, family = sa6, unix.AF_INET6
	default:
		err = errors.New("tcp: proto error")
	}

	return
}

func determineTCPProto(proto string, addr *net.TCPAddr) (string, error) {
	// If the protocol is set to "tcp", we try to determine the actual protocol
	// version from the size of the resolved IP address. Otherwise, we simple use
	// the protcol given to us by the caller.

	if addr.IP.To4() != nil {
		return "tcp4", nil
	}

	if addr.IP.To16() != nil {
		return "tcp6", nil
	}

	switch proto {
	case "tcp", "tcp4", "tcp6":
		return proto, nil
	}

	return "", errors.New("tcp: proto error")
}

// tcpReusablePort creates an endpoint for communication and returns a file descriptor that refers to that endpoint.
// Argument `reusePort` indicates whether the SO_REUSEPORT flag will be assigned.
func TCPSocket(proto, addr string) (fd int, netAddr net.Addr, err error) {
	var (
		family   int
		sockaddr unix.Sockaddr
	)

	if sockaddr, family, netAddr, err = getTCPSockaddr(proto, addr); err != nil {
		return
	}

	if fd, err = sysSocket(family, unix.SOCK_STREAM, unix.IPPROTO_TCP); err != nil {
		err = os.NewSyscallError("socket", err)
		return
	}
	defer func() {
		if err != nil {
			_ = unix.Close(fd)
		}
	}()

	if err = os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)); err != nil {
		return
	}

	if err = os.NewSyscallError("bind", unix.Bind(fd, sockaddr)); err != nil {
		return
	}

	// Set backlog size to the maximum.
	err = os.NewSyscallError("listen", unix.Listen(fd, 128))

	return
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

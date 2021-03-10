package base

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
)

var (
	fnNetInterfaces = net.Interfaces
)

type GlobalID struct {
	listener net.Listener
	port     uint16
	mac      net.HardwareAddr
}

func NewGlobalID() *GlobalID {
	ret := &GlobalID{
		listener: nil,
		port:     0,
		mac:      nil,
	}
	// try to find a port
	for i := 0; i < 1000; i++ {
		testPort := uint16(50000 + rand.Uint64()%15000)
		if ln, e := net.Listen(
			"tcp",
			fmt.Sprintf("localhost:%d", testPort),
		); e != nil {
			ret.port = testPort
			ret.listener = ln
			break
		}
	}

	// try to find a mac
	if interfaces, e := fnNetInterfaces(); e == nil {
		for _, inter := range interfaces {
			interfaceAddrList, _ := inter.Addrs()
			for _, address := range interfaceAddrList {
				ipNet, isValidIpNet := address.(*net.IPNet)
				if isValidIpNet && ipNet.IP.IsGlobalUnicast() {
					ret.mac = inter.HardwareAddr
					break
				}
			}
		}
	}

	return ret
}

func (p *GlobalID) GetID() uint64 {
	if p.port > 0 && len(p.mac) == 6 {
		buffer := make([]byte, 8)
		binary.LittleEndian.PutUint16(buffer, p.port)
		copy(buffer[2:], p.mac)
		return binary.LittleEndian.Uint64(buffer)
	}

	return 0
}

func (p *GlobalID) Close() {
	if p.listener != nil {
		_ = p.listener.Close()
	}
	p.listener = nil
	p.port = 0
	p.mac = nil
}

//// FindMacAddrByIP ...
//func FindMacAddrByIP(ip string) string {
//    if interfaces, e := fnNetInterfaces(); e == nil {
//        for _, inter := range interfaces {
//            interfaceAddrList, _ := inter.Addrs()
//            for _, address := range interfaceAddrList {
//                ipNet, isValidIpNet := address.(*net.IPNet)
//                if isValidIpNet && ipNet.IP.String() == ip {
//                    return inter.HardwareAddr.String()
//                }
//            }
//        }
//    }
//
//    return ""
//}

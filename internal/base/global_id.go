package base

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
)

var (
	fnNetInterfaces = net.Interfaces
	fnNetListen     = net.Listen
)

// GlobalID ...
type GlobalID struct {
	listener net.Listener
	port     uint16
	mac      net.HardwareAddr
}

// NewGlobalID ...
func NewGlobalID() *GlobalID {
	port := uint16(0)
	mac := []byte(nil)
	listener := net.Listener(nil)

	// try to find a port
	for i := 0; i < 1000; i++ {
		testPort := uint16(50000 + rand.Uint64()%15000)
		if ln, e := fnNetListen(
			"tcp",
			fmt.Sprintf("localhost:%d", testPort),
		); e == nil {
			port = testPort
			listener = ln
			break
		}
	}

	// try to find a mac
	if interfaces, e := fnNetInterfaces(); e == nil {
		for _, inter := range interfaces {
			interfaceAddrList, _ := inter.Addrs()
			for _, address := range interfaceAddrList {
				ipNet, isValidIPNet := address.(*net.IPNet)
				if isValidIPNet && ipNet.IP.IsGlobalUnicast() {
					mac = inter.HardwareAddr
					break
				}
			}
		}
	}

	// check
	if port <= 0 || len(mac) != 6 {
		return nil
	}

	return &GlobalID{
		listener: listener,
		port:     port,
		mac:      mac,
	}
}

// GetID ...
func (p *GlobalID) GetID() uint64 {
	if p.port > 0 && len(p.mac) == 6 {
		buffer := make([]byte, 8)
		binary.LittleEndian.PutUint16(buffer, p.port)
		copy(buffer[2:], p.mac)
		return binary.LittleEndian.Uint64(buffer)
	}

	return 0
}

// Close ...
func (p *GlobalID) Close() {
	if p.listener != nil {
		_ = p.listener.Close()
	}
	p.listener = nil
	p.port = 0
	p.mac = nil
}

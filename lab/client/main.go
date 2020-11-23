package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

func main() {
	sendString := "hello world"
	conn, e := net.Dial("tcp", "0.0.0.0:28888")
	if e != nil {
		panic(e)
	}
	conn.(*net.TCPConn).SetNoDelay(false)

	rPos := 0
	rBuf := make([]byte, 1024)
	wBuf := make([]byte, 1024)

	for i := 0; i < 100; i++ {
		binary.LittleEndian.PutUint32(wBuf, uint32(len(sendString)))
		copy(wBuf[4:], sendString)
		wPos := 0
		for wPos < len(sendString)+4 {
			n, e := conn.Write(wBuf[wPos : len(sendString)+4])
			if e != nil {
				fmt.Println(e)
				return
			}
			wPos += n
		}

		for rPos < 4 {
			n, e := conn.Read(rBuf[rPos:])
			if e != nil {
				fmt.Println(e)
				return
			}
			rPos += n
		}

		streamLen := int(binary.LittleEndian.Uint32(rBuf))

		for rPos < streamLen+4 {
			n, e := conn.Read(rBuf[rPos:])
			if e != nil {
				fmt.Println(e)
				return
			}
			rPos += n
		}

		if string(rBuf[4:streamLen+4]) != sendString {
			fmt.Println("tcp error")
			return
		}

		rPos = copy(rBuf, rBuf[streamLen+4:rPos])
	}
}

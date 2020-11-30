package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	//sendString := "hello world"
	waitCH := make(chan bool)
	pCount := 100

	for i := 0; i < pCount; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				conn, e := net.Dial("tcp", "0.0.0.0:8080")
				if e != nil {
					time.Sleep(time.Second)
					fmt.Println(e)
				} else {
					_, _ = conn.Write([]byte("hello"))
					time.Sleep(10 * time.Millisecond)
					_ = conn.Close()
				}
			}

			waitCH <- true
		}()
	}

	for i := 0; i < pCount; i++ {
		<-waitCH
	}

	//conn.(*net.TCPConn).SetNoDelay(false)
	//
	//rPos := 0
	//rBuf := make([]byte, 1024)
	//wBuf := make([]byte, 1024)
	//
	//for i := 0; i < 100; i++ {
	//	binary.LittleEndian.PutUint32(wBuf, uint32(len(sendString)))
	//	copy(wBuf[4:], sendString)
	//	wPos := 0
	//	for wPos < len(sendString)+4 {
	//		n, e := conn.Write(wBuf[wPos : len(sendString)+4])
	//		if e != nil {
	//			fmt.Println(e)
	//			return
	//		}
	//		wPos += n
	//	}
	//
	//	for rPos < 4 {
	//		n, e := conn.Read(rBuf[rPos:])
	//		if e != nil {
	//			fmt.Println(e)
	//			return
	//		}
	//		rPos += n
	//	}
	//
	//	streamLen := int(binary.LittleEndian.Uint32(rBuf))
	//
	//	for rPos < streamLen+4 {
	//		n, e := conn.Read(rBuf[rPos:])
	//		if e != nil {
	//			fmt.Println(e)
	//			return
	//		}
	//		rPos += n
	//	}
	//
	//	if string(rBuf[4:streamLen+4]) != sendString {
	//		fmt.Println("tcp error")
	//		return
	//	}
	//
	//	rPos = copy(rBuf, rBuf[streamLen+4:rPos])
	//}
}

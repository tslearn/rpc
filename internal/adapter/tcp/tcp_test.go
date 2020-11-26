package tcp

import (
	"bytes"
	"fmt"
	"github.com/rpccloud/rpc/internal/core"
	"net"
	"testing"
	"time"
)

func BenchmarkClient_TCP(b *testing.B) {
	requestStream := core.NewStream()
	responseStream := core.NewStream()

	fnServer := func(conn net.Conn) {
		sConn := NewTCPStreamConn(conn)

		for {
			s, e := sConn.ReadStream(time.Second, 1024)
			if e != nil {
				fmt.Println(e)
				return
			}
			if !bytes.Equal(s.GetBufferUnsafe(), requestStream.GetBufferUnsafe()) {
				fmt.Println("net error")
				return
			}
			s.Release()
			if e := sConn.WriteStream(responseStream, time.Second); e != nil {
				fmt.Println(e)
				return
			}
		}
	}

	go func() { // create serve
		server, err := net.Listen("tcp", "0.0.0.0:28888")
		if err != nil {
			panic(err)
		}

		for {
			// Listen for an incoming connection.
			conn, err := server.Accept()
			if err != nil {
				panic(err)
			}

			fnServer(conn)
		}
	}()

	time.Sleep(time.Second)
	conn, e := net.Dial("tcp", "0.0.0.0:28888")
	if e != nil {
		panic(e)
	}
	sConn := NewTCPStreamConn(conn)

	b.ReportAllocs()
	b.ResetTimer()
	b.N = 5000000

	for i := 0; i < b.N; i++ {
		if e := sConn.WriteStream(requestStream, time.Second); e != nil {
			fmt.Println(e)
			return
		}

		s, e := sConn.ReadStream(time.Second, 1024)
		if e != nil {
			fmt.Println(e)
			return
		}
		if !bytes.Equal(s.GetBufferUnsafe(), responseStream.GetBufferUnsafe()) {
			fmt.Println("net error")
			return
		}
		s.Release()
	}
}

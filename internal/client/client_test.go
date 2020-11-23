package client

import (
	"encoding/binary"
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/internal/adapter/tcp"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/server"
	"net"
	"testing"
	"time"
)

func TestClient_Debug(t *testing.T) {
	rpcServer := server.NewServer().ListenTCP("0.0.0.0:28888")

	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(3000 * time.Millisecond)

	rpcClient, err := newClient("tcp://0.0.0.0:28888")

	if err != nil {
		panic(err)
	}

	for i := 0; i < 2; i++ {
		fmt.Println(rpcClient.SendMessage(20*time.Second, "#.test:SayHello", i))
	}

	rpcServer.Close()
}

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) rpc.ActionCacheFunc {
	switch fnString {
	case "":
		return func(rt rpc.Runtime, stream *rpc.Stream, fn interface{}) int {
			if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(rpc.Runtime) rpc.Return)(rt)
				return 0
			}
		}
	default:
		return nil
	}
}

func BenchmarkClient_Debug(b *testing.B) {
	rpcServer := server.NewServer().ListenTCP("0.0.0.0:28888")
	rpcServer.AddService(
		"test",
		core.NewService().On("SayHello", func(rt core.Runtime) core.Return {
			return rt.Reply(true)
		}),
		nil,
	).SetNumOfThreads(1024).SetActionCache(&testFuncCache{})
	go func() {
		rpcServer.Serve()
	}()

	time.Sleep(time.Second)
	conn, e := net.Dial("tcp", "0.0.0.0:28888")
	if e != nil {
		panic(e)
	}
	sConn := tcp.NewTCPStreamConn(conn)

	sendStream := core.NewStream()
	sendStream.SetCallbackID(0)
	sendStream.WriteInt64(core.ControlStreamConnectRequest)
	sendStream.WriteString("")

	fmt.Println("OK")

	if err := sConn.WriteStream(sendStream, 3*time.Second); err != nil {
		panic(err)
	}

	if _, err := sConn.ReadStream(3*time.Second, 1024); err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.N = 1000000

	for i := 0; i < b.N; i++ {
		sendS, _ := core.MakeRequestStream(false, 0, "#.test:SayHello", "@")
		sendS.SetCallbackID(uint64(i) + 1)
		if e := sConn.WriteStream(sendS, time.Second); e != nil {
			fmt.Println(e)
			return
		}
		sendS.Release()

		s, e := sConn.ReadStream(time.Second, 1024)
		if e != nil {
			fmt.Println(e)
			return
		}
		s.Release()
	}

	rpcServer.Close()
	sConn.Close()
}

func BenchmarkClient_TCP(b *testing.B) {
	const sendString = "hello worldhello worldhello worldhello worldhello worldhello worldhello worldhello worldhello worldhello worldhello worldhello world"
	const writeString = "hello worldhello worldhello worldhello worldhello worldhello worldhel"

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

			conn.(*net.TCPConn).SetNoDelay(false)

			go func() {
				rPos := 0
				rBuf := make([]byte, 1024)
				wBuf := make([]byte, 1024)

				for {
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

					binary.LittleEndian.PutUint32(wBuf, uint32(len(writeString)))
					copy(wBuf[4:], writeString)
					wPos := 0
					for wPos < len(writeString)+4 {
						n, e := conn.Write(wBuf[wPos : len(writeString)+4])
						if e != nil {
							fmt.Println(e)
							return
						}
						wPos += n
					}
				}
			}()
		}
	}()

	time.Sleep(time.Second)
	conn, e := net.Dial("tcp", "0.0.0.0:28888")
	if e != nil {
		panic(e)
	}
	conn.(*net.TCPConn).SetNoDelay(false)

	rPos := 0
	rBuf := make([]byte, 1024)
	wBuf := make([]byte, 1024)

	b.ReportAllocs()
	b.ResetTimer()
	b.N = 500000

	for i := 0; i < b.N; i++ {
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

		if string(rBuf[4:streamLen+4]) != writeString {
			fmt.Println("tcp error")
			return
		}

		rPos = copy(rBuf, rBuf[streamLen+4:rPos])
	}

}

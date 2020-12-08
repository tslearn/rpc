package main

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/adapter/async"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"log"
	"net/http"
	_ "net/http/pprof"
)

type receiver struct {
}

func (p *receiver) OnConnOpen(streamConn *adapter.StreamConn) {
	fmt.Println("Server: OnConnOpen")
}

func (p *receiver) OnConnClose(streamConn *adapter.StreamConn) {
	fmt.Println("Server: OnConnClose")
}

func (p *receiver) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *core.Stream,
) {
	fmt.Println("Server: Read Stream")
	streamConn.WriteStream(stream)
	stream.Release()
}

func (p *receiver) OnConnError(
	streamConn *adapter.StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Server: OnConnError", err)
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	serverAdapter := async.NewAsyncServerAdapter(
		"tcp", "0.0.0.0:8080", 1200, 1200, &receiver{},
	)
	serverAdapter.Open()
}

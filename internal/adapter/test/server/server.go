package main

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/adapter/common"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

type receiver struct {
}

func (p *receiver) OnConnOpen(streamConn *common.StreamConn) {
	fmt.Println("Server: OnConnOpen")
}

func (p *receiver) OnConnClose(streamConn *common.StreamConn) {
	fmt.Println("Server: OnConnClose")
}

func (p *receiver) OnConnReadStream(
	streamConn *common.StreamConn,
	stream *core.Stream,
) {
	streamConn.WriteStreamAndRelease(stream)
}

func (p *receiver) OnConnError(
	streamConn *common.StreamConn,
	err *base.Error,
) {
	if streamConn != nil {
		streamConn.Close()
	}

	fmt.Println("Server: OnConnError", err)
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	tlsConfig, err := base.GetTLSServerConfig(
		"../cert/server.pem",
		"../cert/server-key.pem",
	)

	if err != nil {
		panic(err)
	}

	serverAdapter := adapter.NewServerAdapter(
		"tcp", "0.0.0.0:8080", tlsConfig, 1200, 1200, &receiver{},
	)

	if serverAdapter.Open() {
		serverAdapter.Run()
	}
}

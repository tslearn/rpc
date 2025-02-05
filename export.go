// Package rpc provides the necessary structs and functions required to use rpc
// service.
//
// The full document is located at https://github.com/rpccloud/rpc/blob/master/doc/rpc/index.md
package rpc

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/client"
	"github.com/rpccloud/rpc/internal/rpc"
	"github.com/rpccloud/rpc/internal/server"
)

// Bool ...
type Bool = rpc.Bool

// Int64 ...
type Int64 = rpc.Int64

// Uint64 ...
type Uint64 = rpc.Uint64

// Float64 ...
type Float64 = rpc.Float64

// String ...
type String = rpc.String

// Bytes ...
type Bytes = rpc.Bytes

// Array common Array type
type Array = rpc.Array

// Map common Map type
type Map = rpc.Map

// Any common Any type
type Any = rpc.Any

// RTValue ...
type RTValue = rpc.RTValue

// RTArray ...
type RTArray = rpc.RTArray

// RTMap ...
type RTMap = rpc.RTMap

// Return ...
type Return = rpc.Return

// Runtime ...
type Runtime = rpc.Runtime

// Stream ...
type Stream = rpc.Stream

// ActionCache ...
type ActionCache = rpc.ActionCache

// ActionCacheFunc ...
type ActionCacheFunc = rpc.ActionCacheFunc

// Service ...
type Service = rpc.Service

// NewService ...
func NewService() *Service {
	return rpc.NewService()
}

// Server ...
type Server = server.Server

// NewServer ...
func NewServer() *Server {
	return server.NewServer()
}

// Client ...
type Client = client.Client

// Dial ...
func Dial(network string, addr string) *Client {
	return client.Dial(network, addr)
}

// DialTLS ...
func DialTLS(network string, addr string, tlsConfig *tls.Config) *Client {
	return client.DialTLS(network, addr, tlsConfig)
}

// GetTLSServerConfig ...
func GetTLSServerConfig(certFile string, keyFile string) (*tls.Config, error) {
	return base.GetTLSServerConfig(certFile, keyFile)
}

// GetTLSClientConfig ...
func GetTLSClientConfig(
	verifyServerCert bool,
	caFiles []string,
) (*tls.Config, error) {
	return base.GetTLSClientConfig(verifyServerCert, caFiles)
}

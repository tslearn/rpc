package internal

import (
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	assert := NewRPCAssert(t)

	service := NewRPCService()
	assert(service).IsNotNil()
	assert(len(service.(*rpcService).children)).Equals(0)
	assert(len(service.(*rpcService).replies)).Equals(0)
	assert(strings.Contains(
		service.(*rpcService).debug,
		"TestNewService",
	)).IsTrue()
}

func TestRpcService_Add(t *testing.T) {
	assert := NewRPCAssert(t)
	childService := NewRPCService()
	service := NewRPCService().AddChild("user", childService)
	assert(service).IsNotNil()
	assert(len(service.(*rpcService).children)).Equals(1)
	assert(len(service.(*rpcService).replies)).Equals(0)
	assert(service.(*rpcService).children[0].name).Equals("user")
	assert(service.(*rpcService).children[0].serviceMeta).Equals(childService)
	assert(strings.Contains(
		service.(*rpcService).children[0].debug,
		"TestRpcService_Add",
	)).IsTrue()

	// add nil is ok
	assert(service.AddChild("nil", nil)).Equals(service)
}

func TestRpcService_Reply(t *testing.T) {
	assert := NewRPCAssert(t)
	service := NewRPCService().Reply("sayHello", 2345)
	assert(service).IsNotNil()
	assert(len(service.(*rpcService).children)).Equals(0)
	assert(len(service.(*rpcService).replies)).Equals(1)
	assert(service.(*rpcService).replies[0].name).Equals("sayHello")
	assert(service.(*rpcService).replies[0].handler).Equals(2345)

	assert(strings.Contains(
		service.(*rpcService).replies[0].debug,
		"TestRpcService_Reply",
	)).IsTrue()
}

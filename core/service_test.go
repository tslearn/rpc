package core

import (
	"github.com/tslearn/rpcc/util"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	assert := util.NewAssert(t)

	service := NewService()
	assert(service).IsNotNil()
	assert(len(service.(*rpcService).children)).Equals(0)
	assert(len(service.(*rpcService).echos)).Equals(0)
	assert(strings.Contains(
		service.(*rpcService).debug,
		"TestNewService",
	)).IsTrue()
}

func TestRpcService_Add(t *testing.T) {
	assert := util.NewAssert(t)
	childService := NewService()
	service := NewService().AddService("user", childService)
	assert(service).IsNotNil()
	assert(len(service.(*rpcService).children)).Equals(1)
	assert(len(service.(*rpcService).echos)).Equals(0)
	assert(service.(*rpcService).children[0].name).Equals("user")
	assert(service.(*rpcService).children[0].serviceMeta).Equals(childService)
	assert(strings.Contains(
		service.(*rpcService).children[0].debug,
		"TestRpcService_Add",
	)).IsTrue()

	// add nil is ok
	assert(service.AddService("nil", nil)).Equals(service)
}

func TestRpcService_Echo(t *testing.T) {
	assert := util.NewAssert(t)
	service := NewService().Echo("sayHello", true, 2345)
	assert(service).IsNotNil()
	assert(len(service.(*rpcService).children)).Equals(0)
	assert(len(service.(*rpcService).echos)).Equals(1)
	assert(service.(*rpcService).echos[0].name).Equals("sayHello")
	assert(service.(*rpcService).echos[0].export).Equals(true)
	assert(service.(*rpcService).echos[0].handler).Equals(2345)

	assert(strings.Contains(
		service.(*rpcService).echos[0].debug,
		"TestRpcService_Echo",
	)).IsTrue()
}

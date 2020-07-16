package internal

import (
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	assert := NewAssert(t)

	service := NewService()
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(0)
	assert(len(service.replies)).Equals(0)
	assert(strings.Contains(
		service.debug,
		"TestNewService",
	)).IsTrue()
}

func TestRpcService_Add(t *testing.T) {
	assert := NewAssert(t)
	childService := NewService()
	service := NewService().AddChild("user", childService)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(1)
	assert(len(service.replies)).Equals(0)
	assert(service.children[0].name).Equals("user")
	assert(service.children[0].service).Equals(childService)
	assert(strings.Contains(
		service.children[0].debug,
		"TestRpcService_Add",
	)).IsTrue()

	// add nil is ok
	assert(service.AddChild("nil", nil)).Equals(service)
}

func TestRpcService_Reply(t *testing.T) {
	assert := NewAssert(t)
	service := NewService().Reply("sayHello", 2345)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(0)
	assert(len(service.replies)).Equals(1)
	assert(service.replies[0].name).Equals("sayHello")
	assert(service.replies[0].handler).Equals(2345)

	assert(strings.Contains(
		service.replies[0].debug,
		"TestRpcService_Reply",
	)).IsTrue()
}

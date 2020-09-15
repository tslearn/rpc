package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func TestNewServiceMeta(t *testing.T) {
	assert := base.NewAssert(t)
	service := NewService()

	assert(NewServiceMeta("test", service, "debug", nil)).Equals(&ServiceMeta{
		name:     "test",
		service:  service,
		fileLine: "debug",
	})
}

func TestNewService(t *testing.T) {
	assert := base.NewAssert(t)

	// Test(1)
	service, fileLine := NewService(), base.GetFileLine(0)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(0)
	assert(len(service.replies)).Equals(0)
	assert(service.fileLine).Equals(fileLine)
}

func TestService_AddChildService(t *testing.T) {
	assert := base.NewAssert(t)

	// Test(1)
	child := NewService()
	service, fileLine := NewService().
		AddChildService("ch", child, nil), base.GetFileLine(0)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(1)
	assert(len(service.replies)).Equals(0)
	assert(service.children[0].name).Equals("ch")
	assert(service.children[0].service).Equals(child)
	assert(service.children[0].fileLine).Equals(fileLine)
}

func TestService_Reply(t *testing.T) {
	assert := base.NewAssert(t)

	// Test(1)
	service, fileLine := NewService().Reply("sayHello", 2345), base.GetFileLine(0)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(0)
	assert(len(service.replies)).Equals(1)
	assert(service.replies[0].name).Equals("sayHello")
	assert(service.replies[0].handler).Equals(2345)
	assert(service.replies[0].fileLine).Equals(fileLine)
}
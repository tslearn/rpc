package internal

import (
	"testing"
)

func TestNewService(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	service, fileLine := NewService(), GetFileLine(0)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(0)
	assert(len(service.replies)).Equals(0)
	assert(service.fileLine).Equals(fileLine)
}

func TestService_AddChildService(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	child := NewService()
	service, fileLine := NewService().AddChildService("ch", child), GetFileLine(0)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(1)
	assert(len(service.replies)).Equals(0)
	assert(service.children[0].name).Equals("ch")
	assert(service.children[0].service).Equals(child)
	assert(service.children[0].fileLine).Equals(fileLine)
}

func TestService_Reply(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	service, fileLine := NewService().Reply("sayHello", 2345), GetFileLine(0)
	assert(service).IsNotNil()
	assert(len(service.children)).Equals(0)
	assert(len(service.replies)).Equals(1)
	assert(service.replies[0].name).Equals("sayHello")
	assert(service.replies[0].handler).Equals(2345)
	assert(service.replies[0].fileLine).Equals(fileLine)
}

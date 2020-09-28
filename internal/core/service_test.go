package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"testing"
)

func TestNewServiceMeta(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := NewService()
		assert(NewServiceMeta("test", service, "debug", nil)).Equal(&ServiceMeta{
			name:     "test",
			service:  service,
			fileLine: "debug",
		})
	})
}

func TestNewService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service, fileLine := NewService(), base.GetFileLine(0)
		assert(service).IsNotNil()
		assert(len(service.children)).Equal(0)
		assert(len(service.replies)).Equal(0)
		assert(service.fileLine).Equal(fileLine)
	})
}

func TestService_AddChildService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		child := NewService()
		service, fileLine := NewService().
			AddChildService("ch", child, nil), base.GetFileLine(0)
		assert(service).IsNotNil()
		assert(len(service.children)).Equal(1)
		assert(len(service.replies)).Equal(0)
		assert(service.children[0].name).Equal("ch")
		assert(service.children[0].service).Equal(child)
		assert(service.children[0].fileLine).Equal(fileLine)
	})
}

func TestService_Reply(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service, fileLine := NewService().Reply("sayHello", 2345), base.GetFileLine(0)
		assert(service).IsNotNil()
		assert(len(service.children)).Equal(0)
		assert(len(service.replies)).Equal(1)
		assert(service.replies[0].name).Equal("sayHello")
		assert(service.replies[0].handler).Equal(2345)
		assert(service.replies[0].fileLine).Equal(fileLine)
	})
}

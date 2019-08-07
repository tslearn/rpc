package common

import (
	"testing"
)

var service = newServiceMeta().
	Echo("sayHello", true, func(
		ctx Context,
		name RPCString,
	) Return {
		n, _ := name.ToString()
		if n == "true" {
			return ctx.OK(true)
		} else {
			return ctx.Error()
		}
	}).
	Echo("sayGoodBye", true, func(
		ctx Context,
		name RPCString,
	) Return {
		return nil
	})

func TestRpcProcessor_AddService(t *testing.T) {
	logger := NewLogger()
	processor := newProcessor(logger, 16, 16)
	processor.AddService("user", service)
}

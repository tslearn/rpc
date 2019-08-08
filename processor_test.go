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

func getProcessor() *rpcProcessor {
	logger := NewLogger()
	return newProcessor(logger, 16, 16)
}

func TestRpcProcessor_Execute(t *testing.T) {
	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)
	processor.stop()
}

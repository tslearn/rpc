package core

import (
	"testing"
	"time"
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

func TestRPCProcessorProfile(t *testing.T) {
	rpcProcessorProfile()
}

func TestRpcProcessor_Execute(t *testing.T) {
	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)

	stream := NewRPCStream()
	processor.put(stream)

	time.Sleep(time.Second)
	processor.stop()
}

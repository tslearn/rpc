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
		_name, _ := name.ToString()
		return ctx.OK("hello " + _name)
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
	byte16 := make([]byte, 16, 16)
	stream.WriteBytes(byte16)
	stream.WriteString("$.user:sayHello")
	stream.WriteUint64(3)
	stream.WriteString("#")
	stream.WriteInt64(3)

	processor.put(stream)

	time.Sleep(10000 * time.Second)
	processor.stop()
}

package common

import (
	"github.com/rpccloud/rpc/common"
	"testing"
)

var service = newServiceMeta().
	Echo("sayHello", true, func(
		ctx Context,
		name common.RPCString,
	) Return {
		n, _ := name.ToString()
		if n == "true" {
			return ctx.OK()
		} else {
			return ctx.Error()
		}
		return nil
	}).
	Echo("sayGoodBye", true, func(
		ctx Context,
		name common.RPCString,
	) Return {
		return nil
	})

func TestRpcProcessor_AddService(t *testing.T) {
	logger := common.NewLogger()
	processor := newProcessor(logger, 16, 16)
	processor.AddService("user", service)
}

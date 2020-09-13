package system

import (
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/system/seed"
	"github.com/rpccloud/rpc/internal/core"
)

var Service = rpc.NewServiceWithOnMount(
	func(service *core.Service, data interface{}) error {
		service.AddChildService("seed", seed.Service, data)
		return nil
	},
)

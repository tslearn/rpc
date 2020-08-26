package user

import (
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/internal"
)

var Service = rpc.NewServiceWithOnMount(
	func(service *internal.Service, data interface{}) error {
		service.AddChildService("phone", phoneService, data)
		return nil
	},
)

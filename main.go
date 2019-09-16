package main

import (
	"github.com/tslearn/rpcc/core"
	"time"
)

func main() {
	server := core.NewWebSocketServer().
		AddService("user", core.NewService().
			Echo("sayHello", true, func(
				ctx core.Content,
				name string,
			) core.Return {
				time.Sleep(time.Second)
				return ctx.OK("hello " + name)
			}))

	server.SetReadTimeoutMS(800)
	server.Start("0.0.0.0", 12345, "/ws")

}

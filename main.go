package main

import "github.com/tslearn/rpcc/core"

func main() {

	server := core.NewWebSocketServer().
		AddService("user", core.NewService().
			Echo("sayHello", true, func(
				ctx core.Content,
				name string,
			) core.Return {
				return ctx.OK("hello " + name)
			}))

	server.Start("0.0.0.0", 12345, "/ws")

}

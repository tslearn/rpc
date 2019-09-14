package main

import (
	"fmt"
	"github.com/tslearn/rpcc/core"
	"sync/atomic"
	"time"
)

func main() {

	count := uint64(0)

	server := core.NewWebSocketServer().
		AddService("user", core.NewService().
			Echo("sayHello", true, func(
				ctx core.Content,
				name string,
			) core.Return {
				fmt.Println(atomic.AddUint64(&count, 1))
				time.Sleep(time.Second)
				return ctx.OK("hello " + name)
			}))

	server.SetReadTimeoutMS(500)
	server.Start("0.0.0.0", 12345, "/ws")

}

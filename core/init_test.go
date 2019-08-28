package core

//
//func TestMain(m *testing.M) {
//	server := NewWebSocketServer().
//		AddService("user", newServiceMeta().
//			Echo("sayHello", true, func(
//				ctx *rpcContext,
//				name string,
//			) *rpcReturn {
//				return ctx.OK("hello " + name)
//			}))
//	server.StartBackground("0.0.0.0", 12345, "/ws")
//	time.Sleep(100 * time.Millisecond)
//
//	// call flag.Parse() here if TestMain uses flags
//	ret := m.Run()
//
//	_ = server.Close()
//	os.Exit(ret)
//}

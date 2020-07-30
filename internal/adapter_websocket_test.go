package internal

//
//func TestNewWebSocketServerEndPoint(t *testing.T) {
//	assert := internal.NewAssert(t)
//
//	server := NewWebSocketServerAdapter("127.0.0.1:20080", "/test")
//
//	for i := 0; i < 2; i++ {
//		assert(
//			server.Open(func(conn IStreamConn) {
//				fmt.Println(conn)
//			}, func(err Error) {
//				fmt.Println(err)
//			}),
//		).IsTrue()
//
//		time.Sleep(2 * time.Second)
//
//		assert(
//			server.Close(func(err Error) {
//				fmt.Println(err)
//			}),
//		).IsTrue()
//	}
//}
//
//func TestNewWebSocketClientEndPoint(t *testing.T) {
//	assert := internal.NewAssert(t)
//	server := NewWebSocketServerAdapter("127.0.0.1:20080", "/test")
//	server.Open(func(conn IStreamConn) {
//		time.Sleep(3 * time.Second)
//	}, func(err Error) {
//		fmt.Println(err)
//	})
//
//	client := NewWebSocketClientEndPoint("ws://127.0.0.1:20080/test")
//	assert(
//		client.Open(func(conn IStreamConn) {
//			time.Sleep(3 * time.Second)
//		}, func(err Error) {
//			fmt.Println(err)
//		}),
//	).IsTrue()
//
//	time.Sleep(1 * time.Second)
//
//	assert(
//		client.Close(func(err Error) {
//			fmt.Println(err)
//		}),
//	).IsTrue()
//
//	server.Close(func(err Error) {
//		fmt.Println(err)
//	})
//}

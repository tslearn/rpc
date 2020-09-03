package internal

//
//func TestNewRPCLock(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewLock()
//	assert(o1).IsNotNil()
//}
//
//func TestRPCLock_DoWithLock(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewLock()
//	sum := 0
//	waits := make(chan bool)
//	for i := 0; i < 100; i++ {
//		go func() {
//			for n := 0; n < 1000; n++ {
//				o1.DoWithLock(func() {
//					sum += n
//				})
//			}
//			waits <- true
//		}()
//	}
//	for i := 0; i < 100; i++ {
//		<-waits
//	}
//	assert(sum).Equals(49950000)
//}
//
//func TestRPCLock_CallWithLock(t *testing.T) {
//	assert := NewAssert(t)
//
//	// Test(1)
//	o1 := NewLock()
//	sum := 0
//	waits := make(chan bool)
//	for i := 0; i < 100; i++ {
//		go func() {
//			for n := 0; n < 1000; n++ {
//				assert(o1.CallWithLock(func() interface{} {
//					sum += n
//					return true
//				})).Equals(true)
//			}
//			waits <- true
//		}()
//	}
//	for i := 0; i < 100; i++ {
//		<-waits
//	}
//	assert(sum).Equals(49950000)
//}

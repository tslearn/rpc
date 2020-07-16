package internal

import "testing"

func TestNewRPCLock(t *testing.T) {
	assert := NewAssert(t)
	assert(NewRPCLock()).IsNotNil()
}

func TestRPCLock_DoWithLock(t *testing.T) {
	assert := NewAssert(t)
	locker := NewRPCLock()
	waits := make(chan bool)
	sum := 0

	for i := 0; i < 100; i++ {
		go func() {
			for n := 0; n < 1000; n++ {
				locker.DoWithLock(func() {
					sum += n
				})
			}
			waits <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-waits
	}

	assert(sum).Equals(49950000)
}

func TestRPCLock_CallWithLock(t *testing.T) {
	assert := NewAssert(t)
	locker := NewRPCLock()
	waits := make(chan bool)
	sum := 0

	for i := 0; i < 100; i++ {
		go func() {
			for n := 0; n < 1000; n++ {
				assert(locker.CallWithLock(func() interface{} {
					sum += n
					return true
				})).Equals(true)
			}
			waits <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-waits
	}

	assert(sum).Equals(49950000)
}

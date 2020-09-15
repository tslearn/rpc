package base

import (
	"testing"
	"time"
)

func TestStatusManager_SetRunning(t *testing.T) {
	assert := NewAssert(t)

	assert(StatusManager{}.status).Equals(statusManagerClosed)
	assert(StatusManager{}.closeCH).IsNil()

	// Test(1)
	mgr1 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	assert(mgr1.SetRunning(nil)).IsTrue()
	assert(mgr1.status).Equals(statusManagerRunning)
	assert(mgr1.closeCH).IsNil()

	// Test(2)
	mgr2 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	ret2 := make(chan bool, 1)
	assert(mgr2.SetRunning(func() {
		ret2 <- true
	})).IsTrue()
	assert(mgr2.status).Equals(statusManagerRunning)
	assert(mgr2.closeCH).IsNil()
	assert(<-ret2).IsTrue()

	// Test(3)
	mgr3 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	assert(mgr3.SetRunning(nil)).IsFalse()
	assert(mgr3.status).Equals(statusManagerRunning)
	assert(mgr3.closeCH).IsNil()

	// Test(4)
	mgr4 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	assert(mgr4.SetRunning(func() {
		assert().Fail("this code should not eval")
	})).IsFalse()
	assert(mgr4.status).Equals(statusManagerRunning)
	assert(mgr4.closeCH).IsNil()

	// Test(5)
	mgr5 := &StatusManager{status: statusManagerClosing, closeCH: nil}
	assert(mgr5.SetRunning(nil)).IsFalse()
	assert(mgr5.status).Equals(statusManagerClosing)
	assert(mgr5.closeCH).IsNil()

	// Test(6)
	mgr6 := &StatusManager{status: statusManagerClosing, closeCH: nil}
	assert(mgr6.SetRunning(func() {
		assert().Fail("this code should not eval")
	})).IsFalse()
	assert(mgr6.status).Equals(statusManagerClosing)
	assert(mgr6.closeCH).IsNil()
}

func TestStatusManager_SetClosing(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	mgr1 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	assert(mgr1.SetClosing(nil)).IsTrue()
	assert(mgr1.status).Equals(statusManagerClosing)
	assert(mgr1.closeCH).IsNotNil()

	// Test(2)
	mgr2 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	ret2 := make(chan bool, 1)
	assert(mgr2.SetClosing(func(closeCH chan bool) {
		select {
		case <-closeCH:
			assert().Fail("this code should not eval")
		case <-time.After(time.Second):
			ret2 <- true
		}
	})).IsTrue()
	assert(mgr2.status).Equals(statusManagerClosing)
	assert(mgr2.closeCH).IsNotNil()
	assert(<-ret2).IsTrue()

	// Test(3)
	mgr3 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	ret3 := chan bool(nil)
	assert(mgr3.SetClosing(func(closeCH chan bool) {
		ret3 = closeCH
	})).IsTrue()
	assert(mgr3.status).Equals(statusManagerClosing)
	assert(mgr3.closeCH).IsNotNil()
	assert(mgr3.SetClosed(nil)).IsTrue()
	assert(mgr3.closeCH).IsNil()
	assert(<-ret3).IsTrue()

	// Test(4)
	mgr4 := &StatusManager{status: statusManagerClosing, closeCH: nil}
	assert(mgr4.SetClosing(nil)).IsFalse()
	assert(mgr4.status).Equals(statusManagerClosing)
	assert(mgr4.closeCH).IsNil()

	// Test(5)
	mgr5 := &StatusManager{status: statusManagerClosing, closeCH: nil}
	assert(mgr5.SetClosing(func(_ chan bool) {
		assert().Fail("this code should not eval")
	})).IsFalse()
	assert(mgr5.status).Equals(statusManagerClosing)
	assert(mgr5.closeCH).IsNil()

	// Test(6)
	mgr6 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	assert(mgr6.SetClosing(nil)).IsFalse()
	assert(mgr6.status).Equals(statusManagerClosed)
	assert(mgr6.closeCH).IsNil()

	// Test(7)
	mgr7 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	assert(mgr7.SetClosing(func(_ chan bool) {
		assert().Fail("this code should not eval")
	})).IsFalse()
	assert(mgr7.status).Equals(statusManagerClosed)
	assert(mgr7.closeCH).IsNil()
}

func TestStatusManager_SetClosed(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	mgr1 := &StatusManager{
		status:  statusManagerClosing,
		closeCH: make(chan bool, 1),
	}
	assert(mgr1.SetClosed(nil)).IsTrue()
	assert(mgr1.status).Equals(statusManagerClosed)
	assert(mgr1.closeCH).IsNil()

	// Test(2)
	mgr2 := &StatusManager{
		status:  statusManagerClosing,
		closeCH: make(chan bool, 1),
	}
	ret2 := make(chan bool, 1)
	assert(mgr2.SetClosed(func() {
		ret2 <- true
	})).IsTrue()
	assert(mgr2.status).Equals(statusManagerClosed)
	assert(mgr2.closeCH).IsNil()
	assert(<-ret2).IsTrue()

	// Test(3)
	mgr3 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	assert(mgr3.SetClosed(nil)).IsFalse()
	assert(mgr3.status).Equals(statusManagerClosed)
	assert(mgr3.closeCH).IsNil()

	// Test(4)
	mgr4 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	assert(mgr4.SetClosed(func() {
		assert().Fail("this code should not eval")
	})).IsFalse()
	assert(mgr4.status).Equals(statusManagerClosed)
	assert(mgr4.closeCH).IsNil()

	// Test(5)
	mgr5 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	assert(mgr5.SetClosed(nil)).IsFalse()
	assert(mgr5.status).Equals(statusManagerRunning)
	assert(mgr5.closeCH).IsNil()

	// Test(6)
	mgr6 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	assert(mgr6.SetClosed(func() {
		assert().Fail("this code should not eval")
	})).IsFalse()
	assert(mgr6.status).Equals(statusManagerRunning)
	assert(mgr6.closeCH).IsNil()
}

func TestStatusManager_IsRunning(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	mgr1 := &StatusManager{status: statusManagerClosed, closeCH: nil}
	assert(mgr1.IsRunning()).IsFalse()

	// Test(2)
	mgr2 := &StatusManager{status: statusManagerRunning, closeCH: nil}
	assert(mgr2.IsRunning()).IsTrue()

	// Test(3)
	mgr3 := &StatusManager{status: statusManagerClosing, closeCH: nil}
	assert(mgr3.IsRunning()).IsFalse()
}

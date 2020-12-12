package base

import (
	"testing"
	"time"
)

func TestStatusManager(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(statusManagerClosed).Equal(int32(0))
		assert(statusManagerRunning).Equal(int32(1))
		assert(statusManagerClosing).Equal(int32(2))

		assert(StatusManager{}.status).Equal(statusManagerClosed)
		assert(StatusManager{}.closeCH).Equal(nil)
	})
}

func TestStatusManager_SetRunning(t *testing.T) {
	t.Run("statusManagerClosed to statusManagerRunning", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		assert(v1.SetRunning(nil)).IsTrue()
		assert(v1.status, v1.closeCH).Equal(statusManagerRunning, nil)

		v2 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		r2 := make(chan bool, 1)
		assert(v2.SetRunning(func() { r2 <- true })).IsTrue()
		assert(v2.status, v2.closeCH).Equal(statusManagerRunning, nil)
		assert(<-r2).IsTrue()
	})

	t.Run("statusManagerRunning to statusManagerRunning", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		assert(v1.SetRunning(nil)).IsFalse()
		assert(v1.status, v1.closeCH).Equal(statusManagerRunning, nil)

		v2 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		assert(v2.SetRunning(func() { panic("error") })).IsFalse()
		assert(v2.status, v2.closeCH).Equal(statusManagerRunning, nil)
	})

	t.Run("statusManagerClosing to statusManagerRunning", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &StatusManager{status: statusManagerClosing, closeCH: nil}
		assert(v1.SetRunning(nil)).IsFalse()
		assert(v1.status, v1.closeCH).Equal(statusManagerClosing, nil)

		v2 := &StatusManager{status: statusManagerClosing, closeCH: nil}
		assert(v2.SetRunning(func() { panic("error") })).IsFalse()
		assert(v2.status, v2.closeCH).Equal(statusManagerClosing, nil)
	})
}

func TestStatusManager_SetClosing(t *testing.T) {
	t.Run("statusManagerRunning to statusManagerClosing", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		assert(v1.SetClosing(nil)).IsTrue()
		assert(v1.status).Equal(statusManagerClosing)
		assert(v1.closeCH).IsNotNil()

		v2 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		r2 := make(chan bool, 1)
		assert(v2.SetClosing(func(closeCH chan bool) {
			select {
			case <-closeCH:
				panic("error")
			case <-time.After(50 * time.Millisecond):
				r2 <- true
			}
		})).IsTrue()
		assert(v2.status).Equal(statusManagerClosing)
		assert(v2.closeCH).IsNotNil()
		assert(<-r2).IsTrue()

		v3 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		r3 := make(chan bool, 1)
		assert(v3.SetClosing(func(closeCH chan bool) {
			go func() {
				select {
				case <-closeCH:
					r3 <- true
				}
			}()
		})).IsTrue()
		v3.SetClosed(nil)
		assert(<-r3).IsTrue()
	})

	t.Run("statusManagerClosing to statusManagerClosing", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{status: statusManagerClosing, closeCH: nil}
		assert(v1.SetClosing(nil)).IsFalse()
		assert(v1.status, v1.closeCH).Equal(statusManagerClosing, nil)

		v2 := &StatusManager{status: statusManagerClosing, closeCH: nil}
		assert(v2.SetClosing(func(_ chan bool) { panic("error") })).IsFalse()
		assert(v2.status, v2.closeCH).Equal(statusManagerClosing, nil)
	})

	t.Run("statusManagerClosed to statusManagerClosing", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		assert(v1.SetClosing(nil)).IsFalse()
		assert(v1.status, v1.closeCH).Equal(statusManagerClosed, nil)

		// Test(7)
		v2 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		assert(v2.SetClosing(func(_ chan bool) { panic("error") })).IsFalse()
		assert(v2.status, v2.closeCH).Equal(statusManagerClosed, nil)
	})
}

func TestStatusManager_SetClosed(t *testing.T) {
	t.Run("statusManagerClosing to statusManagerClosed", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{
			status:  statusManagerClosing,
			closeCH: make(chan bool, 1),
		}
		assert(v1.SetClosed(nil)).IsTrue()
		assert(v1.status, v1.closeCH).Equal(statusManagerClosed, nil)

		v2 := &StatusManager{
			status:  statusManagerClosing,
			closeCH: make(chan bool, 1),
		}
		r2 := make(chan bool, 1)
		assert(v2.SetClosed(func() { r2 <- true })).IsTrue()
		assert(v2.status, v2.closeCH).Equal(statusManagerClosed, nil)
		assert(<-r2).IsTrue()
	})

	t.Run("statusManagerClosed to statusManagerClosed", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		assert(v1.SetClosed(nil)).IsFalse()
		assert(v1.status, v1.closeCH).Equal(statusManagerClosed, nil)

		v2 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		assert(v2.SetClosed(func() { panic("error") })).IsFalse()
		assert(v2.status, v2.closeCH).Equal(statusManagerClosed, nil)
	})

	t.Run("statusManagerRunning to statusManagerClosed", func(t *testing.T) {
		assert := NewAssert(t)

		// Test(5)
		v1 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		assert(v1.SetClosed(nil)).IsFalse()
		assert(v1.status, v1.closeCH).Equal(statusManagerRunning, nil)

		// Test(6)
		v2 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		assert(v2.SetClosed(func() {
			assert().Fail("this code should not eval")
		})).IsFalse()
		assert(v2.status, v2.closeCH).Equal(statusManagerRunning, nil)
	})
}

func TestStatusManager_IsRunning(t *testing.T) {
	t.Run("test status statusManagerClosed", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &StatusManager{status: statusManagerClosed, closeCH: nil}
		assert(v1.IsRunning()).IsFalse()
	})

	t.Run("test status statusManagerRunning", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &StatusManager{status: statusManagerRunning, closeCH: nil}
		assert(v1.IsRunning()).IsTrue()
	})

	t.Run("test status statusManagerClosing", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &StatusManager{status: statusManagerClosing, closeCH: nil}
		assert(v1.IsRunning()).IsFalse()
	})
}

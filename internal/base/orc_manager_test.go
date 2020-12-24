package base

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestORCManagerBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(orcBitLock).Equal(256)
		assert(orcStatusClosed).Equal(int32(0))
		assert(orcStatusClosing).Equal(int32(1))
		assert(orcStatusReady).Equal(int32(2))
	})
}

func TestORCManager_NewORCManager(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		assert(o).IsNotNil()
		assert(o.status).Equal(orcStatusClosed)
	})
}

func TestORCManager_isRunning(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusClosed
		assert(o.isRunning()).IsFalse()
		o.status = orcBitLock | orcStatusClosed
		assert(o.isRunning()).IsFalse()
		o.status = orcStatusReady
		assert(o.isRunning()).IsTrue()
		o.status = orcBitLock | orcStatusReady
		assert(o.isRunning()).IsTrue()
	})
}

func TestORCManager_Open(t *testing.T) {
	t.Run("nil func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		assert(o.Open(nil)).IsFalse()
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("panic in func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		assert(o.Open(func() bool {
			panic("error")
		})).IsFalse()
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("status orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusReady
		assert(o.Open(func() bool {
			panic("illegal call here")
		})).IsFalse()
		assert(o.status).Equal(orcStatusReady)
	})

	t.Run("status orcBitLock | orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcBitLock | orcStatusReady
		assert(o.Open(func() bool {
			panic("illegal call here")
		})).IsFalse()
		assert(o.status).Equal(orcBitLock | orcStatusReady)
	})

	t.Run("status orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusClosed
		callCount := 0
		assert(o.Open(func() bool {
			callCount++
			return true
		})).IsTrue()
		assert(callCount).Equal(1)
		assert(o.status).Equal(orcStatusReady)
		o.status = orcStatusClosed
		assert(o.Open(func() bool {
			callCount++
			return false
		})).IsFalse()
		assert(callCount).Equal(2)
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("status orcBitLock | orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcBitLock | orcStatusClosing

		go func() {
			time.Sleep(300 * time.Millisecond)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(orcStatusClosed)
		}()

		assert(o.Open(func() bool {
			return true
		})).IsTrue()

		assert(o.status).Equal(orcStatusReady)
	})
}

func TestORCManager_Run(t *testing.T) {
	t.Run("nil func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusReady
		assert(o.Run(nil)).IsTrue()
		assert(o.status).Equal(orcStatusReady)
	})

	t.Run("status orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusClosed
		assert(o.Run(func(isRunning func() bool) {
			panic("illegal call here")
		})).IsFalse()
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("status orcBitLock | orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcBitLock | orcStatusClosed
		assert(o.Run(func(isRunning func() bool) {
			panic("illegal call here")
		})).IsFalse()
		assert(o.status).Equal(orcBitLock | orcStatusClosed)
	})

	t.Run("status orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusReady
		callCount := 0
		assert(o.Run(func(isRunning func() bool) {
			callCount++
		})).IsTrue()
		assert(callCount).Equal(1)
		assert(o.status).Equal(orcStatusReady)
	})

	t.Run("status orcBitLock | orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcBitLock | orcStatusReady

		go func() {
			time.Sleep(300 * time.Millisecond)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(o.status & 0xFF)
		}()

		assert(o.Run(func(isRunning func() bool) {})).IsTrue()
		assert(o.status).Equal(orcStatusReady)
	})
}

func TestORCManager_Close(t *testing.T) {
	t.Run("nil func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusReady
		assert(o.Close(nil, nil)).IsTrue()
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("status orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusClosed
		assert(o.Close(func() {
			panic("illegal call here")
		}, func() {
			panic("illegal call here")
		})).IsFalse()
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("status orcBitLock | orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcBitLock | orcStatusClosed
		assert(o.Close(func() {
			panic("illegal call here")
		}, func() {
			panic("illegal call here")
		})).IsFalse()
		assert(o.status).Equal(orcBitLock | orcStatusClosed)
	})

	t.Run("status orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcStatusReady
		callCount := 0
		assert(o.Close(func() {
			callCount++
		}, func() {
			callCount++
		})).IsTrue()
		assert(callCount).Equal(2)
		assert(o.status).Equal(orcStatusClosed)
	})

	t.Run("status orcBitLock | orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.status = orcBitLock | orcStatusReady

		go func() {
			time.Sleep(300 * time.Millisecond)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(o.status & 0xFF)
		}()

		assert(o.Close(nil, nil)).IsTrue()
		assert(o.status).Equal(orcStatusClosed)
	})
}

func TestORCManagerParallels(t *testing.T) {
	t.Run("test open and close", func(t *testing.T) {
		fnTest := func() (int64, int64, int64) {
			waitCH := make(chan bool)
			o := NewORCManager()

			testCount := 10000
			parallels := 4

			openOK := int64(0)
			runOK := int64(0)
			closeOK := int64(0)

			for n := 0; n < parallels; n++ {
				go func() {
					for i := 0; i < testCount; i++ {
						if o.Open(func() bool {
							time.Sleep(100 * time.Microsecond)
							return true
						}) {
							atomic.AddInt64(&openOK, 1)
						} else {
							time.Sleep(100 * time.Microsecond)
						}
					}
					waitCH <- true
				}()

				go func() {
					for i := 0; i < testCount; i++ {
						if o.Run(func(isRunning func() bool) {
							time.Sleep(100 * time.Microsecond)
						}) {
							atomic.AddInt64(&runOK, 1)
						} else {
							time.Sleep(100 * time.Microsecond)
						}
					}
					waitCH <- true
				}()

				go func() {
					for i := 0; i < testCount; i++ {
						if o.Close(func() {
							time.Sleep(100 * time.Microsecond)
						}, nil) {
							atomic.AddInt64(&closeOK, 1)
						} else {
							time.Sleep(100 * time.Microsecond)
						}
					}
					waitCH <- true
				}()
			}

			for n := 0; n < parallels; n++ {
				<-waitCH
				<-waitCH
				<-waitCH
			}

			if o.Close(nil, nil) {
				closeOK++
			}

			return openOK, runOK, closeOK
		}

		waitCH := make(chan bool)
		for i := 0; i < 30; i++ {
			go func() {
				assert := NewAssert(t)
				openOK, runOK, closeOK := fnTest()
				assert(openOK).Equal(closeOK)
				assert(runOK > 0).Equal(true)
				waitCH <- true
			}()
		}

		for i := 0; i < 30; i++ {
			<-waitCH
		}
	})
}

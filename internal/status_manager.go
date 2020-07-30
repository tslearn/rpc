package internal

import "sync/atomic"

const statusManagerClosed = int32(0)
const statusManagerRunning = int32(1)
const statusManagerClosing = int32(2)

type StatusManager struct {
	status  int32
	closeCH chan struct{}
	lock    Lock
}

func (p *StatusManager) SetRunning(onSuccess func()) bool {
	return p.lock.CallWithLock(func() interface{} {
		if atomic.CompareAndSwapInt32(
			&p.status,
			statusManagerClosed,
			statusManagerRunning,
		) {
			if onSuccess != nil {
				onSuccess()
			}
			return true
		}

		return false
	}).(bool)
}

func (p *StatusManager) SetClosing(onSuccess func(ch chan struct{})) bool {
	return p.lock.CallWithLock(func() interface{} {
		if atomic.CompareAndSwapInt32(
			&p.status,
			statusManagerRunning,
			statusManagerClosing,
		) {
			p.closeCH = make(chan struct{}, 1)
			if onSuccess != nil {
				onSuccess(p.closeCH)
			}
			return true
		}

		return false
	}).(bool)
}

func (p *StatusManager) SetClosed(onSuccess func()) bool {
	return p.lock.CallWithLock(func() interface{} {
		if atomic.CompareAndSwapInt32(
			&p.status,
			statusManagerClosing,
			statusManagerClosed,
		) {
			if onSuccess != nil {
				onSuccess()
			}
			p.closeCH <- struct{}{}
			close(p.closeCH)
			p.closeCH = nil
			return true
		}

		return false
	}).(bool)
}

func (p *StatusManager) IsRunning() bool {
	return atomic.CompareAndSwapInt32(
		&p.status,
		statusManagerRunning,
		statusManagerRunning,
	)
}

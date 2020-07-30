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

func (p *StatusManager) SetRunning(onSuccess func(), onFail func()) bool {
	return p.lock.CallWithLock(func() interface{} {
		ret := atomic.CompareAndSwapInt32(
			&p.status,
			statusManagerClosed,
			statusManagerRunning,
		)

		if ret {
			if onSuccess != nil {
				onSuccess()
			}
		} else {
			if onFail != nil {
				onFail()
			}
		}

		return ret
	}).(bool)
}

func (p *StatusManager) SetClosing(
	onSuccess func(ch chan struct{}),
	onFail func(),
) bool {
	return p.lock.CallWithLock(func() interface{} {
		ret := atomic.CompareAndSwapInt32(
			&p.status,
			statusManagerRunning,
			statusManagerClosing,
		)

		if ret {
			p.closeCH = make(chan struct{}, 1)
			if onSuccess != nil {
				onSuccess(p.closeCH)
			}
		} else {
			if onFail != nil {
				onFail()
			}
		}

		return ret
	}).(bool)
}

func (p *StatusManager) SetClosed(onSuccess func(), onFail func()) bool {
	return p.lock.CallWithLock(func() interface{} {
		ret := atomic.CompareAndSwapInt32(
			&p.status,
			statusManagerClosing,
			statusManagerClosed,
		)

		if ret {
			if onSuccess != nil {
				onSuccess()
			}
			p.closeCH <- struct{}{}
			close(p.closeCH)
			p.closeCH = nil
		} else {
			if onFail != nil {
				onFail()
			}
		}

		return ret
	}).(bool)
}

func (p *StatusManager) IsRunning() bool {
	return atomic.CompareAndSwapInt32(
		&p.status,
		statusManagerRunning,
		statusManagerRunning,
	)
}

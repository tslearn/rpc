package base

import "sync"

var (
	gPanicMutex         = &sync.Mutex{}
	gPanicSubscriptions = make([]*PanicSubscription, 0)
)

type PanicSubscription struct {
	id      int64
	onPanic func(err *Error)
}

func (p *PanicSubscription) Close() bool {
	if p == nil {
		return false
	}

	gPanicMutex.Lock()
	defer gPanicMutex.Unlock()

	for i := 0; i < len(gPanicSubscriptions); i++ {
		if gPanicSubscriptions[i].id == p.id {
			gPanicSubscriptions = append(
				gPanicSubscriptions[:i],
				gPanicSubscriptions[i+1:]...,
			)
			return true
		}
	}
	return false
}

// SubscribePanic ...
func SubscribePanic(onPanic func(*Error)) *PanicSubscription {
	if onPanic == nil {
		return nil
	}

	gPanicMutex.Lock()
	defer gPanicMutex.Unlock()

	ret := &PanicSubscription{
		id:      GetSeed(),
		onPanic: onPanic,
	}
	gPanicSubscriptions = append(gPanicSubscriptions, ret)
	return ret
}

// PublishPanic ...
func PublishPanic(err *Error) {
	defer func() {
		_ = recover()
	}()

	gPanicMutex.Lock()
	defer gPanicMutex.Unlock()

	for _, sub := range gPanicSubscriptions {
		if sub != nil && sub.onPanic != nil {
			sub.onPanic(err)
		}
	}
}

func RunWithCatchPanic(fn func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	fn()
	return
}

func RunWithSubscribePanic(fn func()) *Error {
	ch := make(chan *Error, 1)
	sub := SubscribePanic(func(err *Error) {
		ch <- err
	})
	defer sub.Close()

	fn()

	select {
	case err := <-ch:
		return err
	default:
		return nil
	}
}

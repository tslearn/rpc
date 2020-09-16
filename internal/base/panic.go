package base

import "sync"

var (
	gPanicMutex         = &sync.Mutex{}
	gPanicSubscriptions = make([]*PanicSubscription, 0)
)

// ReportPanic ...
func ReportPanic(v interface{}) {
	defer func() {
		_ = recover()
	}()

	gPanicMutex.Lock()
	defer gPanicMutex.Unlock()

	for _, sub := range gPanicSubscriptions {
		if sub != nil && sub.onPanic != nil {
			sub.onPanic(v)
		}
	}
}

// SubscribePanic ...
func SubscribePanic(onPanic func(interface{})) *PanicSubscription {
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

func TestRunWithCatchPanic(fn func()) (ret interface{}) {
	defer func() {
		ret = recover()
	}()

	fn()
	return
}

func TestRunWithSubscribePanic(fn func()) interface{} {
	ch := make(chan interface{}, 1)
	sub := SubscribePanic(func(v interface{}) {
		ch <- v
	})
	defer sub.Close()

	fn()

	select {
	case v := <-ch:
		return v
	default:
		return nil
	}
}

type PanicSubscription struct {
	id      int64
	onPanic func(v interface{})
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

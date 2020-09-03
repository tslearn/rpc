package internal

//
//// Lock ...
//type Lock struct {
//	mutex sync.Mutex
//}
//
//// NewLock ...
//func NewLock() *Lock {
//	return &Lock{}
//}
//
//// DoWithLock ...
//func (p *Lock) DoWithLock(fn func()) {
//	p.mutex.Lock()
//	defer p.mutex.Unlock()
//	fn()
//}
//
//// CallWithLock ...
//func (p *Lock) CallWithLock(fn func() interface{}) interface{} {
//	p.mutex.Lock()
//	defer p.mutex.Unlock()
//	return fn()
//}

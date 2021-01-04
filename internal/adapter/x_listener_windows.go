package adapter

// TCPListener ...
type XListener struct {
}

// NewXListener ...
func NewXListener(
	network string,
	addr string,
	onAccept func(fd int, localAddr net.Addr, remoteAddr net.Addr),
	onError func(err *base.Error),
) *TCPListener {
	panic("it does not support windows platform")
}

// Close ...
func (p *XListener) Close() {
	if atomic.CompareAndSwapInt32(
		&p.status,
		tcpListenerStatusRunning,
		tcpListenerStatusClosed,
	) {
		_ = unix.Close(p.lnFD)
		p.poller.Close()
	}
}

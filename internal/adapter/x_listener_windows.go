package adapter

// TCPListener ...
type TCPListener struct {
}

// NewTCPListener ...
func NewTCPListener(
	network string,
	addr string,
	onAccept func(fd int, localAddr net.Addr, remoteAddr net.Addr),
	onError func(err *base.Error),
) *TCPListener {
	panic("it does not support windows platform")
}

// OnRead ...
func (p *TCPListener) OnRead(fd int) {
	if fd == p.lnFD {
		if connFD, sa, e := unix.Accept(p.lnFD); e != nil {
			if e != unix.EAGAIN {
				p.onError(errors.ErrTCPListener.AddDebug(e.Error()))
			}
		} else {
			if remoteAddr := sockAddrToTCPAddr(sa); remoteAddr != nil {
				p.onAccept(connFD, p.lnAddr, remoteAddr)
			} else {
				_ = unix.Close(connFD)
			}
		}
	}
}

// Close ...
func (p *TCPListener) Close() {
	if atomic.CompareAndSwapInt32(
		&p.status,
		tcpListenerStatusRunning,
		tcpListenerStatusClosed,
	) {
		_ = unix.Close(p.lnFD)
		p.poller.Close()
	}
}

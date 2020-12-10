package async

import (
	"golang.org/x/sys/unix"
	"os"
	"unsafe"
)

// Poller represents a poller which is in charge of monitoring file-descriptors.
type Poller struct {
	fd     int    // epoll fd
	wfd    int    // wake fd
	wfdBuf []byte // wfd buffer to read packet
}

// OpenPoller instantiates a poller.
func OpenPoller() (poller *Poller, err error) {
	poller = new(Poller)
	if poller.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
		poller = nil
		err = os.NewSyscallError("epoll_create1", err)
		return
	}
	if poller.wfd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC); err != nil {
		_ = poller.Close()
		poller = nil
		err = os.NewSyscallError("eventfd", err)
		return
	}
	poller.wfdBuf = make([]byte, 8)
	if err = poller.AddRead(poller.wfd); err != nil {
		_ = poller.Close()
		poller = nil
		return
	}
	poller.asyncJobQueue = internal.NewAsyncJobQueue()
	return
}

// Close closes the poller.
func (p *Poller) Close() error {
	if err := os.NewSyscallError("close", unix.Close(p.fd)); err != nil {
		return err
	}
	return os.NewSyscallError("close", unix.Close(p.wfd))
}

// Make the endianness of bytes compatible with more linux OSs under different processor-architectures,
// according to http://man7.org/linux/man-pages/man2/eventfd.2.html.
var (
	u uint64 = 1
	b        = (*(*[8]byte)(unsafe.Pointer(&u)))[:]
)

// Trigger wakes up the poller blocked in waiting for network-events and runs jobs in asyncJobQueue.
func (p *Poller) Trigger(job internal.Job) (err error) {
	if p.asyncJobQueue.Push(job) == 1 {
		_, err = unix.Write(p.wfd, b)
	}
	return os.NewSyscallError("write", err)
}

// Polling blocks the current goroutine, waiting for network-events.
func (p *Poller) Polling(callback func(fd int, ev uint32) error) error {
	el := newEventList(InitEvents)
	var wakenUp bool

	for {
		n, err := unix.EpollWait(p.fd, el.events, -1)
		if err != nil && err != unix.EINTR {
			logging.DefaultLogger.Warnf("Error occurs in epoll: %v", os.NewSyscallError("epoll_wait", err))
			continue
		}

		for i := 0; i < n; i++ {
			if fd := int(el.events[i].Fd); fd != p.wfd {
				switch err = callback(fd, el.events[i].Events); err {
				case nil:
				case errors.ErrAcceptSocket, errors.ErrServerShutdown:
					return err
				default:
					logging.DefaultLogger.Warnf("Error occurs in event-loop: %v", err)
				}
			} else {
				wakenUp = true
				_, _ = unix.Read(p.wfd, p.wfdBuf)
			}
		}

		if wakenUp {
			wakenUp = false
			switch err = p.asyncJobQueue.ForEach(); err {
			case nil:
			case errors.ErrServerShutdown:
				return err
			default:
				logging.DefaultLogger.Warnf("Error occurs in user-defined function, %v", err)
			}
		}

		if n == el.size {
			el.increase()
		}
	}
}

const (
	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents     = unix.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
)

// AddReadWrite registers the given file-descriptor with readable and writable events to the poller.
func (p *Poller) AddReadWrite(fd int) error {
	return os.NewSyscallError("epoll_ctl add",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readWriteEvents}))
}

// AddRead registers the given file-descriptor with readable event to the poller.
func (p *Poller) AddRead(fd int) error {
	return os.NewSyscallError("epoll_ctl add",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readEvents}))
}

// AddWrite registers the given file-descriptor with writable event to the poller.
func (p *Poller) AddWrite(fd int) error {
	return os.NewSyscallError("epoll_ctl add",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Fd: int32(fd), Events: writeEvents}))
}

// ModRead renews the given file-descriptor with readable event in the poller.
func (p *Poller) ModRead(fd int) error {
	return os.NewSyscallError("epoll_ctl mod",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readEvents}))
}

// ModReadWrite renews the given file-descriptor with readable and writable events in the poller.
func (p *Poller) ModReadWrite(fd int) error {
	return os.NewSyscallError("epoll_ctl mod",
		unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readWriteEvents}))
}

// Delete removes the given file-descriptor from the poller.
func (p *Poller) Delete(fd int) error {
	return os.NewSyscallError("epoll_ctl del", unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd, nil))
}

// +build linux

package async

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"os"
	"unsafe"
)

const triggerDataAddConn = 1
const triggerDataExit = 2

const pollerStatusRunning = 1
const pollerStatusClosing = 2
const pollerStatusClosed = 0

const (
	ErrEvents = unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
	OutEvents = ErrEvents | unix.EPOLLOUT
	InEvents  = ErrEvents | unix.EPOLLIN | unix.EPOLLPRI
)

// Poller represents a poller which is in charge of monitoring file-descriptors.
type Poller struct {
	status  uint32
	closeCH chan bool
	fd      int    // epoll fd
	wfd     int    // wake fd
	wfdBuf  []byte // wfd buffer to read packet
	events  [128]unix.EpollEvent

	onError      func(err *base.Error)
	onInvokeAdd  func()
	onInvokeExit func()
	onFDRead     func(fd int)
	onFDWrite    func(fd int)
	onFDClose    func(fd int)
}

// OpenPoller instantiates a poller.
func OpenPoller(
	onError func(err *base.Error),
	onInvokeAdd func(),
	onInvokeExit func(),
	onFDRead func(fd int),
	onFDWrite func(fd int),
	onFDClose func(fd int),
) *Poller {

	if pfd, e := unix.EpollCreate1(unix.EPOLL_CLOEXEC); e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	} else if wfd, e := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC); e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	} else {
		ret := &Poller{
			status:       pollerStatusRunning,
			closeCH:      make(chan bool),
			fd:           pfd,
			wfd:          wfd,
			wfdBuf:       make([]byte, 512),
			onError:      onError,
			onInvokeAdd:  onInvokeAdd,
			onInvokeExit: onInvokeExit,
			onFDRead:     onFDRead,
			onFDWrite:    onFDWrite,
			onFDClose:    onFDClose,
		}

		if e := ret.AddRead(ret.wfd); e != nil {
			_ = ret.Close()
			return nil
		}

		go func() {
			ret.run()
		}()

		return ret
	}
}

func (p *Poller) run() {

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
func (p *Poller) Trigger() (err error) {
	_, err = unix.Write(p.wfd, b)
	return os.NewSyscallError("write", err)
}

// Polling blocks the current goroutine, waiting for network-events.
func (p *Poller) Polling(callback func(fd int, ev uint32) error) error {
	var wakenUp bool

	for {
		n, e := unix.EpollWait(p.fd, p.events[:], -1)

		if e != nil {
			if e != unix.EINTR {
				p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
			}
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

// +build linux

package netpoll

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
)

const triggerDataAddConn = 1
const triggerDataExit = 2

const pollerStatusRunning = 1
const pollerStatusClosing = 2
const pollerStatusClosed = 0

const (
	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents     = unix.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
)

const (
	// ErrEvents ...
	ErrEvents = unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
	// OutEvents ...
	OutEvents = unix.EPOLLOUT
	// InEvents ...
	InEvents = unix.EPOLLIN | unix.EPOLLPRI
)

var (
	triggerDataAddConnBuffer = []byte{triggerDataAddConn, 0, 0, 0, 0, 0, 0, 0}
	triggerDataExitBuffer    = []byte{triggerDataExit, 0, 0, 0, 0, 0, 0, 0}
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

// NewPoller instantiates a poller.
func NewPoller(
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
	} else if wfd, e := unix.Eventfd(
		0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC,
	); e != nil {
		_ = unix.Close(pfd)
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

		if e := ret.RegisterFD(ret.wfd); e != nil {
			ret.Close()
			return nil
		}

		go func() {
			ret.run()
		}()

		return ret
	}
}

func (p *Poller) run() {
	for {
		n, e := unix.EpollWait(p.fd, p.events[:], -1)

		if e != nil {
			if e != unix.EINTR {
				p.onError(errors.ErrTemp.AddDebug(e.Error()))
			}
		}

		for i := 0; i < n; i++ {
			evt := &p.events[i]

			if fd, ev := int(evt.Fd), evt.Events; fd != p.wfd {
				if ev&ErrEvents != 0 {
					p.onFDClose(fd)
				} else if ev&InEvents != 0 {
					p.onFDRead(fd)
				} else if ev&OutEvents != 0 {
					p.onFDWrite(fd)
				} else {
					p.onError(errors.ErrTemp.AddDebug("unknown event filter"))
				}
			} else {
				if ev&InEvents != 0 {
					if readN, _ := unix.Read(p.wfd, p.wfdBuf); readN > 0 {
						for j := 0; j < readN; j++ {
							if p.wfdBuf[j] == triggerDataAddConn {
								p.onInvokeAdd()
							} else if p.wfdBuf[j] == triggerDataExit {
								p.onInvokeExit()
								atomic.StoreUint32(&p.status, pollerStatusClosed)
							} else if p.wfdBuf[j] == 0 {
								continue
							} else {
								p.onError(errors.ErrKqueueSystem.AddDebug("unknown event data"))
							}
						}
					}
				}
			}
		}
	}
}

// Close ...
func (p *Poller) Close() {
	if atomic.CompareAndSwapUint32(
		&p.status,
		pollerStatusRunning,
		pollerStatusClosing,
	) {
		go func() {
			for atomic.LoadUint32(&p.status) == pollerStatusClosing {
				if e := p.TriggerExit(); e != nil {
					p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
				}
				time.Sleep(50 * time.Millisecond)
			}
		}()

		<-p.closeCH

		if e := unix.Close(p.wfd); e != nil {
			p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		}

		if e := unix.Close(p.fd); e != nil {
			p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		}
	} else {
		p.onError(errors.ErrKqueueNotRunning)
	}
}

// RegisterFD ...
func (p *Poller) RegisterFD(fd int) error {
	return unix.EpollCtl(
		p.fd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: readEvents},
	)
}

// UnregisterFD ...
func (p *Poller) UnregisterFD(fd int) error {
	e := unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd, nil)
	return os.NewSyscallError("kqueue add", e)
}

// WatchWrite ...
func (p *Poller) WatchWrite(fd int) error {
	return unix.EpollCtl(
		p.fd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: writeEvents},
	)
}

// UnwatchWrite ...
func (p *Poller) UnwatchWrite(fd int) error {
	return unix.EpollCtl(
		p.fd,
		unix.EPOLL_CTL_MOD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: readEvents},
	)
}

// TriggerAddConn ...
func (p *Poller) TriggerAddConn() (err error) {
	_, e := unix.Write(p.wfd, triggerDataAddConnBuffer)
	return os.NewSyscallError("kqueue trigger", e)
}

// TriggerExit ...
func (p *Poller) TriggerExit() (err error) {
	_, e := unix.Write(p.wfd, triggerDataExitBuffer)
	return os.NewSyscallError("kqueue trigger", e)
}

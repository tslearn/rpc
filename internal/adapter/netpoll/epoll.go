// +build linux

package netpoll

import (
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
)

const (
	pollerStatusRunning = 1
	pollerStatusClosing = 2
	pollerStatusClosed  = 0

	readEvents  = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents = unix.EPOLLOUT

	errEvents = unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
	outEvents = unix.EPOLLOUT
	inEvents  = unix.EPOLLIN | unix.EPOLLPRI
)

var triggerSendData = []byte{1, 1, 1, 1, 1, 1, 1, 1}

// Poller ...
type Poller struct {
	status    uint32
	pollFD    int
	triggerFD int
	events    [128]unix.EpollEvent
	onError   func(err *base.Error)
}

// NewPoller instantiates a poller.
func NewPoller(
	onError func(err *base.Error),
	onTrigger func(),
	onFDRead func(fd int),
	onFDWrite func(fd int),
	onFDClose func(fd int),
) *Poller {
	if pollFD, e := unix.EpollCreate1(unix.EPOLL_CLOEXEC); e != nil {
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	} else if triggerFD, e := unix.Eventfd(
		0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC,
	); e != nil {
		_ = unix.Close(pollFD)
		onError(errors.ErrTemp.AddDebug(e.Error()))
		return nil
	} else {
		ret := &Poller{
			status:    pollerStatusRunning,
			pollFD:    pollFD,
			triggerFD: triggerFD,
			onError:   onError,
		}

		if e := ret.RegisterFD(triggerFD); e != nil {
			ret.Close()
			return nil
		}

		go func() {
			triggerReadBuf := make([]byte, 256)

			for atomic.LoadUint32(&ret.status) == pollerStatusRunning {
				n, e := unix.EpollWait(ret.pollFD, ret.events[:], -1)

				if e != nil {
					if e != unix.EINTR {
						onError(errors.ErrTemp.AddDebug(e.Error()))
					}
				}

				for i := 0; i < n; i++ {
					evt := &ret.events[i]
					if fd, ev := int(evt.Fd), evt.Events; fd == triggerFD {
						_, _ = unix.Read(triggerFD, triggerReadBuf)
						onTrigger()
					} else {
						if ev&errEvents != 0 {
							onFDClose(fd)
						} else if ev&inEvents != 0 {
							onFDRead(fd)
						} else if ev&outEvents != 0 {
							onFDWrite(fd)
						} else {
							// ignore
						}
					}
				}
			}

			if e := unix.Close(triggerFD); e != nil {
				onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
			}

			atomic.StoreUint32(&ret.status, pollerStatusClosed)
		}()

		return ret
	}
}

// Close ...
func (p *Poller) Close() {
	if atomic.CompareAndSwapUint32(
		&p.status,
		pollerStatusRunning,
		pollerStatusClosing,
	) {
		for atomic.LoadUint32(&p.status) == pollerStatusClosing {
			time.Sleep(50 * time.Millisecond)
		}

		if e := unix.Close(p.pollFD); e != nil {
			p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		}
	} else {
		p.onError(errors.ErrKqueueNotRunning)
	}
}

// RegisterFD ...
func (p *Poller) RegisterFD(fd int) error {
	return unix.EpollCtl(
		p.pollFD,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: readEvents},
	)
}

// UnregisteredFD ...
func (p *Poller) UnregisteredFD(fd int) error {
	return unix.EpollCtl(
		p.pollFD,
		unix.EPOLL_CTL_DEL,
		fd,
		nil,
	)
}

// AddWrite ...
func (p *Poller) AddWrite(fd int) error {
	return unix.EpollCtl(
		p.pollFD,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: writeEvents},
	)
}

// DelWrite ...
func (p *Poller) DelWrite(fd int) error {
	return unix.EpollCtl(
		p.pollFD,
		unix.EPOLL_CTL_MOD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: readEvents},
	)
}

// Trigger ...
func (p *Poller) Trigger() (err error) {
	_, e := unix.Write(p.triggerFD, triggerSendData)
	return e
}

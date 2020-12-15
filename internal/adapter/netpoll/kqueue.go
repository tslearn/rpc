// +build darwin freebsd dragonfly

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
)

// Poller ...
type Poller struct {
	status  uint32
	pollFD  int
	events  [128]unix.Kevent_t
	onError func(err *base.Error)
}

// NewPoller ...
func NewPoller(
	onError func(err *base.Error),
	onTrigger func(),
	onFDRead func(fd int),
	onFDWrite func(fd int),
	onFDClose func(fd int),
) *Poller {
	if pollFD, e := unix.Kqueue(); e != nil {
		onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		return nil
	} else if _, e := unix.Kevent(pollFD, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Flags:  unix.EV_ADD | unix.EV_CLEAR,
	}}, nil, nil); e != nil {
		onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		return nil
	} else {
		ret := &Poller{
			status:  pollerStatusRunning,
			pollFD:  pollFD,
			onError: onError,
		}

		go func() {
			for atomic.LoadUint32(&ret.status) == pollerStatusRunning {
				n, e := unix.Kevent(ret.pollFD, nil, ret.events[:], nil)

				if e != nil {
					if e != unix.EINTR {
						onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
					}
				}

				for i := 0; i < n; i++ {
					evt := &ret.events[i]
					if fd := int(evt.Ident); fd == 0 {
						onTrigger()
					} else {
						if evt.Flags&unix.EV_EOF != 0 || evt.Flags&unix.EV_ERROR != 0 {
							onFDClose(fd)
						} else if evt.Filter == unix.EVFILT_READ {
							onFDRead(fd)
						} else if evt.Filter == unix.EVFILT_WRITE {
							onFDWrite(fd)
						} else {
							// ignore
						}
					}
				}
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
	_, err := unix.Kevent(p.pollFD, []unix.Kevent_t{{
		Ident:  uint64(fd),
		Flags:  unix.EV_ADD,
		Filter: unix.EVFILT_READ,
	}}, nil, nil)
	return err
}

// UnregisteredFD ...
func (p *Poller) UnregisteredFD(fd int) error {
	return nil
}

// AddWrite ...
func (p *Poller) AddWrite(fd int) error {
	_, err := unix.Kevent(p.pollFD, []unix.Kevent_t{{
		Ident:  uint64(fd),
		Flags:  unix.EV_ADD,
		Filter: unix.EVFILT_WRITE,
	}}, nil, nil)
	return err
}

// DelWrite ...
func (p *Poller) DelWrite(fd int) error {
	_, err := unix.Kevent(p.pollFD, []unix.Kevent_t{{
		Ident:  uint64(fd),
		Flags:  unix.EV_DELETE,
		Filter: unix.EVFILT_WRITE,
	}}, nil, nil)
	return err
}

// Trigger ...
func (p *Poller) Trigger() error {
	_, err := unix.Kevent(p.pollFD, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
		Data:   0,
	}}, nil, nil)
	return err
}

// +build freebsd dragonfly darwin

package xadapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"os"
)

const triggerDataAdd = 1
const triggerDataExit = 2

// Poller ...
type Poller struct {
	fd     int
	events [256]unix.Kevent_t
}

// NewPoller ...
func NewPoller() (poller *Poller, err error) {
	poller = new(Poller)
	if poller.fd, err = unix.Kqueue(); err != nil {
		poller = nil
		err = os.NewSyscallError("kqueue", err)
		return
	}
	if _, err = unix.Kevent(poller.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Flags:  unix.EV_ADD | unix.EV_CLEAR,
	}}, nil, nil); err != nil {
		_ = poller.Close()
		poller = nil
		err = os.NewSyscallError("kqueue add|clear", err)
		return
	}
	return
}

// Close ...
func (p *Poller) Close() error {
	return os.NewSyscallError("close", unix.Close(p.fd))
}

// Polling ...
func (p *Poller) Polling(
	onTriggerAdd func(),
	onRead func(fd int),
	onClose func(fd int),
	onError func(err *base.Error),
	onTriggerExit func(),
) {
	for {
		n, err := unix.Kevent(p.fd, nil, p.events[:], nil)

		if err != nil && err != unix.EINTR {
			onError(errors.ErrKqueueSystem.AddDebug(err.Error()))
		}

		for i := 0; i < n; i++ {
			evt := p.events[i]
			if fd := int(evt.Ident); fd == 0 {
				if evt.Data == triggerDataAdd {
					onTriggerAdd()
				} else if evt.Data == triggerDataExit {
					onTriggerExit()
					return
				} else {
					onError(errors.ErrKqueueSystem.AddDebug("unknown event data"))
				}
			} else {
				if evt.Flags&unix.EV_EOF != 0 || evt.Flags&unix.EV_ERROR != 0 {
					onClose(fd)
				} else {
					onRead(fd)
				}
			}
		}
	}
}

// Add ...
func (p *Poller) RegisterFD(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: uint64(fd), Flags: unix.EV_ADD, Filter: unix.EVFILT_READ},
	}, nil, nil)
	return os.NewSyscallError("kqueue add", err)
}

// Delete ...
func (p *Poller) UnregisterFD(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: uint64(fd), Flags: unix.EV_DELETE, Filter: unix.EVFILT_READ},
	}, nil, nil)
	return os.NewSyscallError("kqueue delete", err)
}

// InvokeAddTrigger ...
func (p *Poller) InvokeAddTrigger() (err error) {
	_, err = unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
		Data:   triggerDataAdd,
	}}, nil, nil)
	return os.NewSyscallError("kqueue trigger", err)
}

// InvokeAddTrigger ...
func (p *Poller) InvokeExitTrigger() (err error) {
	_, err = unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
		Data:   triggerDataExit,
	}}, nil, nil)
	return os.NewSyscallError("kqueue trigger", err)
}

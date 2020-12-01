// +build freebsd dragonfly darwin

package xadapter

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

const TriggerTypeAdd = 1
const triggerTypeShutdown = 2

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
	callback func(fd int, isClose bool),
	addConn func(),
	exit func(),
) error {
	for {
		n, err := unix.Kevent(p.fd, nil, p.events[:], nil)
		if err != nil && err != unix.EINTR {
			return os.NewSyscallError("kqueue wait", err)
		}

		for i := 0; i < n; i++ {
			evt := p.events[i]
			if fd := int(evt.Ident); fd == 0 {
				fmt.Println("evt.Fflags", evt.Data)
				addConn()
			} else {
				callback(fd, evt.Flags&unix.EV_EOF != 0 || evt.Flags&unix.EV_ERROR != 0)
			}
		}
	}
}

// Add ...
func (p *Poller) Add(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: uint64(fd), Flags: unix.EV_ADD, Filter: unix.EVFILT_READ},
	}, nil, nil)
	return os.NewSyscallError("kqueue add", err)
}

// Delete ...
func (p *Poller) Delete(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: uint64(fd), Flags: unix.EV_DELETE, Filter: unix.EVFILT_READ},
	}, nil, nil)
	return os.NewSyscallError("kqueue delete", err)
}

// Trigger ...
func (p *Poller) Trigger(data int64) (err error) {
	_, err = unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: 0, Filter: unix.EVFILT_USER, Fflags: unix.NOTE_TRIGGER, Data: data},
	}, nil, nil)
	return os.NewSyscallError("kqueue trigger", err)
}

// +build freebsd dragonfly darwin

package async

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"os"
	"sync/atomic"
	"time"
)

const triggerDataAddConn = 1
const triggerDataExit = 2

const pollerStatusRunning = 1
const pollerStatusClosing = 2
const pollerStatusClosed = 0

// Poller ...
type Poller struct {
	status  uint32
	closeCH chan bool
	fd      int
	events  [128]unix.Kevent_t

	onError      func(err *base.Error)
	onInvokeAdd  func()
	onInvokeExit func()
	onFDRead     func(fd int)
	onFDWrite    func(fd int)
	onFDClose    func(fd int)
}

// NewPoller ...
func NewPoller(
	onError func(err *base.Error),
	onInvokeAdd func(),
	onInvokeExit func(),
	onFDRead func(fd int),
	onFDWrite func(fd int),
	onFDClose func(fd int),
) *Poller {
	if pfd, e := unix.Kqueue(); e != nil {
		onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		return nil
	} else if _, e := unix.Kevent(pfd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Flags:  unix.EV_ADD | unix.EV_CLEAR,
	}}, nil, nil); e != nil {
		onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		return nil
	} else {
		ret := &Poller{
			status:       pollerStatusRunning,
			closeCH:      make(chan bool),
			fd:           pfd,
			onError:      onError,
			onInvokeAdd:  onInvokeAdd,
			onInvokeExit: onInvokeExit,
			onFDRead:     onFDRead,
			onFDWrite:    onFDWrite,
			onFDClose:    onFDClose,
		}

		go func() {
			ret.run()
		}()

		return ret
	}
}

func (p *Poller) run() {
	for atomic.LoadUint32(&p.status) == pollerStatusRunning {
		n, e := unix.Kevent(p.fd, nil, p.events[:], nil)

		if e != nil {
			if e != unix.EINTR {
				p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
			}
		}

		for i := 0; i < n; i++ {
			evt := &p.events[i]
			if fd := int(evt.Ident); fd == 0 {
				if evt.Data == triggerDataAddConn {
					p.onInvokeAdd()
				} else if evt.Data == triggerDataExit {
					p.onInvokeExit()
					atomic.StoreUint32(&p.status, pollerStatusClosed)
				} else {
					p.onError(errors.ErrKqueueSystem.AddDebug("unknown event data"))
				}
			} else {
				if evt.Flags&unix.EV_EOF != 0 || evt.Flags&unix.EV_ERROR != 0 {
					p.onFDClose(fd)
				} else if evt.Filter == unix.EVFILT_READ {
					p.onFDRead(fd)
				} else if evt.Filter == unix.EVFILT_WRITE {
					p.onFDWrite(fd)
				} else {
					p.onError(errors.ErrKqueueSystem.AddDebug("unknown event filter"))
				}
			}
		}
	}

	p.closeCH <- true
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

		if e := unix.Close(p.fd); e != nil {
			p.onError(errors.ErrKqueueSystem.AddDebug(e.Error()))
		}
	} else {
		p.onError(errors.ErrKqueueNotRunning)
	}
}

// RegisterFD ...
func (p *Poller) RegisterFD(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{
			Ident:  uint64(fd),
			Flags:  unix.EV_ADD | unix.EV_CLEAR,
			Filter: unix.EVFILT_READ,
		},
		{
			Ident:  uint64(fd),
			Flags:  unix.EV_ADD | unix.EV_CLEAR,
			Filter: unix.EVFILT_WRITE,
		},
	}, nil, nil)
	return os.NewSyscallError("kqueue add", err)
}

// TriggerAddConn ...
func (p *Poller) TriggerAddConn() (err error) {
	_, err = unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
		Data:   triggerDataAddConn,
	}}, nil, nil)
	return os.NewSyscallError("kqueue trigger", err)
}

// InvokeAddTrigger ...
func (p *Poller) TriggerExit() (err error) {
	_, err = unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
		Data:   triggerDataExit,
	}}, nil, nil)
	return os.NewSyscallError("kqueue trigger", err)
}

// Delete
func (p *Poller) Delete(fd int) error {
	return nil
}

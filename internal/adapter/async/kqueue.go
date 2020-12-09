// +build freebsd dragonfly darwin

package async

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"os"
	"sync/atomic"
	"time"
)

const triggerDataAddConn = 1
const triggerDataWriteConn = 2
const triggerDataExit = 3

const pollerStatusRunning = 1
const pollerStatusClosing = 2
const pollerStatusClosed = 0

// Poller ...
type Poller struct {
	status  uint32
	closeCH chan bool
	onError func(err *base.Error)
	fd      int
	events  [256]unix.Kevent_t
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
			status:  pollerStatusRunning,
			closeCH: make(chan bool),
			onError: onError,
			fd:      pfd,
		}

		go func() {
			for atomic.LoadUint32(&ret.status) == pollerStatusRunning {
				n, err := unix.Kevent(ret.fd, nil, ret.events[:], nil)

				if err != nil && err != unix.EINTR {
					onError(errors.ErrKqueueSystem.AddDebug(err.Error()))
					continue
				}

				for i := 0; i < n; i++ {
					evt := ret.events[i]
					if fd := int(evt.Ident); fd == 0 {
						if evt.Data == triggerDataAddConn {
							onInvokeAdd()
						} else if evt.Data == triggerDataExit {
							onInvokeExit()
							atomic.StoreUint32(&ret.status, pollerStatusClosed)
							break
						} else {
							onError(errors.ErrKqueueSystem.AddDebug("unknown event data"))
						}
					} else {
						if evt.Flags&unix.EV_EOF != 0 || evt.Flags&unix.EV_ERROR != 0 {
							onFDClose(fd)
						} else if evt.Filter == unix.EVFILT_READ {
							onFDRead(fd)
						} else if evt.Filter == unix.EVFILT_WRITE {
							onFDWrite(fd)
						} else if evt.Filter == unix.EVFILT_USER &&
							evt.Data == triggerDataWriteConn {
							onFDWrite(fd)
						} else {
							onError(errors.ErrKqueueSystem.AddDebug("unknown event filter"))
						}
					}
				}
			}
			ret.closeCH <- true
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

// RegisterReadFD ...
func (p *Poller) RegisterReadFD(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{
			Ident:  uint64(fd),
			Flags:  unix.EV_ADD | unix.EV_CLEAR,
			Filter: unix.EVFILT_READ,
		},
	}, nil, nil)
	return os.NewSyscallError("kqueue add", err)
}

// RegisterWriteFD ...
func (p *Poller) RegisterWriteFD(fd int) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{
			Ident:  uint64(fd),
			Flags:  unix.EV_ADD | unix.EV_CLEAR,
			Filter: unix.EVFILT_WRITE,
		},
		{
			Ident:  uint64(fd),
			Flags:  unix.EV_ADD | unix.EV_CLEAR,
			Filter: unix.EVFILT_USER,
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

func (p *Poller) TriggerWriteConn(fd int) (err error) {
	fmt.Println("TriggerWriteConn", fd)
	_, err = unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  uint64(fd),
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
		Data:   triggerDataWriteConn,
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

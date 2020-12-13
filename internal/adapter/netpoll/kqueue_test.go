// +build darwin freebsd dragonfly

package netpoll

import (
	"github.com/rpccloud/rpc/internal/base"
	"golang.org/x/sys/unix"
	"net"
	"testing"
	"time"
)

type testKqueue struct {
	pollID int
}

func newTestKqueue() *testKqueue {
	pollID, e := unix.Kqueue()

	if e != nil {
		panic(e)
	}

	return &testKqueue{pollID: pollID}
}

func (p *testKqueue) wait() int {
	var events [128]unix.Kevent_t
	n, e := unix.Kevent(p.pollID, nil, events[:], &unix.Timespec{Nsec: 1000})
	if e != nil {
		panic(e)
	}
	return n
}

func (p *testKqueue) RegisterRead(fd int) error {
	_, err := unix.Kevent(p.pollID, []unix.Kevent_t{{
		Ident:  uint64(fd),
		Flags:  unix.EV_ADD,
		Filter: unix.EVFILT_READ,
	}}, nil, nil)

	return err
}

func TestEpoll_CanAsync(t *testing.T) {
	assert := base.NewAssert(t)
	fd := 0
	e := error(nil)

	fd, _, e = listen("tcp", "0.0.0.0:63322")
	if e != nil {
		panic(e)
	}

	poll1 := newTestKqueue()
	poll2 := newTestKqueue()

	if e := poll1.RegisterRead(fd); e != nil {
		panic(e)
	}

	if e := poll2.RegisterRead(fd); e != nil {
		panic(e)
	}

	if _, e := net.Dial("tcp", "127.0.0.1:63322"); e != nil {
		panic(e)
	}

	time.Sleep(time.Second)

	assert(poll1.wait()).Equal(1)
	assert(poll2.wait()).Equal(1)

	if _, _, e := unix.Accept(fd); e != nil {
		panic(e)
	}

	assert(poll1.wait()).Equal(0)
	assert(poll2.wait()).Equal(0)

	_ = unix.Close(fd)
	_ = unix.Close(poll1.pollID)
	_ = unix.Close(poll2.pollID)
}

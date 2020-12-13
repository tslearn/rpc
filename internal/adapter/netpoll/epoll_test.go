// +build linux

package netpoll

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type testEpoll struct {
	eid int
}

func newTestEpoll() *testEpoll {
	eid, e := unix.EpollCreate1(unix.EPOLL_CLOEXEC)

	if e != nil {
		panic(e)
	}

	return &testEpoll{eid: eid}
}

func (p *testEpoll) wait() int {
	var events [128]unix.EpollEvent
	n, e := unix.EpollWait(p.eid, events[:], 1000)
	if e != nil {
		panic(e)
	}
	return n
}

func (p *testEpoll) RegisterRead(fd int) error {
	return unix.EpollCtl(
		p.eid,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Fd: int32(fd), Events: readEvents},
	)
}

func TestEpoll_CanAsync(t *testing.T) {
	redBuf := make([]byte, 1024)
	fd, e := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if e != nil {
		panic(e)
	}

	poll1 := newTestEpoll()
	poll2 := newTestEpoll()

	if e := poll1.RegisterRead(fd); e != nil {
		panic(e)
	}

	if e := poll2.RegisterRead(fd); e != nil {
		panic(e)
	}

	if _, e := unix.Write(fd, triggerDataAddConnBuffer); e != nil {
		panic(e)
	}

	go func() {
		fmt.Println(poll1.wait())
		fmt.Println(poll2.wait())

		fmt.Println("A")
		unix.Read(fd, redBuf)
		fmt.Println("B")

		fmt.Println(poll1.wait())
		fmt.Println(poll2.wait())
	}()

	time.Sleep(10 * time.Second)
}

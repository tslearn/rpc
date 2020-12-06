// +build darwin

package xadapter

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func sysSocket(family, sotype, proto int) (fd int, err error) {
	syscall.ForkLock.RLock()
	if fd, err = unix.Socket(family, sotype, proto); err == nil {
		unix.CloseOnExec(fd)
	}
	syscall.ForkLock.RUnlock()

	if err != nil {
		return
	}

	if err = unix.SetNonblock(fd, true); err != nil {
		_ = unix.Close(fd)
	}

	return
}

// +build linux

package netpoll

import "golang.org/x/sys/unix"

func sysSocket(family int, sotype int, proto int) (int, error) {
	return unix.Socket(family, sotype|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, proto)
}

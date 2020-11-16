package client

import "unsafe"

type Channel struct {
	seq  uint64
	item unsafe.Pointer
}

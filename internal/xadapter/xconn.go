package xadapter

type XConn interface {
	FD() int

	OnRead() error
	OnClose() error

	OnWriteReady() error
}

package server

import (
	"errors"
	"fmt"
	"github.com/rpccloud/rpc/internal/xadapter"
	"github.com/rpccloud/rpc/internal/xadapter/tcp/reuseport"
	"golang.org/x/sys/unix"
)

var ErrAcceptorAlreadyListening = errors.New("it is already listening")
var ErrAcceptorNotListening = errors.New("it is not listening")

type Acceptor struct {
	rootFD int
	proto  string
	addr   string
	poller *xadapter.Poller
}

func NewAcceptor(proto string, addr string) *Acceptor {
	return &Acceptor{
		rootFD: 0,
		proto:  proto,
		addr:   addr,
		poller: nil,
	}
}

func (p *Acceptor) handleEvent(fd int, filter int16) error {
	if fd == p.rootFD {
		connFD, _, err := unix.Accept(fd)
		fmt.Println(p.rootFD, connFD)
		return err
	}

	return nil
}

func (p *Acceptor) Listen() error {
	if p.rootFD != 0 {
		return ErrAcceptorAlreadyListening
	} else if rootFD, _, err := reuseport.TCPSocket(p.proto, p.addr, true); err != nil {
		return err
	} else if poller, err := xadapter.OpenPoller(); err != nil {
		return err
	} else if err := poller.AddRead(rootFD); err != nil {
		return err
	} else {
		p.rootFD = rootFD
		p.poller = poller
		return nil
	}
}

func (p *Acceptor) Serve() error {
	if p.rootFD == 0 {
		return ErrAcceptorNotListening
	}

	fmt.Println("Serve: ", p.rootFD)
	for {
		err := p.poller.Polling(p.handleEvent)
		if err != nil {
			return err
		}
	}
}

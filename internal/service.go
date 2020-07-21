package internal

import "sync"

type rpcReplyCheckStatus int

const (
	rpcReplyCheckStatusNone rpcReplyCheckStatus = iota
	rpcReplyCheckStatusOK
	rpcReplyCheckStatusError
)

type rpcReplyMeta struct {
	name    string      // the name of reply
	handler interface{} // reply handler
	debug   string      // where the reply add in source file
	sync.Map
}

func (p *rpcReplyMeta) GetCheckStatus(key string) rpcReplyCheckStatus {
	if v, ok := p.Load(key); ok {
		return v.(rpcReplyCheckStatus)
	}
	return rpcReplyCheckStatusNone
}

func (p *rpcReplyMeta) SetCheckOK(key string) {
	p.Store(key, rpcReplyCheckStatusOK)
}

func (p *rpcReplyMeta) SetCheckError(key string) {
	p.Store(key, rpcReplyCheckStatusError)
}

type rpcChildMeta struct {
	name    string   // the name of child service
	service *Service // the real service
	debug   string   // where the service add in source file
}

type Service struct {
	children []*rpcChildMeta // all the children node meta pointer
	replies  []*rpcReplyMeta // all the replies meta pointer
	debug    string          // where the service define in source file
	Lock
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: make([]*rpcChildMeta, 0, 0),
		replies:  make([]*rpcReplyMeta, 0, 0),
		debug:    AddFileLine("", 1),
	}
}

// Reply add reply handler
func (p *Service) Reply(
	name string,
	handler interface{},
) *Service {
	// add reply meta
	p.DoWithLock(func() {
		p.replies = append(p.replies, &rpcReplyMeta{
			name:    name,
			handler: handler,
			debug:   AddFileLine("", 1),
		})
	})
	return p
}

// AddChild add child service
func (p *Service) AddChild(name string, service *Service) *Service {
	// add child meta
	p.DoWithLock(func() {
		p.children = append(p.children, &rpcChildMeta{
			name:    name,
			service: service,
			debug:   AddFileLine("", 1),
		})
	})
	return p
}

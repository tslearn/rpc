package internal

import "sync"

type replyMeta struct {
	name    string      // the name of reply
	handler interface{} // reply handler
	debug   string      // where the reply add in source file
}

type childMeta struct {
	name    string   // the name of child service
	service *Service // the real service
	debug   string   // where the service add in source file
}

type Service struct {
	children []*childMeta // all the children node meta pointer
	replies  []*replyMeta // all the replies meta pointer
	debug    string       // where the service define in source file
	sync.Mutex
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: make([]*childMeta, 0, 0),
		replies:  make([]*replyMeta, 0, 0),
		debug:    GetStackString(1),
	}
}

// Reply add reply handler
func (p *Service) Reply(
	name string,
	handler interface{},
) *Service {
	p.Lock()
	defer p.Unlock()
	// add reply meta
	p.replies = append(p.replies, &replyMeta{
		name:    name,
		handler: handler,
		debug:   GetStackString(1),
	})
	return p
}

// AddChild add child service
func (p *Service) AddChild(name string, service *Service) *Service {
	p.Lock()
	defer p.Unlock()

	// add child meta
	p.children = append(p.children, &childMeta{
		name:    name,
		service: service,
		debug:   GetStackString(1),
	})
	return p
}

package internal

type rpcReplyMeta struct {
	name     string      // the name of reply
	handler  interface{} // reply handler
	fileLine string      // where the reply add in source file
}

type rpcChildMeta struct {
	name     string   // the name of child service
	service  *Service // the real service
	fileLine string   // where the service add in source file
}

type Service struct {
	children []*rpcChildMeta // all the children node meta pointer
	replies  []*rpcReplyMeta // all the replies meta pointer
	fileLine string          // where the service define in source file
	Lock
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: make([]*rpcChildMeta, 0, 0),
		replies:  make([]*rpcReplyMeta, 0, 0),
		fileLine: GetFileLine(1),
	}
}

// AddChildService ...
func (p *Service) AddChildService(name string, service *Service) *Service {
	// add child meta
	p.DoWithLock(func() {
		p.children = append(p.children, &rpcChildMeta{
			name:     name,
			service:  service,
			fileLine: GetFileLine(3),
		})
	})
	return p
}

// Reply add reply handler
func (p *Service) Reply(name string, handler interface{}) *Service {
	// add reply meta
	p.DoWithLock(func() {
		p.replies = append(p.replies, &rpcReplyMeta{
			name:     name,
			handler:  handler,
			fileLine: GetFileLine(3),
		})
	})
	return p
}

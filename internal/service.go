package internal

// Service ...
type Service interface {
	Reply(
		name string,
		handler interface{},
	) Service

	AddService(
		name string,
		service Service,
	) Service
}

type rpcReplyMeta struct {
	name    string      // the name of reply
	handler interface{} // reply handler
	debug   string      // where the reply add in source file
}

type rpcNodeMeta struct {
	name        string      // the name of child service
	serviceMeta *rpcService // the real service
	debug       string      // where the service add in source file
}

type rpcService struct {
	children []*rpcNodeMeta  // all the children node meta pointer
	replies  []*rpcReplyMeta // all the replies meta pointer
	debug    string          // where the service define in source file
	RPCLock
}

// NewRPCService define a new service
func NewRPCService() Service {
	return &rpcService{
		children: make([]*rpcNodeMeta, 0, 0),
		replies:  make([]*rpcReplyMeta, 0, 0),
		debug:    GetStackString(1),
	}
}

// Reply add reply handler
func (p *rpcService) Reply(
	name string,
	handler interface{},
) Service {
	p.DoWithLock(func() {
		// add reply meta
		p.replies = append(p.replies, &rpcReplyMeta{
			name:    name,
			handler: handler,
			debug:   GetStackString(3),
		})
	})
	return p
}

// AddService add child service
func (p *rpcService) AddService(name string, service Service) Service {
	serviceMeta, ok := service.(*rpcService)
	if !ok {
		serviceMeta = (*rpcService)(nil)
	}

	// lock
	p.DoWithLock(func() {
		// add child meta
		p.children = append(p.children, &rpcNodeMeta{
			name:        name,
			serviceMeta: serviceMeta,
			debug:       GetStackString(3),
		})
	})

	return p
}

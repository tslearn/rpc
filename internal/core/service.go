package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"sync"
)

type rpcReplyMeta struct {
	name     string      // the name of reply
	handler  interface{} // reply handler
	fileLine string      // where the reply add in source file
}

// ServiceMeta ...
type ServiceMeta struct {
	name     string   // the name of child service
	service  *Service // the real service
	fileLine string   // where the service add in source file
	data     Map
}

// NewServiceMeta ...
func NewServiceMeta(
	name string,
	service *Service,
	fileLine string,
	data Map,
) *ServiceMeta {
	return &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: fileLine,
		data:     data,
	}
}

// Service ...
type Service struct {
	children []*ServiceMeta  // all the children node meta pointer
	replies  []*rpcReplyMeta // all the replies meta pointer
	fileLine string          // where the service define in source file
	sync.Mutex
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: nil,
		replies:  nil,
		fileLine: base.GetFileLine(1),
	}
}

// AddChildService ...
func (p *Service) AddChildService(
	name string,
	service *Service,
	data Map,
) *Service {
	p.Lock()
	defer p.Unlock()
	// add child meta
	p.children = append(p.children, &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: base.GetFileLine(1),
		data:     data,
	})
	return p
}

// Reply add reply handler
func (p *Service) Reply(
	name string,
	handler interface{},
) *Service {
	p.Lock()
	defer p.Unlock()

	// add reply meta
	p.replies = append(p.replies, &rpcReplyMeta{
		name:     name,
		handler:  handler,
		fileLine: base.GetFileLine(1),
	})
	return p
}

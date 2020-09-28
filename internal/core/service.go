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
	name     string      // the name of child service
	service  *Service    // the real service
	fileLine string      // where the service add in source file
	config   interface{} // data
}

// NewServiceMeta ...
func NewServiceMeta(
	name string,
	service *Service,
	fileLine string,
	config interface{},
) *ServiceMeta {
	return &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: fileLine,
		config:   config,
	}
}

// Service ...
type Service struct {
	children []*ServiceMeta  // all the children node meta pointer
	replies  []*rpcReplyMeta // all the replies meta pointer
	fileLine string          // where the service define in source file

	//OnDataUpdate
	//OnMount
	//OnUnmount
	onMount func(service *Service, data interface{}) error
	sync.Mutex
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: nil,
		replies:  nil,
		fileLine: base.GetFileLine(1),
		onMount:  nil,
	}
}

// NewServiceWithOnMount define a new service with onMount
func NewServiceWithOnMount(
	onMount func(service *Service, data interface{}) error,
) *Service {
	return &Service{
		children: nil,
		replies:  nil,
		fileLine: base.GetFileLine(1),
		onMount:  onMount,
	}
}

// AddChildService ...
func (p *Service) AddChildService(
	name string,
	service *Service,
	config interface{},
) *Service {
	p.Lock()
	defer p.Unlock()
	// add child meta
	p.children = append(p.children, &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: base.GetFileLine(1),
		config:   config,
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

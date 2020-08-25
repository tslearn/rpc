package internal

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
	data     interface{} // data
}

// NewServiceMeta ...
func NewServiceMeta(
	name string,
	service *Service,
	fileLine string,
	data interface{},
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
	onMount  func(service *Service, data interface{}) error
	Lock
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: nil,
		replies:  nil,
		fileLine: GetFileLine(1),
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
		fileLine: GetFileLine(1),
		onMount:  onMount,
	}
}

// AddChildService ...
func (p *Service) AddChildService(
	name string,
	service *Service,
	data interface{},
) *Service {
	// add child meta
	p.DoWithLock(func() {
		p.children = append(p.children, &ServiceMeta{
			name:     name,
			service:  service,
			fileLine: GetFileLine(3),
			data:     data,
		})
	})
	return p
}

// Reply add reply handler
func (p *Service) Reply(
	name string,
	handler interface{},
) *Service {
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

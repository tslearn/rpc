package common

import (
	"reflect"
	"sync"
)

type rpcEchoMeta struct {
	name    string      // the name of echo
	export  bool        // weather echo is export to gateway
	handler interface{} // echo handler
	debug   string      // where the echo add in source file
}

type rpcAddMeta struct {
	name        string          // the name of child service
	serviceMeta *rpcServiceMeta // the real service
	debug       string          // where the service add in source file
}

type rpcServiceMeta struct {
	children []*rpcAddMeta  // all the children service meta pointer
	echos    []*rpcEchoMeta // all the echos meta pointer
	debug    string         // where the service define in source file
	sync.Mutex
}

// NewService define a new service
func newServiceMeta() *rpcServiceMeta {
	return &rpcServiceMeta{
		children: make([]*rpcAddMeta, 0, 0),
		echos:    make([]*rpcEchoMeta, 0, 0),
		debug:    GetStackString(1),
	}
}

// Echo add echo handler
func (p *rpcServiceMeta) Echo(
	name string,
	export bool,
	handler interface{},
) *rpcServiceMeta {
	// lock
	p.Lock()
	defer p.Unlock()

	// add echo meta
	p.echos = append(p.echos, &rpcEchoMeta{
		name:    name,
		export:  export,
		handler: handler,
		debug:   GetStackString(1),
	})
	return p
}

// Add add child service
func (p *rpcServiceMeta) Add(
	name string,
	serviceMeta *rpcServiceMeta,
) *rpcServiceMeta {
	// lock
	p.Lock()
	defer p.Unlock()

	// add child meta
	p.children = append(p.children, &rpcAddMeta{
		name:        name,
		serviceMeta: serviceMeta,
		debug:       GetStackString(1),
	})

	return p
}

type rpcEchoNode struct {
	serviceNode *rpcServiceNode
	path        string
	echoMeta    *rpcEchoMeta
	fnCache     fnCacheFunc
	reflectFn   reflect.Value
	callString  string
	debugString string
	argTypes    []reflect.Type
	indicator   *indicator
}

type rpcServiceNode struct {
	path    string
	addMeta *rpcAddMeta
	depth   uint
}

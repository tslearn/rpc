package core

type fnCacheFunc = func(
	ctx Context,
	stream *RPCStream,
	fn interface{},
) bool

type fnProcessorCallback = func(
	stream *RPCStream,
	success bool,
)

// Any common Any type
type Any = interface{}

// Array common Array type
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

var (
	readTypeArray Array
	readTypeMap   Map
)

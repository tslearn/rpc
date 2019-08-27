package core

// RPCStreamWriteErrorCode ...
type RPCStreamWriteErrorCode int

const (
	RPCStreamWriteOK RPCStreamWriteErrorCode = iota
	RPCStreamWriteUnsupportedType
	RPCStreamWriteRPCArrayIsNotAvailable
	RPCStreamWriteRPCArrayError
	RPCStreamWriteRPCMapIsNotAvailable
	RPCStreamWriteRPCMapError
)

type fnCacheFunc = func(
	ctx *rpcContext,
	stream *rpcStream,
	fn interface{},
) bool

type fnProcessorCallback = func(
	stream *rpcStream,
	success bool,
)

// rpcReturn is rpc function return type
type rpcReturn struct{}

// Any common Any type
type Any = interface{}

// Array common Array type
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

//type Service = *rpcServiceMeta
//var NewService = newServiceMeta
//type Content = *rpcContext
//type Return = *rpcReturn

var (
	emptyString = ""
	emptyBytes  = make([]byte, 0, 0)
	nilRPCArray = rpcArray{}
	nilRPCMap   = rpcMap{}
	nilReturn   = (*rpcReturn)(nil)
	nilContext  = (*rpcContext)(nil)
)

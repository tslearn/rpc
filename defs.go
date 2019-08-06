package common


type fnCacheFunc = func(
	ctx Context,
	stream *RPCStream,
	fn interface{},
) bool

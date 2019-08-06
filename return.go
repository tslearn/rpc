package common

// rpcReturn is rpc function return type
type rpcReturn struct{}
type Return = *rpcReturn

var nilReturn = Return(nil)

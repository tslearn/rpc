package common

import "github.com/rpccloud/rpc/common"

func getFCache(fn interface{}) fCacheFunc {
	stringKind, ok := getFuncKind(fn)
	if !ok {
		return nil
	}

	switch stringKind {
	case "B":
		return fcB
	}

	return nil
}

func fcB(c Context, s *common.RPCStream, f interface{}) bool {
	h, a := s.ReadBool()
	if !a && !s.IsReadFinish() {
		return false
	}
	f.(func(Context, bool))(c, h)
	return true
}

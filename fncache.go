package common

func getFCache(fn interface{}) fnCacheFunc {
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

func fcB(c Context, s *RPCStream, f interface{}) bool {
	h, a := s.ReadBool()
	if !a && !s.IsReadFinish() {
		return false
	}
	f.(func(Context, bool))(c, h)
	return true
}

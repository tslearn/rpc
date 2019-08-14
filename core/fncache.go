package core

func getFCache(fn interface{}) fnCacheFunc {
	stringKind, ok := getFuncKind(fn)
	if !ok {
		return nil
	}

	switch stringKind {
	case "I":
		return fcI
	case "S":
		return fcS
	}

	return nil
}

func fcI(c Context, s *rpcStream, f interface{}) bool {
	h, ok := s.ReadInt64()
	if !ok || !s.IsReadFinish() {
		return false
	}
	f.(func(Context, int64) Return)(c, h)
	return true
}

func fcS(c Context, s *rpcStream, f interface{}) bool {
	h, ok := s.ReadString()
	if !ok || !s.IsReadFinish() {
		return false
	}
	f.(func(Context, string) Return)(c, h)
	return true
}

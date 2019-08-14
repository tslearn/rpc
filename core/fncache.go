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

func fcI(c *rpcContext, s *rpcStream, f interface{}) bool {
	h, ok := s.ReadInt64()
	if !ok || !s.IsReadFinish() {
		return false
	}
	f.(func(*rpcContext, int64) *rpcReturn)(c, h)
	return true
}

func fcS(c *rpcContext, s *rpcStream, f interface{}) bool {
	h, ok := s.ReadString()
	if !ok || !s.IsReadFinish() {
		return false
	}
	f.(func(*rpcContext, string) *rpcReturn)(c, h)
	return true
}

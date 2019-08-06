package common

// PubContext ...
type PubContext struct {
	stream *RPCStream
}

// PubControl ...
type PubControl struct {
	ctx *PubContext
}

// OK ...
func (p *PubControl) OK() bool {
	return p == nil || p.ctx != nil
}

// HasContext ...
func (p *PubControl) HasContext() bool {
	return p != nil && p.ctx != nil
}

// Close ...
func (p *PubControl) Close() bool {
	if p.ctx != nil {
		p.ctx = nil
		return true
	}
	return false
}

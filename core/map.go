package core

import "sync"

var rpcMapInnerCache = sync.Pool{
	New: func() interface{} {
		return &rpcMapInner{
			smallMap: make([]rpcMapItem, 0, 16),
			largeMap: nil,
		}
	},
}

type rpcMapItem struct {
	key string
	pos int
}

type rpcMapInner struct {
	smallMap []rpcMapItem
	largeMap map[string]int
}

func (p *rpcMapInner) getItemPos(key string) int {
	if p.largeMap == nil {
		smallMap := p.smallMap
		for i := 0; i < len(smallMap); i++ {
			if smallMap[i].key == key {
				return smallMap[i].pos
			}
		}
	} else {
		if v, ok := p.largeMap[key]; ok {
			return v
		}
	}
	return -1
}

func (p *rpcMapInner) hasKey(key string) bool {
	if p.largeMap == nil {
		smallMap := p.smallMap
		for i := 0; i < len(smallMap); i++ {
			if smallMap[i].key == key {
				return true
			}
		}
		return false
	} else {
		_, ok := p.largeMap[key]
		return ok
	}
}

func (p *rpcMapInner) setItemPos(key string, idx int) bool {
	if p.largeMap == nil {
		smallMap := p.smallMap
		// find the name
		for i := 0; i < len(smallMap); i++ {
			if smallMap[i].key == key {
				smallMap[i].pos = idx
				return true
			}
		}

		// the name is not exist
		if len(smallMap) < 16 {
			p.smallMap = append(p.smallMap, rpcMapItem{
				key: key,
				pos: idx,
			})
		} else {
			p.toLargeMode()
			p.largeMap[key] = idx
		}
	} else {
		p.largeMap[key] = idx
	}

	return true
}

// Delete ...
func (p *rpcMapInner) deleteItem(key string) bool {
	if p.largeMap == nil {
		smallMap := p.smallMap
		for i := 0; i < len(smallMap); i++ {
			if smallMap[i].key == key {
				smallMap[i] = smallMap[len(smallMap)-1]
				p.smallMap = p.smallMap[:len(smallMap)-1]
				return true
			}
		}
	} else {
		if _, ok := p.largeMap[key]; ok {
			delete(p.largeMap, key)
			if len(p.largeMap) <= 16 {
				p.toSmallMode()
			}
			return true
		}
	}
	return false
}

func (p *rpcMapInner) free() {
	p.smallMap = p.smallMap[:0]
	p.largeMap = nil
	rpcMapInnerCache.Put(p)
}

func (p *rpcMapInner) toLargeMode() {
	if p.largeMap == nil {
		p.largeMap = make(map[string]int)
		for _, it := range p.smallMap {
			p.largeMap[it.key] = it.pos
		}
		p.smallMap = p.smallMap[:0]
	}
}

func (p *rpcMapInner) toSmallMode() {
	if p.largeMap != nil && len(p.largeMap) <= 16 {
		for key, value := range p.largeMap {
			p.smallMap = append(p.smallMap, rpcMapItem{
				key: key,
				pos: value,
			})
		}
		p.largeMap = nil
	}
}

// RPCMap ...
type rpcMap struct {
	ctx *rpcContext
	in  *rpcMapInner
}

func newRPCMap(ctx *rpcContext) rpcMap {
	if ctx.ok() {
		return rpcMap{
			ctx: ctx,
			in:  rpcMapInnerCache.Get().(*rpcMapInner),
		}
	}
	return nilRPCMap
}

func newRPCMapByMap(ctx *rpcContext, val Map) rpcMap {
	if val == nil {
		return nilRPCMap
	}
	ret := newRPCMap(ctx)
	for name, value := range val {
		if !ret.Set(name, value) {
			ret.release()
			return nilRPCMap
		}
	}
	return ret
}

func (p rpcMap) ok() bool {
	return p.in != nil && p.ctx.ok()
}

// Release ...
func (p *rpcMap) release() {
	if p.in != nil {
		p.in.free()
		p.in = nil
	}
	p.ctx = nil
}

func (p rpcMap) getIS() (*rpcMapInner, *rpcStream) {
	return p.in, p.ctx.getCacheStream()
}

// Size ...
func (p rpcMap) Size() int {
	if in := p.in; in != nil && p.ctx.ok() {
		if in.largeMap == nil {
			return len(in.smallMap)
		} else {
			return len(in.largeMap)
		}
	}
	return -1
}

// Keys ...
func (p rpcMap) Keys() []string {
	if in, _ := p.getIS(); in != nil {
		if in.largeMap == nil {
			ret := make([]string, 0, len(in.smallMap))
			for _, it := range in.smallMap {
				ret = append(ret, it.key)
			}
			return ret
		} else {
			ret := make([]string, 0, len(in.largeMap))
			for key := range in.largeMap {
				ret = append(ret, key)
			}
			return ret
		}
	}
	return []string{}
}

// GetNil ...
func (p rpcMap) GetNil(name string) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadNil()
		}
	}
	return false
}

// SetNil ...
func (p rpcMap) SetNil(name string) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteNil()
		return in.setItemPos(name, idx)
	}
	return false
}

// GetBool ...
func (p rpcMap) GetBool(name string) (bool, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadBool()
		}
	}
	return false, false
}

// SetBool ...
func (p rpcMap) SetBool(name string, value bool) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteBool(value)
		return in.setItemPos(name, idx)
	}
	return false
}

// GetFloat64 ...
func (p rpcMap) GetFloat64(name string) (float64, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadFloat64()
		}
	}
	return 0, false
}

// SetFloat64 ...
func (p rpcMap) SetFloat64(name string, value float64) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteFloat64(value)
		return in.setItemPos(name, idx)
	}
	return false
}

// GetInt64 ...
func (p rpcMap) GetInt64(name string) (int64, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadInt64()
		}
	}
	return 0, false
}

// SetInt64 ...
func (p rpcMap) SetInt64(name string, value int64) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteInt64(value)
		return in.setItemPos(name, idx)
	}
	return false
}

// GetUint64 ...
func (p rpcMap) GetUint64(name string) (uint64, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadUint64()
		}
	}
	return 0, false
}

// SetUint64 ...
func (p rpcMap) SetUint64(name string, value uint64) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteUint64(value)
		return in.setItemPos(name, idx)
	}
	return false
}

// GetRPCString ...
func (p rpcMap) GetString(name string) (string, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadString()
		}
	}
	return emptyString, false
}

// SetRPCString ...
func (p rpcMap) SetString(name string, value string) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteString(value)
		return in.setItemPos(name, idx)
	}
	return false
}

// GetRPCBytes ...
func (p rpcMap) GetBytes(name string) ([]byte, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadBytes()
		}
	}
	return emptyBytes, false
}

// SetRPCBytes ...
func (p rpcMap) SetBytes(name string, value []byte) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		s.WriteBytes(value)
		return in.setItemPos(name, idx)
	}
	return false
}

// GetRPCArray ...
func (p rpcMap) GetRPCArray(name string) (rpcArray, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadRPCArray(p.ctx)
		}
	}
	return nilRPCArray, false
}

// SetRPCArray ...
func (p rpcMap) SetRPCArray(name string, value rpcArray) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		if s.WriteRPCArray(value) == RPCStreamWriteOK {
			return in.setItemPos(name, idx)
		} else {
			return false
		}
	}
	return false
}

// GetRPCMap ...
func (p rpcMap) GetRPCMap(name string) (rpcMap, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadRPCMap(p.ctx)
		}
	}
	return nilRPCMap, false
}

// SetRPCMap ...
func (p rpcMap) SetRPCMap(name string, value rpcMap) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		if s.WriteRPCMap(value) == RPCStreamWriteOK {
			return in.setItemPos(name, idx)
		} else {
			return false
		}
	}
	return false
}

// Get ...
func (p rpcMap) Get(name string) (interface{}, bool) {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		if idx := in.getItemPos(name); idx > 0 {
			s.setReadPosUnsafe(idx)
			return s.ReadByContext(p.ctx)
		}
	}
	return nil, false
}

// Set ...
func (p rpcMap) Set(name string, value interface{}) bool {
	if in, s := p.getIS(); s != nil && in != nil && name != "" {
		idx := s.GetWritePos()
		if s.Write(value) == RPCStreamWriteOK {
			return in.setItemPos(name, idx)
		} else {
			return false
		}
	}
	return false
}

// Delete ...
func (p rpcMap) Delete(name string) bool {
	if in, _ := p.getIS(); in != nil {
		return in.deleteItem(name)
	}
	return false
}

func (p rpcMap) equals(val rpcMap) bool {
	if p.ctx == nil && p.in == nil && val.ctx == nil && val.in == nil {
		return true
	}
	if !p.ok() || !val.ok() {
		return false
	}
	if p.Size() != val.Size() {
		return false
	}

	for _, key := range p.Keys() {
		lv, ok := p.Get(key)
		if !ok {
			return false
		}
		rv, ok := val.Get(key)
		if !ok {
			return false
		}
		if !rpcEquals(lv, rv) {
			return false
		}
	}
	return true
}

func (p rpcMap) contains(val interface{}) bool {
	for _, key := range p.Keys() {
		lv, ok := p.Get(key)
		if !ok {
			return false
		}
		if rpcEquals(lv, val) {
			return true
		}
	}
	return false
}

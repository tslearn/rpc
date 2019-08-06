package common

import (
	"sync"
	"unsafe"
)

type rpcStatus uint8

const (
	rpcStatusError rpcStatus = iota
	rpcStatusAllocated
	rpcStatusNotAllocated
)

// RPCStreamWriteErrorCode ...
type RPCStreamWriteErrorCode int

const (
	RPCStreamWriteOK RPCStreamWriteErrorCode = iota
	RPCStreamWriteUnsupportedType
	//RPCStreamWriteRPCStringIsNotAvailable
	RPCStreamWriteRPCStringError
	//RPCStreamWriteRPCBytesIsNotAvailable
	RPCStreamWriteRPCBytesError
	RPCStreamWriteRPCArrayIsNotAvailable
	RPCStreamWriteRPCArrayError
	RPCStreamWriteRPCMapIsNotAvailable
	RPCStreamWriteRPCMapError
	RPCStreamWriteOverflow
)

// RPCInt64 ...
type RPCInt64 = int64

// RPCUint64 ...
type RPCUint64 = uint64

// RPCFloat64 ...
type RPCFloat64 = float64

// RPCBool ...
type RPCBool = bool

// RPCString ...
type RPCString struct {
	pub    *PubControl
	status rpcStatus
	bytes  []byte
}

// OK ...
func (p RPCString) OK() bool {
	return p.pub.OK() && p.status != rpcStatusError
}

// ToString ...
func (p RPCString) ToString() (string, bool) {
	if p.pub.OK() {
		if p.status == rpcStatusNotAllocated {
			return string(p.bytes), true
		} else if p.status == rpcStatusAllocated {
			return *(*string)(unsafe.Pointer(&p.bytes)), true
		}
	}
	return "", false
}

// RPCBytes ...
type RPCBytes struct {
	pub    *PubControl
	status rpcStatus
	bytes  []byte
}

// OK ...
func (p RPCBytes) OK() bool {
	return p.pub.OK() && p.status != rpcStatusError
}

// ToBytes ...
func (p RPCBytes) ToBytes() ([]byte, bool) {
	if p.pub.OK() {
		if p.status == rpcStatusNotAllocated {
			ret := make([]byte, len(p.bytes), len(p.bytes))
			copy(ret, p.bytes)
			return ret, true
		} else if p.status == rpcStatusAllocated {
			return p.bytes, true
		}
	}
	return nil, false
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// RPCArray
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var rpcArrayInnerCacheWithStream = sync.Pool{
	New: func() interface{} {
		return &rpcArrayInner{
			items:  make([]int, 0, 64),
			stream: NewRPCStream(),
			seed:   0,
		}
	},
}

var rpcArrayInnerCacheWithoutStream = sync.Pool{
	New: func() interface{} {
		return &rpcArrayInner{
			items:  make([]int, 0, 64),
			stream: nil,
			seed:   0,
		}
	},
}

type rpcArrayInner struct {
	items  []int
	stream *RPCStream
	seed   int
}

func (p *rpcArrayInner) free() {
	if cap(p.items) == 64 {
		p.items = p.items[:0]
	} else {
		p.items = make([]int, 0, 64)
	}

	p.seed++

	if p.stream == nil {
		rpcArrayInnerCacheWithoutStream.Put(p)
	} else {
		p.stream.Reset()
		rpcArrayInnerCacheWithStream.Put(p)
	}
}

// RPCArray ...
type RPCArray struct {
	pub  *PubControl
	in   *rpcArrayInner
	seed int
}

func newRPCArray(pub *PubControl) RPCArray {
	if pub.OK() {
		if pub.HasContext() {
			in := rpcArrayInnerCacheWithoutStream.Get().(*rpcArrayInner)
			return RPCArray{
				pub:  pub,
				in:   in,
				seed: in.seed,
			}
		} else {
			in := rpcArrayInnerCacheWithStream.Get().(*rpcArrayInner)
			return RPCArray{
				pub:  pub,
				in:   in,
				seed: in.seed,
			}
		}
	}
	return nilRPCArray
}

// Release ...
func (p *RPCArray) Release() {
	if p.in != nil && p.seed == p.in.seed {
		p.in.free()
	}

	p.in = nil
}

// OK ...
func (p *RPCArray) OK() bool {
	return p.in != nil && p.seed == p.in.seed && p.pub.OK()
}

// Size ...
func (p *RPCArray) Size() int {
	if p.OK() {
		return len(p.in.items)
	}
	return 0
}

func (p *RPCArray) getStream() *RPCStream {
	if p.in != nil && p.seed == p.in.seed {
		if p.pub == nil {
			return p.in.stream
		} else if p.pub.ctx != nil {
			return p.pub.ctx.stream
		} else {
			return nil
		}
	}
	return nil
}

// GetNil ...
func (p *RPCArray) GetNil(index int) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadNil()
	}
	return false
}

// SetNil ...
func (p *RPCArray) SetNil(index int) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		in.items[index] = stream.GetWritePos()
		stream.WriteNil()
		return true
	}
	return false
}

// AppendNil ...
func (p *RPCArray) AppendNil() bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		in.items = append(in.items, stream.GetWritePos())
		stream.WriteNil()
		return true
	}
	return false
}

// GetBool ...
func (p *RPCArray) GetBool(index int) (bool, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadBool()
	}
	return false, false
}

// SetBool ...
func (p *RPCArray) SetBool(index int, value bool) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		in.items[index] = stream.GetWritePos()
		stream.WriteBool(value)
		return true
	}
	return false
}

// AppendBool ...
func (p *RPCArray) AppendBool(value bool) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		in.items = append(in.items, stream.GetWritePos())
		stream.WriteBool(value)
		return true
	}
	return false
}

// GetInt64 ...
func (p *RPCArray) GetInt64(index int) (int64, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadInt64()
	}
	return 0, false
}

// SetInt64 ...
func (p *RPCArray) SetInt64(index int, value int64) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		in.items[index] = stream.GetWritePos()
		stream.WriteInt64(value)
		return true
	}
	return false
}

// AppendInt64 ...
func (p *RPCArray) AppendInt64(value int64) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		in.items = append(in.items, stream.GetWritePos())
		stream.WriteInt64(value)
		return true
	}
	return false
}

// GetUint64 ...
func (p *RPCArray) GetUint64(index int) (uint64, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadUint64()
	}
	return 0, false
}

// SetUint64 ...
func (p *RPCArray) SetUint64(index int, value uint64) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		in.items[index] = stream.GetWritePos()
		stream.WriteUint64(value)
		return true
	}
	return false
}

// AppendUint64 ...
func (p *RPCArray) AppendUint64(value uint64) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		in.items = append(in.items, stream.GetWritePos())
		stream.WriteUint64(value)
		return true
	}
	return false
}

// GetFloat64 ...
func (p *RPCArray) GetFloat64(index int) (float64, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadFloat64()
	}
	return 0, false
}

// SetFloat64 ...
func (p *RPCArray) SetFloat64(index int, value float64) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		in.items[index] = stream.GetWritePos()
		stream.WriteFloat64(value)
		return true
	}
	return false
}

// AppendFloat64 ...
func (p *RPCArray) AppendFloat64(value float64) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		in.items = append(in.items, stream.GetWritePos())
		stream.WriteFloat64(value)
		return true
	}
	return false
}

// GetRPCString ...
func (p *RPCArray) GetRPCString(index int) RPCString {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadRPCString(p.pub)
	}
	return errorRPCString
}

// SetRPCString ...
func (p *RPCArray) SetRPCString(index int, value RPCString) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		pos := stream.GetWritePos()
		if stream.WriteRPCString(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// AppendRPCString ...
func (p *RPCArray) AppendRPCString(value RPCString) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCString(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

// GetRPCBytes ...
func (p *RPCArray) GetRPCBytes(index int) RPCBytes {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadRPCBytes(p.pub)
	}
	return errorRPCBytes
}

// SetRPCBytes ...
func (p *RPCArray) SetRPCBytes(index int, value RPCBytes) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		pos := stream.GetWritePos()
		if stream.WriteRPCBytes(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// AppendRPCBytes ...
func (p *RPCArray) AppendRPCBytes(value RPCBytes) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCBytes(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

// GetRPCArray ...
func (p *RPCArray) GetRPCArray(index int) (RPCArray, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadRPCArray(p.pub)
	}
	return nilRPCArray, false
}

// SetRPCArray ...
func (p *RPCArray) SetRPCArray(index int, value RPCArray) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		pos := stream.GetWritePos()
		if stream.WriteRPCArray(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// AppendRPCArray ...
func (p *RPCArray) AppendRPCArray(value RPCArray) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCArray(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

// GetRPCMap ...
func (p *RPCArray) GetRPCMap(index int) (RPCMap, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.ReadRPCMap(p.pub)
	}
	return nilRPCMap, false
}

// SetRPCMap ...
func (p *RPCArray) SetRPCMap(index int, value RPCMap) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		pos := stream.GetWritePos()
		if stream.WriteRPCMap(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// AppendRPCMap ...
func (p *RPCArray) AppendRPCMap(value RPCMap) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCMap(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

// Get ...
func (p *RPCArray) Get(index int) (interface{}, bool) {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		stream.setReadPosUnsafe(in.items[index])
		return stream.Read(p.pub)
	}
	return nil, false
}

// Set ...
func (p *RPCArray) Set(index int, value interface{}) bool {
	in := p.in
	if stream := p.getStream(); stream != nil && index >= 0 && index < len(in.items) {
		pos := stream.GetWritePos()
		if stream.Write(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// Append ...
func (p *RPCArray) Append(value interface{}) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.Write(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// RPCMap
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var rpcMapInnerCacheWithStream = sync.Pool{
	New: func() interface{} {
		return &rpcMapInner{
			smallMap: make([]rpcMapItem, 0, 16),
			largeMap: nil,
			stream:   NewRPCStream(),
			seed:     0,
		}
	},
}

var rpcMapInnerCacheWithoutStream = sync.Pool{
	New: func() interface{} {
		return &rpcMapInner{
			smallMap: make([]rpcMapItem, 0, 16),
			largeMap: nil,
			stream:   nil,
			seed:     0,
		}
	},
}

type rpcMapItem struct {
	name string
	pos  int
}

type rpcMapInner struct {
	smallMap []rpcMapItem
	largeMap map[string]int
	stream   *RPCStream
	seed     int
}

func (p *rpcMapInner) free() {
	p.smallMap = p.smallMap[:0]
	p.largeMap = nil

	p.seed++

	if p.stream == nil {
		rpcMapInnerCacheWithoutStream.Put(p)
	} else {
		p.stream.Reset()
		rpcMapInnerCacheWithStream.Put(p)
	}
}

func (p *rpcMapInner) toLargeMode() {
	p.largeMap = make(map[string]int)
	for _, it := range p.smallMap {
		p.largeMap[it.name] = it.pos
	}
	p.smallMap = p.smallMap[:0]
}

func (p *rpcMapInner) toSmallMode() {
	for key, value := range p.largeMap {
		p.smallMap = append(p.smallMap, rpcMapItem{
			name: key,
			pos:  value,
		})
	}
	p.largeMap = nil
}

// RPCMap ...
type RPCMap struct {
	pub  *PubControl
	in   *rpcMapInner
	seed int
}

func newRPCMap(pub *PubControl) RPCMap {
	if pub.OK() {
		if pub.HasContext() {
			in := rpcMapInnerCacheWithoutStream.Get().(*rpcMapInner)
			return RPCMap{
				pub:  pub,
				in:   in,
				seed: in.seed,
			}
		}
		in := rpcMapInnerCacheWithStream.Get().(*rpcMapInner)
		return RPCMap{
			pub:  pub,
			in:   in,
			seed: in.seed,
		}
	}
	return nilRPCMap
}

// Release ...
func (p *RPCMap) Release() {
	if p.in != nil && p.seed == p.in.seed {
		p.in.free()
	}

	p.in = nil
}

// OK ...
func (p *RPCMap) OK() bool {
	return p.in != nil && p.seed == p.in.seed && p.pub.OK()
}

// Size ...
func (p *RPCMap) Size() int {
	if p.OK() {
		if p.in.largeMap == nil {
			return len(p.in.smallMap)
		}
		return len(p.in.largeMap)
	}
	return 0
}

func (p *RPCMap) getStream() *RPCStream {
	if p.in != nil && p.seed == p.in.seed {
		if p.pub == nil {
			return p.in.stream
		} else if p.pub.ctx != nil {
			return p.pub.ctx.stream
		} else {
			return nil
		}
	}
	return nil
}

// Keys ...
func (p *RPCMap) Keys() []string {
	if p.OK() {
		in := p.in
		if in.largeMap == nil {
			ret := make([]string, 0)
			for _, it := range in.smallMap {
				ret = append(ret, it.name)
			}
			return ret
		}
		ret := make([]string, 0)
		for key := range in.largeMap {
			ret = append(ret, key)
		}
		return ret
	}
	return []string{}
}

// GetNil ...
func (p *RPCMap) GetNil(name string) bool {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadNil()
		}
	}
	return false
}

// SetNil ...
func (p *RPCMap) SetNil(name string) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		stream.WriteNil()
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetBool ...
func (p *RPCMap) GetBool(name string) (bool, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadBool()
		}
	}
	return false, false
}

// SetBool ...
func (p *RPCMap) SetBool(name string, value bool) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		stream.WriteBool(value)
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetFloat64 ...
func (p *RPCMap) GetFloat64(name string) (float64, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadFloat64()
		}
	}
	return 0, false
}

// SetFloat64 ...
func (p *RPCMap) SetFloat64(name string, value float64) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		stream.WriteFloat64(value)
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetInt64 ...
func (p *RPCMap) GetInt64(name string) (int64, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadInt64()
		}
	}
	return 0, false
}

// SetInt64 ...
func (p *RPCMap) SetInt64(name string, value int64) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		stream.WriteInt64(value)
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetUint64 ...
func (p *RPCMap) GetUint64(name string) (uint64, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadUint64()
		}
	}
	return 0, false
}

// SetUint64 ...
func (p *RPCMap) SetUint64(name string, value uint64) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		stream.WriteUint64(value)
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetRPCString ...
func (p *RPCMap) GetRPCString(name string) RPCString {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadRPCString(p.pub)
		}
	}
	return errorRPCString
}

// SetRPCString ...
func (p *RPCMap) SetRPCString(name string, value RPCString) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCString(value) != RPCStreamWriteOK {
			return false
		}
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetRPCBytes ...
func (p *RPCMap) GetRPCBytes(name string) RPCBytes {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadRPCBytes(p.pub)
		}
	}
	return errorRPCBytes
}

// SetRPCBytes ...
func (p *RPCMap) SetRPCBytes(name string, value RPCBytes) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCBytes(value) != RPCStreamWriteOK {
			return false
		}
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetRPCArray ...
func (p *RPCMap) GetRPCArray(name string) (RPCArray, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadRPCArray(p.pub)
		}
	}
	return nilRPCArray, false
}

// SetRPCArray ...
func (p *RPCMap) SetRPCArray(name string, value RPCArray) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCArray(value) != RPCStreamWriteOK {
			return false
		}
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// GetRPCMap ...
func (p *RPCMap) GetRPCMap(name string) (RPCMap, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.ReadRPCMap(p.pub)
		}
	}
	return nilRPCMap, false
}

// SetRPCMap ...
func (p *RPCMap) SetRPCMap(name string, value RPCMap) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.WriteRPCMap(value) != RPCStreamWriteOK {
			return false
		}
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// Get ...
func (p *RPCMap) Get(name string) (interface{}, bool) {
	if stream := p.getStream(); stream != nil && name != "" {
		in := p.in
		idx := 0
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					idx = smallMap[i].pos
					break
				}
			}
		} else {
			if v, ok := in.largeMap[name]; ok {
				idx = v
			}
		}
		if idx > 0 {
			stream.setReadPosUnsafe(idx)
			return stream.Read(p.pub)
		}
	}
	return nil, false
}

// Set ...
func (p *RPCMap) Set(name string, value interface{}) bool {
	if stream := p.getStream(); stream != nil {
		in := p.in
		pos := stream.GetWritePos()
		if stream.Write(value) != RPCStreamWriteOK {
			return false
		}
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i].pos = pos
					return true
				}
			}

			if len(smallMap) < 16 {
				in.smallMap = append(in.smallMap, rpcMapItem{
					name: name,
					pos:  pos,
				})
			} else {
				in.toLargeMode()
				in.largeMap[name] = pos
			}
		} else {
			in.largeMap[name] = pos
		}
		return true
	}
	return false
}

// Delete ...
func (p *RPCMap) Delete(name string) bool {
	if p.OK() {
		in := p.in
		smallMap := in.smallMap
		if in.largeMap == nil {
			for i := 0; i < len(smallMap); i++ {
				if smallMap[i].name == name {
					smallMap[i] = smallMap[len(smallMap)-1]
					in.smallMap = in.smallMap[:len(smallMap)-1]
					return true
				}
			}
		} else {
			if _, ok := in.largeMap[name]; ok {
				delete(in.largeMap, name)
				if len(in.largeMap) == 16 {
					in.toSmallMode()
				}
				return true
			}
		}
	}
	return false
}

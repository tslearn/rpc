package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"math/rand"
	"testing"
)

var streamTestSuccessCollections = map[string][][2]interface{}{
	"nil": {
		{nil, []byte{0x01}},
	},
	"bool": {
		{true, []byte{0x02}},
		{false, []byte{0x03}},
	},
	"float64": {
		{float64(0), []byte{0x04}},
		{float64(100), []byte{
			0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0x40,
		}},
		{3.1415926, []byte{
			0x05, 0x4A, 0xD8, 0x12, 0x4D, 0xFB, 0x21, 0x09, 0x40,
		}},
		{-3.1415926, []byte{
			0x05, 0x4A, 0xD8, 0x12, 0x4D, 0xFB, 0x21, 0x09, 0xC0,
		}},
		{float64(-100), []byte{
			0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0xC0,
		}},
	},
	"int64": {
		{int64(-9223372036854775808), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}},
		{int64(-9007199254740992), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE0, 0x7F,
		}},
		{int64(-9007199254740991), []byte{
			0x08, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe0, 0x7F,
		}},
		{int64(-9007199254740990), []byte{
			0x08, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE0, 0x7F,
		}},
		{int64(-2147483649), []byte{
			0x08, 0xFF, 0xFF, 0xFF, 0x7F, 0xFF, 0xFF, 0xFF, 0x7F,
		}},
		{int64(-2147483648), []byte{0x07, 0x00, 0x00, 0x00, 0x00}},
		{int64(-32769), []byte{0x07, 0xFF, 0x7F, 0xFF, 0x7F}},
		{int64(-32768), []byte{0x06, 0x00, 0x00}},
		{int64(-8), []byte{0x06, 0xF8, 0x7F}},
		{int64(-7), []byte{0x0E}},
		{int64(-1), []byte{0x14}},
		{int64(0), []byte{0x15}},
		{int64(1), []byte{0x16}},
		{int64(32), []byte{0x35}},
		{int64(33), []byte{0x06, 0x21, 0x80}},
		{int64(32767), []byte{0x06, 0xFF, 0xFF}},
		{int64(32768), []byte{0x07, 0x00, 0x80, 0x00, 0x80}},
		{int64(2147483647), []byte{0x07, 0xFF, 0xFF, 0xFF, 0xFF}},
		{int64(2147483648), []byte{
			0x08, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x80,
		}},
		{int64(9007199254740990), []byte{
			0x08, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x80,
		}},
		{int64(9007199254740991), []byte{
			0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x80,
		}},
		{int64(9007199254740992), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x80,
		}},
		{int64(9223372036854775807), []byte{
			0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}},
	},
	"uint64": {
		{uint64(0), []byte{0x36}},
		{uint64(9), []byte{0x3F}},
		{uint64(10), []byte{0x09, 0x0A, 0x00}},
		{uint64(65535), []byte{0x09, 0xFF, 0xFF}},
		{uint64(65536), []byte{0x0A, 0x00, 0x00, 0x01, 0x00}},
		{uint64(4294967295), []byte{0x0A, 0xFF, 0xFF, 0xFF, 0xFF}},
		{uint64(4294967296), []byte{
			0x0B, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		}},
		{uint64(9007199254740990), []byte{
			0x0B, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x00,
		}},
		{uint64(9007199254740991), []byte{
			0x0B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x00,
		}},
		{uint64(9007199254740992), []byte{
			0x0B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x00,
		}},
		{uint64(18446744073709551615), []byte{
			0x0B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}},
	},
	"string": {
		{"", []byte{0x80}},
		{"a", []byte{0x81, 0x61, 0x00}},
		{"😀☘️🀄️©️🌈🎩", []byte{
			0x9E, 0xF0, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
			0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
			0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
			0xA9, 0x00,
		}},
		{"😀中☘️文🀄️©️🌈🎩测试a\n\r\b", []byte{
			0xAE, 0xF0, 0x9F, 0x98, 0x80, 0xE4, 0xB8, 0xAD, 0xE2, 0x98,
			0x98, 0xEF, 0xB8, 0x8F, 0xE6, 0x96, 0x87, 0xF0, 0x9F, 0x80,
			0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0,
			0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xE6, 0xB5, 0x8B,
			0xE8, 0xAF, 0x95, 0x61, 0x0A, 0x0D, 0x08, 0x00,
		}},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", []byte{
			0xBF, 0x45, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		}},
		{"😀☘️🀄️©️🌈🎩😛👩‍👩‍👦👨‍👩‍👦‍👦👼🗣👑👚👹👺🌳🍊", []byte{
			0xBF, 0x73, 0x00, 0x00, 0x00, 0xF0, 0x9F, 0x98, 0x80, 0xE2,
			0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
			0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
			0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
			0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9,
			0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
			0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80,
			0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xE2, 0x80, 0x8D, 0xF0, 0x9F,
			0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xBC, 0xF0, 0x9F, 0x97, 0xA3,
			0xF0, 0x9F, 0x91, 0x91, 0xF0, 0x9F, 0x91, 0x9A, 0xF0, 0x9F,
			0x91, 0xB9, 0xF0, 0x9F, 0x91, 0xBA, 0xF0, 0x9F, 0x8C, 0xB3,
			0xF0, 0x9F, 0x8D, 0x8A, 0x00,
		}},
	},
	"bytes": {
		{([]byte)(nil), []byte{0x01}},
		{[]byte{}, []byte{0xC0}},
		{[]byte{0xDA}, []byte{0xC1, 0xDA}},
		{[]byte{
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61,
		}, []byte{
			0xFF, 0x44, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		}},
	},
	"array": {
		{Array(nil), []byte{0x01}},
		{Array{}, []byte{64}},
		{Array{true}, []byte{
			65, 6, 0, 0, 0, 2,
		}},
		{Array{
			true, false,
		}, []byte{
			66, 7, 0, 0, 0, 2, 3,
		}},
		{Array{
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
		}, []byte{
			94, 35, 0, 0, 0, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2,
		}},
		{Array{
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true,
		}, []byte{
			95, 40, 0, 0, 0, 31, 0, 0, 0, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		}},
		{Array{
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true,
		}, []byte{
			95, 41, 0, 0, 0, 32, 0, 0, 0, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2,
		}},
	},
	"map": {
		//{Map(nil), []byte{0x01}},
		//{Map{}, []byte{0x60}},
		{Map{"1": true}, []byte{
			0x61, 0x09, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02,
		}},
		//{Map{
		//	"1": true, "2": true, "3": true, "4": true,
		//	"5": true, "6": true, "7": true, "8": true,
		//	"9": true, "a": true, "b": true, "c": true,
		//	"d": true, "e": true, "f": true, "g": true,
		//	"h": true, "i": true, "j": true, "k": true,
		//	"l": true, "m": true, "n": true, "o": true,
		//	"p": true, "q": true, "r": true, "s": true,
		//	"t": true, "u": true,
		//}, []byte{
		//	0x7E, 0x7D, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x81,
		//	0x32, 0x00, 0x02, 0x81, 0x33, 0x00, 0x02, 0x81, 0x34, 0x00,
		//	0x02, 0x81, 0x35, 0x00, 0x02, 0x81, 0x36, 0x00, 0x02, 0x81,
		//	0x37, 0x00, 0x02, 0x81, 0x38, 0x00, 0x02, 0x81, 0x39, 0x00,
		//	0x02, 0x81, 0x61, 0x00, 0x02, 0x81, 0x62, 0x00, 0x02, 0x81,
		//	0x63, 0x00, 0x02, 0x81, 0x64, 0x00, 0x02, 0x81, 0x65, 0x00,
		//	0x02, 0x81, 0x66, 0x00, 0x02, 0x81, 0x67, 0x00, 0x02, 0x81,
		//	0x68, 0x00, 0x02, 0x81, 0x69, 0x00, 0x02, 0x81, 0x6A, 0x00,
		//	0x02, 0x81, 0x6B, 0x00, 0x02, 0x81, 0x6C, 0x00, 0x02, 0x81,
		//	0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00, 0x02, 0x81, 0x6F, 0x00,
		//	0x02, 0x81, 0x70, 0x00, 0x02, 0x81, 0x71, 0x00, 0x02, 0x81,
		//	0x72, 0x00, 0x02, 0x81, 0x73, 0x00, 0x02, 0x81, 0x74, 0x00,
		//	0x02, 0x81, 0x75, 0x00, 0x02,
		//}},
		//{Map{
		//	"1": true, "2": true, "3": true, "4": true,
		//	"5": true, "6": true, "7": true, "8": true,
		//	"9": true, "a": true, "b": true, "c": true,
		//	"d": true, "e": true, "f": true, "g": true,
		//	"h": true, "i": true, "j": true, "k": true,
		//	"l": true, "m": true, "n": true, "o": true,
		//	"p": true, "q": true, "r": true, "s": true,
		//	"t": true, "u": true, "v": true,
		//}, []byte{
		//	0x7F, 0x85, 0x00, 0x00, 0x00, 0x1F, 0x00, 0x00, 0x00, 0x81,
		//	0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00,
		//	0x02, 0x81, 0x34, 0x00, 0x02, 0x81, 0x35, 0x00, 0x02, 0x81,
		//	0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00,
		//	0x02, 0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81,
		//	0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02, 0x81, 0x64, 0x00,
		//	0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81,
		//	0x67, 0x00, 0x02, 0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00,
		//	0x02, 0x81, 0x6A, 0x00, 0x02, 0x81, 0x6B, 0x00, 0x02, 0x81,
		//	0x6C, 0x00, 0x02, 0x81, 0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00,
		//	0x02, 0x81, 0x6F, 0x00, 0x02, 0x81, 0x70, 0x00, 0x02, 0x81,
		//	0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00,
		//	0x02, 0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81,
		//	0x76, 0x00, 0x02,
		//}},
		//{Map{
		//	"1": true, "2": true, "3": true, "4": true,
		//	"5": true, "6": true, "7": true, "8": true,
		//	"9": true, "a": true, "b": true, "c": true,
		//	"d": true, "e": true, "f": true, "g": true,
		//	"h": true, "i": true, "j": true, "k": true,
		//	"l": true, "m": true, "n": true, "o": true,
		//	"p": true, "q": true, "r": true, "s": true,
		//	"t": true, "u": true, "v": true, "w": true,
		//}, []byte{
		//	0x7F, 0x89, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x81,
		//	0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00,
		//	0x02, 0x81, 0x34, 0x00, 0x02, 0x81, 0x35, 0x00, 0x02, 0x81,
		//	0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00,
		//	0x02, 0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81,
		//	0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02, 0x81, 0x64, 0x00,
		//	0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81,
		//	0x67, 0x00, 0x02, 0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00,
		//	0x02, 0x81, 0x6A, 0x00, 0x02, 0x81, 0x6B, 0x00, 0x02, 0x81,
		//	0x6C, 0x00, 0x02, 0x81, 0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00,
		//	0x02, 0x81, 0x6F, 0x00, 0x02, 0x81, 0x70, 0x00, 0x02, 0x81,
		//	0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00,
		//	0x02, 0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81,
		//	0x76, 0x00, 0x02, 0x81, 0x77, 0x00, 0x02,
		//}},
	},
}

func getTestRange(start uint, end uint, head uint, tail uint, step uint) []int {
	ret := make([]int, 0)
	for i := start; i <= end; i++ {
		pos := i % streamBlockSize
		if i < start+head {
			ret = append(ret, int(i))
		} else if pos < head {
			ret = append(ret, int(i))
		} else if pos > streamBlockSize-tail {
			ret = append(ret, int(i))
		} else {
			if i%step == 0 {
				ret = append(ret, int(i))
			}
		}
	}
	return ret
}

func TestStream(t *testing.T) {
	t.Run("test constant", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(streamVersion).Equal(1)
		assert(streamBlockSize).Equal(512)
		assert(streamFrameArrayInitSize).Equal(4)
		assert(streamPosVersion).Equal(0)
		assert(streamPosStatusBit).Equal(1)
		assert(streamPosTargetID).Equal(2)
		assert(streamPosSourceID).Equal(10)
		assert(streamPosZoneID).Equal(18)
		assert(streamPosIPMap).Equal(20)
		assert(streamPosSessionID).Equal(28)
		assert(streamPosCallbackID).Equal(36)
		assert(streamPosDepth).Equal(44)
		assert(streamPosCheck).Equal(46)
		assert(streamPosBody).Equal(54)
		assert(streamStatusBitDebug).Equal(0)
		assert(StreamWriteOK).Equal("")
	})

	t.Run("test initStreamFrame0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(len(initStreamFrame0)).Equal(streamBlockSize)
		assert(cap(initStreamFrame0)).Equal(streamBlockSize)
		assert(initStreamFrame0[0]).Equal(byte(streamVersion))
		for i := 1; i < streamBlockSize; i++ {
			assert(initStreamFrame0[i]).Equal(uint8(0))
		}
	})

	t.Run("test initStreamFrameN", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(len(initStreamFrameN)).Equal(streamBlockSize)
		assert(cap(initStreamFrameN)).Equal(streamBlockSize)
		for i := 0; i < streamBlockSize; i++ {
			assert(initStreamFrameN[i]).Equal(uint8(0))
		}
	})

	t.Run("test streamCache", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := streamCache.Get().(*Stream)
		assert(len(stream.frames)).Equal(1)
		assert(cap(stream.frames)).Equal(streamFrameArrayInitSize)
		assert(stream.readSeg).Equal(0)
		assert(stream.readIndex).Equal(streamPosBody)
		assert(stream.readFrame).Equal(*(stream.frames[0]))
		assert(stream.writeSeg).Equal(0)
		assert(stream.writeIndex).Equal(streamPosBody)
		assert(stream.writeFrame).Equal(*(stream.frames[0]))
		assert(len(stream.bufferBytes)).Equal(streamBlockSize)
		assert(cap(stream.bufferBytes)).Equal(streamBlockSize)
		assert(len(stream.bufferFrames)).Equal(streamFrameArrayInitSize)
		assert(cap(stream.bufferFrames)).Equal(streamFrameArrayInitSize)
		assert(*stream.frames[0]).Equal(stream.bufferBytes[0:])
		streamCache.Put(stream)
	})

	t.Run("test frameCache", func(t *testing.T) {
		assert := base.NewAssert(t)
		frame := frameCache.Get().(*[]byte)
		assert(len(*frame)).Equal(streamBlockSize)
		assert(cap(*frame)).Equal(streamBlockSize)
		frameCache.Put(frame)
	})

	t.Run("test readSkipArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(readSkipArray).Equal([256]int{
			-0x01, +0x01, +0x01, +0x01, +0x01, +0x09, +0x03, +0x05,
			+0x09, +0x03, +0x05, +0x09, -0x01, -0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x01, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08, +0x09,
			+0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10, +0x11,
			+0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18, +0x19,
			+0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20, +0x21,
			+0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28, +0x29,
			+0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30, +0x31,
			+0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38, +0x39,
			+0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x40, +0x00,
			+0x01, +0x02, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08,
			+0x09, +0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10,
			+0x11, +0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18,
			+0x19, +0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20,
			+0x21, +0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28,
			+0x29, +0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30,
			+0x31, +0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38,
			+0x39, +0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x00,
		})
	})
}

func TestStream_Reset(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()

		for i := 0; i < streamFrameArrayInitSize*(streamBlockSize+2); i++ {
			assert(len(stream.frames)).Equal(1)
			assert(cap(stream.frames)).Equal(streamFrameArrayInitSize)
			assert(stream.readSeg).Equal(0)
			assert(stream.readIndex).Equal(streamPosBody)
			assert(stream.readFrame).Equal(*stream.frames[0])
			assert(stream.writeSeg).Equal(0)
			assert(stream.writeIndex).Equal(streamPosBody)
			assert(stream.writeFrame).Equal(*stream.frames[0])
			assert(*stream.frames[0]).Equal(initStreamFrame0)

			for j := 0; j < i; j++ {
				stream.PutBytes([]byte{12})
				if stream.writeIndex == 0 {
					assert(stream.writeFrame).Equal(initStreamFrameN)
				}
			}
			stream.SetReadPos(i)
			stream.Reset()
		}

		stream.Release()
	})
}

func TestStream_Release(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		mp := map[string]bool{}
		for i := 0; i < 1000; i++ {
			v := NewStream()
			mp[fmt.Sprintf("%p", v)] = true
			v.Release()
		}
		assert(len(mp) < 1000).IsTrue()
	})
}

func TestStream_GetVersion(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		assert(v.GetVersion()).Equal(uint8(1))
		v.Release()
	})
}

func TestStream_SetVersion(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			v.SetVersion(uint8(i))
			assert(v.GetVersion()).Equal(uint8(i))
			v.Release()
		}
	})
}

func TestStream_HasStatusBitDebug(t *testing.T) {
	t.Run("test bit set", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			v.SetStatusBitDebug()
			assert(v.HasStatusBitDebug()).IsTrue()
			v.Release()
		}
	})

	t.Run("test bit unset", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			v.ClearStatusBitDebug()
			assert(v.HasStatusBitDebug()).IsFalse()
			v.Release()
		}
	})
}

func TestStream_SetStatusBitDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			if !v.HasStatusBitDebug() {
				v.SetStatusBitDebug()
				assert(v.HasStatusBitDebug()).IsTrue()
				v.ClearStatusBitDebug()
			}
			assert((*v.frames[0])[streamPosStatusBit]).Equal(byte(i))
			v.Release()
		}
	})
}

func TestStream_ClearStatusBitDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			if v.HasStatusBitDebug() {
				v.ClearStatusBitDebug()
				assert(v.HasStatusBitDebug()).IsFalse()
				v.SetStatusBitDebug()
			}
			assert((*v.frames[0])[streamPosStatusBit]).Equal(byte(i))
			v.Release()
		}
	})
}

func TestStream_GetTargetID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetTargetID(id)
			assert(v.GetTargetID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_SetTargetID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetTargetID(id)
			assert(v.GetTargetID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_GetSourceID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSourceID(id)
			assert(v.GetSourceID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_SetSourceID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSourceID(id)
			assert(v.GetSourceID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_GetZoneID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := uint16(i)
			v.SetZoneID(id)
			assert(v.GetZoneID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_SetZoneID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := uint16(i)
			v.SetZoneID(id)
			assert(v.GetZoneID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_GetIPMap(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			ipMap := rand.Uint64()
			v.SetIPMap(ipMap)
			assert(v.GetIPMap()).Equal(ipMap)
			v.Release()
		}
	})
}

func TestStream_SetIPMap(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			ipMap := rand.Uint64()
			v.SetIPMap(ipMap)
			assert(v.GetIPMap()).Equal(ipMap)
			v.Release()
		}
	})
}

func TestStream_GetSessionID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSessionID(id)
			assert(v.GetSessionID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_SetSessionID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSessionID(id)
			assert(v.GetSessionID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_GetCallbackID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetCallbackID(id)
			assert(v.GetCallbackID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_SetCallbackID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetCallbackID(id)
			assert(v.GetCallbackID()).Equal(id)
			v.Release()
		}
	})
}

func TestStream_GetDepth(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			depth := uint16(i)
			v.SetDepth(depth)
			assert(v.GetDepth()).Equal(depth)
			v.Release()
		}
	})
}

func TestStream_SetDepth(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			depth := uint16(i)
			v.SetDepth(depth)
			assert(v.GetDepth()).Equal(depth)
			v.Release()
		}
	})
}

func TestStream_GetReadPos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i <= 3*streamBlockSize; i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.GetReadPos()).Equal(streamPosBody)
			} else {
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.GetReadPos()).Equal(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetReadPos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 3*streamBlockSize; i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.SetReadPos(i)).IsFalse()
				assert(stream.GetReadPos()).Equal(streamPosBody)
			} else if i == streamPosBody {
				stream.SetWritePos(i)
				assert(stream.SetReadPos(i)).IsTrue()
				assert(stream.GetReadPos()).Equal(i)
				assert(stream.SetReadPos(i + 1)).IsFalse()
				assert(stream.GetReadPos()).Equal(i)
			} else {
				stream.SetWritePos(i)
				assert(stream.SetReadPos(i - 1)).IsTrue()
				assert(stream.GetReadPos()).Equal(i - 1)
				assert(stream.SetReadPos(i)).IsTrue()
				assert(stream.GetReadPos()).Equal(i)
				assert(stream.SetReadPos(i + 1)).IsFalse()
				assert(stream.GetReadPos()).Equal(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetReadPosToBodyStart(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 3*streamBlockSize; i++ {
			stream := NewStream()
			if i >= streamPosBody {
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.GetReadPos()).Equal(i)
			}
			stream.SetReadPosToBodyStart()
			assert(stream.GetReadPos()).Equal(streamPosBody)
			stream.Release()
		}
	})
}

func TestStream_GetWritePos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i <= 3*streamBlockSize; i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.GetWritePos()).Equal(streamPosBody)
			} else {
				stream.SetWritePos(i)
				assert(stream.GetWritePos()).Equal(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetWritePos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < streamFrameArrayInitSize*(streamBlockSize+2); i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.SetWritePos(i)).IsFalse()
				assert(stream.GetWritePos()).Equal(streamPosBody)
			} else {
				assert(stream.SetWritePos(i)).IsTrue()
				assert(stream.GetWritePos()).Equal(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetWritePosToBodyStart(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < streamFrameArrayInitSize*(streamBlockSize+2); i++ {
			stream := NewStream()
			if i >= streamPosBody {
				stream.SetWritePos(i)
				assert(stream.GetWritePos()).Equal(i)
			}
			stream.SetWritePosToBodyStart()
			assert(stream.GetWritePos()).Equal(streamPosBody)
			stream.Release()
		}
	})
}

func TestStream_CanRead(t *testing.T) {
	t.Run("test can not read", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			assert(stream.SetReadPos(i)).IsTrue()
			assert(stream.CanRead()).IsFalse()
			if (i+1)%streamBlockSize != 0 {
				stream.readIndex = (i + 1) % streamBlockSize
				assert(stream.CanRead()).IsFalse()
			}
			stream.Release()
		}
	})

	t.Run("test can read", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i + 1)
			stream.SetReadPos(i)
			assert(stream.CanRead()).IsTrue()
			stream.Release()
		}
	})
}

func TestStream_IsReadFinish(t *testing.T) {
	t.Run("test read is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := streamPosBody + 1; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i - 1)
			assert(stream.IsReadFinish()).IsFalse()
			if (i+1)%streamBlockSize != 0 {
				stream.readIndex = (i + 1) % streamBlockSize
				assert(stream.IsReadFinish()).IsFalse()
			}
			stream.Release()
		}
	})

	t.Run("test read is finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			assert(stream.IsReadFinish()).IsTrue()
			stream.Release()
		}
	})
}

func TestStream_gotoNextWriteFrame(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < streamBlockSize*(streamFrameArrayInitSize+2); i++ {
			stream := NewStream()
			if i >= streamPosBody {
				stream.SetWritePos(i)
			}

			currSeg := stream.writeSeg
			assert(stream.writeSeg).Equal(len(stream.frames) - 1)
			stream.gotoNextWriteFrame()
			assert(stream.writeSeg).Equal(len(stream.frames) - 1)
			assert(stream.writeSeg).Equal(currSeg + 1)
			stream.Release()
		}
	})
}

func TestStream_gotoNextReadFrameUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		stream := NewStream()
		stream.SetWritePos(streamBlockSize * (streamFrameArrayInitSize + 3))
		for i := streamPosBody; i < testEnd; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			stream.gotoNextReadFrameUnsafe()
			assert(stream.GetReadPos()).
				Equal((i/streamBlockSize + 1) * streamBlockSize)
		}
		stream.Release()
	})
}

func TestStream_gotoNextReadByteUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		stream := NewStream()
		stream.SetWritePos(streamBlockSize * (streamFrameArrayInitSize + 3))
		for i := streamPosBody; i < testEnd; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			stream.gotoNextReadByteUnsafe()
			assert(stream.GetReadPos()).Equal(i + 1)
		}
		stream.Release()
	})
}

func TestStream_hasNBytesToRead(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetWritePos(3 * streamBlockSize)
		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			for n := 0; n < 3*streamBlockSize; n++ {
				assert(stream.hasNBytesToRead(n)).Equal(i+n <= 3*streamBlockSize)
			}
		}
		stream.Release()
	})
}

func TestStream_isSafetyReadNBytesInCurrentFrame(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetWritePos(3 * streamBlockSize)
		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			for n := 0; n < 2*streamBlockSize; n++ {
				assert(stream.isSafetyReadNBytesInCurrentFrame(n)).
					Equal(streamBlockSize-i%streamBlockSize > n)
			}
		}
		stream.Release()
	})
}

func TestStream_peekNBytesCrossFrameUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(n int) {
			if n > streamBlockSize || n <= 0 {
				panic("error")
			}

			buf := make([]byte, n)
			for i := 0; i < n; i++ {
				buf[i] = byte(i)
			}

			for i := 2*streamBlockSize - n; i < 2*streamBlockSize; i++ {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes(buf)
				assert(stream.peekNBytesCrossFrameUnsafe(n)).Equal(buf)
				assert(stream.GetReadPos()).Equal(i)
				stream.Release()
			}
		}

		for _, i := range getTestRange(1, streamBlockSize, 12, 1, 93) {
			fnTest(i)
		}
	})
}

func TestStream_readNBytesCrossFrameUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(n int) {
			if n > streamBlockSize || n <= 0 {
				panic("error")
			}

			buf := make([]byte, n)
			for i := 0; i < n; i++ {
				buf[i] = byte(i)
			}

			for i := 2*streamBlockSize - n; i < 2*streamBlockSize; i++ {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes(buf)
				assert(stream.readNBytesCrossFrameUnsafe(n)).Equal(buf)
				assert(stream.GetReadPos()).Equal(i + n)
				stream.Release()
			}
		}

		for _, i := range getTestRange(1, streamBlockSize, 12, 1, 93) {
			fnTest(i)
		}
	})
}

func TestRpcStream_peekSkip(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)

		testCollection := Array{
			Array{[]byte{0}, 0, byte(0)},
			Array{[]byte{1}, 1, byte(1)},
			Array{[]byte{2}, 1, byte(2)},
			Array{[]byte{3}, 1, byte(3)},
			Array{[]byte{4}, 1, byte(4)},
			Array{[]byte{5}, 9, byte(5)},
			Array{[]byte{6}, 3, byte(6)},
			Array{[]byte{7}, 5, byte(7)},
			Array{[]byte{8}, 9, byte(8)},
			Array{[]byte{9}, 3, byte(9)},
			Array{[]byte{10}, 5, byte(10)},
			Array{[]byte{11}, 9, byte(11)},
			Array{[]byte{12}, 0, byte(0)},
			Array{[]byte{13}, 0, byte(0)},
			Array{[]byte{14}, 1, byte(14)},
			Array{[]byte{63}, 1, byte(63)},
			Array{[]byte{64}, 1, byte(64)},
			Array{[]byte{65, 6, 0, 0, 0}, 6, byte(65)},
			Array{[]byte{94, 6, 0, 0, 0}, 6, byte(94)},
			Array{[]byte{95, 6, 0, 0, 0}, 6, byte(95)},
			Array{[]byte{96, 6, 0, 0, 0}, 1, byte(96)},
			Array{[]byte{97, 6, 0, 0, 0}, 6, byte(97)},
			Array{[]byte{126, 6, 0, 0, 0}, 6, byte(126)},
			Array{[]byte{127, 6, 0, 0, 0}, 6, byte(127)},
			Array{[]byte{128, 6, 0, 0, 0}, 1, byte(128)},
			Array{[]byte{129, 6, 0, 0, 0}, 3, byte(129)},
			Array{[]byte{190, 6, 0, 0, 0}, 64, byte(190)},
			Array{[]byte{191, 80, 0, 0, 0}, 80, byte(191)},
			Array{[]byte{192, 6, 0, 0, 0}, 1, byte(192)},
			Array{[]byte{193, 6, 0, 0, 0}, 2, byte(193)},
			Array{[]byte{254, 6, 0, 0, 0}, 63, byte(254)},
			Array{[]byte{255, 80, 0, 0, 0}, 80, byte(255)},
			Array{[]byte{255, 80, 0}, 0, byte(0)},
		}

		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			for _, item := range testCollection {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes(item.(Array)[0].([]byte))
				assert(stream.peekSkip()).Equal(item.(Array)[1], item.(Array)[2])
				assert(stream.GetReadPos()).Equal(i)
				stream.Release()
			}
		}
	})
}

func TestRpcStream_readSkipItem(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 2*streamBlockSize, 16, 16, 61)
		for _, i := range testRange {
			for j := 0; j < 2*streamBlockSize; j++ {
				// skip > 0
				bytes := make([]byte, j, j)
				for n := 0; n < j; n++ {
					bytes[n] = byte(n)
				}
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.WriteBytes(bytes)
				assert(stream.readSkipItem(stream.GetWritePos() - 1)).Equal(-1)
				assert(stream.GetReadPos()).Equal(i)
				assert(stream.readSkipItem(stream.GetWritePos())).Equal(i)
				assert(stream.GetReadPos()).Equal(stream.GetWritePos())

				// skip == 0
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes([]byte{13})
				assert(stream.readSkipItem(stream.GetWritePos())).Equal(-1)

				stream.Release()
			}
		}
	})
}

func TestRpcStream_writeStreamNext(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 2*streamBlockSize, 16, 16, 61)

		for _, i := range testRange {
			bytes := make([]byte, i, i)
			dataStream := NewStream()
			for n := 0; n < i; n++ {
				bytes[n] = byte(n)
			}
			dataStream.WriteBytes(bytes)

			// invalid code
			bugStream0 := NewStream()
			bugStream0.PutBytes([]byte{13})

			// length overflow
			bugStream1 := NewStream()
			bugStream1.PutBytes([]byte{65, 6, 0, 0, 0})

			for j := streamPosBody; j < streamBlockSize+20; j++ {
				stream := NewStream()
				stream.SetWritePos(j)
				dataStream.SetReadPos(streamPosBody)

				// dataStream
				assert(stream.writeStreamNext(dataStream)).IsTrue()
				assert(dataStream.GetReadPos()).Equal(dataStream.GetWritePos())
				assert(stream.GetWritePos()).
					Equal(dataStream.GetWritePos() + j - streamPosBody)
				// bugStream0
				assert(stream.writeStreamNext(bugStream0)).IsFalse()
				assert(bugStream0.GetReadPos()).Equal(streamPosBody)
				assert(stream.GetWritePos()).
					Equal(dataStream.GetWritePos() + j - streamPosBody)
				// bugStream1
				assert(stream.writeStreamNext(bugStream1)).IsFalse()
				assert(bugStream1.GetReadPos()).Equal(streamPosBody)
				assert(stream.GetWritePos()).
					Equal(dataStream.GetWritePos() + j - streamPosBody)

				stream.Release()
			}
		}
	})
}

func TestStream_GetBuffer(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		bytes := make([]byte, testEnd)
		for n := 0; n < testEnd; n++ {
			bytes[n] = byte(n)
		}
		for i := 0; i < testEnd; i++ {
			stream := NewStream()
			stream.PutBytes(bytes[:i])
			assert(stream.GetBuffer()[streamPosBody:]).Equal(bytes[:i])
			stream.Release()
		}
	})
}

func TestStream_GetBufferUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		bytes := make([]byte, testEnd)
		for n := 0; n < testEnd; n++ {
			bytes[n] = byte(n)
		}
		for i := 0; i < testEnd; i++ {
			stream := NewStream()
			stream.PutBytes(bytes[:i])
			assert(stream.GetBufferUnsafe()[streamPosBody:]).Equal(bytes[:i])
			stream.Release()
		}
	})
}

func TestStream_PutBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 10, 10, 93)
		for _, i := range testRange {
			for _, n := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				bytes := make([]byte, n)
				for z := 0; z < n; z++ {
					bytes[z] = byte(z)
				}
				stream.PutBytes(bytes)
				assert(stream.GetBuffer()[i:]).Equal(bytes)
				stream.Release()
			}
		}
	})
}

func TestStream_PutBytesTo(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(0, 3*streamBlockSize, 10, 10, 93)

		for _, i := range testRange {
			bytes := make([]byte, i)
			for n := 0; n < i; n++ {
				bytes[n] = byte(n)
			}

			for _, j := range testRange {
				stream := NewStream()
				if i+j < streamPosBody {
					assert(stream.PutBytesTo(bytes, j)).IsFalse()
				} else {
					assert(stream.PutBytesTo(bytes, j)).IsTrue()
					assert(stream.GetBuffer()[j:]).Equal(bytes)
					assert(stream.GetWritePos()).Equal(i + j)
				}

				stream.Release()
			}
		}
	})
}

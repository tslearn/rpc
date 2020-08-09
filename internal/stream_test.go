package internal

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

var streamTestCollections = map[string][][2]interface{}{
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
			0xBF, 0x3F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		}},
		{"😀☘️🀄️©️🌈🎩😛👩‍👩‍👦👨‍👩‍👦‍👦👼🗣👑👚👹👺🌳🍊", []byte{
			0xBF, 0x6D, 0x00, 0x00, 0x00, 0xF0, 0x9F, 0x98, 0x80, 0xE2,
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
			0xFF, 0x3F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
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
		{Map(nil), []byte{0x01}},
		{Map{}, []byte{0x60}},
		{Map{"1": true}, []byte{
			0x61, 0x09, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02,
		}},
		{Map{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true,
		}, []byte{
			0x7E, 0x7D, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x81,
			0x32, 0x00, 0x02, 0x81, 0x33, 0x00, 0x02, 0x81, 0x34, 0x00,
			0x02, 0x81, 0x35, 0x00, 0x02, 0x81, 0x36, 0x00, 0x02, 0x81,
			0x37, 0x00, 0x02, 0x81, 0x38, 0x00, 0x02, 0x81, 0x39, 0x00,
			0x02, 0x81, 0x61, 0x00, 0x02, 0x81, 0x62, 0x00, 0x02, 0x81,
			0x63, 0x00, 0x02, 0x81, 0x64, 0x00, 0x02, 0x81, 0x65, 0x00,
			0x02, 0x81, 0x66, 0x00, 0x02, 0x81, 0x67, 0x00, 0x02, 0x81,
			0x68, 0x00, 0x02, 0x81, 0x69, 0x00, 0x02, 0x81, 0x6A, 0x00,
			0x02, 0x81, 0x6B, 0x00, 0x02, 0x81, 0x6C, 0x00, 0x02, 0x81,
			0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00, 0x02, 0x81, 0x6F, 0x00,
			0x02, 0x81, 0x70, 0x00, 0x02, 0x81, 0x71, 0x00, 0x02, 0x81,
			0x72, 0x00, 0x02, 0x81, 0x73, 0x00, 0x02, 0x81, 0x74, 0x00,
			0x02, 0x81, 0x75, 0x00, 0x02,
		}},
		{Map{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true, "v": true,
		}, []byte{
			0x7F, 0x85, 0x00, 0x00, 0x00, 0x1F, 0x00, 0x00, 0x00, 0x81,
			0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00,
			0x02, 0x81, 0x34, 0x00, 0x02, 0x81, 0x35, 0x00, 0x02, 0x81,
			0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00,
			0x02, 0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81,
			0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02, 0x81, 0x64, 0x00,
			0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81,
			0x67, 0x00, 0x02, 0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00,
			0x02, 0x81, 0x6A, 0x00, 0x02, 0x81, 0x6B, 0x00, 0x02, 0x81,
			0x6C, 0x00, 0x02, 0x81, 0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00,
			0x02, 0x81, 0x6F, 0x00, 0x02, 0x81, 0x70, 0x00, 0x02, 0x81,
			0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00,
			0x02, 0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81,
			0x76, 0x00, 0x02,
		}},
		{Map{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true, "v": true, "w": true,
		}, []byte{
			0x7F, 0x89, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x81,
			0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00,
			0x02, 0x81, 0x34, 0x00, 0x02, 0x81, 0x35, 0x00, 0x02, 0x81,
			0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00,
			0x02, 0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81,
			0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02, 0x81, 0x64, 0x00,
			0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81,
			0x67, 0x00, 0x02, 0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00,
			0x02, 0x81, 0x6A, 0x00, 0x02, 0x81, 0x6B, 0x00, 0x02, 0x81,
			0x6C, 0x00, 0x02, 0x81, 0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00,
			0x02, 0x81, 0x6F, 0x00, 0x02, 0x81, 0x70, 0x00, 0x02, 0x81,
			0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00,
			0x02, 0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81,
			0x76, 0x00, 0x02, 0x81, 0x77, 0x00, 0x02,
		}},
	},
}

func TestIsUTF8Bytes(t *testing.T) {
	assert := NewAssert(t)

	assert(isUTF8Bytes(([]byte)("abc"))).IsTrue()
	assert(isUTF8Bytes(([]byte)("abc！#@¥#%#%#¥%"))).IsTrue()
	assert(isUTF8Bytes(([]byte)("中文"))).IsTrue()
	assert(isUTF8Bytes(([]byte)("🀄️文👃d"))).IsTrue()
	assert(isUTF8Bytes(([]byte)("🀄️文👃"))).IsTrue()

	assert(isUTF8Bytes(([]byte)(`
    😀 😁 😂 🤣 😃 😄 😅 😆 😉 😊 😋 😎 😍 😘 🥰 😗 😙 😚 ☺️ 🙂 🤗 🤩 🤔 🤨
    🙄 😏 😣 😥 😮 🤐 😯 😪 😫 😴 😌 😛 😜 😝 🤤 😒 😓 😔 😕 🙃 🤑 😲 ☹️ 🙁
    😤 😢 😭 😦 😧 😨 😩 🤯 😬 😰 😱 🥵 🥶 😳 🤪 😵 😡 😠 🤬 😷 🤒 🤕 🤢
    🤡 🥳 🥴 🥺 🤥 🤫 🤭 🧐 🤓 😈 👿 👹 👺 💀 👻 👽 🤖 💩 😺 😸 😹 😻 😼 😽
    👶 👧 🧒 👦 👩 🧑 👨 👵 🧓 👴 👲 👳‍♀️ 👳‍♂️ 🧕 🧔 👱‍♂️ 👱‍♀️ 👨‍🦰 👩‍🦰 👨‍🦱 👩‍🦱 👨‍🦲 👩‍🦲 👨‍🦳
    👩‍🦳 🦸‍♀️ 🦸‍♂️ 🦹‍♀️ 🦹‍♂️ 👮‍♀️ 👮‍♂️ 👷‍♀️ 👷‍♂️ 💂‍♀️ 💂‍♂️ 🕵️‍♀️ 🕵️‍♂️ 👩‍⚕️ 👨‍⚕️ 👩‍🌾 👨‍🌾 👩‍🍳
    👨‍🍳 👩‍🎓 👨‍🎓 👩‍🎤 👨‍🎤 👩‍🏫 👨‍🏫 👩‍🏭 👨‍🏭 👩‍💻 👨‍💻 👩‍💼 👨‍💼 👩‍🔧 👨‍🔧 👩‍🔬 👨‍🔬 👩‍🎨 👨‍🎨 👩‍🚒 👨‍🚒 👩‍✈️ 👨‍✈️ 👩‍🚀
    👩‍⚖️ 👨‍⚖️ 👰 🤵 👸 🤴 🤶 🎅 🧙‍♀️ 🧙‍♂️ 🧝‍♀️ 🧝‍♂️ 🧛‍♀️ 🧛‍♂️ 🧟‍♀️ 🧟‍♂️ 🧞‍♀️ 🧞‍♂️ 🧜‍♀️
    🧜‍♂️ 🧚‍♀️ 🧚‍♂️ 👼 🤰 🤱 🙇‍♀️ 🙇‍♂️ 💁‍♀️ 💁‍♂️ 🙅‍♀️ 🙅‍♂️ 🙆‍♀️ 🙆‍♂️ 🙋‍♀️ 🙋‍♂️ 🤦‍♀️ 🤦‍♂️
    🤷‍♀️ 🤷‍♂️ 🙎‍♀️ 🙎‍♂️ 🙍‍♀️ 🙍‍♂️ 💇‍♀️ 💇‍♂️ 💆‍♀️ 💆‍♂️ 🧖‍♀️ 🧖‍♂️ 💅 🤳 💃 🕺 👯‍♀️ 👯‍♂️
    🕴 🚶‍♀️ 🚶‍♂️ 🏃‍♀️ 🏃‍♂️ 👫 👭 👬 💑 👩‍❤️‍👩 👨‍❤️‍👨 💏 👩‍❤️‍💋‍👩 👨‍❤️‍💋‍👨 👪 👨‍👩‍👧 👨‍👩‍👧‍👦 👨‍👩‍👦‍👦
    👨‍👩‍👧‍👧 👩‍👩‍👦 👩‍👩‍👧 👩‍👩‍👧‍👦 👩‍👩‍👦‍👦 👩‍👩‍👧‍👧 👨‍👨‍👦 👨‍👨‍👧 👨‍👨‍👧‍👦 👨‍👨‍👦‍👦 👨‍👨‍👧‍👧 👩‍👦 👩‍👧 👩‍👧‍👦 👩‍👦‍👦 👩‍👧‍👧 👨‍👦 👨‍👧 👨‍👧‍👦 👨‍👦‍👦 👨‍👧‍👧 🤲 👐
    👎 👊 ✊ 🤛 🤜 🤞 ✌️ 🤟 🤘 👌 👈 👉 👆 👇 ☝️ ✋ 🤚 🖐 🖖 👋 🤙 💪 🦵 🦶
    💍 💄 💋 👄 👅 👂 👃 👣 👁 👀 🧠 🦴 🦷 🗣 👤 👥
  `))).IsTrue()

	assert(isUTF8Bytes([]byte{0xC1})).IsFalse()
	assert(isUTF8Bytes([]byte{0xC1, 0x01})).IsFalse()

	assert(isUTF8Bytes([]byte{0xE1, 0x80})).IsFalse()
	assert(isUTF8Bytes([]byte{0xE1, 0x01, 0x81})).IsFalse()
	assert(isUTF8Bytes([]byte{0xE1, 0x80, 0x01})).IsFalse()

	assert(isUTF8Bytes([]byte{0xF1, 0x80, 0x80})).IsFalse()
	assert(isUTF8Bytes([]byte{0xF1, 0x70, 0x80, 0x80})).IsFalse()
	assert(isUTF8Bytes([]byte{0xF1, 0x80, 0x70, 0x80})).IsFalse()
	assert(isUTF8Bytes([]byte{0xF1, 0x80, 0x80, 0x70})).IsFalse()

	assert(isUTF8Bytes([]byte{0xFF, 0x80, 0x80, 0x70})).IsFalse()
}

func TestCheckValue(t *testing.T) {
	assert := NewAssert(t)
	assert(checkValue(nil, "value", 1)).Equals("")
	assert(checkValue(Array(nil), "value", 1)).Equals("")
	assert(checkValue(Map(nil), "value", 1)).Equals("")
	assert(checkValue(true, "value", 1)).Equals("")
	assert(checkValue(1, "value", 1)).Equals("")
	assert(checkValue(int8(1), "value", 1)).Equals("")
	assert(checkValue(int16(1), "value", 1)).Equals("")
	assert(checkValue(int32(1), "value", 1)).Equals("")
	assert(checkValue(int64(1), "value", 1)).Equals("")
	assert(checkValue(uint(1), "value", 1)).Equals("")
	assert(checkValue(uint8(1), "value", 1)).Equals("")
	assert(checkValue(uint16(1), "value", 1)).Equals("")
	assert(checkValue(uint32(1), "value", 1)).Equals("")
	assert(checkValue(uint64(1), "value", 1)).Equals("")
	assert(checkValue(float32(1), "value", 1)).Equals("")
	assert(checkValue(float64(1), "value", 1)).Equals("")
	assert(checkValue("", "value", 1)).Equals("")
	assert(checkValue(Bytes{}, "value", 1)).Equals("")
	assert(checkValue(Array{}, "value", 1)).Equals("")
	assert(checkValue(Map{}, "value", 1)).Equals("")

	assert(checkValue(nil, "value", 0)).
		Equals("value is too complicated")
	assert(checkValue(make(chan bool), "value", 1)).
		Equals("value type (chan bool) is not supported")

	assert(checkValue(Array{true}, "value", 1)).
		Equals("value[0] is too complicated")
	assert(checkValue(Array{true}, "value", 2)).Equals("")
	assert(checkValue(Map{"key": "value"}, "value", 1)).
		Equals("value[\"key\"] is too complicated")
	assert(checkValue(Map{"key": "value"}, "value", 2)).Equals("")
}

func TestStream_basic(t *testing.T) {
	assert := NewAssert(t)

	// test streamCache
	stream := streamCache.Get().(*Stream)
	assert(len(stream.frames)).Equals(1)
	assert(cap(stream.frames)).Equals(8)
	assert(stream.readSeg).Equals(0)
	assert(stream.readIndex).Equals(streamBodyPos)
	assert(stream.readFrame).Equals(*stream.frames[0])
	assert(stream.writeSeg).Equals(0)
	assert(stream.writeIndex).Equals(streamBodyPos)
	assert(stream.writeFrame).Equals(*stream.frames[0])

	// test frameCache
	frame := frameCache.Get().(*[]byte)
	assert(frame).IsNotNil()
	assert(len(*frame)).Equals(512)
	assert(cap(*frame)).Equals(512)
}

func TestStream_newRPCStream_Release_Reset(t *testing.T) {
	assert := NewAssert(t)

	// test streamCache
	for i := 0; i < 5000; i++ {
		stream := NewStream()
		assert(len(stream.frames)).Equals(1)
		assert(cap(stream.frames)).Equals(8)
		assert(stream.readSeg).Equals(0)
		assert(stream.readIndex).Equals(streamBodyPos)
		assert(stream.readFrame).Equals(*stream.frames[0])
		assert(stream.writeSeg).Equals(0)
		assert(stream.writeIndex).Equals(streamBodyPos)
		assert(stream.writeFrame).Equals(*stream.frames[0])

		for n := 0; n < i; n++ {
			stream.PutBytes([]byte{9})
		}
		stream.Release()
	}
}

func TestStream_getServerCallbackID_setServerCallbackID(t *testing.T) {
	assert := NewAssert(t)

	stream := NewStream()
	bytes8 := make([]byte, 8, 8)

	for i := 0; i < 1000; i++ {
		v := rand.Uint64()
		binary.LittleEndian.PutUint64(bytes8, v)
		stream.SetCallbackID(v)
		assert(stream.header[0:8]).Equals(bytes8)
		assert(stream.GetBufferUnsafe()[1:9]).Equals(bytes8)
		assert(stream.GetCallbackID()).Equals(v)
	}
}

func TestStream_GetSessionId_SetSessionId(t *testing.T) {
	assert := NewAssert(t)

	stream := NewStream()
	bytes8 := make([]byte, 8, 8)

	for i := 0; i < 1000; i++ {
		v := rand.Uint64()
		binary.LittleEndian.PutUint64(bytes8, v)
		stream.SetSessionID(v)
		assert(stream.header[8:16]).Equals(bytes8)
		assert(stream.GetBufferUnsafe()[9:17]).Equals(bytes8)
		assert(stream.GetSessionID()).Equals(v)
	}
}

func TestStream_GetSequence_SetSequence(t *testing.T) {
	assert := NewAssert(t)

	stream := NewStream()
	bytes8 := make([]byte, 8, 8)

	for i := 0; i < 1000; i++ {
		v := rand.Uint64()
		binary.LittleEndian.PutUint64(bytes8, v)
		stream.SetSequence(v)
		assert(stream.header[8:16]).Equals(bytes8)
		assert(stream.GetBufferUnsafe()[9:17]).Equals(bytes8)
		assert(stream.GetSequence()).Equals(v)
	}
}

func TestStream_getMachineID_setMachineID(t *testing.T) {
	assert := NewAssert(t)

	stream := NewStream()
	bytes8 := make([]byte, 8, 8)

	for i := 0; i < 1000; i++ {
		v := rand.Uint64()
		binary.LittleEndian.PutUint64(bytes8, v)
		stream.SetMachineID(v)
		assert(stream.header[16:24]).Equals(bytes8)
		assert(stream.GetBufferUnsafe()[17:25]).Equals(bytes8)
		assert(stream.GetMachineID()).Equals(v)
	}
}

func TestStream_GetHeader(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 5000; i++ {
		bytes := make([]byte, i, i)
		for n := 0; n < i; n++ {
			bytes[n] = byte(n)
		}

		stream := NewStream()
		stream.PutBytes(bytes)
		assert(stream.GetHeader()).Equals(stream.header)
		stream.Release()
	}
}

func TestStream_GetBuffer(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 5000; i++ {
		bytes := make([]byte, i, i)
		for n := 0; n < i; n++ {
			bytes[n] = byte(n)
		}

		stream := NewStream()
		stream.PutBytes(bytes)
		assert(stream.GetBuffer()[0]).Equals(byte(1))
		assert(stream.GetBuffer()[streamBodyPos:]).Equals(bytes)
		stream.Release()
	}
}

func TestStream_GetBufferUnsafe(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 5000; i++ {
		bytes := make([]byte, i, i)
		for n := 0; n < i; n++ {
			bytes[n] = byte(n)
		}

		stream := NewStream()
		stream.PutBytes(bytes)
		assert(stream.GetBufferUnsafe()[0]).Equals(byte(1))
		assert(stream.GetBufferUnsafe()[1:streamBodyPos]).Equals(stream.header)
		assert(stream.GetBufferUnsafe()[streamBodyPos:]).Equals(bytes)
		stream.Release()
	}
}

func TestStream_GetReadPos(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 5000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)
		stream.SetReadPos(i)
		assert(stream.GetReadPos()).Equals(i)
		stream.Release()
	}
}

func TestStream_SetReadPos(t *testing.T) {
	assert := NewAssert(t)

	for i := 1; i < 5000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)
		assert(stream.SetReadPos(-1)).IsFalse()
		assert(stream.SetReadPos(i - 1)).IsTrue()
		assert(stream.SetReadPos(i)).IsTrue()
		assert(stream.SetReadPos(i + 1)).IsFalse()
		stream.Release()
	}
}

func TestStream_SetReadPosToBodyStart(t *testing.T) {
	assert := NewAssert(t)

	for i := streamBodyPos; i < streamBodyPos+5000; i++ {
		stream := NewStream()
		stream.SetWritePos(i + 1)
		stream.SetReadPos(i)
		assert(stream.GetReadPos()).Equals(i)
		stream.SetReadPosToBodyStart()
		assert(stream.GetReadPos()).Equals(streamBodyPos)
	}
}

func TestStream_setReadPosUnsafe(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.setReadPosUnsafe(i)
		assert(stream.GetReadPos()).Equals(i)
	}

	stream.Release()
}

func TestStream_GetWritePos(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.setWritePosUnsafe(i)
		assert(stream.GetWritePos()).Equals(i)
	}
}

func TestStream_SetWritePos(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 10000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)
		assert(stream.GetWritePos()).Equals(i)
		stream.Release()
	}
}

func TestStream_SetWritePosToBodyStart(t *testing.T) {
	assert := NewAssert(t)

	for i := 1; i < 5000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)
		assert(stream.GetWritePos()).Equals(i)
		stream.SetWritePosToBodyStart()
		assert(stream.GetWritePos()).Equals(streamBodyPos)
	}
}

func TestStream_setWritePosUnsafe(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(10000)

	for i := 0; i < 10000; i++ {
		stream.setWritePosUnsafe(i)
		assert(stream.GetWritePos()).Equals(i)
	}
}

func TestStream_CanRead(t *testing.T) {
	assert := NewAssert(t)

	for i := 1; i < 10000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)

		stream.setReadPosUnsafe(i - 1)
		assert(stream.CanRead()).IsTrue()

		stream.setReadPosUnsafe(i)
		assert(stream.CanRead()).IsFalse()

		if (i+1)%512 != 0 {
			stream.setReadPosUnsafe(i + 1)
			assert(stream.CanRead()).IsFalse()
		}
	}
}

func TestStream_IsReadFinish(t *testing.T) {
	assert := NewAssert(t)

	for i := 1; i < 10000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)

		stream.setReadPosUnsafe(i - 1)
		assert(stream.IsReadFinish()).IsFalse()

		stream.setReadPosUnsafe(i)
		assert(stream.IsReadFinish()).IsTrue()

		if (i+1)%512 != 0 {
			stream.setReadPosUnsafe(i + 1)
			assert(stream.IsReadFinish()).IsFalse()
		}
	}
}

func TestStream_gotoNextReadFrameUnsafe(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(10000)

	for i := 0; i < 8000; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		stream.gotoNextReadFrameUnsafe()
		assert(stream.GetReadPos()).Equals((i/512 + 1) * 512)
	}
}

func TestStream_gotoNextReadByteUnsafe(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(10000)

	for i := 0; i < 8000; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		stream.gotoNextReadByteUnsafe()
		assert(stream.GetReadPos()).Equals(i + 1)
	}
}

func TestStream_hasOneByteToRead(t *testing.T) {
	assert := NewAssert(t)

	for i := 1; i < 2000; i++ {
		stream := NewStream()
		stream.SetWritePos(i)

		for n := 0; n < i; n++ {
			assert(stream.SetReadPos(n))
			assert(stream.hasOneByteToRead()).IsTrue()
		}

		assert(stream.SetReadPos(i))
		assert(stream.hasOneByteToRead()).IsFalse()
		stream.Release()
	}
}

func TestStream_hasNBytesToRead(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(1100)

	for i := 0; i < 1000; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 1600; n++ {
			assert(stream.hasNBytesToRead(n)).Equals(i+n <= 1100)
		}
	}
}

func TestStream_isSafetyReadNBytesInCurrentFrame(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyReadNBytesInCurrentFrame(n)).
				Equals(512-i%512 > n)
		}
	}
}

func TestStream_isSafetyRead3BytesInCurrentFrame(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyRead3BytesInCurrentFrame()).
				Equals(512-i%512 > 3)
		}
	}
}

func TestStream_isSafetyRead5BytesInCurrentFrame(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyRead5BytesInCurrentFrame()).
				Equals(512-i%512 > 5)
		}
	}
}

func TestStream_isSafetyRead9BytesInCurrentFrame(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	stream.SetWritePos(1100)

	for i := 0; i < 800; i++ {
		assert(stream.SetReadPos(i)).IsTrue()
		for n := 0; n < 800; n++ {
			assert(stream.isSafetyRead9BytesInCurrentFrame()).
				Equals(512-i%512 > 9)
		}
	}
}

func TestStream_putBytes(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 600; i++ {
		for n := 0; n < 600; n++ {
			stream := NewStream()
			stream.SetWritePos(i)
			bytes := make([]byte, n, n)
			for z := 0; z < n; z++ {
				bytes[z] = byte(z)
			}
			stream.PutBytes(bytes)
			assert(stream.GetBuffer()[i:]).Equals(bytes)
			stream.Release()
		}
	}
}

func TestStream_putString(t *testing.T) {
	assert := NewAssert(t)

	for i := 0; i < 600; i++ {
		for n := 0; n < 600; n++ {
			stream := NewStream()
			stream.SetWritePos(i)
			bytes := make([]byte, n, n)
			for z := 0; z < n; z++ {
				bytes[z] = byte(z)
			}
			strVal := string(bytes)
			stream.PutString(strVal)
			assert(stream.GetBuffer()[i:]).Equals(bytes)
			stream.Release()
		}
	}
}

func TestStream_read3BytesCrossFrameUnsafe(t *testing.T) {
	assert := NewAssert(t)

	stream0 := NewStream()
	stream0.SetWritePos(508)
	stream0.PutBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	stream0.SetReadPos(509)
	assert(stream0.read3BytesCrossFrameUnsafe()).Equals([]byte{2, 3, 4})
	assert(stream0.GetReadPos()).Equals(512)
	stream0.SetReadPos(510)
	assert(stream0.read3BytesCrossFrameUnsafe()).Equals([]byte{3, 4, 5})
	assert(stream0.GetReadPos()).Equals(513)
	stream0.SetReadPos(511)
	assert(stream0.read3BytesCrossFrameUnsafe()).Equals([]byte{4, 5, 6})
	assert(stream0.GetReadPos()).Equals(514)
}

func TestStream_peek5BytesCrossFrameUnsafe(t *testing.T) {
	assert := NewAssert(t)

	stream0 := NewStream()
	stream0.SetWritePos(506)
	stream0.PutBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	stream0.SetReadPos(507)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{2, 3, 4, 5, 6})
	assert(stream0.GetReadPos()).Equals(507)
	stream0.SetReadPos(508)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{3, 4, 5, 6, 7})
	assert(stream0.GetReadPos()).Equals(508)
	stream0.SetReadPos(509)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{4, 5, 6, 7, 8})
	assert(stream0.GetReadPos()).Equals(509)
	stream0.SetReadPos(510)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{5, 6, 7, 8, 9})
	assert(stream0.GetReadPos()).Equals(510)
	stream0.SetReadPos(511)
	assert(stream0.peek5BytesCrossFrameUnsafe()).Equals([]byte{6, 7, 8, 9, 10})
	assert(stream0.GetReadPos()).Equals(511)
}

func TestStream_read5BytesCrossFrameUnsafe(t *testing.T) {
	assert := NewAssert(t)

	stream0 := NewStream()
	stream0.SetWritePos(506)
	stream0.PutBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	stream0.SetReadPos(507)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{2, 3, 4, 5, 6})
	assert(stream0.GetReadPos()).Equals(512)
	stream0.SetReadPos(508)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{3, 4, 5, 6, 7})
	assert(stream0.GetReadPos()).Equals(513)
	stream0.SetReadPos(509)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{4, 5, 6, 7, 8})
	assert(stream0.GetReadPos()).Equals(514)
	stream0.SetReadPos(510)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{5, 6, 7, 8, 9})
	assert(stream0.GetReadPos()).Equals(515)
	stream0.SetReadPos(511)
	assert(stream0.read5BytesCrossFrameUnsafe()).Equals([]byte{6, 7, 8, 9, 10})
	assert(stream0.GetReadPos()).Equals(516)
}

func TestStream_read9BytesCrossFrameUnsafe(t *testing.T) {
	assert := NewAssert(t)

	stream0 := NewStream()
	stream0.SetWritePos(502)
	stream0.PutBytes([]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	})
	stream0.SetReadPos(503)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{2, 3, 4, 5, 6, 7, 8, 9, 10})
	assert(stream0.GetReadPos()).Equals(512)
	stream0.SetReadPos(504)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{3, 4, 5, 6, 7, 8, 9, 10, 11})
	assert(stream0.GetReadPos()).Equals(513)
	stream0.SetReadPos(505)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{4, 5, 6, 7, 8, 9, 10, 11, 12})
	assert(stream0.GetReadPos()).Equals(514)
	stream0.SetReadPos(506)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{5, 6, 7, 8, 9, 10, 11, 12, 13})
	assert(stream0.GetReadPos()).Equals(515)
	stream0.SetReadPos(507)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{6, 7, 8, 9, 10, 11, 12, 13, 14})
	assert(stream0.GetReadPos()).Equals(516)
	stream0.SetReadPos(508)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{7, 8, 9, 10, 11, 12, 13, 14, 15})
	assert(stream0.GetReadPos()).Equals(517)
	stream0.SetReadPos(509)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{8, 9, 10, 11, 12, 13, 14, 15, 16})
	assert(stream0.GetReadPos()).Equals(518)
	stream0.SetReadPos(510)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{9, 10, 11, 12, 13, 14, 15, 16, 17})
	assert(stream0.GetReadPos()).Equals(519)
	stream0.SetReadPos(511)
	assert(stream0.read9BytesCrossFrameUnsafe()).
		Equals([]byte{10, 11, 12, 13, 14, 15, 16, 17, 18})
	assert(stream0.GetReadPos()).Equals(520)
}

func TestStream_readNBytesUnsafe(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	for i := 0; i < 2000; i++ {
		stream.PutBytes([]byte{byte(i)})
	}
	streamBuf := stream.GetBuffer()

	for i := 1; i < 600; i++ {
		for n := 0; n < 1100; n++ {
			stream.SetReadPos(i)
			assert(stream.readNBytesUnsafe(n)).
				Equals(streamBuf[i : i+n])
		}
	}
}

func TestStream_WriteNil(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["nil"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteNil()
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(i + 1)
			stream.Release()
		}
	}
}

func TestStream_WriteBool(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["bool"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteBool(testData[0].(bool))
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestStream_WriteFloat64(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["float64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteFloat64(testData[0].(float64))
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestStream_WriteInt64(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["int64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteInt64(testData[0].(int64))
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestStream_WriteUInt64(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["uint64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteUint64(testData[0].(uint64))
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestStream_WriteString(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["string"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteString(testData[0].(string))
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestStream_WriteBytes(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["bytes"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.WriteBytes(testData[0].([]byte))
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}
	}
}

func TestStream_WriteArray(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["array"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			assert(stream.WriteArray(testData[0].(Array))).Equals(StreamWriteOK)
			assert(stream.GetBuffer()[i:]).Equals(testData[1])
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// error type
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			assert(
				stream.WriteArray(Array{true, true, true, make(chan bool), true}),
			).Equals(StreamWriteUnsupportedType)
			assert(stream.GetWritePos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_WriteMap(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["map"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			assert(stream.WriteMap(testData[0].(Map))).Equals(StreamWriteOK)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)

			stream.SetReadPos(i)
			assert(stream.ReadMap()).Equals(testData[0].(Map), true)

			stream.Release()
		}

		// error type
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			assert(stream.WriteMap(Map{"0": 0, "1": make(chan bool)})).
				Equals(StreamWriteUnsupportedType)
			assert(stream.GetWritePos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_Write(t *testing.T) {
	assert := NewAssert(t)
	stream := NewStream()
	assert(stream.Write(nil)).Equals(StreamWriteOK)
	assert(stream.Write(true)).Equals(StreamWriteOK)
	assert(stream.Write(0)).Equals(StreamWriteOK)
	assert(stream.Write(int8(0))).Equals(StreamWriteOK)
	assert(stream.Write(int16(0))).Equals(StreamWriteOK)
	assert(stream.Write(int32(0))).Equals(StreamWriteOK)
	assert(stream.Write(int64(0))).Equals(StreamWriteOK)
	assert(stream.Write(uint(0))).Equals(StreamWriteOK)
	assert(stream.Write(uint8(0))).Equals(StreamWriteOK)
	assert(stream.Write(uint16(0))).Equals(StreamWriteOK)
	assert(stream.Write(uint32(0))).Equals(StreamWriteOK)
	assert(stream.Write(uint64(0))).Equals(StreamWriteOK)
	assert(stream.Write(float32(0))).Equals(StreamWriteOK)
	assert(stream.Write(float64(0))).Equals(StreamWriteOK)
	assert(stream.Write("")).Equals(StreamWriteOK)
	assert(stream.Write([]byte{})).Equals(StreamWriteOK)
	assert(stream.Write(Array{})).Equals(StreamWriteOK)
	assert(stream.Write(Map{})).Equals(StreamWriteOK)
	assert(stream.Write(make(chan bool))).Equals(StreamWriteUnsupportedType)
	stream.Release()
}

func TestStream_ReadNil(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["nil"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])

			assert(stream.ReadNil()).Equals(true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadNil()).IsFalse()
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadNil()).IsFalse()
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_ReadBool(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["bool"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadBool()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadBool()).Equals(false, false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadBool()).Equals(false, false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_ReadFloat64(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["float64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadFloat64()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadFloat64()).Equals(float64(0), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadFloat64()).Equals(float64(0), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_ReadInt64(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["int64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadInt64()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadInt64()).Equals(int64(0), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadInt64()).Equals(int64(0), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_ReadUint64(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["uint64"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadUint64()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadUint64()).Equals(uint64(0), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadUint64()).Equals(uint64(0), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_ReadString(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["string"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadString()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadString()).Equals("", false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadString()).Equals("", false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// read tail is not zero
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			stream.SetWritePos(stream.GetWritePos() - 1)
			stream.PutBytes([]byte{1})
			assert(stream.ReadString()).Equals("", false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}

	// read string utf8 error
	stream1 := NewStream()
	stream1.PutBytes([]byte{
		0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
		0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
		0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
		0xA9, 0x00,
	})
	assert(stream1.ReadString()).Equals("", false)
	assert(stream1.GetReadPos()).Equals(streamBodyPos)

	// read string utf8 error
	stream2 := NewStream()
	stream2.PutBytes([]byte{
		0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
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
	})
	assert(stream2.ReadString()).Equals("", false)
	assert(stream2.GetReadPos()).Equals(streamBodyPos)

	// read string length error
	stream3 := NewStream()
	stream3.PutBytes([]byte{
		0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
	})
	assert(stream3.ReadString()).Equals("", false)
	assert(stream3.GetReadPos()).Equals(streamBodyPos)
}

func TestStream_ReadUnsafeString(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["string"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadUnsafeString()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadUnsafeString()).Equals("", false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadUnsafeString()).Equals("", false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// read tail is not zero
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			stream.SetWritePos(stream.GetWritePos() - 1)
			stream.PutBytes([]byte{1})
			assert(stream.ReadUnsafeString()).Equals("", false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}

	// read string utf8 error
	stream1 := NewStream()
	stream1.PutBytes([]byte{
		0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
		0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
		0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
		0xA9, 0x00,
	})
	assert(stream1.ReadUnsafeString()).Equals("", false)
	assert(stream1.GetReadPos()).Equals(streamBodyPos)

	// read string utf8 error
	stream2 := NewStream()
	stream2.PutBytes([]byte{
		0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
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
	})
	assert(stream2.ReadUnsafeString()).Equals("", false)
	assert(stream2.GetReadPos()).Equals(streamBodyPos)

	// read string length error
	stream3 := NewStream()
	stream3.PutBytes([]byte{
		0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
	})
	assert(stream3.ReadUnsafeString()).Equals("", false)
	assert(stream3.GetReadPos()).Equals(streamBodyPos)
}

func TestStream_ReadBytes(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["bytes"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadBytes()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadBytes()).Equals(Bytes(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadBytes()).Equals(Bytes(nil), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}

	// read bytes length error
	stream1 := NewStream()
	stream1.PutBytes([]byte{
		0xFF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
	})
	assert(stream1.ReadBytes()).Equals(Bytes(nil), false)
	assert(stream1.GetReadPos()).Equals(streamBodyPos)
}

func TestStream_ReadUnsafeBytes(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["bytes"] {
		// ok
		for i := 1; i < 1100; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			assert(stream.ReadUnsafeBytes()).Equals(testData[0], true)
			assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
			stream.Release()
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadUnsafeBytes()).Equals(Bytes(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadUnsafeBytes()).Equals(Bytes(nil), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}

	// read bytes length error
	stream1 := NewStream()
	stream1.PutBytes([]byte{
		0xFF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
	})
	assert(stream1.ReadUnsafeBytes()).Equals(Bytes(nil), false)
	assert(stream1.GetReadPos()).Equals(streamBodyPos)
}

func TestStream_ReadArray(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["array"] {
		// ok
		for i := 1; i < 530; i++ {
			for j := 1; j < 530; j++ {
				// skip for performance
				if j > 10 && j < 500 {
					continue
				}

				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadArray()).Equals(testData[0].(Array), true)
				assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}

		// overflow
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadArray()).Equals(Array(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadArray()).Equals(Array(nil), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// error in stream
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			if len(testData[0].(Array)) > 0 {
				stream.SetWritePos(stream.GetWritePos() - 1)
				stream.PutBytes([]byte{13})
				assert(stream.ReadArray()).Equals(Array(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// error in stream
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{0x41, 0x07, 0x00, 0x00, 0x00, 0x02, 0x02})
			assert(stream.ReadArray()).Equals(Array(nil), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	}
}

func TestStream_ReadMap(t *testing.T) {
	assert := NewAssert(t)

	for _, testData := range streamTestCollections["map"] {
		// ok
		for i := 1; i < 530; i++ {
			for j := 1; j < 530; j++ {
				// skip for performance
				if j > 10 && j < 500 {
					continue
				}
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadMap()).Equals(testData[0], true)
				assert(stream.GetWritePos()).Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}

		// overflow
		for i := 1; i < 530; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			writePos := stream.GetWritePos()
			for idx := i; idx < writePos-1; idx++ {
				stream.SetReadPos(i)
				stream.setWritePosUnsafe(idx)
				assert(stream.ReadMap()).Equals(Map(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// type not match
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadMap()).Equals(Map(nil), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// error in stream
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			if len(testData[0].(Map)) > 0 {
				stream.SetWritePos(stream.GetWritePos() - 1)
				stream.PutBytes([]byte{13})
				assert(stream.ReadMap()).Equals(Map(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}

		// error in stream, length error
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{
				0x61, 0x0A, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x02,
			})
			assert(stream.ReadMap()).Equals(Map(nil), false)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}

		// error in stream, key error
		for i := 1; i < 550; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.Write(testData[0])
			wPos := stream.GetWritePos()
			mapSize := len(testData[0].(Map))

			if mapSize > 30 {
				stream.setWritePosUnsafe(i + 9)
				stream.PutBytes([]byte{13})
				stream.setWritePosUnsafe(wPos)
				assert(stream.ReadMap()).Equals(Map(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			} else if mapSize > 0 {
				stream.setWritePosUnsafe(i + 5)
				stream.PutBytes([]byte{13})
				stream.setWritePosUnsafe(wPos)
				assert(stream.ReadMap()).Equals(Map(nil), false)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}
	}
}

func TestStream_Read(t *testing.T) {
	assert := NewAssert(t)

	testCollections := make([][2]interface{}, 0, 0)

	for key := range streamTestCollections {
		for _, v := range streamTestCollections[key] {
			testCollections = append(testCollections, v)
		}
	}

	for _, item := range testCollections {
		stream := NewStream()
		stream.PutBytes(item[1].([]byte))
		if isNil(item[0]) {
			assert(stream.Read()).Equals(nil, true)
		} else {
			assert(stream.Read()).Equals(item[0], true)
		}
	}

	stream := NewStream()
	stream.PutBytes([]byte{12})
	assert(stream.Read()).Equals(nil, false)

	stream = NewStream()
	stream.PutBytes([]byte{13})
	assert(stream.Read()).Equals(nil, false)
}

func BenchmarkRPCStream_ReadString(b *testing.B) {
	stream := NewStream()
	stream.WriteString("#.user.login:isUserARight")

	b.ReportAllocs()
	b.N = 10000000
	for i := 0; i < b.N; i++ {
		stream.SetReadPos(streamBodyPos)
		stream.ReadString()
	}
}

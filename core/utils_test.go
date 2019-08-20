package core

import (
	"testing"
	"unsafe"
)

type testObject struct {
	test string
}

func Test_GetStackString(t *testing.T) {
	assert := newAssert(t)
	assert(FindLinesByPrefix(
		GetStackString(0),
		"-01",
	)[0]).Contains("Test_GetStackString")
	assert(FindLinesByPrefix(
		GetStackString(0),
		"-01",
	)[0]).Contains("utils_test")
}

func Test_FindLinesByPrefix(t *testing.T) {
	assert := newAssert(t)

	ret := FindLinesByPrefix("", "")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals("")

	ret = FindLinesByPrefix("", "hello")
	assert(len(ret)).Equals(0)

	ret = FindLinesByPrefix("hello", "dd")
	assert(len(ret)).Equals(0)

	ret = FindLinesByPrefix("  ddhello", "dd")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals("  ddhello")

	ret = FindLinesByPrefix(" \t ddhello", "dd")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals(" \t ddhello")

	ret = FindLinesByPrefix(" \t ddhello\ndd\n", "dd")
	assert(len(ret)).Equals(2)
	assert(ret[0]).Equals(" \t ddhello")
	assert(ret[1]).Equals("dd")
}

func Test_GetByteArrayDebugString(t *testing.T) {
	assert := newAssert(t)
	assert(GetByteArrayDebugString([]byte{})).Equals(
		"",
	)
	assert(GetByteArrayDebugString([]byte{1, 2})).Equals(
		"0000: 0x01 0x02 ",
	)
	assert(GetByteArrayDebugString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})).Equals(
		"0000: 0x01 0x02 0x03 0x04 0x05 0x06 0x07 0x08 0x09 0x0a 0x0b 0x0c 0x0d 0x0e 0x0f 0x10 ",
	)
	assert(GetByteArrayDebugString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17})).Equals(
		"0000: 0x01 0x02 0x03 0x04 0x05 0x06 0x07 0x08 0x09 0x0a 0x0b 0x0c 0x0d 0x0e 0x0f 0x10 \n0016: 0x11 ",
	)
}

func Test_GetUrlBySchemeHostPortAndPath(t *testing.T) {
	assert := newAssert(t)

	assert(GetURLBySchemeHostPortAndPath("", "127.0.0.1", 8080, "/world")).
		Equals("")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "")).
		Equals("ws://127.0.0.1:8080/")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "/")).
		Equals("ws://127.0.0.1:8080/")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "world")).
		Equals("ws://127.0.0.1:8080/world")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "/world")).
		Equals("ws://127.0.0.1:8080/world")
}

func Test_ConvertOrdinalToString(t *testing.T) {
	assert := newAssert(t)

	assert(ConvertOrdinalToString(0)).Equals("")
	assert(ConvertOrdinalToString(1)).Equals("1st")
	assert(ConvertOrdinalToString(2)).Equals("2nd")
	assert(ConvertOrdinalToString(3)).Equals("3rd")
	assert(ConvertOrdinalToString(4)).Equals("4th")
	assert(ConvertOrdinalToString(10)).Equals("10th")
	assert(ConvertOrdinalToString(100)).Equals("100th")
}

func Test_GetObjectFieldPointer(t *testing.T) {
	assert := newAssert(t)
	obj := &testObject{
		test: "hi",
	}
	assert(GetObjectFieldPointer(obj, "test")).Equals(unsafe.Pointer(&obj.test))
}

func Test_AddPrefixPerLine(t *testing.T) {
	assert := newAssert(t)

	assert(AddPrefixPerLine("", "")).Equals("")
	assert(AddPrefixPerLine("a", "")).Equals("a")
	assert(AddPrefixPerLine("\n", "")).Equals("\n")
	assert(AddPrefixPerLine("a\n", "")).Equals("a\n")
	assert(AddPrefixPerLine("a\nb", "")).Equals("a\nb")
	assert(AddPrefixPerLine("", "-")).Equals("-")
	assert(AddPrefixPerLine("a", "-")).Equals("-a")
	assert(AddPrefixPerLine("\n", "-")).Equals("-\n-")
	assert(AddPrefixPerLine("a\n", "-")).Equals("-a\n-")
	assert(AddPrefixPerLine("a\nb", "-")).Equals("-a\n-b")
}

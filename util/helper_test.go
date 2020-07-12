package util

import (
	"testing"
)

func TestAddPrefixPerLine(t *testing.T) {
	assert := NewAssert(t)

	assert(AddPrefixPerLine("", "")).Equals("")
	assert(AddPrefixPerLine("a", "")).Equals("a")
	assert(AddPrefixPerLine("\n", "")).Equals("\n")
	assert(AddPrefixPerLine("a\n", "")).Equals("a\n")
	assert(AddPrefixPerLine("a\nb", "")).Equals("a\nb")
	assert(AddPrefixPerLine("", "-")).Equals("-")
	assert(AddPrefixPerLine("a", "-")).Equals("-a")
	assert(AddPrefixPerLine("\n", "-")).Equals("-\n")
	assert(AddPrefixPerLine("a\n", "-")).Equals("-a\n")
	assert(AddPrefixPerLine("a\nb", "-")).Equals("-a\n-b")
}

func TestConcatString(t *testing.T) {
	assert := NewAssert(t)

	assert(ConcatString("", "")).Equals("")
	assert(ConcatString("a", "")).Equals("a")
	assert(ConcatString("", "b")).Equals("b")
	assert(ConcatString("a", "b")).Equals("ab")
	assert(ConcatString("a", "b", "")).Equals("ab")
	assert(ConcatString("a", "b", "c")).Equals("abc")
}

func BenchmarkAddPrefixPerLine(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		AddPrefixPerLine("a\nb\nc", "test")
	}
}

func BenchmarkConcatString(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		ConcatString("a", "b")
	}
}

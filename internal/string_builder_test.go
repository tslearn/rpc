package internal

import (
	"testing"
)

func TestNewStringBuilder(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sb1 := NewStringBuilder()
	defer sb1.Release()
	assert(len(sb1.buffer)).Equals(0)
	assert(cap(sb1.buffer)).Equals(stringBuilderBufferSize)
}

func TestStringBuilder_Reset(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sb1 := NewStringBuilder()
	defer sb1.Release()
	for i := 0; i < stringBuilderBufferSize; i++ {
		sb1.AppendString("S")
	}
	sb1.Reset()
	assert(len(sb1.buffer)).Equals(0)
	assert(cap(sb1.buffer)).Equals(stringBuilderBufferSize)

	// Test(2)
	sb2 := NewStringBuilder()
	defer sb2.Release()
	for i := 0; i < 2*stringBuilderBufferSize; i++ {
		sb2.AppendString("S")
	}
	sb2.Reset()
	assert(len(sb2.buffer)).Equals(0)
	assert(cap(sb2.buffer)).Equals(stringBuilderBufferSize)
}

func TestStringBuilder_AppendByte(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sb1 := NewStringBuilder()
	defer sb1.Release()
	sb1.AppendByte('a')
	sb1.AppendByte('b')
	sb1.AppendByte('c')
	assert(sb1.String()).Equals("abc")
}

func TestStringBuilder_AppendBytes(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	longString1 := GetRandString(128)
	sb1 := NewStringBuilder()
	defer sb1.Release()
	sb1.AppendBytes([]byte(longString1))
	assert(sb1.String()).Equals(longString1)
}

func TestStringBuilder_AppendString(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	longString1 := getTestLongString()
	var testCollection1 = [][2]interface{}{
		{[]string{""}, ""},
		{[]string{"a"}, "a"},
		{[]string{"中国"}, "中国"},
		{[]string{"🀄🀄🀄🀄🀄🀄🀄️"}, "🀄🀄🀄🀄🀄🀄🀄️"},
		{[]string{"", "🀄🀄🀄🀄🀄🀄🀄️"}, "🀄🀄🀄🀄🀄🀄🀄️"},
		{[]string{"中国", "🀄🀄🀄🀄🀄🀄🀄️"}, "中国🀄🀄🀄🀄🀄🀄🀄️"},
		{[]string{longString1}, longString1},
	}
	for _, item := range testCollection1 {
		builder := NewStringBuilder()
		for i := 0; i < len(item[0].([]string)); i++ {
			builder.AppendString(item[0].([]string)[i])
		}
		assert(builder.String()).Equals(item[1])
		builder.Release()
	}
}

func TestStringBuilder_Merge(t *testing.T) {
	assert := NewAssert(t)

	sb1 := NewStringBuilder()
	sb2 := NewStringBuilder()

	sb1.Merge(sb2)
	assert(sb1.String()).Equals("")
	sb2.AppendString("123")

	sb1.Merge(sb2)
	assert(sb1.String()).Equals("123")

	sb1.Merge(sb2)
	assert(sb1.String()).Equals("123123")

	sb1.Merge(nil)
	assert(sb1.String()).Equals("123123")
}

func TestStringBuilder_IsEmpty(t *testing.T) {
	assert := NewAssert(t)

	builder := NewStringBuilder()
	assert(builder.IsEmpty()).IsTrue()
	builder.AppendString("a")
	assert(builder.IsEmpty()).IsFalse()
}

func TestStringBuilder_String(t *testing.T) {
	assert := NewAssert(t)

	builder := NewStringBuilder()
	builder.AppendString("a")
	assert(builder.String()).Equals("a")
}

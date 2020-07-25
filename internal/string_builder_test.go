package internal

import (
	"math/rand"
	"testing"
	"time"
)

func getRandomTestBuilder() (*StringBuilder, string) {
	rand.Seed(time.Now().UnixNano())
	str := GetRandString(rand.Int() % 10000)
	sb := NewStringBuilder()
	sb.AppendString(str)
	return sb, str
}

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
	for i := 0; i < 2000; i++ {
		sb1, str1 := getRandomTestBuilder()
		sb1.AppendByte('a')
		assert(sb1.String()).Equals(str1 + "a")
		sb1.Release()
	}
}

func TestStringBuilder_AppendBytes(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	for i := 0; i < 2000; i++ {
		sb1, str1 := getRandomTestBuilder()
		add1 := GetRandString(rand.Int() % 10000)
		sb1.AppendBytes(([]byte)(add1))
		assert(sb1.String()).Equals(str1 + add1)
		sb1.Release()
	}
}

func TestStringBuilder_AppendString(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	for i := 0; i < 2000; i++ {
		sb1, str1 := getRandomTestBuilder()
		add1 := GetRandString(rand.Int() % 10000)
		sb1.AppendString(add1)
		assert(sb1.String()).Equals(str1 + add1)
		sb1.Release()
	}
}

func TestStringBuilder_Merge(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	for i := 0; i < 2000; i++ {
		sb1, str1 := getRandomTestBuilder()
		mergeSb1, mergeStr1 := getRandomTestBuilder()
		sb1.Merge(mergeSb1)
		assert(sb1.String()).Equals(str1 + mergeStr1)
		sb1.Release()
		mergeSb1.Release()
	}

	// Test(2)
	for i := 0; i < 2000; i++ {
		sb1, str1 := getRandomTestBuilder()
		sb1.Merge(nil)
		assert(sb1.String()).Equals(str1)
		sb1.Release()
	}
}

func TestStringBuilder_IsEmpty(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sb1 := NewStringBuilder()
	assert(sb1.IsEmpty()).IsTrue()
	sb1.AppendString("a")
	assert(sb1.IsEmpty()).IsFalse()
}

func TestStringBuilder_String(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sb1 := NewStringBuilder()
	sb1.AppendString("a")
	assert(sb1.String()).Equals("a")
}

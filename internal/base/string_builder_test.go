package base

import (
	"math/rand"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

func getRandomTestBuilder() (*StringBuilder, string) {
	rand.Seed(time.Now().UnixNano())
	str := GetRandString(rand.Int() % 5000)
	sb := NewStringBuilder()
	sb.AppendString(str)
	return sb, str
}

func TestNewStringBuilder(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewStringBuilder()
		defer v1.Release()
		assert(len(v1.buffer)).Equal(0)
		assert(cap(v1.buffer)).Equal(512)
	})
}

func TestStringBuilder_Reset(t *testing.T) {
	t.Run("buffer is enlarged", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewStringBuilder()
		defer v1.Release()
		v1.AppendString(GetRandString(2 * stringBuilderBufferSize))
		v1.Reset()
		assert(len(v1.buffer)).Equal(0)
		assert(cap(v1.buffer)).Equal(stringBuilderBufferSize)
	})

	t.Run("buffer is not enlarged", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewStringBuilder()
		defer v1.Release()
		v1.AppendString(GetRandString(stringBuilderBufferSize))
		v1.Reset()
		assert(len(v1.buffer)).Equal(0)
		assert(cap(v1.buffer)).Equal(stringBuilderBufferSize)
	})
}

func TestStringBuilder_Release(t *testing.T) {
	t.Run("gc without release", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewStringBuilder()
		r1 := make(chan bool)
		runtime.SetFinalizer(v1, func(f *StringBuilder) {
			r1 <- true
		})
		debug.SetGCPercent(1)
		runtime.GC()
		assert(<-r1).IsTrue()
	})

	t.Run("gc with release", func(t *testing.T) {
		v1 := NewStringBuilder()
		runtime.SetFinalizer(v1, func(f *StringBuilder) { panic("error") })
		v1.Release()
		debug.SetGCPercent(1)
		runtime.GC()
	})
}

func TestStringBuilder_AppendByte(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		for i := 0; i < 1000; i++ {
			v1, s1 := getRandomTestBuilder()
			v1.AppendByte('a')
			assert(v1.String()).Equal(s1 + "a")
			v1.Release()
		}
	})
}

func TestStringBuilder_AppendBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		for i := 0; i < 1000; i++ {
			v1, s1 := getRandomTestBuilder()
			add1 := GetRandString(rand.Int() % 5000)
			v1.AppendBytes(([]byte)(add1))
			assert(v1.String()).Equal(s1 + add1)
			v1.Release()
		}
	})

}

func TestStringBuilder_AppendString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		for i := 0; i < 1000; i++ {
			v1, s1 := getRandomTestBuilder()
			add1 := GetRandString(rand.Int() % 5000)
			v1.AppendString(add1)
			assert(v1.String()).Equal(s1 + add1)
			v1.Release()
		}
	})
}

func TestStringBuilder_Merge(t *testing.T) {
	t.Run("merge nil", func(t *testing.T) {
		assert := NewAssert(t)
		for i := 0; i < 1000; i++ {
			v1, s1 := getRandomTestBuilder()
			v1.Merge(nil)
			assert(v1.String()).Equal(s1)
			v1.Release()
		}
	})

	t.Run("merge not nil", func(t *testing.T) {
		assert := NewAssert(t)
		for i := 0; i < 2000; i++ {
			v1, s1 := getRandomTestBuilder()
			v2, s2 := getRandomTestBuilder()
			v1.Merge(v2)
			assert(v1.String()).Equal(s1 + s2)
			v1.Release()
			v2.Release()
		}
	})
}

func TestStringBuilder_IsEmpty(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewStringBuilder()
		assert(v1.IsEmpty()).IsTrue()
		v1.AppendString("a")
		assert(v1.IsEmpty()).IsFalse()
	})
}

func TestStringBuilder_String(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewStringBuilder()
		v1.AppendString("a")
		assert(v1.String()).Equal("a")
	})
}

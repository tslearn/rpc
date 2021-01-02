package base

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestIsNil(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsNil(nil)).IsTrue()
		assert(IsNil(t)).IsFalse()
		assert(IsNil(3)).IsFalse()
		assert(IsNil(0)).IsFalse()
		assert(IsNil(uintptr(0))).IsFalse()
		assert(IsNil(uintptr(1))).IsFalse()
		assert(IsNil(unsafe.Pointer(nil))).IsTrue()
		assert(IsNil(unsafe.Pointer(t))).IsFalse()
	})
}

func TestMinInt(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(MinInt(1, 2)).Equal(1)
		assert(MinInt(2, 1)).Equal(1)
	})
}

func TestMaxInt(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(MaxInt(1, 2)).Equal(2)
		assert(MaxInt(2, 1)).Equal(2)
	})
}

func TestStringToBytesUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(cap(StringToBytesUnsafe("hello"))).Equal(5)
		assert(len(StringToBytesUnsafe("hello"))).Equal(5)
		assert(string(StringToBytesUnsafe("hello"))).Equal("hello")
	})
}

func TestBytesToStringUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(len(BytesToStringUnsafe([]byte("hello")))).Equal(5)
		assert(BytesToStringUnsafe([]byte("hello"))).Equal("hello")
	})
}

func TestIsUTF8Bytes(t *testing.T) {
	t.Run("invalid utf8", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsUTF8Bytes([]byte{0xC1})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xC1, 0x01})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xE1, 0x80})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xE1, 0x01, 0x81})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xE1, 0x80, 0x01})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xF1, 0x80, 0x80})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xF1, 0x70, 0x80, 0x80})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xF1, 0x80, 0x70, 0x80})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xF1, 0x80, 0x80, 0x70})).IsFalse()
		assert(IsUTF8Bytes([]byte{0xFF, 0x80, 0x80, 0x70})).IsFalse()
	})

	t.Run("valid utf8", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsUTF8Bytes(([]byte)("abc"))).IsTrue()
		assert(IsUTF8Bytes(([]byte)("abc！#@¥#%#%#¥%"))).IsTrue()
		assert(IsUTF8Bytes(([]byte)("中文"))).IsTrue()
		assert(IsUTF8Bytes(([]byte)("🀄️文👃d"))).IsTrue()
		assert(IsUTF8Bytes(([]byte)("🀄️文👃"))).IsTrue()
		assert(IsUTF8Bytes(([]byte)(`
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
	})
}

func TestGetSeed(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		seed := GetSeed()
		assert(seed > 10000).IsTrue()
		for i := int64(0); i < 500; i++ {
			assert(GetSeed()).Equal(seed + 1 + i)
		}
	})
}

func TestGetRandString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(GetRandString(-1)).Equal("")
		for i := 0; i < 500; i++ {
			assert(len(GetRandString(i))).Equal(i)
		}
	})
}

func TestAddPrefixPerLine(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(AddPrefixPerLine("", "")).Equal("")
		assert(AddPrefixPerLine("a", "")).Equal("a")
		assert(AddPrefixPerLine("\n", "")).Equal("\n")
		assert(AddPrefixPerLine("a\n", "")).Equal("a\n")
		assert(AddPrefixPerLine("a\nb", "")).Equal("a\nb")
		assert(AddPrefixPerLine("", "-")).Equal("-")
		assert(AddPrefixPerLine("a", "-")).Equal("-a")
		assert(AddPrefixPerLine("\n", "-")).Equal("-\n")
		assert(AddPrefixPerLine("a\n", "-")).Equal("-a\n")
		assert(AddPrefixPerLine("a\nb", "-")).Equal("-a\n-b")
	})
}

func TestConcatString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(ConcatString("", "")).Equal("")
		assert(ConcatString("a", "")).Equal("a")
		assert(ConcatString("", "b")).Equal("b")
		assert(ConcatString("a", "b")).Equal("ab")
		assert(ConcatString("a", "b", "")).Equal("ab")
		assert(ConcatString("a", "b", "c")).Equal("abc")
	})
}

func TestGetFileLine(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		fileLine1 := GetFileLine(0)
		assert(strings.Contains(fileLine1, "base_test.go")).IsTrue()
	})
}

func TestAddFileLine(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := AddFileLine("", 0)
		assert(strings.HasPrefix(v1, " ")).IsFalse()
		assert(strings.Contains(v1, "base_test.go")).IsTrue()
	})

	t.Run("skip overflow", func(t *testing.T) {
		assert := NewAssert(t)
		assert(AddFileLine("header", 1000)).Equal("header")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := AddFileLine("header", 0)
		assert(strings.HasPrefix(v1, "header ")).IsTrue()
		assert(strings.Contains(v1, "base_test.go")).IsTrue()
	})
}

func TestConvertOrdinalToString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(ConvertOrdinalToString(0)).Equal("")
		assert(ConvertOrdinalToString(1)).Equal("1st")
		assert(ConvertOrdinalToString(2)).Equal("2nd")
		assert(ConvertOrdinalToString(3)).Equal("3rd")
		assert(ConvertOrdinalToString(4)).Equal("4th")
		assert(ConvertOrdinalToString(10)).Equal("10th")
		assert(ConvertOrdinalToString(100)).Equal("100th")
	})
}

func TestWaitAtLeastDurationWhenRunning(t *testing.T) {
	t.Run("test isRunning return true", func(t *testing.T) {
		waitCH := make(chan bool)
		for i := 0; i < 1000; i++ {
			go func() {
				assert := NewAssert(t)
				startTime := TimeNow()
				WaitAtLeastDurationWhenRunning(TimeNow().UnixNano(), func() bool {
					return true
				}, 500*time.Millisecond)
				interval := TimeNow().Sub(startTime)
				fmt.Println(interval)
				assert(interval > 480*time.Millisecond).IsTrue()
				assert(interval < 580*time.Millisecond).IsTrue()
				waitCH <- true
			}()
		}
		for i := 0; i < 1000; i++ {
			<-waitCH
		}
	})

	t.Run("test isRunning return false 1", func(t *testing.T) {
		waitCH := make(chan bool)
		for i := 0; i < 1000; i++ {
			go func() {
				assert := NewAssert(t)
				startTime := TimeNow()
				WaitAtLeastDurationWhenRunning(TimeNow().UnixNano(), func() bool {
					return false
				}, 500*time.Millisecond)
				interval := TimeNow().Sub(startTime)
				assert(interval >= 0).IsTrue()
				assert(interval < 80*time.Millisecond).IsTrue()
				waitCH <- true
			}()
		}
		for i := 0; i < 1000; i++ {
			<-waitCH
		}
	})

	t.Run("test isRunning return false 2", func(t *testing.T) {
		waitCH := make(chan bool)
		for i := 0; i < 1000; i++ {
			go func() {
				assert := NewAssert(t)
				startTime := TimeNow()
				count := 0
				WaitAtLeastDurationWhenRunning(TimeNow().UnixNano(), func() bool {
					count++
					return count < 3
				}, 500*time.Millisecond)
				interval := TimeNow().Sub(startTime)
				assert(interval >= 180*time.Millisecond).IsTrue()
				assert(interval < 280*time.Millisecond).IsTrue()
				waitCH <- true
			}()
		}
		for i := 0; i < 1000; i++ {
			<-waitCH
		}
	})
}

func TestIsTCPPortOccupied(t *testing.T) {
	t.Run("not occupied", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsTCPPortOccupied(65535)).Equal(false)
	})

	t.Run("occupied", func(t *testing.T) {
		assert := NewAssert(t)
		Listener, _ := net.Listen("tcp", "127.0.0.1:65535")
		assert(IsTCPPortOccupied(65535)).Equal(true)
		_ = Listener.Close()
	})
}

func TestReadFromFile(t *testing.T) {
	t.Run("file not exist", func(t *testing.T) {
		assert := NewAssert(t)
		v1, err1 := ReadFromFile("./no_file")
		assert(v1).Equal("")
		assert(err1).IsNotNil()
		assert(strings.Contains(err1.Error(), "no_file")).IsTrue()
	})

	t.Run("file exist", func(t *testing.T) {
		assert := NewAssert(t)
		_ = ioutil.WriteFile("./tmp_file", []byte("hello"), 0666)
		assert(ReadFromFile("./tmp_file")).Equal("hello", nil)
		_ = os.Remove("./tmp_file")
	})
}

func BenchmarkGetFileLine(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			AddFileLine("hello", 1)
		}
	})
}

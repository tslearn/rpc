package internal

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestDebug(t *testing.T) {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, 0x112233445566)

	fmt.Println(buf)
}

func BenchmarkDebug(b *testing.B) {
	b.ReportAllocs()

	stream := NewStream()
	stream.SetWritePos(200)

	bytes := make([]byte, 10)

	for i := 0; i < b.N; i++ {
		stream.PutBytesTo(bytes, 300)
	}

}

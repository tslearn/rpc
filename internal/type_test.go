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

func TestRTPosRecord(t *testing.T) {
	assert := NewAssert(t)

	posArray := []int64{
		0,
		1,
		1000,
		0x00FFFFFFFFFFFFFE,
		0x00FFFFFFFFFFFFFF,
		0x7FFFFFFFFFFFFFFE,
		0x7FFFFFFFFFFFFFFF,
	}

	flagArray := []bool{
		false,
		true,
	}

	for i := 0; i < len(posArray); i++ {
		for j := 0; j < len(flagArray); j++ {
			pos := posArray[i]
			flag := flagArray[j]

			record := makePosRecord(pos, flag)

			fmt.Println("record", record)
			assert(record.getPos()).Equals(pos)
			assert(record.isString()).Equals(flag)

			fmt.Println(record)
		}
	}
}

func BenchmarkRTPosRecord(b *testing.B) {
	pos := int64(128)
	flag := true
	record := posRecord(0)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		record = makePosRecord(pos, flag)
		pos = record.getPos()
		flag = record.isString()
	}

	fmt.Println(record)
}

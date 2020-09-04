package internal

import (
	"fmt"
	"testing"
)

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
			assert(record.needBuffer()).Equals(flag)

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
		flag = record.needBuffer()
	}

	fmt.Println(record)
}

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

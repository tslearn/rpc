package util

import (
	"math/rand"
	"time"
)

const (
	base64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
)

// GetRandString get random string
func GetRandString(strLen int) string {
	sb := NewStringBuilder()
	for strLen > 0 {
		rand64 := rand.Uint64()
		for used := 0; used < 10 && strLen > 0; used++ {
			sb.AppendByte(base64String[rand64%64])
			rand64 = rand64 / 64
			strLen--
		}
	}
	ret := sb.String()
	sb.Release()
	return ret
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

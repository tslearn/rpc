package router

import (
	"fmt"
	"testing"
)

func TestDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		mp := make(map[string]int)
		mp["hi"] = 3
		fmt.Println(mp)
	})
}

package lab

import (
	"testing"
	"time"
)

func TestCallbackLogWriter_Write(t *testing.T) {
	for {
		time.Sleep(100 * time.Microsecond)
		//runtime.Gosched()
	}
}

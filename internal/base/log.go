package base

import (
	"io"
	"os"
	"sync/atomic"
	"unsafe"
)

type logHelper struct {
	writer io.Writer
}

var logWriter = unsafe.Pointer(&logHelper{
	writer: os.Stdout,
})

func Log(str string) {
	if helper := (*logHelper)(atomic.LoadPointer(&logWriter)); helper != nil {
		_, _ = helper.writer.Write(StringToBytesUnsafe(str))
	} else {
		_, _ = os.Stdout.WriteString(str)
	}
}

func SetLogWriter(writer io.Writer) {
	if writer != nil {
		atomic.StorePointer(&logWriter, unsafe.Pointer(&logHelper{
			writer: writer,
		}))
	}
}

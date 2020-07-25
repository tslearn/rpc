package internal

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
)

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) ReplyCacheFunc {
	switch fnString {
	case "S":
		return func(c *ContextObject, s *Stream, f interface{}) bool {
			h, ok := s.ReadString()
			if !ok || s.CanRead() {
				return false
			}
			f.(func(*ContextObject, string) *ReturnObject)(c, h)
			return true
		}
	default:
		return nil
	}
}

func getNewProcessor() *Processor {
	return NewProcessor(
		true,
		8192,
		16,
		32,
		nil,
	)
}

func readStringFromFile(filePath string) (string, error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func TestBuildFuncCache(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
	}()

	// Test(1)
	tmpFile1 := path.Join(path.Dir(file), "_tmp_/test-cache-01.go")
	snapshotFile1 := path.Join(path.Dir(file), "snapshot/test-cache-01.snapshot")
	assert(buildFuncCache("pkgName", tmpFile1, []string{})).IsNil()
	assert(readStringFromFile(tmpFile1)).Equals(readStringFromFile(snapshotFile1))

	// Test(2)
	tmpFile2 := path.Join(path.Dir(file), "_tmp_/test-cache-02.go")
	snapshotFile2 := path.Join(path.Dir(file), "snapshot/test-cache-02.snapshot")
	assert(buildFuncCache("pkgName", tmpFile2, []string{
		"BMUF", "UUIB", "MSXA", "FFFFF",
		"",
		"B", "I", "U", "F", "S", "X", "A", "M",
		"BU", "FI", "AU", "FM", "SX", "BX", "MA", "MI",
		"BUF", "ABM", "UFS", "XAA", "MMS", "MMM", "AAA", "MFF",
		"BIUFSXAM", "AAAAAAAA", "MAXSFUIB",
	})).IsNil()
	assert(readStringFromFile(tmpFile2)).Equals(readStringFromFile(snapshotFile2))
}

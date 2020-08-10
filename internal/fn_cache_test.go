package internal

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
)

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
	assert(testReadFromFile(tmpFile1)).Equals(testReadFromFile(snapshotFile1))

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
	assert(testReadFromFile(tmpFile2)).Equals(testReadFromFile(snapshotFile2))

	// Test(3)
	tmpFile3 := path.Join(path.Dir(file), "_tmp_/test-cache-03.go")
	assert(buildFuncCache("pkgName", tmpFile3, []string{"A", "A"})).
		Equals(NewKernelPanic("duplicate kind A"))

	// Test(4)
	tmpFile4 := path.Join(path.Dir(file), "_tmp_/test-cache-04.go")
	assert(buildFuncCache("pkgName", tmpFile4, []string{"Y", "A"})).
		Equals(NewKernelPanic("error kind Y"))

	// Test(5)
	tmpFile5 := path.Join(path.Dir(file), "fn_cache_test.go", "test-cache-05.go")
	fmt.Println(buildFuncCache("pkgName", tmpFile5, []string{"A"}))
	assert(strings.Contains(
		buildFuncCache("pkgName", tmpFile5, []string{"A"}).Error(),
		"fn_cache_test.go: not a directory",
	)).IsTrue()

	// Test(6)
	tmpFile6 := path.Join(path.Dir(file), "_tmp_")
	fmt.Println(buildFuncCache("pkgName", tmpFile6, []string{"A"}))
	assert(strings.Contains(
		buildFuncCache("pkgName", tmpFile6, []string{"A"}).Error(),
		"_tmp_: is a directory",
	)).IsTrue()
}

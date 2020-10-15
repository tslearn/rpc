package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestBuildFuncCache(t *testing.T) {
	_, currFile, _, _ := runtime.Caller(0)
	currDir := path.Dir(currFile)
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(currFile), "_tmp_"))
	}()

	t.Run("duplicate kind", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(currDir, "_tmp_/duplicate-kind.go")
		assert(buildFuncCache("pkgName", filePath, []string{"A", "A"})).
			Equal(errors.ErrFnCacheDuplicateKindString.AddDebug("duplicate kind A"))
	})

	t.Run("illegal kind", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(currDir, "_tmp_/illegal-kind.go")
		assert(buildFuncCache("pkgName", filePath, []string{"T", "A"})).
			Equal(errors.ErrFnCacheIllegalKindString.AddDebug("illegal kind T"))
	})

	t.Run("mkdir error", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(currDir, "fn_cache_test.go", "error.go")
		assert(buildFuncCache("pkgName", filePath, []string{"A"})).
			Equal(errors.ErrCacheMkdirAll)
	})

	t.Run("write to file error", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(currDir, "_snapshot_")
		assert(buildFuncCache("pkgName", filePath, []string{"A"})).
			Equal(errors.ErrCacheWriteFile)
	})

	t.Run("kinds is empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(currDir, "_tmp_/test-cache-01.go")
		snapPath := path.Join(currDir, "_snapshot_/test-cache-01.snapshot")
		assert(buildFuncCache("pkgName", filePath, []string{})).IsNil()
		assert(base.ReadFromFile(filePath)).Equal(base.ReadFromFile(snapPath))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(currDir, "_tmp_/test-cache-02.go")
		snapPath := path.Join(currDir, "_snapshot_/test-cache-02.snapshot")
		assert(buildFuncCache("pkgName", filePath, []string{
			"",
			"BIUFSXAMVYZ",
			"B", "I", "U", "F", "S", "X", "A", "M", "V", "Y", "Z",
			"BY", "IY", "UY", "FY", "SY", "XY", "AY", "MY", "VY", "YY", "ZY",
			"BUF", "ABM", "UFS", "XAA", "MMS", "MMM", "AAA", "MFF",
			"BIUFSXAM", "AAAAAAAA", "MAXSFUIB",
			"BIUFSXAMYZ", "BIUSXAMVYZ", "BIUFXAMVYZ",
		})).IsNil()
		assert(base.ReadFromFile(filePath)).Equal(base.ReadFromFile(snapPath))
	})
}

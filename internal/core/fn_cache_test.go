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

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) ActionCacheFunc {
	switch fnString {
	case "":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime) Return)(rt)
				return true
			}
		}
	case "S":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadString(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, String) Return)(rt, arg0)
				return true
			}
		}
	case "I":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadInt64(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Int64) Return)(rt, arg0)
				return true
			}
		}
	case "M":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadMap(); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Map) Return)(rt, arg0)
				return true
			}
		}
	case "V":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadRTValue(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTValue) Return)(rt, arg0)
				return true
			}
		}
	case "Y":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadRTArray(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTArray) Return)(rt, arg0)
				return true
			}
		}
	case "Z":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadRTMap(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTMap) Return)(rt, arg0)
				return true
			}
		}
	case "IV":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadInt64(); err != nil {
				return false
			} else if arg1, err := stream.ReadRTValue(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Int64, RTValue) Return)(rt, arg0, arg1)
				return true
			}
		}
	case "BIUFSXAMVYZ":
		return func(rt Runtime, stream *Stream, fn interface{}) bool {
			if arg0, err := stream.ReadBool(); err != nil {
				return false
			} else if arg1, err := stream.ReadInt64(); err != nil {
				return false
			} else if arg2, err := stream.ReadUint64(); err != nil {
				return false
			} else if arg3, err := stream.ReadFloat64(); err != nil {
				return false
			} else if arg4, err := stream.ReadString(); err != nil {
				return false
			} else if arg5, err := stream.ReadBytes(); err != nil {
				return false
			} else if arg6, err := stream.ReadArray(); err != nil {
				return false
			} else if arg7, err := stream.ReadMap(); err != nil {
				return false
			} else if arg8, err := stream.ReadRTValue(rt); err != nil {
				return false
			} else if arg9, err := stream.ReadRTArray(rt); err != nil {
				return false
			} else if arg10, err := stream.ReadRTMap(rt); err != nil {
				return false
			} else if !stream.IsReadFinish() {
				return false
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(
					Runtime, Bool, Int64, Uint64, Float64, String, Bytes,
					Array, Map, RTValue, RTArray, RTMap,
				) Return)(
					rt, arg0, arg1, arg2, arg3, arg4, arg5,
					arg6, arg7, arg8, arg9, arg10,
				)
				return true
			}
		}
	default:
		return nil
	}
}

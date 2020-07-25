package internal

import (
	"fmt"
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
	_, file, _, _ := runtime.Caller(0)
	processor1 := getNewProcessor()
	_ = processor1.AddService(
		"user",
		NewService().
			Reply("hello", func(ctx Context,
				b bool, i int64, u uint64, f float64, s string,
				x Bytes, a Array, m Map,
			) Return {
				return ctx.OK(true)
			}).
			Reply("ok", func(ctx Context) Return {
				return ctx.OK(true)
			}),
		"",
	)
	processor1.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-1.go"),
	)
}

func TestFnCache_basic(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)

	processor0 := getNewProcessor()
	assert(processor0.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-0.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-0.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-0.go")))

	processor1 := getNewProcessor()
	_ = processor1.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject) Return {
			return ctx.OK(true)
		}), "")
	assert(processor1.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-1.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-1.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-1.go")))

	processor2 := getNewProcessor()
	_ = processor2.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Bool) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor2.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-2.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-2.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-2.go")))

	processor3 := getNewProcessor()
	_ = processor3.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Int64) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor3.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-3.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-3.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-3.go")))

	processor4 := getNewProcessor()
	_ = processor4.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Uint64) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor4.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-4.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-4.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-4.go")))

	processor5 := getNewProcessor()
	_ = processor5.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Float64) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor5.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-5.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-5.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-5.go")))

	processor6 := getNewProcessor()
	_ = processor6.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ String) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor6.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-6.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-6.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-6.go")))

	processor7 := getNewProcessor()
	_ = processor7.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Bytes) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor7.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-7.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-7.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-7.go")))

	processor8 := getNewProcessor()
	_ = processor8.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Array) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor8.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-8.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-8.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-8.go")))

	processor9 := getNewProcessor()
	_ = processor9.AddService("abc", NewService().
		Reply("sayHello", func(ctx *ContextObject, _ Map) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor9.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-9.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-9.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-9.go")))

	processor10 := getNewProcessor()
	_ = processor10.AddService("abc", NewService().
		Reply("sayHello", func(
			ctx *ContextObject, _ Bool, _ Int64, _ Uint64, _ Float64, _ String,
			_ Bytes, _ Array, _ Map,
		) *ReturnObject {
			return ctx.OK(true)
		}), "")
	assert(processor10.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-10.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-10.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-10.go")))

	_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
}

func Test_debug(t *testing.T) {
	metas, _ := getFuncMetas([]string{"AM", "AI", "U", "AMI", ""})

	for _, meta := range metas {
		fmt.Println(meta.name)
		fmt.Println(meta.identifier)
		fmt.Println(meta.body)
	}

	// fmt.Println(getFuncBodyByKind("fnCache1", "AM"))
}

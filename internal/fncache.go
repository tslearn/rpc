package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// ReplyCache ...
type ReplyCache interface {
	Get(fnString string) ReplyCacheFunc
}

// ReplyCacheFunc ...
type ReplyCacheFunc = func(
	ctx *ContextObject,
	stream *Stream,
	fn interface{},
) bool

type fnCache struct{}

func (p *fnCache) getParamName(idx int) string {
	if idx < 6 {
		return []string{
			"a", "b", "c", "d", "e", "f",
		}[idx]
	}
	return fmt.Sprintf("pa%d", idx)
}

func (p *fnCache) getOKName(idx int) string {
	if idx < 6 {
		return []string{
			"g", "h", "i", "j", "k", "l",
		}[idx]
	}
	return fmt.Sprintf("ok%d", idx)
}

func (p *fnCache) writeHeader(
	pkgName string,
	sb *StringBuilder,
	kinds []string,
) {
	kindMap := make(map[int32]bool)
	for _, kind := range kinds {
		for _, char := range kind {
			kindMap[char] = true
		}
	}
	sb.AppendString(fmt.Sprintf("package %s\n\n", pkgName))
	sb.AppendString("import \"github.com/rpccloud/rpc/core\"\n\n")

	sb.AppendString("type rpcCache struct{}\n\n")

	sb.AppendString("// NewRPCCache ...\n")
	sb.AppendString("func NewRPCCache() common.ReplyCache {\n")
	sb.AppendString("\treturn &rpcCache{}\n")
	sb.AppendString("}\n\n")

	sb.AppendString("// Get ...\n")
	sb.AppendString(
		"func (p *rpcCache) Get(fnString string) common.ReplyCacheFunc {\n",
	)
	sb.AppendString("\treturn getFCache(fnString)\n")
	sb.AppendString("}\n\n")
	sb.AppendString("type n = bool\n")
	sb.AppendString("type o = common.ContextObject\n")
	sb.AppendString("type p = common.ReturnObject\n")
	sb.AppendString("type q = *common.Stream\n")
	if _, ok := kindMap['B']; ok {
		sb.AppendString("type r = common.Bool\n")
	}
	if _, ok := kindMap['I']; ok {
		sb.AppendString("type s = common.Int64\n")
	}
	if _, ok := kindMap['U']; ok {
		sb.AppendString("type t = common.Uint64\n")
	}
	if _, ok := kindMap['F']; ok {
		sb.AppendString("type u = common.Float64\n")
	}
	if _, ok := kindMap['S']; ok {
		sb.AppendString("type v = common.String\n")
	}
	if _, ok := kindMap['X']; ok {
		sb.AppendString("type w = common.Bytes\n")
	}
	if _, ok := kindMap['A']; ok {
		sb.AppendString("type x = common.Array\n")
	}
	if _, ok := kindMap['M']; ok {
		sb.AppendString("type y = common.Map\n")
	}
	sb.AppendString("type z = interface{}\n\n")
	sb.AppendString("const af = false\n")
	sb.AppendString("const at = true\n")
}

func (p *fnCache) writeGetFunc(sb *StringBuilder, kinds []string) {
	sb.AppendString("\nfunc getFCache(fnString string) common.ReplyCacheFunc {\n")
	sb.AppendString("\tswitch fnString {\n")

	for _, kind := range kinds {
		sb.AppendString(fmt.Sprintf("\tcase \"%s\":\n", kind))
		sb.AppendString(fmt.Sprintf("\t\treturn fc%s\n", kind))
	}

	sb.AppendString("\t}\n\n")
	sb.AppendString("\treturn nil\n")

	sb.AppendString("}\n")
}

func (p *fnCache) writeFunctions(sb *StringBuilder, kinds []string) {
	for _, kind := range kinds {
		p.writeFunc(sb, kind)
	}
}

func (p *fnCache) writeFunc(sb *StringBuilder, kind string) {
	sb.AppendString(fmt.Sprintf("\nfunc fc%s(m o, q q, z z) n {\n", kind))

	sbBody := NewStringBuilder()
	sbType := NewStringBuilder()
	sbParam := NewStringBuilder()
	sbOK := NewStringBuilder()
	for idx, c := range kind {
		paramName := p.getParamName(idx)
		okName := p.getOKName(idx)
		sbParam.AppendString(fmt.Sprintf(", %s", paramName))
		sbOK.AppendString(fmt.Sprintf("!%s || ", okName))
		switch c {
		case 'B':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadBool()\n", paramName, okName),
			)
			sbType.AppendString(", r")
		case 'I':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadInt64()\n", paramName, okName),
			)
			sbType.AppendString(", s")
		case 'U':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadUint64()\n", paramName, okName),
			)
			sbType.AppendString(", t")
		case 'F':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadFloat64()\n", paramName, okName),
			)
			sbType.AppendString(", u")
		case 'S':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadString()\n", paramName, okName),
			)
			sbType.AppendString(", v")
		case 'X':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadBytes()\n", paramName, okName),
			)
			sbType.AppendString(", w")
		case 'A':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadArray()\n", paramName, okName),
			)
			sbType.AppendString(", x")
		case 'M':
			sbBody.AppendString(
				fmt.Sprintf("\t%s, %s := q.ReadMap()\n", paramName, okName),
			)
			sbType.AppendString(", y")
		}
	}

	sb.AppendString(sbBody.String())
	sb.AppendString(fmt.Sprintf("\tif %sq.CanRead() {\n", sbOK.String()))
	sb.AppendString("\t\treturn af\n")
	sb.AppendString("\t}\n")

	sb.AppendString(fmt.Sprintf(
		"\tz.(func(o%s) p)(m%s)\n",
		sbType.String(),
		sbParam.String(),
	))
	sb.AppendString("\treturn at\n")
	sb.AppendString("}\n")

	sbBody.Release()
	sbType.Release()
	sbParam.Release()
	sbOK.Release()
}

func buildFuncCache(pkgName string, output string, kinds []string) Error {
	sb := NewStringBuilder()
	defer sb.Release()
	cache := &fnCache{}
	cache.writeHeader(pkgName, sb, kinds)
	cache.writeGetFunc(sb, kinds)
	cache.writeFunctions(sb, kinds)

	if err := os.MkdirAll(path.Dir(output), os.ModePerm); err != nil {
		return NewError(ErrKindFromNone, err.Error())
	} else if err := ioutil.WriteFile(
		output,
		[]byte(sb.String()),
		0666,
	); err != nil {
		return NewError(ErrKindFromNone, err.Error())
	} else {
		return nil
	}
}

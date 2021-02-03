package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

type rpcFuncMeta struct {
	name       string
	body       string
	identifier string
}

func getFuncBodyByKind(name string, kind string) (string, *base.Error) {
	sb := base.NewStringBuilder()
	defer sb.Release()

	sb.AppendString(fmt.Sprintf(
		"func %s(rt rpc.Runtime, stream *rpc.Stream, fn interface{}) int {\n",
		name,
	))

	argArray := []string{"rt"}
	typeArray := []string{"rpc.Runtime"}

	if kind == "" {
		sb.AppendString("\tif !stream.IsReadFinish() {\n\t\treturn -1\n\t}")
	} else {
		for idx, c := range kind {
			argName := "arg" + strconv.Itoa(idx)
			argArray = append(argArray, argName)
			callString := ""

			switch c {
			case vkBool:
				callString = "stream.ReadBool()"
				typeArray = append(typeArray, "rpc.Bool")
			case vkInt64:
				callString = "stream.ReadInt64()"
				typeArray = append(typeArray, "rpc.Int64")
			case vkUint64:
				callString = "stream.ReadUint64()"
				typeArray = append(typeArray, "rpc.Uint64")
			case vkFloat64:
				callString = "stream.ReadFloat64()"
				typeArray = append(typeArray, "rpc.Float64")
			case vkString:
				callString = "stream.ReadString()"
				typeArray = append(typeArray, "rpc.String")
			case vkBytes:
				callString = "stream.ReadBytes()"
				typeArray = append(typeArray, "rpc.Bytes")
			case vkArray:
				callString = "stream.ReadArray()"
				typeArray = append(typeArray, "rpc.Array")
			case vkRTArray:
				callString = "stream.ReadRTArray(rt)"
				typeArray = append(typeArray, "rpc.RTArray")
			case vkMap:
				callString = "stream.ReadMap()"
				typeArray = append(typeArray, "rpc.Map")
			case vkRTMap:
				callString = "stream.ReadRTMap(rt)"
				typeArray = append(typeArray, "rpc.RTMap")
			case vkRTValue:
				callString = "stream.ReadRTValue(rt)"
				typeArray = append(typeArray, "rpc.RTValue")
			default:
				return "", base.ErrFnCacheIllegalKindString.
					AddDebug(fmt.Sprintf("illegal kind %s", kind))
			}

			condString := " else if"

			if idx == 0 {
				condString = "\tif"
			}

			sb.AppendString(fmt.Sprintf(
				"%s %s, err := %s; err != nil {\n\t\treturn %d\n\t}",
				condString,
				argName,
				callString,
				idx+1,
			))
		}

		sb.AppendString(" else if !stream.IsReadFinish() {\n\t\treturn -1\n\t}")
	}

	sb.AppendString(fmt.Sprintf(
		" else {"+
			"\n\t\tstream.SetWritePosToBodyStart()"+
			"\n\t\tfn.(func(%s) rpc.Return)(%s)\n\t\t"+
			"return 0\n\t}\n}",
		strings.Join(typeArray, ", "),
		strings.Join(argArray, ", "),
	))
	return sb.String(), nil
}

func getFuncMetas(kinds []string) ([]*rpcFuncMeta, *base.Error) {
	sortKinds := make([]string, len(kinds))
	copy(sortKinds, kinds)
	sort.SliceStable(sortKinds, func(i, j int) bool {
		if len(sortKinds[i]) < len(sortKinds[j]) {
			return true
		} else if len(sortKinds[i]) > len(sortKinds[j]) {
			return false
		} else {
			return strings.Compare(sortKinds[i], sortKinds[j]) < 0
		}
	})

	funcMap := make(map[string]bool)
	ret := make([]*rpcFuncMeta, 0)

	for idx, kind := range sortKinds {
		fnName := "fnCache" + strconv.Itoa(idx)
		if _, ok := funcMap[kind]; ok {
			return nil, base.ErrFnCacheDuplicateKindString.AddDebug(
				fmt.Sprintf("duplicate kind %s", kind),
			)
		} else if fnBody, err := getFuncBodyByKind(fnName, kind); err != nil {
			return nil, err
		} else {
			funcMap[kind] = true
			ret = append(ret, &rpcFuncMeta{
				name:       fnName,
				body:       fnBody,
				identifier: kind,
			})
		}
	}

	return ret, nil
}

func buildFuncCache(pkgName string, output string, kinds []string) *base.Error {
	sb := base.NewStringBuilder()
	defer sb.Release()
	if metas, err := getFuncMetas(kinds); err == nil {
		sb.AppendString(fmt.Sprintf("package %s\n\n", pkgName))
		sb.AppendString("import \"github.com/rpccloud/rpc\"\n\n")

		sb.AppendString("type rpcCache struct{}\n\n")

		sb.AppendString("// NewRPCCache ...\n")
		sb.AppendString("func NewRPCCache() rpc.ActionCache {\n")
		sb.AppendString("\treturn &rpcCache{}\n")
		sb.AppendString("}\n\n")

		sb.AppendString("// Get ...\n")
		sb.AppendString(
			"func (p *rpcCache) Get(fnString string) rpc.ActionCacheFunc {\n",
		)
		sb.AppendString(
			"\tswitch fnString {\n",
		)
		for _, meta := range metas {
			sb.AppendString(
				fmt.Sprintf("\tcase \"%s\":\n", meta.identifier),
			)
			sb.AppendString(
				fmt.Sprintf("\t\treturn %s\n", meta.name),
			)
		}
		sb.AppendString(
			"\tdefault:\n\t\treturn nil\n\t}\n}\n\n",
		)

		for _, meta := range metas {
			sb.AppendString(
				fmt.Sprintf("%s\n\n", meta.body),
			)
		}
	} else {
		return err
	}

	if err := os.MkdirAll(path.Dir(output), os.ModePerm); err != nil {
		return base.ErrCacheMkdirAll
	} else if err := ioutil.WriteFile(
		output,
		[]byte(sb.String()),
		0666,
	); err != nil {
		return base.ErrCacheWriteFile
	} else {
		return nil
	}
}

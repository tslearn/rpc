package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func getFuncBodyByKind(name string, kind string) (string, error) {
	sb := NewStringBuilder()
	defer sb.Release()

	sb.AppendString(fmt.Sprintf(
		"func %s(ctx rpc.Context, stream rpc.Stream, fn interface{}) bool {\n",
		name,
	))

	argArray := []string{"ctx"}
	typeArray := []string{"rpc.Context"}

	if kind == "" {
		sb.AppendString("\tif !stream.IsReadFinish() {\n\t\treturn true\n\t}")
	} else {
		for idx, c := range kind {
			argName := "arg" + strconv.Itoa(idx)
			argArray = append(argArray, argName)
			callString := ""

			switch c {
			case 'B':
				callString = "ReadBool"
				typeArray = append(typeArray, "rpc.Bool")
			case 'I':
				callString = "ReadInt64"
				typeArray = append(typeArray, "rpc.Int64")
			case 'U':
				callString = "ReadUint64"
				typeArray = append(typeArray, "rpc.Uint64")
			case 'F':
				callString = "ReadFloat64"
				typeArray = append(typeArray, "rpc.Float64")
			case 'S':
				callString = "ReadString"
				typeArray = append(typeArray, "rpc.String")
			case 'X':
				callString = "ReadBytes"
				typeArray = append(typeArray, "rpc.Bytes")
			case 'A':
				callString = "ReadArray"
				typeArray = append(typeArray, "rpc.Array")
			case 'M':
				callString = "ReadMap"
				typeArray = append(typeArray, "rpc.Map")
			default:
				return "nil", errors.New(fmt.Sprintf("error kind %s", kind))
			}

			condString := " else if"

			if idx == 0 {
				condString = "\tif"
			}

			sb.AppendString(fmt.Sprintf(
				"%s %s, ok := stream.%s(); !ok {\n\t\treturn false\n\t}",
				condString,
				argName,
				callString,
			))
		}

		sb.AppendString(" else if !stream.IsReadFinish() {\n\t\t return true\n\t}")
	}

	sb.AppendString(fmt.Sprintf(
		" else {\n\t\tfn.(func(%s) rpc.Return)(%s)\n\t\treturn true\n\t}\n}",
		strings.Join(typeArray, ", "),
		strings.Join(argArray, ", "),
	))
	return sb.String(), nil
}

func getFuncMetas(kinds []string) ([]*rpcFuncMeta, error) {
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
			return nil, errors.New(fmt.Sprintf("duplicate kind %s", kind))
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

func buildFuncCache(pkgName string, output string, kinds []string) error {
	sb := NewStringBuilder()
	defer sb.Release()
	if metas, err := getFuncMetas(kinds); err != nil {
		return err
	} else {
		sb.AppendString(fmt.Sprintf("package %s\n\n", pkgName))
		sb.AppendString("import \"github.com/rpccloud/rpc\"\n\n")

		sb.AppendString("type rpcCache struct{}\n\n")

		sb.AppendString("// NewRPCCache ...\n")
		sb.AppendString("func NewRPCCache() rpc.ReplyCache {\n")
		sb.AppendString("\treturn &rpcCache{}\n")
		sb.AppendString("}\n\n")

		sb.AppendString("// Get ...\n")
		sb.AppendString(
			"func (p *rpcCache) Get(fnString string) rpc.ReplyCacheFunc {\n",
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
			"\tdefault: \n\t\treturn nil\n\t}\n}\n\n",
		)

		for _, meta := range metas {
			sb.AppendString(
				fmt.Sprintf("%s\n\n", meta.body),
			)
		}
	}

	if err := os.MkdirAll(path.Dir(output), os.ModePerm); err != nil {
		return NewBaseError(err.Error())
	} else if err := ioutil.WriteFile(
		output,
		[]byte(sb.String()),
		0666,
	); err != nil {
		return NewBaseError(err.Error())
	} else {
		return nil
	}
}

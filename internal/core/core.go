// Package core ...
package core

import (
	"reflect"

	"github.com/rpccloud/rpc/internal/base"
)

func getFuncKind(fn reflect.Value) (string, *base.Error) {
	if fn.Kind() != reflect.Func {
		return "", base.ErrActionHandler.
			AddDebug("handler must be a function")
	} else if fn.Type().NumIn() < 1 ||
		fn.Type().In(0) != reflect.ValueOf(Runtime{}).Type() {
		return "", base.ErrActionHandler.AddDebug(base.ConcatString(
			"handler 1st argument type must be ",
			convertTypeToString(runtimeType)),
		)
	} else if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(emptyReturn).Type() {
		return "", base.ErrActionHandler.AddDebug(base.ConcatString(
			"handler return type must be ",
			convertTypeToString(returnType),
		))
	} else {
		sb := base.NewStringBuilder()
		defer sb.Release()

		for i := 1; i < fn.Type().NumIn(); i++ {
			switch fn.Type().In(i) {
			case boolType:
				sb.AppendByte(vkBool)
			case int64Type:
				sb.AppendByte(vkInt64)
			case uint64Type:
				sb.AppendByte(vkUint64)
			case float64Type:
				sb.AppendByte(vkFloat64)
			case stringType:
				sb.AppendByte(vkString)
			case bytesType:
				sb.AppendByte(vkBytes)
			case arrayType:
				sb.AppendByte(vkArray)
			case mapType:
				sb.AppendByte(vkMap)
			case rtValueType:
				sb.AppendByte(vkRTValue)
			case rtArrayType:
				sb.AppendByte(vkRTArray)
			case rtMapType:
				sb.AppendByte(vkRTMap)
			default:
				return "", base.ErrActionHandler.AddDebug(
					base.ConcatString(
						"handler ",
						base.ConvertOrdinalToString(1+uint(i)),
						" argument type ",
						fn.Type().In(i).String(),
						" is not supported",
					))
			}
		}

		return sb.String(), nil
	}
}

func convertTypeToString(reflectType reflect.Type) string {
	switch reflectType {
	case nil:
		return "<nil>"
	case runtimeType:
		return "rpc.Runtime"
	case returnType:
		return "rpc.Return"
	case boolType:
		return "rpc.Bool"
	case int64Type:
		return "rpc.Int64"
	case uint64Type:
		return "rpc.Uint64"
	case float64Type:
		return "rpc.Float64"
	case stringType:
		return "rpc.String"
	case bytesType:
		return "rpc.Bytes"
	case arrayType:
		return "rpc.Array"
	case mapType:
		return "rpc.Map"
	case rtValueType:
		return "rpc.RTValue"
	case rtArrayType:
		return "rpc.RTArray"
	case rtMapType:
		return "rpc.RTMap"
	default:
		return reflectType.String()
	}
}

func getFastKey(s string) uint32 {
	bytes := base.StringToBytesUnsafe(s)

	if size := len(bytes); size > 0 {
		return uint32(bytes[0])<<16 |
			uint32(bytes[size>>1])<<8 |
			uint32(bytes[size-1])
	}

	return 0
}

// MakeRequestStream ...
func MakeRequestStream(
	debug bool,
	depth uint16,
	target string,
	from string,
	args ...interface{},
) (*Stream, *base.Error) {
	stream := NewStream()
	// set debug bit
	if debug {
		stream.SetStatusBitDebug()
	} else {
		stream.ClearStatusBitDebug()
	}
	// set depth
	stream.SetDepth(depth)
	// write target
	stream.WriteString(target)
	// write from
	stream.WriteString(from)
	// write args
	for i := 0; i < len(args); i++ {
		if reason := stream.Write(args[i]); reason != StreamWriteOK {
			stream.Release()
			return nil, base.ErrUnsupportedValue.AddDebug(
				base.ConcatString(
					base.ConvertOrdinalToString(uint(i)+2),
					" argument: ",
					reason,
				))
		}
	}
	return stream, nil
}

// ParseResponseStream ...
func ParseResponseStream(stream *Stream) (Any, *base.Error) {
	if errCode, err := stream.ReadUint64(); err != nil {
		return nil, err
	} else if errCode == 0 {
		return stream.Read()
	} else if message, err := stream.ReadString(); err != nil {
		return nil, err
	} else if !stream.IsReadFinish() {
		return nil, base.ErrStream
	} else {
		return nil, base.NewError(errCode, message)
	}
}

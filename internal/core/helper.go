package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

func MakeRequestStream(target string, args ...interface{}) (*Stream, *base.Error) {
	stream := NewStream()
	stream.SetDepth(0)
	// write target
	stream.WriteString(target)
	// write from
	stream.WriteString("")
	// write args
	for i := 0; i < len(args); i++ {
		if reason := stream.Write(args[i]); reason != StreamWriteOK {
			return nil, errors.ErrRuntimeArgumentNotSupported.
				AddDebug(base.ConcatString("value", reason))
		}
	}

	return stream, nil
}

func ReadResponseStream(stream *Stream) (interface{}, *base.Error) {
	// parse the stream
	if errCode, ok := stream.ReadUint64(); !ok {
		return nil, errors.ErrBadStream
	} else if errCode == 0 {
		if ret, ok := stream.Read(); ok {
			return ret, nil
		}
		return nil, errors.ErrBadStream
	} else if message, ok := stream.ReadString(); !ok {
		return nil, errors.ErrBadStream
	} else if !stream.IsReadFinish() {
		return nil, errors.ErrBadStream
	} else {
		return nil, base.NewError(errCode, message)
	}
}

package rpc

import (
	"strconv"

	"github.com/rpccloud/rpc/internal/base"
)

func getErrorString(
	machineID uint64,
	sessionID uint64,
	err *base.Error,
) string {
	if err == nil {
		return ""
	}

	machineString := ""
	if machineID != 0 {
		machineString = base.ConcatString(
			"<target:",
			strconv.FormatUint(machineID, 10),
			"> ",
		)
	}
	sessionString := ""
	if sessionID != 0 {
		sessionString = base.ConcatString(
			"<session:",
			strconv.FormatUint(sessionID, 10),
			"> ",
		)
	}

	return base.ConcatString(
		base.ConvertToIsoDateString(base.TimeNow()),
		" ",
		machineString,
		sessionString,
		err.Error(),
		"\n",
	)
}

// StreamHub ...
type StreamHub struct {
	onConnectRequestStream    func(stream *Stream)
	onConnectResponseStream   func(stream *Stream)
	onPingStream              func(stream *Stream)
	onPongStream              func(stream *Stream)
	onRPCRequestStream        func(stream *Stream)
	onRPCResponseOKStream     func(stream *Stream)
	onRPCResponseErrorStream  func(stream *Stream)
	onRPCBoardCastStream      func(stream *Stream)
	onSystemErrorReportStream func(stream *Stream)

	logger   *base.Logger
	logLevel base.ErrorLevel
}

// NewStreamHub ...
func NewStreamHub(
	isLogErrorToScreen bool,
	logFile string,
	logLevel base.ErrorLevel,
	onConnectRequestStream func(stream *Stream),
	onConnectResponseStream func(stream *Stream),
	onPingStream func(stream *Stream),
	onPongStream func(stream *Stream),
	onRPCRequestStream func(stream *Stream),
	onRPCResponseOKStream func(stream *Stream),
	onRPCResponseErrorStream func(stream *Stream),
	onRPCBoardCastStream func(stream *Stream),
	onSystemErrorReportStream func(stream *Stream),
) *StreamHub {
	logger, err := base.NewLogger(isLogErrorToScreen, logFile)

	ret := &StreamHub{
		onConnectRequestStream:    onConnectRequestStream,
		onConnectResponseStream:   onConnectResponseStream,
		onPingStream:              onPingStream,
		onPongStream:              onPongStream,
		onRPCRequestStream:        onRPCRequestStream,
		onRPCResponseOKStream:     onRPCResponseOKStream,
		onRPCResponseErrorStream:  onRPCResponseErrorStream,
		onRPCBoardCastStream:      onRPCBoardCastStream,
		onSystemErrorReportStream: onSystemErrorReportStream,
		logger:                    logger,
		logLevel:                  logLevel,
	}

	if err != nil {
		errStream := MakeSystemErrorStream(err)
		ret.OnReceiveStream(errStream)
	}

	return ret
}

// OnReceiveStream ...
func (p *StreamHub) OnReceiveStream(stream *Stream) {
	if stream != nil {
		fn := (func(stream *Stream))(nil)
		switch stream.GetKind() {
		case StreamKindConnectRequest:
			fn = p.onConnectRequestStream
		case StreamKindConnectResponse:
			fn = p.onConnectResponseStream
		case StreamKindPing:
			fn = p.onPingStream
		case StreamKindPong:
			fn = p.onPongStream
		case StreamKindRPCRequest:
			fn = p.onRPCRequestStream
		case StreamKindRPCResponseOK:
			fn = p.onRPCResponseOKStream
		case StreamKindRPCResponseError:
			fn = p.onRPCResponseErrorStream
		case StreamKindRPCBoardCast:
			fn = p.onRPCBoardCastStream
		case StreamKindSystemErrorReport:
			fn = p.onSystemErrorReportStream
			_, err := ParseResponseStream(stream)

			if err.GetLevel()&p.logLevel == 0 {
				return
			}

			sessionID := stream.GetSessionID()
			p.logger.Log(getErrorString(0, sessionID, err))

			// exchange the stream
			stream, _ = MakeInternalRequestStream(
				false,
				0,
				"$sys.log.ReportError",
				"@",
				err.GetCode(),
				err.GetMessage(),
			)
			stream.SetSessionID(sessionID)
		}

		if fn != nil {
			fn(stream)
		}
	}
}

// Close ...
func (p *StreamHub) Close() bool {
	if err := p.logger.Close(); err != nil {
		errStream := MakeSystemErrorStream(err)
		p.OnReceiveStream(errStream)
		return false
	}

	return true
}

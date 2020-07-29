package internal

// ErrorKind ...
type ErrorKind uint64

// ErrStringRunOutOfReplyScope ...
const ErrStringRunOutOfReplyScope = "run out of reply goroutine"

// ErrStringBadStream ...
const ErrStringBadStream = "bad stream"

const (
	// ErrorKindNone ...
	ErrorKindNone ErrorKind = 0
	// ErrorKindProtocol ...
	ErrorKindProtocol ErrorKind = 1
	// ErrorKindTransport ...
	ErrorKindTransport ErrorKind = 2
	// ErrorKindReply ...
	ErrorKindReply ErrorKind = 3
	// ErrorKindReplyPanic ...
	ErrorKindReplyPanic ErrorKind = 4
	// ErrorKindRuntimePanic ...
	ErrorKindRuntimePanic ErrorKind = 5
	// ErrorKindKernelPanic ...
	ErrorKindKernelPanic ErrorKind = 6
)

var (
	gPanicLocker        = NewLock()
	gPanicSubscriptions = make([]*rpcPanicSubscription, 0)
)

// reportPanic ...
func reportPanic(err Error) {
	defer func() {
		recover()
	}()

	gPanicLocker.DoWithLock(func() {
		for _, sub := range gPanicSubscriptions {
			if sub != nil && sub.onPanic != nil {
				sub.onPanic(err)
			}
		}
	})
}

// subscribePanic ...
func subscribePanic(onPanic func(Error)) *rpcPanicSubscription {
	if onPanic == nil {
		return nil
	}

	return gPanicLocker.CallWithLock(func() interface{} {
		ret := &rpcPanicSubscription{
			id:      GetSeed(),
			onPanic: onPanic,
		}
		gPanicSubscriptions = append(gPanicSubscriptions, ret)
		return ret
	}).(*rpcPanicSubscription)
}

type rpcPanicSubscription struct {
	id      int64
	onPanic func(err Error)
}

func (p *rpcPanicSubscription) Close() bool {
	if p == nil {
		return false
	}

	return gPanicLocker.CallWithLock(func() interface{} {
		for i := 0; i < len(gPanicSubscriptions); i++ {
			if gPanicSubscriptions[i].id == p.id {
				gPanicSubscriptions = append(
					gPanicSubscriptions[:i],
					gPanicSubscriptions[i+1:]...,
				)
				return true
			}
		}
		return false
	}).(bool)
}

// Error ...
type Error interface {
	GetKind() ErrorKind
	GetMessage() string
	GetDebug() string
	AddDebug(debug string) Error
	Error() string
}

// NewError ...
func NewError(kind ErrorKind, message string, debug string) Error {
	return &rpcError{
		kind:    kind,
		message: message,
		debug:   debug,
	}
}

// NewBaseError ...
func NewBaseError(message string) Error {
	return NewError(ErrorKindRuntimePanic, message, "")
}

// NewProtocolError ...
func NewProtocolError(message string) Error {
	return NewError(ErrorKindProtocol, message, "")
}

// NewTransportError ...
func NewTransportError(message string) Error {
	return NewError(ErrorKindTransport, message, "")
}

// NewReplyError ...
func NewReplyError(message string) Error {
	return NewError(ErrorKindReply, message, "")
}

// NewReplyPanic ...
func NewReplyPanic(message string) Error {
	return NewError(ErrorKindReplyPanic, message, "")
}

// NewRuntimePanic ...
func NewRuntimePanic(message string) Error {
	return NewError(ErrorKindRuntimePanic, message, "")
}

// NewKernelPanic ...
func NewKernelPanic(message string) Error {
	return NewError(ErrorKindKernelPanic, message, "")
}

// ConvertToError convert interface{} to Error if type matches
func ConvertToError(v interface{}) Error {
	if ret, ok := v.(Error); ok {
		return ret
	}

	return nil
}

type rpcError struct {
	kind    ErrorKind
	message string
	debug   string
}

func (p *rpcError) GetKind() ErrorKind {
	return p.kind
}

func (p *rpcError) GetMessage() string {
	return p.message
}

func (p *rpcError) GetDebug() string {
	return p.debug
}

func (p *rpcError) AddDebug(debug string) Error {
	if p.debug == "" {
		p.debug = debug
	} else {
		p.debug += "\n"
		p.debug += debug
	}

	return p
}

func (p *rpcError) Error() string {
	sb := NewStringBuilder()
	defer sb.Release()

	if p.message != "" {
		sb.AppendString(p.message)
	}

	if p.debug != "" {
		if !sb.IsEmpty() {
			sb.AppendByte('\n')
		}
		sb.AppendString(p.debug)
	}

	return sb.String()
}

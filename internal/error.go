package internal

type ErrorKind uint64

const ErrStringUnexpectedNil = "rpc: unexpected nil"
const ErrStringRunOutOfScope = "rpc: run out of reply goroutine"
const ErrStringBadStream = "rpc: bad stream"
const ErrStringTimeout = "rpc: timeout"

const (
	ErrKindNone       ErrorKind = 0
	ErrKindReply      ErrorKind = 1
	ErrKindReplyPanic ErrorKind = 2
	ErrKindRuntime    ErrorKind = 3
	ErrKindProtocol   ErrorKind = 4
	ErrKindTransport  ErrorKind = 5
	ErrKindKernel     ErrorKind = 5
)

var (
	gPanicLocker        = NewLock()
	gPanicSubscriptions = make([]*rpcPanicSubscription, 0)
)

func ReportPanic(err Error) {
	defer func() {
		recover()
	}()

	gPanicLocker.DoWithLock(func() {
		for _, sub := range gPanicSubscriptions {
			if sub != nil && sub.onFatal != nil {
				sub.onFatal(err)
			}
		}
	})
}

func SubscribePanic(onFatal func(Error)) *rpcPanicSubscription {
	if onFatal == nil {
		return nil
	}

	return gPanicLocker.CallWithLock(func() interface{} {
		ret := &rpcPanicSubscription{
			id:      GetSeed(),
			onFatal: onFatal,
		}
		gPanicSubscriptions = append(gPanicSubscriptions, ret)
		return ret
	}).(*rpcPanicSubscription)
}

type rpcPanicSubscription struct {
	id      int64
	onFatal func(err Error)
}

func (p *rpcPanicSubscription) Close() bool {
	if p == nil {
		return false
	} else {
		return gPanicLocker.CallWithLock(func() interface{} {
			for i := 0; i < len(gPanicSubscriptions); i++ {
				if gPanicSubscriptions[i].id == p.id {
					p.id = 0
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
}

// Error ...
type Error interface {
	GetKind() ErrorKind
	GetMessage() string
	GetDebug() string
	AddDebug(debug string) Error
	Error() string
}

func newError(kind ErrorKind, message string, debug string) Error {
	return &rpcError{
		kind:    kind,
		message: message,
		debug:   debug,
	}
}

// NewError ...
func NewError(kind ErrorKind, message string, debug string) Error {
	return newError(kind, message, debug)
}

// NewBaseError ...
func NewBaseError(message string) Error {
	return newError(ErrKindNone, message, "")
}

// NewReplyError ...
func NewReplyError(message string) Error {
	return newError(ErrKindReply, message, "")
}

// NewReplyPanic ...
func NewReplyPanic(message string) Error {
	return newError(ErrKindReplyPanic, message, "")
}

// NewRuntimeError ...
func NewRuntimeError(message string) Error {
	return newError(ErrKindRuntime, message, "")
}

// NewProtocolError ...
func NewProtocolError(message string) Error {
	return newError(ErrKindProtocol, message, "")
}

// NewTransportError ...
func NewTransportError(message string) Error {
	return newError(ErrKindTransport, message, "")
}

// NewKernelError ...
func NewKernelError(message string) Error {
	return newError(ErrKindKernel, message, "")
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

	if len(p.message) > 0 {
		sb.AppendString(p.message)
	}

	if len(p.debug) > 0 {
		if !sb.IsEmpty() {
			sb.AppendByte('\n')
		}
		sb.AppendString(p.debug)
	}

	return sb.String()
}

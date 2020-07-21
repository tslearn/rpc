package internal

type ErrKind uint64

const (
	ErrKindFromNone      ErrKind = 0
	ErrKindFromProtocol  ErrKind = 1
	ErrKindFromTransport ErrKind = 2
	ErrKindFromTimeout   ErrKind = 3
	ErrKindFromAccess    ErrKind = 4
	ErrKindFromKernel    ErrKind = 5
)

var gFatalErrorReporter = newErrorReporter()

func ReportFatalError(err Error) {
	gFatalErrorReporter.fatalError(err)
}

func SubscribeFatalError(onFatal func(Error)) *rpcFatalSubscription {
	return gFatalErrorReporter.subscribe(onFatal)
}

type rpcErrorReporter struct {
	subscriptions []*rpcFatalSubscription
	Lock
}

func newErrorReporter() *rpcErrorReporter {
	return &rpcErrorReporter{
		subscriptions: make([]*rpcFatalSubscription, 0),
	}
}

func (p *rpcErrorReporter) subscribe(
	onFatal func(Error),
) *rpcFatalSubscription {
	if p == nil || onFatal == nil {
		return nil
	}

	return p.CallWithLock(func() interface{} {
		ret := &rpcFatalSubscription{
			id:       GetSeed(),
			reporter: p,
			onFatal:  onFatal,
		}
		p.subscriptions = append(p.subscriptions, ret)
		return ret
	}).(*rpcFatalSubscription)
}

func (p *rpcErrorReporter) removeSubscription(id int64) bool {
	if p == nil {
		return false
	}

	return p.CallWithLock(func() interface{} {
		for i := 0; i < len(p.subscriptions); i++ {
			if p.subscriptions[i].id == id {
				p.subscriptions[i].id = 0
				array := p.subscriptions
				p.subscriptions = append(array[:i], array[i+1:]...)
				return true
			}
		}
		return false
	}).(bool)
}

func (p *rpcErrorReporter) fatalError(err Error) {
	if p != nil {
		subscriptions := p.CallWithLock(func() interface{} {
			return p.subscriptions
		}).([]*rpcFatalSubscription)

		for _, sub := range subscriptions {
			if sub != nil && sub.onFatal != nil {
				sub.onFatal(err)
			}
		}
	}
}

// LogSubscription ...
type rpcFatalSubscription struct {
	id       int64
	reporter *rpcErrorReporter
	onFatal  func(err Error)
}

// Close ...
func (p *rpcFatalSubscription) Close() bool {
	if p == nil {
		return false
	} else if reporter := p.reporter; reporter == nil {
		return false
	} else {
		return reporter.removeSubscription(p.id)
	}
}

// Error ...
type Error interface {
	GetKind() ErrKind
	GetMessage() string
	GetDebug() string
	AddDebug(debug string) Error
	Error() string
}

func newError(kind ErrKind, message string) Error {
	return &rpcError{
		kind:    kind,
		message: message,
		debug:   "",
	}
}

// NewError ...
func NewError(message string) Error {
	return newError(ErrKindFromNone, message)
}

// NewProtocolError ...
func NewProtocolError(message string) Error {
	return newError(ErrKindFromProtocol, message)
}

// NewTransportError ...
func NewTransportError(message string) Error {
	return newError(ErrKindFromTransport, message)
}

// NewTimeoutError ...
func NewTimeoutError(message string) Error {
	return newError(ErrKindFromTimeout, message)
}

// NewAccessError ...
func NewAccessError(message string) Error {
	return newError(ErrKindFromAccess, message)
}

// NewKernelError ...
func NewKernelError(message string) Error {
	return newError(ErrKindFromKernel, message)
}

// ConvertToError convert interface{} to Error if type matches
func ConvertToError(v interface{}) Error {
	if ret, ok := v.(Error); ok {
		return ret
	}

	return nil
}

type rpcError struct {
	message string
	debug   string
	kind    ErrKind
}

func (p *rpcError) GetKind() ErrKind {
	return p.kind
}

func (p *rpcError) GetMessage() string {
	return p.message
}

func (p *rpcError) GetDebug() string {
	return p.debug
}

func (p *rpcError) AddDebug(debug string) Error {
	if p.debug != "" {
		p.debug += "\n"
	}
	p.debug += debug
	return p
}

func (p *rpcError) Error() string {
	sb := NewStringBuilder()
	defer sb.Release()

	if len(p.message) > 0 {
		sb.AppendString(p.message)
		sb.AppendByte('\n')
	}

	if len(p.debug) > 0 {
		sb.AppendString("Debug:\n")
		sb.AppendString(AddPrefixPerLine(p.debug, "\t"))
		sb.AppendByte('\n')
	}

	return sb.String()
}

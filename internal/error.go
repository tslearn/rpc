package internal

type ErrKind uint64

const (
	ErrKindFromNone      ErrKind = 0
	ErrKindFromProtocol  ErrKind = 4
	ErrKindFromTransport ErrKind = 5
	ErrKindFromTimeout   ErrKind = 7
	ErrKindFromAccess    ErrKind = 6
	ErrKindFromKernel    ErrKind = 3
)

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

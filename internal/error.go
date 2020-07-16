package internal

// Error ...
type Error interface {
	GetMessage() string
	GetDebug() string
	AddDebug(debug string)
	GetExtra() string
	SetExtra(extra string)
	Error() string
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
	extra   string
}

// NewError create new error
func NewError(message string) Error {
	return &rpcError{
		message: message,
		debug:   "",
		extra:   "",
	}
}

// NewErrorByDebug create new error
func NewErrorByDebug(message string, debug string) Error {
	return &rpcError{
		message: message,
		debug:   debug,
		extra:   "",
	}
}

func (p *rpcError) GetMessage() string {
	return p.message
}

func (p *rpcError) GetDebug() string {
	return p.debug
}

func (p *rpcError) AddDebug(debug string) {
	if p.debug != "" {
		p.debug += "\n"
	}
	p.debug += debug
}

func (p *rpcError) GetExtra() string {
	return p.extra
}

func (p *rpcError) SetExtra(extra string) {
	p.extra = extra
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

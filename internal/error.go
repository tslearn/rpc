package internal

// Error ...
type Error interface {
	GetMessage() string
	GetDebug() string
	GetExtra() string
	AddDebug(debug string) Error
	SetExtra(extra string)
	Error() string
}

// NewError create new error
func NewError(message string) Error {
	return &rpcError{
		message: message,
		debug:   "",
		extra:   "",
	}
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

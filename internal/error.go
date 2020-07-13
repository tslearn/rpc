package internal

// RPCError ...
type RPCError interface {
	GetMessage() string
	GetDebug() string
	AddDebug(debug string)
	GetExtra() string
	SetExtra(extra string)
	Error() string
}

// ConvertToRPCError convert interface{} to RPCError if type matches
func ConvertToRPCError(v interface{}) RPCError {
	if ret, ok := v.(RPCError); ok {
		return ret
	}

	return nil
}

type rpcError struct {
	message string
	debug   string
	extra   string
}

// NewRPCError create new error
func NewRPCError(message string) RPCError {
	return &rpcError{
		message: message,
		debug:   "",
		extra:   "",
	}
}

// NewRPCErrorByDebug create new error
func NewRPCErrorByDebug(message string, debug string) RPCError {
	return &rpcError{
		message: message,
		debug:   debug,
		extra:   "",
	}
}

// NewRPCErrorByError ...
func NewRPCErrorByError(err error) RPCError {
	if err == nil {
		return nil
	}

	return NewRPCError(err.Error())
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

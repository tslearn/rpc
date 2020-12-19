package adapter

// ICloseable ...
type ICloseable interface {
	Close() error
}

// EmptyCloseable ...
type EmptyCloseable struct{}

// NewEmptyCloseable ...
func NewEmptyCloseable() *EmptyCloseable {
	return &EmptyCloseable{}
}

// Close ...
func (p *EmptyCloseable) Close() error {
	return nil
}

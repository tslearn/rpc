package adapter

type LoopManager struct {
	channels    []*LoopChannel
	receiver    XReceiver
	currChannel *LoopChannel
	currRemains uint64
}

func NewLoopManager(channels int, receiver XReceiver) (*LoopManager, error) {
	if channels < 1 {
		channels = 1
	}

	ret := &LoopManager{
		receiver:    receiver,
		channels:    make([]*LoopChannel, channels),
		currChannel: nil,
		currRemains: 0,
	}

	for i := 0; i < channels; i++ {
		if channel, err := NewLoopChannel(ret); err != nil {
			return nil, err
		} else {
			ret.channels[i] = channel
		}

	}

	return ret, nil
}

func (p *LoopManager) Open() {
	for i := 0; i < len(p.channels); i++ {
		go func(idx int) {
			p.channels[idx].Open()
		}(i)
	}
}

func (p *LoopManager) Close() error {
	return nil
}

func (p *LoopManager) AllocChannel() *LoopChannel {
	if p.currRemains <= 0 {
		maxConn := int64(-1)
		for i := 0; i < len(p.channels); i++ {
			if connCount := p.channels[i].GetActiveConnCount(); connCount > maxConn {
				p.currChannel = p.channels[i]
				maxConn = connCount
			}
		}
		p.currRemains = 256
	}

	p.currRemains--
	return p.currChannel
}

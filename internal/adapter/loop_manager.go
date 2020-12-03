package adapter

type LoopManager struct {
	channels    []*LoopChannel
	receiver    XReceiver
	currChannel *LoopChannel
	currRemains uint64
}

func NewLoopManager(channels int, receiver XReceiver) *LoopManager {
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
		channel := NewLoopChannel(ret)

		if channel != nil {
			ret.channels[i] = NewLoopChannel(ret)
		} else {
			// clean up and return nil
			for j := 0; j < i; j++ {
				ret.channels[j].Close()
				ret.channels[j] = nil
			}
			return nil
		}
	}

	return ret
}

func (p *LoopManager) Close() {
	waitCH := make(chan bool)
	channelSize := len(p.channels)
	for i := 0; i < channelSize; i++ {
		go func(idx int) {
			p.channels[idx].Close()
			waitCH <- true
		}(i)
	}

	for i := 0; i < channelSize; i++ {
		<-waitCH
	}
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

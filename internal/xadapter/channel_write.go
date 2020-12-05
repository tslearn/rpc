package xadapter

type WriteChannel struct {
	channel *Channel
	poller  *Poller
	connMap map[int]*PollConn
	connCH  chan *PollConn
}

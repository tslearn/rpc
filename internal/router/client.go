package router

import (
	"crypto/tls"
	"encoding/binary"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
	"time"
)

type Client struct {
	addr       string
	tlsConfig  *tls.Config
	slot       *Slot
	orcManager *base.ORCManager
	streamHub  rpc.IStreamHub
	id         *base.GlobalID
}

func NewClient(addr string, tlsConfig *tls.Config, streamHub rpc.IStreamHub) *Client {
	return &Client{
		addr:       addr,
		tlsConfig:  tlsConfig,
		slot:       NewSlotManager(streamHub),
		orcManager: base.NewORCManager(),
		streamHub:  streamHub,
	}
}

func (p *Client) Open() bool {
	return p.orcManager.Open(func() bool {
		p.id = base.NewGlobalID()
		return p.id != nil
	})
}

func (p *Client) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) bool {
		for isRunning() {
			startMS := base.TimeNow().UnixNano()
			frees := p.slot.GetFreeChannels()
			for _, freeChannelID := range frees {
				var conn net.Conn
				var e error

				if p.tlsConfig == nil {
					conn, e = net.Dial("tcp", p.addr)
				} else {
					conn, e = tls.Dial("tcp", p.addr, p.tlsConfig)
				}

				if e != nil {
					p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
						base.ErrRouterClientConnect.AddDebug(e.Error()),
					))
					break
				}

				// init comm
				buffer := make([]byte, 12)
				binary.LittleEndian.PutUint16(buffer, rpc.StreamKindConnectRequest)
				binary.LittleEndian.PutUint16(buffer[2:], freeChannelID)
				binary.LittleEndian.PutUint64(buffer[4:], p.id.GetID())
				if _, e := conn.Write(buffer); e != nil {
					p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
						base.ErrRouterClientConnect.AddDebug(e.Error()),
					))
					break
				}

				// run conn
				p.slot.RunAt(freeChannelID, conn)
			}

			base.WaitAtLeastDurationWhenRunning(startMS, isRunning, time.Second)
		}

		return true
	})
}

func (p *Client) Close() bool {
	return p.orcManager.Close(func() bool {
		p.slot.Close()
		return true
	}, func() {
		p.id = nil
	})
}

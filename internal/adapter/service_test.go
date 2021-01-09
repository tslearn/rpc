package adapter

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

func getFieldPointer(ptr interface{}, fileName string) unsafe.Pointer {
	val := reflect.Indirect(reflect.ValueOf(ptr))
	return unsafe.Pointer(val.FieldByName(fileName).UnsafeAddr())
}

func TestSyncTCPServerService_Open(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("addr error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "error", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsFalse()
		assert(receiver.GetError()).IsNotNil()
		v.Close()
	})

	t.Run("test tcp", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		assert(v.ln).IsNotNil()

		v.Close()
	})

	t.Run("test tls", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		tlsConfig, e := base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server.crt"),
			path.Join(curDir, "_cert_", "server.key"),
		)
		if e != nil {
			panic(e)
		}
		v := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "0.0.0.0:65432", tlsConfig, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		assert(v.ln).IsNotNil()
		v.Close()
	})
}

func TestSyncTCPServerService_Run(t *testing.T) {
	//_, curFile, _, _ := runtime.Caller(0)
	//curDir := path.Dir(curFile)

	t.Run("test tcp", func(t *testing.T) {
		assert := base.NewAssert(t)

		fnStartClient := func(network string, addr string) {
			receiver := newTestSingleReceiver()
			client := &syncClientService{
				adapter: NewClientAdapter(
					network, addr, nil, 1200, 1200, receiver,
				),
				conn:       nil,
				orcManager: base.NewORCManager(),
			}
			client.Open()
			go func() {
				assert(client.Run()).IsTrue()
			}()
			for client.conn == nil {
				time.Sleep(50 * time.Millisecond)
			}
			client.Close()
		}

		receiver := newTestSingleReceiver()
		server := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(server.Open()).IsTrue()
		go func() {
			assert(server.Run()).IsTrue()
		}()

		fnStartClient("tcp", "0.0.0.0:65432")

		server.Close()
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestSyncTCPServerService_Close(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		tcpLn := v.ln.(*net.TCPListener)
		fdPtr := (*unsafe.Pointer)(getFieldPointer(tcpLn, "fd"))
		originFD := *fdPtr
		*fdPtr = nil
		v.Close()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceClose.AddDebug("invalid argument"),
		)
		*fdPtr = originFD
		_ = tcpLn.Close()
	})

	t.Run("test tcp", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		v.Close()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("test tls", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		tlsConfig, e := base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server.crt"),
			path.Join(curDir, "_cert_", "server.key"),
		)
		if e != nil {
			panic(e)
		}
		v := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "0.0.0.0:65432", tlsConfig, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		v.Close()
		assert(receiver.GetError()).IsNil()
	})
}

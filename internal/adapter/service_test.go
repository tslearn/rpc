package adapter

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"net"
	"path"
	"reflect"
	"runtime"
	"testing"
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
			adapter: NewClientAdapter(
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
			adapter: NewClientAdapter(
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
			adapter: NewClientAdapter(
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

}

func TestSyncTCPServerService_Close(t *testing.T) {
	t.Run("close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewClientAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		*(*unsafe.Pointer)(getFieldPointer(
			v.ln.(*net.TCPListener),
			"fd",
		)) = nil
		assert(v.Close()).IsTrue()

		fmt.Println(receiver.GetError())

	})

	t.Run("test tcp", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := newTestSingleReceiver()
		v := &syncTCPServerService{
			adapter: NewClientAdapter(
				"tcp", "0.0.0.0:65432", nil, 1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		assert(v.Open()).IsTrue()
		assert(v.Close()).IsTrue()

	})
}

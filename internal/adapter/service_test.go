package adapter

import (
	"crypto/tls"
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

func getFieldPointer(ptr interface{}, fieldName string) unsafe.Pointer {
	val := reflect.Indirect(reflect.ValueOf(ptr))
	return unsafe.Pointer(val.FieldByName(fieldName).UnsafeAddr())
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
	fnTest := func(isTLS bool, fakeError bool) (*testSingleReceiver, bool) {
		_, curFile, _, _ := runtime.Caller(0)
		curDir := path.Dir(curFile)

		tlsClientConfig := (*tls.Config)(nil)
		tlsServerConfig := (*tls.Config)(nil)

		if isTLS {
			tlsServerConfig, _ = base.GetTLSServerConfig(
				path.Join(curDir, "_cert_", "server.crt"),
				path.Join(curDir, "_cert_", "server.key"),
			)
			tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
				path.Join(curDir, "_cert_", "ca.crt"),
			})
		}

		waitCH := make(chan bool)
		runOK := false
		receiver := newTestSingleReceiver()
		server := &syncTCPServerService{
			adapter: NewServerAdapter(
				"tcp", "127.0.0.1:65432", tlsServerConfig,
				1200, 1200, receiver,
			),
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
		server.Open()
		go func() {
			fdPtr := (*unsafe.Pointer)(nil)
			if isTLS {
				lnPtr := (*net.Listener)(getFieldPointer(server.ln, "Listener"))
				fdPtr = (*unsafe.Pointer)(getFieldPointer(*lnPtr, "fd"))
			} else {
				fdPtr = (*unsafe.Pointer)(getFieldPointer(server.ln, "fd"))
			}

			originFD := *fdPtr
			if fakeError {
				*fdPtr = nil
			}

			waitCH <- true

			go func() {
				if fakeError {
					for receiver.GetOnErrorCount() == 0 {
						time.Sleep(50 * time.Millisecond)
					}
				} else {
					for receiver.GetOnOpenCount() == 0 {
						time.Sleep(50 * time.Millisecond)
					}
				}

				*fdPtr = originFD
				waitCH <- true
			}()

			runOK = server.Run()
		}()

		// Start client
		<-waitCH
		go func() {
			client := &syncClientService{
				adapter: NewClientAdapter(
					"tcp", "127.0.0.1:65432", tlsClientConfig,
					1200, 1200, newTestSingleReceiver(),
				),
				conn:       nil,
				orcManager: base.NewORCManager(),
			}
			client.Open()
			go func() {
				client.Run()
			}()

			// wait server signal ...
			if fakeError {
				for receiver.GetOnErrorCount() == 0 {
					time.Sleep(50 * time.Millisecond)
				}
			} else {
				for receiver.GetOnOpenCount() == 0 {
					time.Sleep(50 * time.Millisecond)
				}
			}
			// close
			client.Close()
		}()

		<-waitCH
		server.Close()

		for receiver.GetOnOpenCount() != receiver.GetOnCloseCount() {
			time.Sleep(50 * time.Millisecond)
		}

		return receiver, runOK
	}

	t.Run("test tcp run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := fnTest(false, true)

		assert(runOK).Equal(true)
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceAccept.AddDebug(
				"invalid argument",
			),
		)
		assert(receiver.GetOnOpenCount()).Equal(0)
		assert(receiver.GetOnCloseCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("test tcp", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := fnTest(false, false)
		assert(runOK).Equal(true)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("test tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := fnTest(true, true)

		assert(runOK).Equal(true)
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceAccept.AddDebug(
				"invalid argument",
			),
		)
		assert(receiver.GetOnOpenCount()).Equal(0)
		assert(receiver.GetOnCloseCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("test tls", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := fnTest(true, false)
		assert(runOK).Equal(true)

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

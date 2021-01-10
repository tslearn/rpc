package adapter

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"net"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func getFieldPointer(ptr interface{}, fieldName string) unsafe.Pointer {
	val := reflect.Indirect(reflect.ValueOf(ptr))
	return unsafe.Pointer(val.FieldByName(fieldName).UnsafeAddr())
}

func SyncServerTestOpen(
	network string,
	isTLS bool,
	fakeError bool,
) (*testSingleReceiver, bool) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	addr := ""
	if fakeError {
		addr = "error-addr"
	} else {
		addr = "127.0.0.1:65432"
	}

	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server.crt"),
			path.Join(curDir, "_cert_", "server.key"),
		)
	}

	receiver := newTestSingleReceiver()

	v := NewSyncServerService(NewServerAdapter(
		network, addr, tlsServerConfig, 1200, 1200, receiver,
	))
	openOK := v.Open()
	v.Close()
	return receiver, openOK
}

func SyncServerTestRun(
	network string,
	isTLS bool,
	fakeError bool,
) (*testSingleReceiver, bool) {
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
	server := NewSyncServerService(NewServerAdapter(
		network, "127.0.0.1:65432", tlsServerConfig,
		1200, 1200, receiver,
	))
	server.Open()
	ln := (net.Listener)(nil)
	if strings.HasPrefix(network, "tcp") {
		ln = server.(*syncTCPServerService).ln
	} else {
		ln = server.(*syncWSServerService).ln
	}
	go func() {
		fdPtr := (*unsafe.Pointer)(nil)
		if isTLS {
			lnPtr := (*net.Listener)(getFieldPointer(ln, "Listener"))
			fdPtr = (*unsafe.Pointer)(getFieldPointer(*lnPtr, "fd"))
		} else {
			fdPtr = (*unsafe.Pointer)(getFieldPointer(ln, "fd"))
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
	if !strings.HasPrefix(network, "tcp") && fakeError {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsClientConfig,
			},
		}
		if isTLS {
			_, _ = client.Get("https://127.0.0.1:65432")
		} else {
			_, _ = client.Get("http://127.0.0.1:65432")
		}
	} else {
		go func() {
			client := NewSyncClientService(NewClientAdapter(
				network, "127.0.0.1:65432", tlsClientConfig,
				1200, 1200, newTestSingleReceiver(),
			))
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
	}

	<-waitCH
	server.Close()

	for receiver.GetOnOpenCount() != receiver.GetOnCloseCount() {
		time.Sleep(50 * time.Millisecond)
	}

	return receiver, runOK
}

func SyncServerTestClose(
	network string,
	isTLS bool,
	fakeError bool,
) (*testSingleReceiver, bool) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server.crt"),
			path.Join(curDir, "_cert_", "server.key"),
		)
	}

	receiver := newTestSingleReceiver()
	v := NewSyncServerService(NewServerAdapter(
		network, "127.0.0.1:65432", tlsServerConfig, 1200, 1200, receiver,
	))
	v.Open()

	go func() {
		v.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	fdPtr := (*unsafe.Pointer)(nil)
	ln := (net.Listener)(nil)
	if strings.HasPrefix(network, "tcp") {
		ln = v.(*syncTCPServerService).ln
	} else {
		ln = v.(*syncWSServerService).ln
	}

	if isTLS {
		lnPtr := (*net.Listener)(getFieldPointer(ln, "Listener"))
		fdPtr = (*unsafe.Pointer)(getFieldPointer(*lnPtr, "fd"))
	} else {
		fdPtr = (*unsafe.Pointer)(getFieldPointer(ln, "fd"))
	}
	originFD := *fdPtr

	if fakeError {
		*fdPtr = nil
	}

	waitCH := make(chan bool, 1)
	closeOK := false
	go func() {
		closeOK = v.Close()
		waitCH <- true
	}()

	if fakeError {
		for receiver.GetOnErrorCount() == 0 {
			time.Sleep(50 * time.Millisecond)
		}
		*fdPtr = originFD
		_ = ln.Close()
	}

	<-waitCH
	return receiver, closeOK
}

func TestSyncTCPServerService_Open(t *testing.T) {
	t.Run("tcp open error", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("tcp", false, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("tcp open ok", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("tcp", false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("tls open error", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("tcp", true, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("tls open ok", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("tcp", false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncTCPServerService_Run(t *testing.T) {
	t.Run("tcp run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("tcp", false, true)

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

	t.Run("tcp run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("tcp", false, false)
		assert(runOK).Equal(true)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("tcp", true, true)

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

	t.Run("tls run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("tcp", true, false)
		assert(runOK).Equal(true)

		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})
}

func TestSyncTCPServerService_Close(t *testing.T) {
	t.Run("tcp close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("tcp", false, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("tcp close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("tcp", false, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("tls close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("tcp", true, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("tls close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("tcp", true, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncWSServerService_Open(t *testing.T) {
	t.Run("ws open error", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("ws", false, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("ws open ok", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("ws", false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("wss open error", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("wss", true, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("wss open ok", func(t *testing.T) {
		receiver, ok := SyncServerTestOpen("wss", true, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncWSServerService_Run(t *testing.T) {
	t.Run("ws run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("ws", false, true)

		assert(runOK).Equal(true)
		assert(receiver.GetError()).Equal(
			errors.ErrSyncWSServerServiceServe.AddDebug(
				"invalid argument",
			),
		)
		assert(receiver.GetOnOpenCount()).Equal(0)
		assert(receiver.GetOnCloseCount()).Equal(0)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("ws run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("ws", false, false)
		assert(runOK).Equal(true)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("tcp", true, true)

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

	t.Run("tls run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := SyncServerTestRun("tcp", true, false)
		assert(runOK).Equal(true)

		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

}

func TestSyncWSServerService_Close(t *testing.T) {
	t.Run("ws close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("ws", false, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncWSServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("ws close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("ws", false, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("wss close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("wss", true, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncWSServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("wss close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := SyncServerTestClose("wss", true, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

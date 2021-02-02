package adapter

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
)

func getFieldPointer(ptr interface{}, fieldName string) unsafe.Pointer {
	val := reflect.Indirect(reflect.ValueOf(ptr))
	return unsafe.Pointer(val.FieldByName(fieldName).UnsafeAddr())
}

func syncServerTestOpen(
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
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
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

func syncServerTestRun(
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
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
		)
		tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
			path.Join(curDir, "_cert_", "ca", "ca.pem"),
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
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				for receiver.GetOnOpenCount() == 0 {
					time.Sleep(10 * time.Millisecond)
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
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				for receiver.GetOnOpenCount() == 0 {
					time.Sleep(10 * time.Millisecond)
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

func syncServerTestClose(
	network string,
	isTLS bool,
	fakeError bool,
) (*testSingleReceiver, bool) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
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

	var fdPtr *unsafe.Pointer
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

func syncClientTest(
	isTLS bool,
	serverNetwork string,
	clientNetwork string,
	addr string,
) (*testSingleReceiver, *syncClientService) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	tlsClientConfig := (*tls.Config)(nil)
	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
		)
		tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
			path.Join(curDir, "_cert_", "ca", "ca.pem"),
		})
	}

	server := NewSyncServerService(NewServerAdapter(
		serverNetwork, addr, tlsServerConfig,
		1200, 1200, newTestSingleReceiver(),
	))
	server.Open()
	defer server.Close()
	go func() {
		server.Run()
	}()

	clientReceiver := newTestSingleReceiver()
	client := &syncClientService{
		adapter: NewClientAdapter(
			clientNetwork, addr, tlsClientConfig,
			1200, 1200, clientReceiver,
		),
		orcManager: base.NewORCManager(),
	}

	waitCH := make(chan bool)
	client.Open()
	go func() {
		for clientReceiver.GetOnOpenCount() == 0 &&
			clientReceiver.GetOnErrorCount() == 0 {
			time.Sleep(50 * time.Millisecond)
		}
		client.Close()
		waitCH <- true
	}()
	client.Run()
	<-waitCH
	return clientReceiver, client
}

func TestNewSyncClientService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for _, network := range []string{
			"tcp", "tcp4", "tcp6", "ws", "wss",
		} {
			adapter := NewClientAdapter(
				network, "localhost", nil, 1200, 1200, newTestSingleReceiver(),
			)

			service := NewSyncClientService(adapter).(*syncClientService)

			assert(service.adapter).Equal(adapter)
			assert(service.conn).IsNil()
			assert(service.orcManager).IsNotNil()
		}
	})

	t.Run("protocol error", func(t *testing.T) {
		assert := base.NewAssert(t)

		receiver := newTestSingleReceiver()
		adapter := NewClientAdapter(
			"err", "localhost", nil, 1200, 1200, receiver,
		)

		service := NewSyncClientService(adapter)
		assert(service).IsNil()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrUnsupportedProtocol.AddDebug(
				"unsupported protocol err",
			))
	})
}

func TestNewSyncServerService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for _, network := range []string{"tcp", "tcp4", "tcp6"} {
			adapter := NewClientAdapter(
				network, "localhost", nil, 1200, 1200, newTestSingleReceiver(),
			)

			service := NewSyncServerService(adapter).(*syncTCPServerService)
			assert(service.adapter).Equal(adapter)
			assert(service.ln).IsNil()
			assert(service.orcManager).IsNotNil()
		}

		for _, network := range []string{"ws", "wss"} {
			adapter := NewClientAdapter(
				network, "localhost", nil, 1200, 1200, newTestSingleReceiver(),
			)

			service := NewSyncServerService(adapter).(*syncWSServerService)
			assert(service.adapter).Equal(adapter)
			assert(service.ln).IsNil()
			assert(service.server).IsNil()
			assert(service.orcManager).IsNotNil()
		}
	})

	t.Run("protocol error", func(t *testing.T) {
		assert := base.NewAssert(t)

		receiver := newTestSingleReceiver()
		adapter := NewClientAdapter(
			"err", "localhost", nil, 1200, 1200, receiver,
		)

		service := NewSyncServerService(adapter)
		assert(service).IsNil()
		assert(receiver.GetOnErrorCount()).Equal(1)
		assert(receiver.GetError()).
			Equal(errors.ErrUnsupportedProtocol.AddDebug(
				"unsupported protocol err",
			))
	})
}

func TestSyncTCPServerService_Open(t *testing.T) {
	t.Run("tcp open error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", false, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("tcp open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("tls open error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", true, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("tls open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncTCPServerService_Run(t *testing.T) {
	t.Run("tcp run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := syncServerTestRun("tcp", false, true)

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
		receiver, runOK := syncServerTestRun("tcp", false, false)
		assert(runOK).Equal(true)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := syncServerTestRun("tcp", true, true)

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
		receiver, runOK := syncServerTestRun("tcp", true, false)
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
		receiver, ok := syncServerTestClose("tcp", false, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("tcp close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", false, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("tls close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", true, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncTCPServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("tls close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", true, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncWSServerService_Open(t *testing.T) {
	t.Run("ws open error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("ws", false, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("ws open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("ws", false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("wss open error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("wss", true, true)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("wss open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("wss", true, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncWSServerService_Run(t *testing.T) {
	t.Run("ws run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := syncServerTestRun("ws", false, true)

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
		receiver, runOK := syncServerTestRun("ws", false, false)
		assert(runOK).Equal(true)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equal(1)
		assert(receiver.GetOnCloseCount()).Equal(1)
		assert(receiver.GetOnStreamCount()).Equal(0)
		assert(receiver.GetOnErrorCount()).Equal(0)
	})

	t.Run("tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, runOK := syncServerTestRun("tcp", true, true)

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
		receiver, runOK := syncServerTestRun("tcp", true, false)
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
		receiver, ok := syncServerTestClose("ws", false, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncWSServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("ws close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("ws", false, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("wss close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("wss", true, true)
		assert(ok).IsTrue()
		assert(receiver.GetError()).Equal(
			errors.ErrSyncWSServerServiceClose.AddDebug("invalid argument"),
		)
	})

	t.Run("wss close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("wss", true, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncClientService(t *testing.T) {
	type testItem struct {
		network string
		isTLS   bool
	}
	addr := "localhost:65432"

	t.Run("tcp4 ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for _, it := range []testItem{
			{network: "tcp", isTLS: false},
			{network: "tcp", isTLS: true},
			{network: "tcp4", isTLS: false},
			{network: "tcp4", isTLS: true},
			{network: "tcp6", isTLS: false},
			{network: "tcp6", isTLS: true},
			{network: "ws", isTLS: false},
			{network: "wss", isTLS: true},
		} {
			receiver, client := syncClientTest(
				it.isTLS, it.network, it.network, addr,
			)
			assert(receiver.GetOnOpenCount()).Equal(1)
			assert(receiver.GetOnCloseCount()).Equal(1)
			assert(receiver.GetOnErrorCount()).Equal(0)
			assert(receiver.GetOnStreamCount()).Equal(0)
			assert(client.conn).IsNil()
		}
	})

	t.Run("tcp4 addr error", func(t *testing.T) {
		assert := base.NewAssert(t)

		for _, it := range []testItem{
			{network: "tcp", isTLS: false},
			{network: "tcp", isTLS: true},
			{network: "tcp4", isTLS: false},
			{network: "tcp4", isTLS: true},
			{network: "tcp6", isTLS: false},
			{network: "tcp6", isTLS: true},
			{network: "ws", isTLS: false},
			{network: "wss", isTLS: true},
		} {
			receiver, client := syncClientTest(
				it.isTLS, it.network, it.network, "addr-error",
			)
			if it.network == "ws" || it.network == "wss" {
				assert(receiver.GetError()).
					Equal(errors.ErrSyncClientServiceDial.AddDebug(
						"dial tcp: lookup addr-error: no such host",
					))
			} else {
				assert(receiver.GetError()).
					Equal(errors.ErrSyncClientServiceDial.AddDebug(fmt.Sprintf(
						"dial %s: address addr-error: missing port in address",
						it.network,
					)))
			}

			assert(receiver.GetOnOpenCount()).Equal(0)
			assert(receiver.GetOnCloseCount()).Equal(0)
			assert(receiver.GetOnErrorCount()).Equal(1)
			assert(receiver.GetOnStreamCount()).Equal(0)
			assert(client.conn).IsNil()
		}
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		assert := base.NewAssert(t)

		for _, it := range []testItem{
			{network: "err", isTLS: false},
			{network: "err", isTLS: true},
		} {
			receiver, client := syncClientTest(
				it.isTLS, "tcp", it.network, addr,
			)

			assert(receiver.GetError()).
				Equal(errors.ErrUnsupportedProtocol.AddDebug(
					"unsupported protocol err",
				))
			assert(receiver.GetOnOpenCount()).Equal(0)
			assert(receiver.GetOnCloseCount()).Equal(0)
			assert(receiver.GetOnErrorCount()).Equal(1)
			assert(receiver.GetOnStreamCount()).Equal(0)
			assert(client.conn).IsNil()
		}
	})
}

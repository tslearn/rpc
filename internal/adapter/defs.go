package adapter

import (
	"errors"
	"github.com/rpccloud/rpc/internal/base"
	"net"
	"reflect"
)

type IAdapter interface {
	Open(receiver XReceiver) *base.Error
	Close() *base.Error
}

func GetFD(conn net.Conn) (int, error) {
	switch tpConn := conn.(type) {
	case *net.TCPConn:
		tcpConn := reflect.Indirect(reflect.ValueOf(tpConn)).FieldByName("conn")
		fdVal := tcpConn.FieldByName("fd")
		pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
		return int(pfdVal.FieldByName("Sysfd").Int()), nil
	default:
		return 0, errors.New("unknown conn type")
	}
}

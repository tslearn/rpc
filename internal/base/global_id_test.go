package base

import (
	"fmt"
	"testing"
)

func TestNewGlobalID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		fmt.Printf("%x\n", NewGlobalID().GetID())
	})
}

//
//func TestFindMacAddrByIP(t *testing.T) {
//    t.Run("ip is nil", func(t *testing.T) {
//        assert := NewAssert(t)
//        assert(FindMacAddrByIP("")).Equal("")
//    })
//
//    t.Run("ip is not local", func(t *testing.T) {
//        assert := NewAssert(t)
//        assert(FindMacAddrByIP("8.8.8.8")).Equal("")
//    })
//
//    t.Run("fake net.Interface error", func(t *testing.T) {
//        assert := NewAssert(t)
//        saveFNNetInterfaces := fnNetInterfaces
//        fnNetInterfaces = func() ([]net.Interface, error) {
//            return nil, errors.New("error")
//        }
//        defer func() {
//            fnNetInterfaces = saveFNNetInterfaces
//        }()
//        assert(FindMacAddrByIP("127.0.0.1")).Equal("")
//    })
//
//    t.Run("test ok", func(t *testing.T) {
//        assert := NewAssert(t)
//        testCount := 0
//        addrList, e := net.InterfaceAddrs()
//        if e != nil {
//            assert().Fail(e.Error())
//        }
//        for _, addr := range addrList {
//            if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.IsGlobalUnicast() {
//                assert(len(FindMacAddrByIP(ipNet.IP.String()))).Equal(17)
//                testCount++
//            }
//        }
//        assert(testCount > 0).IsTrue()
//    })
//}

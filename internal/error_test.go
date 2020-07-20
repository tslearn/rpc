package internal

//
//import (
//	"testing"
//)
//
//func TestNewError(t *testing.T) {
//	assert := NewAssert(t)
//
//	assert(NewError("hello").GetMessage()).Equals("hello")
//	assert(NewError("hello").GetDebug()).Equals("")
//}
//
//func TestNewErrorByDebug(t *testing.T) {
//	assert := NewAssert(t)
//
//	var testCollection = [][2]interface{}{
//		{
//			NewError("").AddDebug(""),
//			"",
//		}, {
//			NewError("message").AddDebug(""),
//			"message\n",
//		}, {
//			NewError("").AddDebug("debug"),
//			"Debug:\n\tdebug\n",
//		}, {
//			NewError("message").AddDebug("debug"),
//			"message\nDebug:\n\tdebug\n",
//		},
//	}
//	for _, item := range testCollection {
//		assert(item[0].(*rpcError).Error()).Equals(item[1])
//	}
//}
//
//func TestConvertToError(t *testing.T) {
//	assert := NewAssert(t)
//
//	assert(ConvertToError(0)).IsNil()
//	assert(ConvertToError(make(chan bool))).IsNil()
//	assert(ConvertToError(nil)).IsNil()
//	assert(ConvertToError(NewError("test"))).IsNotNil()
//}
//
//func TestRpcError_GetMessage(t *testing.T) {
//	assert := NewAssert(t)
//	assert(NewError("message").AddDebug("debug").GetMessage()).Equals("message")
//}
//
//func TestRpcError_GetDebug(t *testing.T) {
//	assert := NewAssert(t)
//	assert(NewError("message").AddDebug("debug").GetDebug()).Equals("debug")
//}
//
//func TestRpcError_AddDebug(t *testing.T) {
//	assert := NewAssert(t)
//
//	err := NewError("message")
//	err.AddDebug("m1")
//	assert(err.GetDebug()).Equals("m1")
//	err.AddDebug("m2")
//	assert(err.GetDebug()).Equals("m1\nm2")
//}
//
//func TestRpcError_GetExtra(t *testing.T) {
//	assert := NewAssert(t)
//
//	err := NewError("message")
//	err.SetExtra("extra")
//	assert(err.GetExtra()).Equals("extra")
//}
//
//func TestRpcError_SetExtra(t *testing.T) {
//	assert := NewAssert(t)
//
//	err := NewError("message")
//	err.SetExtra("extra")
//	assert(err.(*rpcError).extra).Equals("extra")
//}

package util

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger(nil)
	assert(logger.level).Equals(LogMaskAll)
	assert(len(logger.writers)).Equals(1)
}

func TestLogger_SetLevel(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger(nil)

	assert(logger.SetLevel(LogMaskNone - 1)).IsFalse()
	assert(logger.level).Equals(LogMaskAll)

	assert(logger.SetLevel(LogMaskAll + 1)).IsFalse()
	assert(logger.level).Equals(LogMaskAll)

	assert(logger.SetLevel(LogMaskNone)).IsTrue()
	assert(logger.level).Equals(LogMaskNone)

	assert(logger.SetLevel(LogMaskAll)).IsTrue()
	assert(logger.level).Equals(LogMaskAll)

	// test all level and logs
	fnTestLogLevel := func(level int32) int32 {
		ret := int32(0)

		logger := NewLogger([]LogWriter{NewCallbackLogWriter(
			func(_ string, tag string, msg string, _ string) {
				if msg == "message" {
					switch tag {
					case "Debug":
						atomic.AddInt32(&ret, LogMaskDebug)
					case "Info":
						atomic.AddInt32(&ret, LogMaskInfo)
					case "Warn":
						atomic.AddInt32(&ret, LogMaskWarn)
					case "Error":
						atomic.AddInt32(&ret, LogMaskError)
					case "Fatal":
						atomic.AddInt32(&ret, LogMaskFatal)
					}
				}
			},
		)})
		logger.SetLevel(level)
		logger.Debug("message")
		logger.Info("message")
		logger.Warn("message")
		logger.Error("message")
		logger.Fatal("message")
		time.Sleep(30 * time.Millisecond)
		return atomic.LoadInt32(&ret)
	}

	assert(fnTestLogLevel(LogMaskNone - 1)).Equals(LogMaskAll)
	for i := int32(0); i < 32; i++ {
		assert(fnTestLogLevel(i)).Equals(i)
	}
	assert(fnTestLogLevel(LogMaskAll + 1)).Equals(LogMaskAll)
}

//
//func TestLogger_log(t *testing.T) {
//  assert := NewAssert(t)
//  logger := NewLogger()
//
//  var buf bytes.Buffer
//  log.SetOutput(&buf)
//  logger.Subscribe()
//
//  logger.Debug("")
//  assert(strings.Contains(buf.String(), "Debug")).IsTrue()
//  logger.Info("")
//  assert(strings.Contains(buf.String(), "Info")).IsTrue()
//  logger.Warn("")
//  assert(strings.Contains(buf.String(), "Warn")).IsTrue()
//  logger.Error("")
//  assert(strings.Contains(buf.String(), "Error")).IsTrue()
//  logger.Fatal("")
//  assert(strings.Contains(buf.String(), "Fatal")).IsTrue()
//
//  log.SetOutput(os.Stderr)
//}
//
//func TestLogger_Debug(t *testing.T) {
//  assert := NewAssert(t)
//  logger := NewLogger()
//
//  var buf1 bytes.Buffer
//  log.SetOutput(&buf1)
//  logger.Debug("message")
//  regex1 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Debug:(\\s)message\n$"
//  assert(regexp.MatchString(regex1, buf1.String())).Equals(true, nil)
//
//  var buf3 bytes.Buffer
//  log.SetOutput(&buf3)
//
//  logger.Debugf("message")
//  regex3 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Debug:(\\s)message\n$"
//  assert(regexp.MatchString(regex3, buf3.String())).Equals(true, nil)
//
//  var buf4 bytes.Buffer
//  log.SetOutput(&buf4)
//  logger.Debugf("message %d", 1)
//  regex4 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Debug:(\\s)message 1\n$"
//  assert(regexp.MatchString(regex4, buf4.String())).Equals(true, nil)
//
//  log.SetOutput(os.Stderr)
//}
//
//func TestLogger_Info(t *testing.T) {
//  assert := NewAssert(t)
//  logger := NewLogger()
//
//  var buf1 bytes.Buffer
//  log.SetOutput(&buf1)
//  logger.Info("message")
//  regex1 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Info:(\\s)message\n$"
//  assert(regexp.MatchString(regex1, buf1.String())).Equals(true, nil)
//
//  var buf3 bytes.Buffer
//  log.SetOutput(&buf3)
//  logger.Infof("message")
//  regex3 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Info:(\\s)message\n$"
//  assert(regexp.MatchString(regex3, buf3.String())).Equals(true, nil)
//
//  var buf4 bytes.Buffer
//  log.SetOutput(&buf4)
//  logger.Infof("message %d", 1)
//  regex4 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Info:(\\s)message 1\n$"
//  assert(regexp.MatchString(regex4, buf4.String())).Equals(true, nil)
//
//  log.SetOutput(os.Stderr)
//}
//
//func TestLogger_Warn(t *testing.T) {
//  assert := NewAssert(t)
//  logger := NewLogger()
//
//  var buf1 bytes.Buffer
//  log.SetOutput(&buf1)
//  logger.Warn("message")
//  regex1 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Warn:(\\s)message\n$"
//  assert(regexp.MatchString(regex1, buf1.String())).Equals(true, nil)
//
//  var buf3 bytes.Buffer
//  log.SetOutput(&buf3)
//  logger.Warnf("message")
//  regex3 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Warn:(\\s)message\n$"
//  assert(regexp.MatchString(regex3, buf3.String())).Equals(true, nil)
//
//  var buf4 bytes.Buffer
//  log.SetOutput(&buf4)
//  logger.Warnf("message %d", 1)
//  regex4 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Warn:(\\s)message 1\n$"
//  assert(regexp.MatchString(regex4, buf4.String())).Equals(true, nil)
//
//  log.SetOutput(os.Stderr)
//}
//
//func TestLogger_Error(t *testing.T) {
//  assert := NewAssert(t)
//  logger := NewLogger()
//
//  var buf1 bytes.Buffer
//  log.SetOutput(&buf1)
//  logger.Error("message")
//  regex1 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Error:(\\s)message\n$"
//  assert(regexp.MatchString(regex1, buf1.String())).Equals(true, nil)
//
//  var buf3 bytes.Buffer
//  log.SetOutput(&buf3)
//  logger.Errorf("message")
//  regex3 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Error:(\\s)message\n$"
//  assert(regexp.MatchString(regex3, buf3.String())).Equals(true, nil)
//
//  var buf4 bytes.Buffer
//  log.SetOutput(&buf4)
//  logger.Errorf("message %d", 1)
//  regex4 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Error:(\\s)message 1\n$"
//  assert(regexp.MatchString(regex4, buf4.String())).Equals(true, nil)
//
//  log.SetOutput(os.Stderr)
//}
//
//func TestLogger_Fatal(t *testing.T) {
//  assert := NewAssert(t)
//  logger := NewLogger()
//
//  var buf1 bytes.Buffer
//  log.SetOutput(&buf1)
//  logger.Fatal("message")
//  regex1 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Fatal:(\\s)message\n$"
//  assert(regexp.MatchString(regex1, buf1.String())).Equals(true, nil)
//
//  var buf3 bytes.Buffer
//  log.SetOutput(&buf3)
//  logger.Fatalf("message")
//  regex3 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Fatal:(\\s)message\n$"
//  assert(regexp.MatchString(regex3, buf3.String())).Equals(true, nil)
//
//  var buf4 bytes.Buffer
//  log.SetOutput(&buf4)
//  logger.Fatalf("message %d", 1)
//  regex4 := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}" +
//    "\\+\\d{2}:\\d{2}(\\s)Fatal:(\\s)message 1\n$"
//  assert(regexp.MatchString(regex4, buf4.String())).Equals(true, nil)
//
//  log.SetOutput(os.Stderr)
//}

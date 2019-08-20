package core

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestLogSubscription_Close(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	subscription := logger.Subscribe()
	subscription.Debug = func(msg string) {}
	subscription.Info = func(msg string) {}
	subscription.Warning = func(msg string) {}
	subscription.Error = func(msg string) {}
	subscription.Fatal = func(msg string) {}
	assert(subscription.Close()).IsNil()
	assert(subscription.id == 0).IsTrue()
	assert(subscription.logger).IsNil()
	assert(subscription.Debug).IsNil()
	assert(subscription.Info).IsNil()
	assert(subscription.Warning).IsNil()
	assert(subscription.Error).IsNil()
	assert(subscription.Fatal).IsNil()
	assert(len(logger.subscriptions)).Equals(0)
	assert(len(logger.cache)).Equals(0)

	subscription = logger.Subscribe()
	subscription.id = 0
	assert(subscription.Close().GetMessage()).Equals("id is not exist")

	subscription = logger.Subscribe()
	subscription.logger = nil
	assert(subscription.Close().GetMessage()).Equals("logger is nil")
}

func TestNewLogger(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()
	assert(logger.level).Equals(LogLevelAll)
	assert(len(logger.subscriptions)).Equals(0)
	assert(len(logger.cache)).Equals(0)
}

func TestLogger_SetLevel(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	assert(logger.SetLevel(-1).GetMessage()).Equals(
		"level(-1) not supported",
	)
	assert(logger.level).Equals(LogLevelAll)

	assert(logger.SetLevel(32).GetMessage()).Equals(
		"level(32) not supported",
	)
	assert(logger.level).Equals(LogLevelAll)

	assert(logger.SetLevel(0)).IsNil()
	assert(logger.level).Equals(0)

	assert(logger.SetLevel(31)).IsNil()
	assert(logger.level).Equals(31)
	// test all level and logs
	fnTestLogLevel := func(level int) int {
		logger := NewLogger()
		logger.SetLevel(level)
		ret := 0
		subscription := logger.Subscribe()
		subscription.Debug = func(msg string) {
			if strings.Contains(msg, "message") &&
				strings.Contains(msg, "Debug") {
				ret += logMaskDebug
			}
		}
		subscription.Info = func(msg string) {
			if strings.Contains(msg, "message") &&
				strings.Contains(msg, "Info") {
				ret += logMaskInfo
			}
		}
		subscription.Warning = func(msg string) {
			if strings.Contains(msg, "message") &&
				strings.Contains(msg, "Warning") {
				ret += logMaskWarning
			}
		}
		subscription.Error = func(msg string) {
			if strings.Contains(msg, "message") &&
				strings.Contains(msg, "Error") {
				ret += logMaskError
			}
		}
		subscription.Fatal = func(msg string) {
			if strings.Contains(msg, "message") &&
				strings.Contains(msg, "Fatal") {
				ret += logMaskFatal
			}
		}
		logger.Debug("message")
		logger.Info("message")
		logger.Warning("message")
		logger.Error("message")
		logger.Fatal("message")
		subscription.Close()
		return ret
	}

	assert(fnTestLogLevel(LogLevelOff)).Equals(LogLevelOff)
	assert(fnTestLogLevel(LogLevelFatal)).Equals(LogLevelFatal)
	assert(fnTestLogLevel(LogLevelError)).Equals(LogLevelError)
	assert(fnTestLogLevel(LogLevelWarning)).Equals(LogLevelWarning)
	assert(fnTestLogLevel(LogLevelInfo)).Equals(LogLevelInfo)
	assert(fnTestLogLevel(LogLevelAll)).Equals(LogLevelAll)

	assert(fnTestLogLevel(-1)).Equals(31)
	for i := 0; i < 32; i++ {
		assert(fnTestLogLevel(i)).Equals(i)
	}
	assert(fnTestLogLevel(32)).Equals(31)
}

func TestLogger_Subscribe(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	subscription := logger.Subscribe()
	assert(subscription.id > 0).IsTrue()
	assert(subscription.logger).Equals(logger)
	assert(subscription.Debug).IsNil()
	assert(subscription.Info).IsNil()
	assert(subscription.Warning).IsNil()
	assert(subscription.Error).IsNil()
	assert(subscription.Fatal).IsNil()

	assert(len(logger.subscriptions)).Equals(1)
	assert(len(logger.cache)).Equals(1)

	assert(subscription.Close()).IsNil()
	assert(subscription.id == 0).IsTrue()
	assert(subscription.logger).IsNil()
	assert(subscription.Debug).IsNil()
	assert(subscription.Info).IsNil()
	assert(subscription.Warning).IsNil()
	assert(subscription.Error).IsNil()
	assert(subscription.Fatal).IsNil()
	assert(len(logger.subscriptions)).Equals(0)
	assert(len(logger.cache)).Equals(0)

	assert(subscription.Close().GetMessage()).Equals("logger is nil")
}

func TestLogger_Debug(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	logger.Debug("message")
	log.SetOutput(os.Stderr)

	regex := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{0,3}" +
		"\\+\\d{2}:\\d{2}(\\s)Debug:(\\s)message\n$"
	assert(regexp.MatchString(regex, buf.String())).Equals(true, nil)
}

func TestLogger_Info(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	logger.Info("message")
	log.SetOutput(os.Stderr)
	regex := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{0,3}" +
		"\\+\\d{2}:\\d{2}(\\s)Info:(\\s)message\n$"
	assert(regexp.MatchString(regex, buf.String())).Equals(true, nil)
}

func TestLogger_Warning(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	logger.Warning("message")
	log.SetOutput(os.Stderr)

	regex := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{0,3}" +
		"\\+\\d{2}:\\d{2}(\\s)Warning:(\\s)message\n$"
	assert(regexp.MatchString(regex, buf.String())).Equals(true, nil)
}

func TestLogger_Error(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	logger.Error("message")
	log.SetOutput(os.Stderr)

	regex := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{0,3}" +
		"\\+\\d{2}:\\d{2}(\\s)Error:(\\s)message\n$"
	assert(regexp.MatchString(regex, buf.String())).Equals(true, nil)
}

func TestLogger_Fatal(t *testing.T) {
	assert := NewAssert(t)
	logger := NewLogger()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	logger.Fatal("message")
	log.SetOutput(os.Stderr)

	regex := "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{0,3}" +
		"\\+\\d{2}:\\d{2}(\\s)Fatal:(\\s)message\n$"
	assert(regexp.MatchString(regex, buf.String())).Equals(true, nil)
}

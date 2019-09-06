package core

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// LogLevelOff this level logs nothing
	LogLevelOff = 0
	// LogLevelFatal this level only logs Fatal
	LogLevelFatal = 1
	// LogLevelError this level logs Error and Fatal
	LogLevelError = 3
	// LogLevelWarn this level logs Warn, Error and Fatal
	LogLevelWarn = 7
	// LogLevelInfo this level logs Info, Warn, Error and Fatal
	LogLevelInfo = 15
	// LogLevelAll this level logs Debug, Info, Warn, Error and Fatal
	LogLevelAll = 31

	logMaskFatal = 1
	logMaskError = 2
	logMaskWarn  = 4
	logMaskInfo  = 8
	logMaskDebug = 16
)

type LogSubscription struct {
	id     int64
	logger *Logger
	Debug  func(msg string)
	Info   func(msg string)
	Warn   func(msg string)
	Error  func(msg string)
	Fatal  func(msg string)
}

func (p *LogSubscription) Close() bool {
	if p.logger == nil {
		return false
	}

	ret := p.logger.removeSubscription(p.id)
	p.id = 0
	p.logger = nil
	p.Debug = nil
	p.Info = nil
	p.Warn = nil
	p.Error = nil
	p.Fatal = nil
	return ret
}

// Logger common logger
type Logger struct {
	level         int
	subscriptions []*LogSubscription
	sync.Mutex
}

// NewLogger create new Logger
func NewLogger() *Logger {
	return &Logger{
		level:         LogLevelAll,
		subscriptions: make([]*LogSubscription, 0, 0),
	}
}

// SetLevel set the log level, such as LogLevelOff LogLevelFatal
// LogLevelError LogLevelWarn LogLevelInfo LogLevelDebug LogLevelAll
func (p *Logger) SetLevel(level int) bool {
	if level >= LogLevelOff && level <= LogLevelAll {
		p.Lock()
		p.level = level
		p.Unlock()
		return true
	} else {
		return false
	}
}

func (p *Logger) removeSubscription(id int64) bool {
	ret := false
	p.Lock()
	for i := 0; i < len(p.subscriptions); i++ {
		if p.subscriptions[i].id == id {
			array := p.subscriptions
			p.subscriptions = append(array[:i], array[i+1:]...)
			ret = true
			break
		}
	}
	p.Unlock()
	return ret
}

// Subscribe add a subscription to logger
func (p *Logger) Subscribe() *LogSubscription {
	p.Lock()
	subscription := &LogSubscription{
		id:     GetSeed(),
		logger: p,
		Debug:  nil,
		Info:   nil,
		Warn:   nil,
		Error:  nil,
		Fatal:  nil,
	}
	p.subscriptions = append(p.subscriptions, subscription)
	p.Unlock()
	return subscription
}

// Debug log debug
func (p *Logger) Debug(v ...interface{}) {
	p.log(logMaskDebug, "Debug", fmt.Sprint(v...))
}

// Debugf log debug
func (p *Logger) Debugf(format string, v ...interface{}) {
	p.log(logMaskDebug, "Debug", fmt.Sprintf(format, v...))
}

// Info log info
func (p *Logger) Info(v ...interface{}) {
	p.log(logMaskInfo, "Info", fmt.Sprint(v...))
}

// Infof log info
func (p *Logger) Infof(format string, v ...interface{}) {
	p.log(logMaskInfo, "Info", fmt.Sprintf(format, v...))
}

// Warn log warn
func (p *Logger) Warn(v ...interface{}) {
	p.log(logMaskWarn, "Warn", fmt.Sprint(v...))
}

// Warnf log warn
func (p *Logger) Warnf(format string, v ...interface{}) {
	p.log(logMaskWarn, "Warn", fmt.Sprintf(format, v...))
}

// Error log error
func (p *Logger) Error(v ...interface{}) {
	p.log(logMaskError, "Error", fmt.Sprint(v...))
}

// Errorf log error
func (p *Logger) Errorf(format string, v ...interface{}) {
	p.log(logMaskError, "Error", fmt.Sprintf(format, v...))
}

// Fatal log fatal
func (p *Logger) Fatal(v ...interface{}) {
	p.log(logMaskFatal, "Fatal", fmt.Sprint(v...))
}

// Fatalf log fatal
func (p *Logger) Fatalf(format string, v ...interface{}) {
	p.log(logMaskFatal, "Fatal", fmt.Sprintf(format, v...))
}

func (p *Logger) log(outputLevel int, tag string, msg string) {
	p.Lock()
	level := p.level
	subscriptions := p.subscriptions
	p.Unlock()

	if level&outputLevel > 0 {
		logMsg := fmt.Sprintf(
			"%s %s: %s",
			time.Now().Format("2006-01-02T15:04:05.999Z07:00"),
			tag,
			msg,
		)

		if len(subscriptions) == 0 {
			flags := log.Flags()
			log.SetFlags(0)
			log.Printf(logMsg)
			log.SetFlags(flags)
		} else {
			for i := 0; i < len(subscriptions); i++ {
				var fn func(msg string) = nil
				switch outputLevel {
				case logMaskDebug:
					fn = subscriptions[i].Debug
				case logMaskInfo:
					fn = subscriptions[i].Info
				case logMaskWarn:
					fn = subscriptions[i].Warn
				case logMaskError:
					fn = subscriptions[i].Error
				case logMaskFatal:
					fn = subscriptions[i].Fatal
				}
				if fn != nil {
					fn(logMsg)
				} else {
					flags := log.Flags()
					log.SetFlags(0)
					log.Printf(logMsg)
					log.SetFlags(flags)
				}
			}
		}
	}
}

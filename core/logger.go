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
	// LogLevelWarning this level logs Warning, Error and Fatal
	LogLevelWarning = 7
	// LogLevelInfo this level logs Info, Warning, Error and Fatal
	LogLevelInfo = 15
	// LogLevelAll this level logs Debug, Info, Warning, Error and Fatal
	LogLevelAll = 31

	logMaskFatal   = 1
	logMaskError   = 2
	logMaskWarning = 4
	logMaskInfo    = 8
	logMaskDebug   = 16
)

type LogSubscription struct {
	id      int64
	logger  *Logger
	Debug   func(msg string)
	Info    func(msg string)
	Warning func(msg string)
	Error   func(msg string)
	Fatal   func(msg string)
}

func (p *LogSubscription) Close() RPCError {
	if p.logger == nil {
		return NewRPCError("logger is nil")
	}

	ret := p.logger.removeSubscription(p.id)
	p.id = 0
	p.logger = nil
	p.Debug = nil
	p.Info = nil
	p.Warning = nil
	p.Error = nil
	p.Fatal = nil
	return ret
}

// Logger common logger
type Logger struct {
	level         int
	subscriptions map[int64]*LogSubscription
	cache         []*LogSubscription
	sync.Mutex
}

// NewLogger create new Logger
func NewLogger() *Logger {
	return &Logger{
		level:         LogLevelAll,
		subscriptions: make(map[int64]*LogSubscription),
		cache:         make([]*LogSubscription, 0, 0),
	}
}

// SetLevel set the log level, such as LogLevelOff LogLevelFatal
// LogLevelError LogLevelWarning LogLevelInfo LogLevelDebug LogLevelAll
func (p *Logger) SetLevel(level int) RPCError {
	if level < LogLevelOff || level > LogLevelAll {
		return NewRPCError(
			fmt.Sprintf("level(%d) not supported", level),
		)
	}
	p.level = level
	return nil
}

func (p *Logger) refreshCache() {
	cache := make([]*LogSubscription, 0)
	for _, value := range p.subscriptions {
		cache = append(cache, value)
	}
	p.cache = cache
}

func (p *Logger) removeSubscription(id int64) RPCError {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.subscriptions[id]; ok {
		delete(p.subscriptions, id)
		p.refreshCache()
		return nil
	}

	return NewRPCError("id is not exist")
}

// Subscribe add a subscription to logger
func (p *Logger) Subscribe() *LogSubscription {
	p.Lock()
	defer p.Unlock()

	id := GetSeed()
	subscription := &LogSubscription{
		id:      id,
		logger:  p,
		Debug:   nil,
		Info:    nil,
		Warning: nil,
		Error:   nil,
		Fatal:   nil,
	}

	p.subscriptions[id] = subscription
	p.refreshCache()
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

// Warning log warning
func (p *Logger) Warning(v ...interface{}) {
	p.log(logMaskWarning, "Warning", fmt.Sprint(v...))
}

// Warningf log warning
func (p *Logger) Warningf(format string, v ...interface{}) {
	p.log(logMaskWarning, "Warning", fmt.Sprintf(format, v...))
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
	if p.level&outputLevel > 0 {
		p.Lock()
		defer p.Unlock()

		logMsg := fmt.Sprintf(
			"%s %s: %s",
			time.Now().Format("2006-01-02T15:04:05.999Z07:00"),
			tag,
			msg,
		)

		if len(p.cache) > 0 {
			for i := 0; i < len(p.cache); i++ {
				var fn func(msg string) = nil

				switch outputLevel {
				case logMaskDebug:
					fn = p.cache[i].Debug
					break
				case logMaskInfo:
					fn = p.cache[i].Info
					break
				case logMaskWarning:
					fn = p.cache[i].Warning
					break
				case logMaskError:
					fn = p.cache[i].Error
					break
				case logMaskFatal:
					fn = p.cache[i].Fatal
					break
				}

				if fn != nil {
					fn(logMsg)
				}
			}
		} else {
			flags := log.Flags()
			log.SetFlags(0)
			log.Printf(logMsg)
			log.SetFlags(flags)
		}
	}
}

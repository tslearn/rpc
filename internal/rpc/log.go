package rpc

import (
	"os"
	"sync"

	"github.com/rpccloud/rpc/internal/base"
)

// ErrorLogV ...
type ErrorLog struct {
	level         base.ErrorLevel
	onLogToServer func(err *base.Error)
	file          *os.File
	closeCH       chan bool
	errorCH       chan *base.Error
	sync.Mutex
}

// func getErrorString(err *base.Error) string {
// 	if err == nil {
// 		return ""
// 	}

// 	return fmt.Sprintf(
// 		"%s %s \n",
// 		base.ConvertToIsoDateString(base.TimeNow()),
// 		err.Error(),
// 	)
// }

// // NewErrorLog ...
// func NewErrorLog(
// 	outFile string,
// 	level base.ErrorLevel,
// 	onLogToServer func(err *base.Error),
// ) *ErrorLog {
// 	if onLogToServer == nil {
// 		_, _ = os.Stderr.WriteString(getErrorString(
// 			ErrLogServerHandlerIsNil.AddDebug("onLogToServer is nil"),
// 		))
// 		return nil
// 	}

// 	file, e := func() (*os.File, error) {
// 		if outFile == "" {
// 			return os.Stderr, nil
// 		}

// 		return os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	}()

// 	if e != nil {
// 		_, _ = os.Stderr.WriteString(getErrorString(
// 			ErrLogOpenFile.AddDebug(e.Error()),
// 		))
// 		return nil
// 	}

// 	ret := &ErrorLog{
// 		level:         level,
// 		onLogToServer: onLogToServer,
// 		file:          file,
// 		closeCH:       make(chan bool, 1),
// 		errorCH:       make(chan *Error, 65536),
// 	}

// 	go func() {
// 		for {
// 			select {
// 			case <-ret.closeCH:
// 				return
// 			case err := <-ret.errorCH:
// 				onLogToServer(err)
// 			}
// 		}
// 	}()

// 	return ret
// }

// func (p *ErrorLog) logToFile(err *Error) {
// 	if errLevel := err.GetLevel(); errLevel&p.level != 0 && p.file != nil {
// 		if _, e := p.file.WriteString(getErrorString(err)); e != nil {
// 			_, _ = os.Stderr.WriteString(getErrorString(
// 				ErrLogWriteFile.AddDebug(e.Error()),
// 			))
// 		}
// 	}
// }

// func (p *ErrorLog) LogToFile(err *Error) {
// 	p.Lock()
// 	defer p.Unlock()
// 	p.logToFile(err)
// }

// func (p *ErrorLog) LogToServer(err *Error) {
// 	p.Lock()
// 	defer p.Unlock()

// 	if errLevel := err.GetLevel(); errLevel&p.level != 0 && p.file != nil {
// 		select {
// 		case p.errorCH <- err:
// 			break
// 		default:
// 			p.logToFile(ErrLogChannelFull)
// 		}
// 	}
// }

// // Log ....
// func (p *ErrorLog) Log(err *Error) {
// 	p.LogToFile(err)
// 	p.LogToServer(err)
// }

// func (p *ErrorLog) Close() bool {
// 	p.Lock()
// 	defer p.Unlock()

// 	if p.file != nil {
// 		if e := p.file.Close(); e != nil {
// 			_, _ = os.Stderr.WriteString(getErrorString(
// 				ErrLogCloseFile.AddDebug(e.Error()),
// 			))
// 			return false
// 		}
// 		p.file = nil
// 		p.closeCH <- true
// 		return true
// 	}

// 	return false
// }

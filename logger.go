package nblogger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type Logger interface {
	Trace(format string, v ...any)
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
	SetLogLevel(level int)
	GetLogLevel() int
	Close()
}

const (
	Trace = iota
	Debug
	Info
	Warn
	Error
)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LUTC
	Lblocking
	Lstdout
	LstdFlags = Ldate | Ltime
)

const (
	write = iota
	exit
)

type logMessage struct {
	cmd int
	buf []byte
}

type BasicLogger struct {
	level  int
	flags  int
	writer io.Writer
	ch     chan logMessage
	lock   sync.Mutex
	wg     sync.WaitGroup
}

var levelStringMap = map[int]string{
	Trace: "[TRACE] ",
	Debug: "[DEBUG] ",
	Info:  "[INFO]  ",
	Warn:  "[WARN]  ",
	Error: "[ERROR] ",
}

func init() {
}

func NewLogger(path string, level int, bufferSize int, flags int) (Logger, error) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("err: %v", err)
		return nil, errors.New("file creation fail")
	}

	var writer io.Writer
	if flags&Lstdout != 0 {
		writer = io.MultiWriter(logFile, os.Stdout)
	} else {
		writer = io.MultiWriter(logFile)
	}

	logger := &BasicLogger{
		level:  level,
		flags:  flags,
		writer: writer,
		ch:     make(chan logMessage, bufferSize),
		lock:   sync.Mutex{},
		wg:     sync.WaitGroup{},
	}

	if logger.flags&Lblocking == 0 {
		logger.wg.Add(1)
		go logger.server()
	}

	return logger, nil
}

func (logger *BasicLogger) server() {
	defer logger.wg.Done()
	for {
		if message, err := <-logger.ch; err && message.cmd == write {
			logger.writer.Write(message.buf)
		} else {
			break
		}
	}
}

func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse orde.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (logger *BasicLogger) formatHeader(buf *[]byte, level int, t time.Time, file string, line int) {
	*buf = append(*buf, levelStringMap[level]...)

	if logger.flags&(Ldate|Ltime|Lmicroseconds) != 0 {
		if logger.flags&LUTC != 0 {
			t = t.UTC()
		}
		if logger.flags&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if logger.flags&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if logger.flags&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if logger.flags&(Lshortfile|Llongfile) != 0 {
		if logger.flags&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (logger *BasicLogger) logging(level int, format string, v ...any) {
	if logger.level > level {
		return
	}

	now := time.Now()
	logger.lock.Lock()
	defer logger.lock.Unlock()
	var file string
	var line int

	if logger.flags&(Lshortfile|Llongfile) != 0 {
		logger.lock.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(2)
		if !ok {
			file = "unknown"
			line = 0
		}
		logger.lock.Lock()
	}

	s := fmt.Sprintf(format, v...)
	var buf []byte
	logger.formatHeader(&buf, level, now, file, line)
	buf = append(buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}

	if logger.flags&Lblocking == 0 {
		message := logMessage{
			cmd: write,
			buf: buf,
		}

		logger.ch <- message
	} else {
		logger.writer.Write(buf)
	}
}

func (logger *BasicLogger) Trace(format string, v ...any) {
	logger.logging(Trace, format, v...)
}
func (logger *BasicLogger) Debug(format string, v ...any) {
	logger.logging(Debug, format, v...)
}
func (logger *BasicLogger) Info(format string, v ...any) {
	logger.logging(Info, format, v...)
}
func (logger *BasicLogger) Warn(format string, v ...any) {
	logger.logging(Warn, format, v...)
}
func (logger *BasicLogger) Error(format string, v ...any) {
	logger.logging(Error, format, v...)
}
func (logger *BasicLogger) SetLogLevel(level int) {
	logger.level = level
}
func (logger *BasicLogger) GetLogLevel() int {
	return logger.level
}

func (logger *BasicLogger) Close() {
	logger.ch <- logMessage{cmd: exit}
	logger.wg.Wait()
}

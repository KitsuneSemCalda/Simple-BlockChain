package p2p

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

type Logger struct {
	level  Level
	output io.Writer
	mu     sync.Mutex
}

var defaultLogger = &Logger{
	level:  LevelInfo,
	output: os.Stdout,
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) log(level Level, prefix string, format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("15:04:05")
	
	var colorCode string
	switch level {
	case LevelDebug:
		colorCode = "\033[36m" // Cyan
	case LevelInfo:
		colorCode = "\033[32m" // Green
	case LevelWarn:
		colorCode = "\033[33m" // Yellow
	case LevelError:
		colorCode = "\033[31m" // Red
	}

	if colorCode != "" {
		fmt.Fprintf(l.output, "%s[%s] %s %s\033[0m\n", colorCode, prefix, timestamp, msg)
	} else {
		fmt.Fprintf(l.output, "[%s] %s %s\n", prefix, timestamp, msg)
	}
}

func Debug(prefix, format string, v ...any) { defaultLogger.log(LevelDebug, prefix, format, v...) }
func Info(prefix, format string, v ...any)  { defaultLogger.log(LevelInfo, prefix, format, v...) }
func Warn(prefix, format string, v ...any)  { defaultLogger.log(LevelWarn, prefix, format, v...) }
func Error(prefix, format string, v ...any) { defaultLogger.log(LevelError, prefix, format, v...) }

func SetLogLevel(level Level) { defaultLogger.SetLevel(level) }

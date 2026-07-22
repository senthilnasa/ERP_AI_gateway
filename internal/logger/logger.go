package logger

import (
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var globalLogger *Logger

func Init(levelStr string) *Logger {
	var level LogLevel
	switch strings.ToLower(levelStr) {
	case "debug":
		level = LevelDebug
	case "warn", "warning":
		level = LevelWarn
	case "error":
		level = LevelError
	default:
		level = LevelInfo
	}

	l := &Logger{
		level:  level,
		logger: log.New(os.Stdout, "[AI-GATEWAY] ", log.LstdFlags|log.Lmicroseconds),
	}
	globalLogger = l
	return l
}

func Get() *Logger {
	if globalLogger == nil {
		return Init("info")
	}
	return globalLogger
}

func (l *Logger) Debug(format string, v ...any) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

func (l *Logger) Info(format string, v ...any) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

func (l *Logger) Warn(format string, v ...any) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

func (l *Logger) Error(format string, v ...any) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] "+format, v...)
	}
}

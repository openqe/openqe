package common

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// LogLevelDebug enables debug, info, and error logs
	LogLevelDebug LogLevel = iota
	// LogLevelInfo enables info and error logs (default)
	LogLevelInfo
	// LogLevelError enables only error logs
	LogLevelError
)

// Logger provides structured logging with different log levels
type Logger struct {
	level    LogLevel
	debugLog *log.Logger
	infoLog  *log.Logger
	errorLog *log.Logger
	prefix   string
}

// NewLogger creates a new logger with the specified level and prefix
func NewLogger(level LogLevel, prefix string) *Logger {
	var debugOut, infoOut io.Writer

	// If not debug level, discard debug logs
	if level <= LogLevelDebug {
		debugOut = os.Stdout
	} else {
		debugOut = io.Discard
	}

	// If only error level, discard info logs
	if level <= LogLevelInfo {
		infoOut = os.Stdout
	} else {
		infoOut = io.Discard
	}

	debugPrefix := ""
	infoPrefix := ""
	errorPrefix := ""

	if prefix != "" {
		debugPrefix = fmt.Sprintf("[%s][DEBUG] ", prefix)
		infoPrefix = fmt.Sprintf("[%s] ", prefix)
		errorPrefix = fmt.Sprintf("[%s][ERROR] ", prefix)
	}

	return &Logger{
		level:    level,
		debugLog: log.New(debugOut, debugPrefix, 0),
		infoLog:  log.New(infoOut, infoPrefix, 0),
		errorLog: log.New(os.Stderr, errorPrefix, 0),
		prefix:   prefix,
	}
}

// NewLoggerFromOptions creates a logger based on GlobalOptions
func NewLoggerFromOptions(opts *GlobalOptions, prefix string) *Logger {
	level := LogLevelInfo
	if opts.Verbose {
		level = LogLevelDebug
	}
	return NewLogger(level, prefix)
}

// Debug logs a debug-level message (only visible when verbose mode is enabled)
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLog.Printf(format, v...)
}

// Info logs an info-level message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLog.Printf(format, v...)
}

// Error logs an error-level message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
}

// Debugln logs a debug-level message with a newline
func (l *Logger) Debugln(v ...interface{}) {
	l.debugLog.Println(v...)
}

// Infoln logs an info-level message with a newline
func (l *Logger) Infoln(v ...interface{}) {
	l.infoLog.Println(v...)
}

// Errorln logs an error-level message with a newline
func (l *Logger) Errorln(v ...interface{}) {
	l.errorLog.Println(v...)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level

	// Update debug output
	if level <= LogLevelDebug {
		l.debugLog.SetOutput(os.Stdout)
	} else {
		l.debugLog.SetOutput(io.Discard)
	}

	// Update info output
	if level <= LogLevelInfo {
		l.infoLog.SetOutput(os.Stdout)
	} else {
		l.infoLog.SetOutput(io.Discard)
	}
}

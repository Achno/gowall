package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	RedColor   = "\033[31m"
	ResetColor = "\033[0m"
)

type Logger struct {
	mu        sync.RWMutex
	quiet     bool
	outWriter io.Writer
	errWriter io.Writer
}

// Default global logger instance
var defaultLogger = New(false)

func New(quiet bool) *Logger {
	return &Logger{
		quiet:     quiet,
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}
}

func (l *Logger) SetQuiet(quiet bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.quiet = quiet
	if quiet {
		l.outWriter = io.Discard
	} else {
		l.outWriter = os.Stdout
	}
}

// SetQuiet sets the global logger's quiet state
func SetQuiet(quiet bool) {
	defaultLogger.SetQuiet(quiet)
}

func (l *Logger) Print(v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.quiet {
		fmt.Fprintln(l.outWriter, v...)
	}
}

func (l *Logger) Printf(format string, v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.quiet {
		fmt.Fprintf(l.outWriter, format+"\n", v...)
	}
}

// Error outputs an error message (always logs to stderr)
func (l *Logger) Error(v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprint(v...)
	fmt.Fprintln(l.errWriter, RedColor+message+ResetColor)
}

// Errorf outputs a formatted error message (always logs to stderr)
func (l *Logger) Errorf(format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(format, v...)
	fmt.Fprintln(l.errWriter, RedColor+message+ResetColor)
}

// Fatal logs an error message and exits with status code 1 (always logs to stderr)
func (l *Logger) Fatal(v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprint(v...)
	fmt.Fprintln(l.errWriter, RedColor+message+ResetColor)
	os.Exit(1)
}

// Fatalf logs a formatted error message and exits with status code 1 (always logs to stderr)
func (l *Logger) Fatalf(format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(format, v...)
	fmt.Fprintln(l.errWriter, RedColor+message+ResetColor)
	os.Exit(1)
}

// Global functions using the default logger

// Logs message to stdout if setQuiet was called with false
func Print(v ...any) {
	defaultLogger.Print(v...)
}

// Logs format message to stdout if setQuiet was called with false
func Printf(format string, v ...any) {
	defaultLogger.Printf(format, v...)
}

// Logs message to sderr
func Error(v ...any) {
	defaultLogger.Error(v...)
}

// Logs format message to stderr
func Errorf(format string, v ...any) {
	defaultLogger.Errorf(format, v...)
}

// Logs messages to stderr and exits with code 1
func Fatal(v ...any) {
	defaultLogger.Fatal(v...)
}

// Logs formated message to stderr and exits with code 1
func Fatalf(format string, v ...any) {
	defaultLogger.Fatalf(format, v...)
}

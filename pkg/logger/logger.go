package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Level represents the logging level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger is a thread-safe logging utility writing to file.
type Logger struct {
	mu     sync.Mutex
	prefix string
	level  Level
	file   *os.File
}

// Default logger instance
var defaultLogger *Logger
var logFilePath string

// Global logging functions
func Debug(format string, args ...interface{}) {
	getDefault().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	getDefault().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	getDefault().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	getDefault().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	getDefault().Error(format, args...)
	os.Exit(1)
}

func getDefault() *Logger {
	if defaultLogger == nil {
		defaultLogger = New("app", INFO)
	}
	return defaultLogger
}

// SetLevel sets the minimum logging level.
func SetLevel(level Level) {
	getDefault().SetLevel(level)
}

// SetPrefix sets the logger prefix.
func SetPrefix(prefix string) {
	getDefault().SetPrefix(prefix)
}

// New creates a new logger that writes to file only.
func New(prefix string, level Level) *Logger {
	logDir := os.Getenv("VIBE_LOG_DIR")
	if logDir == "" {
		logDir = os.TempDir()
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath = filepath.Join(logDir, fmt.Sprintf("vibe-echo_%s.log", timestamp))

	file, err := os.Create(logFilePath)
	if err != nil {
		fmt.Printf("create log failed: %v\n", err)
	}

	return &Logger{
		prefix: prefix,
		level:  level,
		file:   file,
	}
}

// GetLogFilePath returns the current log file path.
func GetLogFilePath() string {
	return logFilePath
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	// Write to file only (no console output)
	fileLine := fmt.Sprintf("%s[%s] %s: %s", l.prefix, ts, levelNames[level], msg)
	if l.file != nil {
		io.WriteString(l.file, fileLine+"\n")
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Close closes the log file.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
}

// DumpEnv prints all environment variables (useful for debugging).
func DumpEnv(prefixes []string) {
	getDefault().Info("=== Environment Variables ===")
	for _, e := range os.Environ() {
		for _, p := range prefixes {
			if strings.HasPrefix(e, p) {
				getDefault().Info("  %s", e)
				break
			}
		}
	}
	getDefault().Info("==============================")
}

// DumpWithPrefix prints all variables that start with given prefixes.
func DumpWithPrefix(prefixes ...string) {
	DumpEnv(prefixes)
}
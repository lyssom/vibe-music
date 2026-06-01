package logger

import (
	"fmt"
	"os"
	"path/filepath"
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
var initOnce sync.Once

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
	initOnce.Do(func() {
		// Default to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			cwd = os.TempDir()
		}
		
		logDir := os.Getenv("VIBE_LOG_DIR")
		if logDir == "" {
			logDir = cwd
		}
		
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFilePath = filepath.Join(logDir, fmt.Sprintf("vibe-echo_%s.log", timestamp))
		
		file, err := os.Create(logFilePath)
		if err != nil {
			fmt.Printf("[LOGGER] create log file failed: %v\n", err)
		}
		
		defaultLogger = &Logger{
			prefix: "main",
			level:  DEBUG,
			file:   file,
		}
		
		fmt.Printf("[LOGGER] Log file: %s\n", logFilePath)
	})
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

// New creates a new logger. All loggers write to the same log file.
func New(prefix string, level Level) *Logger {
	// Ensure default logger is initialized
	l := getDefault()
	l.prefix = prefix
	l.level = level
	return l
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
	
	prefix := l.prefix
	if prefix == "" {
		prefix = "app"
	}

	logLine := fmt.Sprintf("%s[%s] %s: %s", ts, prefix, levelNames[level], msg)
	
	// Write to file
	if l.file != nil {
		fmt.Fprintln(l.file, logLine)
		l.file.Sync() // Ensure write is flushed
	}
	
	// Also print to stderr for real-time viewing
	fmt.Fprintln(os.Stderr, logLine)
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
		l.file = nil
	}
}

// DumpWithPrefix prints environment variables matching prefixes.
func DumpWithPrefix(prefixes ...string) {
	Info("=== Environment Variables ===")
	for _, e := range os.Environ() {
		for _, p := range prefixes {
			if len(p) > 0 && len(e) > len(p) && e[:len(p)] == p {
				Info("  %s", e)
				break
			}
		}
	}
	Info("==============================")
}
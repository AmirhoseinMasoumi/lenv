package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents log severity levels.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l Level) Color() string {
	switch l {
	case DEBUG:
		return "\x1b[36m" // cyan
	case INFO:
		return "\x1b[32m" // green
	case WARN:
		return "\x1b[33m" // yellow
	case ERROR:
		return "\x1b[31m" // red
	default:
		return "\x1b[0m"
	}
}

// Logger handles structured logging with file and console output.
type Logger struct {
	level      Level
	fileWriter io.WriteCloser
	mu         sync.Mutex
	useColor   bool
	component  string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger with the given state directory.
// Should be called early in application startup.
func Init(stateDir string) error {
	var initErr error
	once.Do(func() {
		level := parseLevel(os.Getenv("LENV_LOG_LEVEL"))

		logPath := os.Getenv("LENV_LOG_FILE")
		if logPath == "" && stateDir != "" {
			logPath = filepath.Join(stateDir, "lenv.log")
		}

		var fileWriter io.WriteCloser
		if logPath != "" {
			if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
				initErr = fmt.Errorf("create log directory: %w", err)
				return
			}
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				initErr = fmt.Errorf("open log file: %w", err)
				return
			}
			fileWriter = f
		}

		defaultLogger = &Logger{
			level:      level,
			fileWriter: fileWriter,
			useColor:   hasColorSupport(),
		}

		// Log startup
		defaultLogger.Info("lenv logger initialized", "log_level", level.String(), "log_file", logPath)
	})
	return initErr
}

// Close closes the log file. Should be called on application shutdown.
func Close() {
	if defaultLogger != nil && defaultLogger.fileWriter != nil {
		defaultLogger.fileWriter.Close()
	}
}

// WithComponent creates a logger with a component tag for structured logging.
func WithComponent(component string) *Logger {
	if defaultLogger == nil {
		return &Logger{level: INFO, component: component, useColor: hasColorSupport()}
	}
	return &Logger{
		level:      defaultLogger.level,
		fileWriter: defaultLogger.fileWriter,
		useColor:   defaultLogger.useColor,
		component:  component,
	}
}

// Debug logs a debug message.
func Debug(msg string, keyvals ...interface{}) {
	log(DEBUG, "", msg, keyvals...)
}

// Info logs an info message.
func Info(msg string, keyvals ...interface{}) {
	log(INFO, "", msg, keyvals...)
}

// Warn logs a warning message.
func Warn(msg string, keyvals ...interface{}) {
	log(WARN, "", msg, keyvals...)
}

// Error logs an error message.
func Error(msg string, keyvals ...interface{}) {
	log(ERROR, "", msg, keyvals...)
}

// Debug logs a debug message with component context.
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	log(DEBUG, l.component, msg, keyvals...)
}

// Info logs an info message with component context.
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	log(INFO, l.component, msg, keyvals...)
}

// Warn logs a warning message with component context.
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	log(WARN, l.component, msg, keyvals...)
}

// Error logs an error message with component context.
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	log(ERROR, l.component, msg, keyvals...)
}

func log(level Level, component, msg string, keyvals ...interface{}) {
	if defaultLogger == nil {
		// Fallback: just print to stderr if logger not initialized
		if level >= WARN {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", level, msg)
		}
		return
	}

	if level < defaultLogger.level {
		return
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Build key-value pairs string
	var kvStr string
	if len(keyvals) > 0 {
		pairs := make([]string, 0, len(keyvals)/2)
		for i := 0; i < len(keyvals)-1; i += 2 {
			pairs = append(pairs, fmt.Sprintf("%v=%v", keyvals[i], keyvals[i+1]))
		}
		kvStr = " " + strings.Join(pairs, " ")
	}

	// Get caller info for debug logs
	var caller string
	if level == DEBUG {
		if _, file, line, ok := runtime.Caller(2); ok {
			caller = fmt.Sprintf(" [%s:%d]", filepath.Base(file), line)
		}
	}

	// Component tag
	var compStr string
	if component != "" {
		compStr = fmt.Sprintf("[%s] ", component)
	}

	// Write to file (no colors)
	if defaultLogger.fileWriter != nil {
		fileLine := fmt.Sprintf("%s [%s]%s %s%s%s\n",
			timestamp, level.String(), caller, compStr, msg, kvStr)
		defaultLogger.fileWriter.Write([]byte(fileLine))
	}

	// Write to console: DEBUG/INFO when verbose, always WARN and ERROR
	showOnConsole := level >= WARN || defaultLogger.level == DEBUG
	if showOnConsole {
		var consoleLine string
		if defaultLogger.useColor {
			consoleLine = fmt.Sprintf("%s[%s]\x1b[0m %s%s%s\n",
				level.Color(), level.String(), compStr, msg, kvStr)
		} else {
			consoleLine = fmt.Sprintf("[%s] %s%s%s\n",
				level.String(), compStr, msg, kvStr)
		}
		os.Stderr.WriteString(consoleLine)
	}
}

func parseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func hasColorSupport() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	return term != "" && term != "dumb"
}

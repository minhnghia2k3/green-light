package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

// Initialize constants which represent a specific severity level.
// iota keyword as a shortcut to assign integer values to the constants
const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

// Return a human-friendly string for the severity level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger define a custom Logger type.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// New constructor returns a new Logger instance which writes log entries at or above a minimum severity
// level to a specific output destination
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// Helper functions

func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

// Print is an internal method for writing the log entry
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	// Declare anonymous struct holding the data for the log entry.
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// Include stack trace at level ERROR or FATAL
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	// Lock the mutex so that no two writes to the output destination can happen concurrently. If don't do this
	// it's possible that the text for two or more log entries will be collapsed in the output
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write the log entry followed by a newline
	return l.out.Write(append(line, '\n'))
}

// Implement a Write() method so that it satisfied the io.Writer interface.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}

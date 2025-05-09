package logger

import (
	"codacy/cli-v2/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// CustomTextFormatter is our custom formatter for logs
type CustomTextFormatter struct{}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Get the caller information
	var fileLocation string
	if entry.HasCaller() {
		// Get relative path by removing the workspace root
		fullPath := entry.Caller.File
		if workspaceRoot, err := filepath.Abs("."); err == nil {
			if rel, err := filepath.Rel(workspaceRoot, fullPath); err == nil {
				fullPath = rel
			}
		}
		fileLocation = fmt.Sprintf("%s:%d", fullPath, entry.Caller.Line)
	}

	// Format timestamp
	timestamp := entry.Time.Format("2006-01-02T15:04:05-07:00")

	// Format fields
	var fields string
	if len(entry.Data) > 0 {
		var fieldStrings []string
		for k, v := range entry.Data {
			fieldStrings = append(fieldStrings, fmt.Sprintf("%s=%v", k, v))
		}
		fields = " " + strings.Join(fieldStrings, " ")
	}

	// Format the log message
	var location string
	if fileLocation != "" {
		location = fmt.Sprintf(" (%s)", fileLocation)
	}

	logMessage := fmt.Sprintf("%s [%s]%s %s%s\n",
		timestamp,
		strings.ToUpper(entry.Level.String()),
		location,
		entry.Message,
		fields,
	)

	return []byte(logMessage), nil
}

var fileLogger *logrus.Logger

// Initialize sets up the file logger with the given log directory
func Initialize(logsDir string) error {
	// Create a new logger instance
	fileLogger = logrus.New()

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Set up log rotation using lumberjack
	logFile := filepath.Join(logsDir, "codacy-cli.log")

	// Try to create/open the log file to test permissions
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, utils.DefaultFilePerms)
	if err != nil {
		return fmt.Errorf("failed to create/open log file: %w", err)
	}
	f.Close()

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // megabytes
		MaxBackups: 3,    // number of backups to keep
		MaxAge:     28,   // days
		Compress:   true, // compress old files
	}

	// Configure logrus to use our custom formatter
	fileLogger.SetFormatter(&CustomTextFormatter{})
	fileLogger.SetOutput(lumberjackLogger)
	fileLogger.SetLevel(logrus.DebugLevel)

	// Enable caller information for file location
	fileLogger.SetReportCaller(true)

	return nil
}

// Log logs a message with the given level and fields
func Log(level logrus.Level, msg string, fields logrus.Fields) {
	if fileLogger != nil {
		entry := fileLogger.WithFields(fields)
		entry.Log(level, msg)
	}
}

// Info logs an info level message
func Info(msg string, fields ...logrus.Fields) {
	var f logrus.Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	Log(logrus.InfoLevel, msg, f)
}

// Error logs an error level message
func Error(msg string, fields ...logrus.Fields) {
	var f logrus.Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	Log(logrus.ErrorLevel, msg, f)
}

// Debug logs a debug level message
func Debug(msg string, fields ...logrus.Fields) {
	var f logrus.Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	Log(logrus.DebugLevel, msg, f)
}

// Warn logs a warning level message
func Warn(msg string, fields ...logrus.Fields) {
	var f logrus.Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	Log(logrus.WarnLevel, msg, f)
}

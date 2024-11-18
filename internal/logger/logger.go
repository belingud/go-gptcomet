package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    false,
	ReportTimestamp: false,
	Prefix:          "[GPTCommit]",
})

func Info(msg string, args ...interface{}) {
	logger.Info(msg, args...)
}

func Error(msg string, args ...interface{}) {
	logger.Error(msg, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(msg string, args ...interface{}) {
	logger.Fatal(msg, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

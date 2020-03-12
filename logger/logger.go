package logger

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

var DefaultLogFile = "nancy.combined.log"

var Logger *logrus.Logger

func NewLogger() *logrus.Logger {
	file, err := os.OpenFile(getLogFileLocation(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		Logger.Fatal(err)
	}

	logger := &logrus.Logger{
		Out:       file,
		Level:     logrus.DebugLevel,
		Formatter: &logrus.JSONFormatter{},
	}
	Logger = logger
	return Logger
}

func getLogFileLocation() (result string) {
	result, _ = os.UserHomeDir()
	result = path.Join(result, ".ossindex", DefaultLogFile)
	return
}

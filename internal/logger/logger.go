//
// Copyright 2018-present Sonatype Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package logger has functions to obtain a logger, and helpers for setting up where the logger writes
package logger

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
)

const defaultLogFilename = "nancy.combined.log"

var DefaultLogFile = defaultLogFilename

var logLady *logrus.Logger

// GetLogger will either return the existing logger, or setup a new logger
func GetLogger(loggerFilename string, level int) *logrus.Logger {
	if logLady == nil {
		logLevel := getLoggerLevelFromConfig(level)
		err := setupLogger(loggerFilename, &logLevel)
		if err != nil {
			panic(err)
		}
	}
	return logLady
}

// LogFileLocation will return the location on disk of the log file
func LogFileLocation() (result string, err error) {
	result, _ = os.UserHomeDir()
	err = os.MkdirAll(path.Join(result, types.OssIndexDirName), os.ModePerm)
	if err != nil {
		return
	}
	result = path.Join(result, types.OssIndexDirName, DefaultLogFile)
	return
}

func setupLogger(loggerFilename string, level *logrus.Level) (err error) {
	logLady = logrus.New()

	if loggerFilename != "" {
		DefaultLogFile = loggerFilename
	} else {
		DefaultLogFile = defaultLogFilename
	}

	if level == nil {
		logLady.Level = logrus.ErrorLevel
	} else {
		// Done because report caller adds 20-40% overhead per logrus docs, only set this when we want to debug
		if *level > logrus.DebugLevel {
			logLady.SetReportCaller(true)
		}
		logLady.Level = *level
	}

	logLady.Formatter = &logrus.JSONFormatter{DisableHTMLEscape: true}

	location, err := LogFileLocation()
	if err != nil {
		return
	}

	file, err := os.OpenFile(location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	logLady.Out = file

	return
}

func getLoggerLevelFromConfig(level int) logrus.Level {
	switch level {
	case 1:
		return logrus.WarnLevel
	case 2:
		return logrus.InfoLevel
	case 3:
		return logrus.DebugLevel
	case 4:
		return logrus.TraceLevel
	default:
		return logrus.ErrorLevel
	}
}

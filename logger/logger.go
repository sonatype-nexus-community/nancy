// Copyright 2020 Sonatype Inc.
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

// Package logger has functions to obtain a logger, and helpers for setting up where the logger writes
package logger

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// DefaultLogFile can be overriden to use a different file name for upstream consumers
var DefaultLogFile = "nancy.combined.log"

// Logger can be obtained from outside the package
var Logger *logrus.Logger

func init() {
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
}

func getLogFileLocation() (result string) {
	result, _ = os.UserHomeDir()
	result = path.Join(result, ".ossindex", DefaultLogFile)
	return
}

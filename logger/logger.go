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
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
)

// DefaultLogFile can be overriden to use a different file name for upstream consumers
var DefaultLogFile = "nancy.combined.log"

// LogLady can be obtained from outside the package, the name is a reference to the brilliant
// actress in Twin Peaks
var LogLady = logrus.New()

func init() {
	if len(os.Args) > 1 && os.Args[1][:5] == "-test" {
		DefaultLogFile = "nancy.test.log"
	}
	file, err := os.OpenFile(GetLogFileLocation(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Print(err)
	}

	LogLady.Out = file
	LogLady.Level = logrus.InfoLevel
	LogLady.Formatter = &logrus.JSONFormatter{}
}

// GetLogFileLocation will return the location on disk of the log file
func GetLogFileLocation() (result string) {
	result, _ = os.UserHomeDir()
	os.MkdirAll(path.Join(result, types.OssIndexDirName), os.ModePerm)
	result = path.Join(result, types.OssIndexDirName, DefaultLogFile)
	return
}

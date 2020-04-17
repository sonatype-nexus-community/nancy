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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
)

const DefaultLogFilename = "nancy.combined.log"
const TestLogfilename = "nancy.test.log"

// DefaultLogFile can be overridden to use a different file name for upstream consumers
var DefaultLogFile = DefaultLogFilename

// LogLady can be obtained from outside the package, the name is a reference to the brilliant
// actress in Twin Peaks
var LogLady = logrus.New()

func init() {
	err := doInit(os.Args)
	if err != nil {
		panic(err)
	}
}

func doInit(args []string) (err error) {
	if useTestLogFile(args) {
		DefaultLogFile = TestLogfilename
	}

	LogLady.Level = logrus.InfoLevel
	LogLady.Formatter = &logrus.JSONFormatter{}

	location, err := LogFileLocation()
	if err != nil {
		return
	}

	file, err := os.OpenFile(location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	LogLady.Out = file

	return
}

func stringPrefixInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(b, a) {
			return true
		}
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func useTestLogFile(args []string) bool {
	if stringPrefixInSlice("-test.", args) && !stringInSlice("-iq", args) {
		return true
	}
	return false
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

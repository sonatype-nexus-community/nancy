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
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	homeDir, restoreHomeDirFunc := setTempHomeDir(t)
	defer restoreHomeDirFunc()

	if !strings.Contains(GetLogFileLocation(), DefaultLogFile) {
		t.Errorf("Nancy test file not in log file location. args: %+v", os.Args)
	}

	if !strings.Contains(GetLogFileLocation(), homeDir) {
		t.Errorf("Nancy test file not in expected temporary home directory. args: %+v", os.Args)
	}

	LogLady.Info("Test")

	dat, err := ioutil.ReadFile(GetLogFileLocation())
	if err != nil {
		t.Error("Unable to open log file", err)
	}

	var logTest LogTest

	err = json.Unmarshal(dat, &logTest)
	if err != nil {
		t.Error("Improperly written log, should be valid json")
	}

	if logTest.Level != "info" {
		t.Error("Log level not set properly")
	}

	if logTest.Msg != "Test" {
		t.Error("Message not written to log correctly")
	}

}

type LogTest struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Time  string `json:"time"`
}

func setTempHomeDir(t *testing.T) (string, func()) {
	originalHomedir, err := os.UserHomeDir()
	if err != nil {
		t.Error("unable to get home directory", err)
	}
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("unable to create temp directory", err)
	}

	envVar := "HOME"
	if runtime.GOOS == "windows" {
		envVar = "USERPROFILE"
	}
	err = os.Setenv(envVar, tempDir)
	doInit()

	if err != nil {
		t.Error("unable to set temporary home directory")
	}

	restoreHomeDirFunc := func() {
		os.Setenv(envVar, originalHomedir)
		doInit()
	}
	return tempDir, restoreHomeDirFunc
}

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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	location, _ := LogFileLocation()
	if !strings.Contains(location, TestLogfilename) {
		t.Errorf("Nancy test file not in log file location. args: %+v", os.Args)
	}

	LogLady.Info("Test")

	dat, err := ioutil.ReadFile(location)
	if err != nil {
		t.Error("Unable to open log file")
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

func setupTestCase(t *testing.T) func(t *testing.T) {
	origDefault := DefaultLogFile
	t.Logf("setup test case, origDefault: %+v", origDefault)
	DefaultLogFile = DefaultLogFilename
	return func(t *testing.T) {
		t.Logf("teardown test case, origDefault: %+v", origDefault)
		DefaultLogFile = origDefault
	}
}

func TestInit(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	err := doInit([]string{"yadda", "-test.v"})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, TestLogfilename, DefaultLogFile)
}

func TestInitIQ(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	err := doInit([]string{"yadda", "-iq", "-test.v"})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, DefaultLogFilename, DefaultLogFile)
}

func TestInitDefault(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	err := doInit([]string{"yadda", "-yadda"})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, DefaultLogFilename, DefaultLogFile)
}

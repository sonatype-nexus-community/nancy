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

package main

import (
	"bytes"
	"fmt"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBadArgs(t *testing.T) {
	var err error
	cmd := exec.Command("./nancy", "-bad")
	out, err := cmd.CombinedOutput()
	sout := string(out) // because out is []byte
	if err != nil && !strings.Contains(sout, "flag provided but not defined: -bad") {
		fmt.Println(sout) // so we can see the full output
		t.Errorf("%v", err)
	}
}

func TestProcessIQConfigHelp(t *testing.T) {
	err := processIQConfig(configuration.IqConfiguration{Help: true})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 0, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestProcessIQConfigVersion(t *testing.T) {
	err := processIQConfig(configuration.IqConfiguration{Version: true})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 0, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestProcessIQConfigApplicationMissing(t *testing.T) {
	// NOTE: Usage func will not have been setup here, since configuration.ParseIQ() has not yet been called
	err := processIQConfig(configuration.IqConfiguration{})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 3, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestDoIqExitError(t *testing.T) {
	// NOTE: Usage func will be setup by call to configuration.ParseIQ()
	err := doIq([]string{"server-url"})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 3, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestDoIqInvalidFlag(t *testing.T) {
	// This case would call os.Exit() due to Flag parsing set to flag.ExitOnError
	// therefore, changed to flag.ContinueOnError
	err := doIq([]string{"-badflag"})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 2, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestDoConfigInvalidType(t *testing.T) {
	var stdin bytes.Buffer
	stdin.Write([]byte("yadda\n"))
	err := doConfig(&stdin)
	// TODO this should probably return an error
	assert.NoError(t, err)
}

func TestDoConfigInvalidHomeDir(t *testing.T) {
	configuration.HomeDir = "/no/such/dir"
	defer resetConfig(t)

	var stdin bytes.Buffer
	stdin.Write([]byte("ossindex\nmyOssiUsername\nmyOssiToken\n"))
	err := doConfig(&stdin)
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, "Unable to set config for Nancy", exiterr.Message)
		if errCause, ok := exiterr.Err.(*os.PathError); ok {
			assert.Equal(t, "mkdir", errCause.Op)
			assert.Equal(t, "/no/such/dir/.ossindex", errCause.Path)
		} else {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func setupConfig(t *testing.T) (tempDir string) {
	tempDir, err := ioutil.TempDir("", "config-test")
	assert.NoError(t, err)
	configuration.HomeDir = tempDir
	return tempDir
}

func resetConfig(t *testing.T) {
	var err error
	configuration.HomeDir, err = os.UserHomeDir()
	assert.NoError(t, err)
}

func TestDoConfigOssIndex(t *testing.T) {
	tempDir := setupConfig(t)
	defer resetConfig(t)
	//noinspection GoUnhandledErrorResult
	defer os.RemoveAll(tempDir) // clean up

	var stdin bytes.Buffer
	stdin.Write([]byte("ossindex\nmyOssiUsername\nmyOssiToken\n"))
	err := doConfig(&stdin)
	assert.NoError(t, err)
}

func TestDoConfigIq(t *testing.T) {
	tempDir := setupConfig(t)
	defer resetConfig(t)
	//noinspection GoUnhandledErrorResult
	defer os.RemoveAll(tempDir) // clean up

	var stdin bytes.Buffer
	stdin.Write([]byte("iq\nmyIqServer\nmyIqUsername\nmyIqToken\n"))
	err := doConfig(&stdin)
	assert.NoError(t, err)
}

func TestProcessConfigHelp(t *testing.T) {
	err := processConfig(configuration.Configuration{Help: true})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 0, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestProcessConfigVersion(t *testing.T) {
	err := processConfig(configuration.Configuration{Version: true})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 0, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestProcessConfigCleanCache(t *testing.T) {
	err := processConfig(configuration.Configuration{CleanCache: true})
	assert.NoError(t, err)
}

func TestCheckStdInInvalid(t *testing.T) {
	err := checkStdIn()
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 1, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

func TestCheckStdInValid(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "fakeStdIn")
	assert.Nil(t, err)
	//noinspection GoUnhandledErrorResult
	defer os.Remove(tmpfile.Name()) // clean up

	content := []byte("yadda\n")
	if _, err := tmpfile.Write(content); err != nil {
		assert.NoError(t, err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		assert.NoError(t, err)
	}
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin
	os.Stdin = tmpfile

	err = checkStdIn()
	assert.NoError(t, err)
}

func TestCheckOSSIndexNoneVulnerable(t *testing.T) {
	// TODO find way to mock ossindex url for this test, probably just move checkOSSIndex() method to ossindex package
	purls := []string{"pkg:github/BurntSushi/toml@0.3.1"}
	invalidPurls := []string{"invalidPurl"}
	noColor := true
	quiet := true
	config := configuration.Configuration{Formatter: &audit.AuditLogTextFormatter{Quiet: &quiet, NoColor: &noColor}}
	err := checkOSSIndex(purls, invalidPurls, config)
	assert.NoError(t, err)
}

func TestDoOssi(t *testing.T) {
	err := doOssi([]string{""})
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 1, exiterr.ExitCode)
	} else {
		t.Fail()
	}
}

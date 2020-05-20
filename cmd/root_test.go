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

package cmd

import (
	"bytes"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func checkStringContains(t *testing.T, got, substr string) {
	if !strings.Contains(got, substr) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", substr, got)
	}
}

func TestRootCommandUnknownCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "one", "two")
	checkStringContains(t, output, "Error: unknown command \"one\" for \"nancy\"")
	assert.NotNil(t, err)
	checkStringContains(t, err.Error(), "unknown command \"one\" for \"nancy\"")
}

func TestRootCommandNoArgsInvalidStdInErrorExit(t *testing.T) {
	_, err := executeCommand(rootCmd, "")

	serr, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, 1, serr.ExitCode)
}

func TestRootCommandOssi(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	// setup fake Stdin
	tmpfile, err := ioutil.TempFile("", "example")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte("Tom")); err != nil {
		assert.NoError(t, err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		assert.NoError(t, err)
	}
	os.Stdin = tmpfile
	defer tmpfile.Close()

	// run test in separate process
	if os.Getenv("BE_CRASHER") == "1" {
		_, _ = executeCommand(rootCmd, "")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestRootCommandOssi")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err = cmd.Run()
	assert.NoError(t, err)
}

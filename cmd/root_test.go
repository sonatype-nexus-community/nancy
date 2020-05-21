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
	"fmt"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

func TestRootCommandOssiWithPathArgGopkglockOutsideGopath(t *testing.T) {
	dirToGopkglock := "../packages/testdata"
	pathToGopkglock := dirToGopkglock + "/Gopkg.lock"
	_, err := executeCommand(rootCmd, pathToGopkglock)
	assert.Error(t, err)
	if exiterr, ok := err.(customerrors.ErrorExit); ok {
		assert.Equal(t, 3, exiterr.ExitCode)
		assert.Equal(t, fmt.Sprintf("both %s and %s are not within any known GOPATH", dirToGopkglock, dirToGopkglock), exiterr.Err.Error())
		assert.Equal(t, fmt.Sprintf("could not read lock at path %s", pathToGopkglock), exiterr.Message)
	} else {
		t.Fail()
	}
}

func TestRootCommandOssiWithPathArgGosum(t *testing.T) {
	_, err := executeCommand(rootCmd, "../packages/testdata/go.sum")
	assert.NoError(t, err)
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

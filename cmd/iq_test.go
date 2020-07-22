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
	"fmt"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestIqApplicationFlagMissing(t *testing.T) {
	output, err := executeCommand(rootCmd, "iq")
	checkStringContains(t, output, "Error: required flag(s) \"application\" not set")
	assert.NotNil(t, err)
	checkStringContains(t, err.Error(), "required flag(s) \"application\" not set")
}

func TestIqHelp(t *testing.T) {
	output, err := executeCommand(rootCmd, "iq", "--help")
	checkStringContains(t, output, "go list -m -json all | nancy iq --application your_public_application_id --server ")
	assert.Nil(t, err)
}

func TestInitIQConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	cfgDir := path.Join(tempDir, types.IQServerDirName)
	assert.Nil(t, os.Mkdir(cfgDir, 0700))

	cfgFile = path.Join(tempDir, types.IQServerDirName, types.IQServerConfigFileName)

	const credentials = "username: iqUsername\n" +
		"token: iqToken\n" +
		"server: iqServer"
	assert.Nil(t, ioutil.WriteFile(cfgFile, []byte(credentials), 0644))

	initIQConfig()

	assert.Equal(t, "iqUsername", viper.GetString("username"))
	assert.Equal(t, "iqToken", viper.GetString("token"))
	assert.Equal(t, "iqServer", viper.GetString("server"))
}

var testPurls = []string{
	"pkg:golang/github.com/go-yaml/yaml@v2.2.2",
	"pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
}

type iqFactoryMock struct {
	mockIqServer iIQServer
}

func (f iqFactoryMock) create() iIQServer {
	return f.mockIqServer
}

type mockIqServer struct {
	apStatusUrlResult iq.StatusURLResult
	apErr             error
}

//noinspection GoUnusedParameter
func (s mockIqServer) AuditPackages(purls []string, applicationID string) (statusUrlResult iq.StatusURLResult, err error) {
	return s.apStatusUrlResult, s.apErr
}

func TestAuditWithIQServerAuditPackagesError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady = logger.GetLogger("", configOssi.LogLevel)

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{}, apErr: fmt.Errorf("forced error")}}

	err := auditWithIQServer(testPurls, "testapp")

	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, "Uh oh! There was an error with your request to Nexus IQ Server", typedError.Message)
	assert.Equal(t, 3, typedError.ExitCode)
}

func TestAuditWithIQServerResponseError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady = logger.GetLogger("", configOssi.LogLevel)

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{IsError: true, ErrorMessage: "resErrMsg"}}}

	err := auditWithIQServer(testPurls, "testapp")

	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, "Uh oh! There was an error with your request to Nexus IQ Server", typedError.Message)
	assert.Equal(t, 3, typedError.ExitCode)
	assert.Equal(t, "resErrMsg", typedError.Err.Error())
}

func TestAuditWithIQServerPolicyActionNotFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady = logger.GetLogger("", configOssi.LogLevel)

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{}}}

	err := auditWithIQServer(testPurls, "testapp")

	assert.Nil(t, err)
}

func TestAuditWithIQServerPolicyActionFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady = logger.GetLogger("", configOssi.LogLevel)

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{PolicyAction: "Failure"}}}

	err := auditWithIQServer(testPurls, "testapp")

	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1}, typedError)
}

func TestDoIqInvalidStdIn(t *testing.T) {
	err := doIQ(iqCmd, []string{})
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1, Message: "StdIn is invalid, either empty or another reason"}, err)
}

func TestDoIqParseGoListError(t *testing.T) {
	oldStdIn, tmpFile := createFakeStdInWithString(t, "!   ")
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	err := doIQ(iqCmd, []string{})
	assert.NotNil(t, err)
	checkStringContains(t, err.Error(), "index out of range")
}

func TestDoIqAuditError(t *testing.T) {
	oldStdIn, tmpFile := createFakeStdIn(t)
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	err := doIQ(iqCmd, []string{})
	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, "Uh oh! There was an error with your request to Nexus IQ Server", typedError.Message)
	assert.Equal(t, 3, typedError.ExitCode)
}

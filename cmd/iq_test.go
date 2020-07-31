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
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
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
	checkStringContains(t, output, "Error: required flag(s) \""+flagNameIqApplication+"\" not set")
	assert.NotNil(t, err)
	checkStringContains(t, err.Error(), "required flag(s) \""+flagNameIqApplication+"\" not set")
}

func TestIqHelp(t *testing.T) {
	output, err := executeCommand(rootCmd, "iq", "--help")
	checkStringContains(t, output, "go list -m -json all | nancy iq --"+flagNameIqApplication+" your_public_application_id --"+flagNameIqServerUrl+" ")
	assert.Nil(t, err)
}

func setupIQConfigFile(t *testing.T, tempDir string) {
	cfgDirIQ := path.Join(tempDir, types.IQServerDirName)
	assert.Nil(t, os.Mkdir(cfgDirIQ, 0700))

	cfgFileIQ = path.Join(tempDir, types.IQServerDirName, types.IQServerConfigFileName)
}
func resetIQConfigFile() {
	cfgFileIQ = ""
}

func TestInitIQConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	setupTestOSSIConfigFileValues(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()

	setupIQConfigFile(t, tempDir)
	defer func() {
		resetIQConfigFile()
	}()

	const credentials = configuration.YamlKeyIQUsername + ": iqUsernameValue\n" +
		configuration.YamlKeyIQToken + ": iqTokenValue\n" +
		configuration.YamlKeyIQServer + ": iqServerValue"
	assert.Nil(t, ioutil.WriteFile(cfgFileIQ, []byte(credentials), 0644))

	// init order is not guaranteed
	initIQConfig()
	initConfig()

	// verify the OSSI stuff, since we will call both OSSI and IQ
	assert.Equal(t, "ossiUsernameValue", viper.GetString(configuration.YamlKeyUsername))
	assert.Equal(t, "ossiTokenValue", viper.GetString(configuration.YamlKeyToken))
	// verify the IQ stuff
	assert.Equal(t, "iqUsernameValue", viper.GetString(configuration.YamlKeyIQUsername))
	assert.Equal(t, "iqTokenValue", viper.GetString(configuration.YamlKeyIQToken))
	assert.Equal(t, "iqServerValue", viper.GetString(configuration.YamlKeyIQServer))
}

var testPurls = []string{
	"pkg:golang/github.com/go-yaml/yaml@v2.2.2",
	"pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
}

type iqFactoryMock struct {
	mockIqServer iq.IServer
}

func (f iqFactoryMock) create() iq.IServer {
	return f.mockIqServer
}

type mockIqServer struct {
	apStatusUrlResult iq.StatusURLResult
	apErr             error
}

//noinspection GoUnusedParameter
func (s mockIqServer) AuditPackages(purls []string, applicationID string) (iq.StatusURLResult, error) {
	return s.apStatusUrlResult, s.apErr
}

// use compiler to ensure interface is implemented by mock
var _ iq.IServer = (*mockIqServer)(nil)

func TestAuditWithIQServerAuditPackagesError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	expectedErr := fmt.Errorf("forced error")
	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apErr: expectedErr}}

	err := auditWithIQServer(testPurls, "testapp")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAuditWithIQServerResponseError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{IsError: true, ErrorMessage: "resErrMsg"}}}

	err := auditWithIQServer(testPurls, "testapp")

	assert.Error(t, err)
	assert.Equal(t, errors.New("resErrMsg"), err)
}

func TestAuditWithIQServerPolicyActionNotFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{}}}

	err := auditWithIQServer(testPurls, "testapp")

	assert.Nil(t, err)
}

func TestAuditWithIQServerPolicyActionFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{apStatusUrlResult: iq.StatusURLResult{PolicyAction: "Failure"}}}

	err := auditWithIQServer(testPurls, "testapp")

	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1}, typedError)
}

func TestDoIqInvalidStdIn(t *testing.T) {
	err := doIQ(iqCmd, []string{})
	assert.Equal(t, customerrors.ErrorShowLogPath{Err: stdInInvalid}, err)
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

func TestIqCreatorDefaultOptions(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	// setup empty config files
	setupTestOSSIConfigFile(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()
	setupIQConfigFile(t, tempDir)
	defer func() {
		resetIQConfigFile()
	}()

	logLady, _ = test.NewNullLogger()

	bindViperIq(iqCmd)

	iqServer := iqCreator.create()

	ossIndexServer, ok := iqServer.(*iq.Server)
	assert.True(t, ok)
	assert.Equal(t, "admin", ossIndexServer.Options.User)
	assert.Equal(t, "admin123", ossIndexServer.Options.Token)
	assert.Equal(t, "http://localhost:8070", ossIndexServer.Options.Server)
	assert.Equal(t, "", ossIndexServer.Options.OSSIndexUser)
	assert.Equal(t, "", ossIndexServer.Options.OSSIndexToken)
}

func TestIqCreatorOptionsLogging(t *testing.T) {
	logLady, _ = test.NewNullLogger()
	logLady.Level = logrus.DebugLevel
	iqCreator.create()
}

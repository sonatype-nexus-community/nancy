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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sonatype-nexus-community/go-sona-types/configuration"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	ossIndexTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIqApplicationFlagMissing(t *testing.T) {
	output, err := executeCommand(rootCmd, iqCmd.Use)
	assert.Contains(t, output, "Error: required flag(s) \""+flagNameIqApplication+"\" not set")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \""+flagNameIqApplication+"\" not set")
}

func TestIqHelp(t *testing.T) {
	output, err := executeCommand(rootCmd, iqCmd.Use, "--help")
	assert.Contains(t, output, "go list -json -m all | nancy iq --"+flagNameIqApplication+" your_public_application_id --"+flagNameIqServerUrl+" ")
	assert.Nil(t, err)
}

func TestIqCommandPathInvalidName(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	// TODO debug side effects. calling executeCommand() fails as part of test suite, but is fine when run individually.
	//_, err := executeCommand(rootCmd, iqCmd.Use, "--path", "invalidPath", "-a", "appId")
	configOssi = types.Configuration{Path: "invalidPath"}
	err := doIQ(iqCmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("invalid path value. must point to '%s' file. path: ", GopkgLockFilename))
}

func TestIqCommandPathInvalidFile(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	// TODO debug side effects. calling executeCommand() fails as part of test suite, but is fine when run individually.
	//_, err := executeCommand(rootCmd, iqCmd.Use, "--path", GopkgLockFilename, "-a", "appId")
	configOssi = types.Configuration{Path: GopkgLockFilename}
	err := doIQ(iqCmd, []string{})

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "could not find project"), err.Error())
}

func setupIQConfigFile(t *testing.T, tempDir string) {
	cfgDirIQ := path.Join(tempDir, ossIndexTypes.IQServerDirName)
	assert.Nil(t, os.Mkdir(cfgDirIQ, 0700))

	cfgFileIQ = ossIndexTypes.GetIQServerConfigFile(tempDir)
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

	credentials := fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n",
		configuration.ViperKeyIQUsername, "iqUsernameValue",
		configuration.ViperKeyIQToken, "iqTokenValue",
		configuration.ViperKeyIQServer, "iqServerValue")
	assert.Nil(t, ioutil.WriteFile(cfgFileIQ, []byte(credentials), 0644))

	// init order is not guaranteed
	initIQConfig()
	initConfig()

	// verify the OSSI stuff, since we will call both OSSI and IQ
	assert.Equal(t, "ossiUsernameValue", viper.GetString(configuration.ViperKeyUsername))
	assert.Equal(t, "ossiTokenValue", viper.GetString(configuration.ViperKeyToken))
	// verify the IQ stuff
	assert.Equal(t, "iqUsernameValue", viper.GetString(configuration.ViperKeyIQUsername))
	assert.Equal(t, "iqTokenValue", viper.GetString(configuration.ViperKeyIQToken))
	assert.Equal(t, "iqServerValue", viper.GetString(configuration.ViperKeyIQServer))
}

func TestInitIQConfigWithNoConfigFile(t *testing.T) {
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
	credentials := fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n",
		configuration.ViperKeyIQUsername, "iqUsernameValue",
		configuration.ViperKeyIQToken, "iqTokenValue",
		configuration.ViperKeyIQServer, "iqServerValue")
	assert.Nil(t, ioutil.WriteFile(cfgFileIQ, []byte(credentials), 0644))

	// delete the config files
	assert.NoError(t, os.Remove(cfgFile))
	assert.NoError(t, os.Remove(cfgFileIQ))

	// init order is not guaranteed
	initIQConfig()
	initConfig()

	// verify the OSSI stuff, since we will call both OSSI and IQ
	assert.Equal(t, "", viper.GetString(configuration.ViperKeyUsername))
	assert.Equal(t, "", viper.GetString(configuration.ViperKeyToken))
	// verify the IQ stuff
	assert.Equal(t, "", viper.GetString(configuration.ViperKeyIQUsername))
	assert.Equal(t, "", viper.GetString(configuration.ViperKeyIQToken))
	assert.Equal(t, "", viper.GetString(configuration.ViperKeyIQServer))
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
	auditPackagesStatusURLResult iq.StatusURLResult
	auditPackagesErr             error
}

//noinspection GoUnusedParameter
func (s mockIqServer) AuditPackages(purls []string) (iq.StatusURLResult, error) {
	return s.auditPackagesStatusURLResult, s.auditPackagesErr
}

//noinspection GoUnusedParameter
func (s mockIqServer) AuditWithSbom(sbom string) (iq.StatusURLResult, error) {
	return iq.StatusURLResult{}, fmt.Errorf("mock AuditWithSbom not implemented")
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
	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesErr: expectedErr}}

	err := auditWithIQServer(testPurls)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAuditWithIQServerResponseError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: iq.StatusURLResult{IsError: true, ErrorMessage: "resErrMsg"}}}

	err := auditWithIQServer(testPurls)

	assert.Error(t, err)
	assert.Equal(t, errors.New("resErrMsg"), err)
}

func TestAuditWithIQServerPolicyActionNotFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: iq.StatusURLResult{}}}

	err := auditWithIQServer(testPurls)

	assert.Nil(t, err)
}

func TestAuditWithIQServerPolicyActionFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: iq.StatusURLResult{PolicyAction: "Failure"}}}

	err := auditWithIQServer(testPurls)

	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1}, typedError)
}

func TestAuditWithIQServerPolicyActionWarning(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: iq.StatusURLResult{PolicyAction: "Warning"}}}

	err := auditWithIQServer(testPurls)

	assert.Nil(t, err)
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
	assert.Contains(t, err.Error(), "index out of range")
}

func TestDoIqWithIqServerMissingAppIdError(t *testing.T) {
	oldStdIn, tmpFile := createFakeStdInWithString(t, "")
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	err := doIQ(iqCmd, []string{})
	assert.NotNil(t, err)

	typedError, ok := err.(customerrors.ErrorShowLogPath)
	assert.True(t, ok)

	assert.Contains(t, typedError.Err.Error(), "missing options.Application", typedError)
}

func TestDoIqWithIqServerError(t *testing.T) {
	oldStdIn, tmpFile := createFakeStdInWithString(t, "")
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	origConfigIqApplication := configIQ.IQApplication
	defer func() {
		configIQ.IQApplication = origConfigIqApplication
	}()
	configIQ.IQApplication = "testapp"

	bindViperIq(iqCmd)

	err := doIQ(iqCmd, []string{})
	assert.NotNil(t, err)

	typedError, ok := err.(customerrors.ErrorShowLogPath)
	assert.True(t, ok)

	assert.Contains(t, typedError.Error(), "There was an error communicating with Nexus IQ Server to get your internal application ID")
}

func TestDoIqHappyPath(t *testing.T) {
	oldStdIn, tmpFile := createFakeStdInWithString(t, "")
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{}}

	err := doIQ(iqCmd, []string{})
	assert.Nil(t, err)
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

	origConfigIqApplication := configIQ.IQApplication
	defer func() {
		configIQ.IQApplication = origConfigIqApplication
	}()
	configIQ.IQApplication = "testapp"

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
	origConfigIqApplication := configIQ.IQApplication
	defer func() {
		configIQ.IQApplication = origConfigIqApplication
	}()
	configIQ.IQApplication = "testapp"

	bindViperIq(iqCmd)

	logLady, _ = test.NewNullLogger()
	logLady.Level = logrus.DebugLevel
	assert.NotNil(t, iqCreator.create())
}

func Test_showPolicyActionMessage(t *testing.T) {
	verifyReportURL(t, "anythingElse") //default policy action
	verifyReportURL(t, iq.PolicyActionWarning)
	verifyReportURL(t, iq.PolicyActionFailure)
}

func verifyReportURL(t *testing.T, policyAction string) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	theURL := "someURL"
	showPolicyActionMessage(iq.StatusURLResult{AbsoluteReportHTMLURL: theURL, PolicyAction: policyAction}, bufWriter)
	bufWriter.Flush()
	assert.True(t, strings.Contains(buf.String(), "Report URL:  "+theURL), buf.String())
}

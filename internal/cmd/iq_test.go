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
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	internaliq "github.com/sonatype-nexus-community/nancy/internal/iq"
	localossindex "github.com/sonatype-nexus-community/nancy/internal/ossindex"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLifecycleApplicationFlagMissing(t *testing.T) {
	output, err := executeCommand(rootCmd, lifecycleCmd.Use)
	assert.Contains(t, output, "Error: required flag(s) \""+flagNameLifecycleApplication+"\" not set")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \""+flagNameLifecycleApplication+"\" not set")
}

func TestLifecycleHelp(t *testing.T) {
	output, err := executeCommand(rootCmd, lifecycleCmd.Use, "--help")
	assert.Contains(t, output, "go list -json -deps ./... | nancy lifecycle --"+flagNameLifecycleApplication)
	assert.Nil(t, err)
}

func TestLifecycleCommandPathInvalidName(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = types.Configuration{Path: "invalidPath", SkipUpdateCheck: true}
	err := doLifecycle(lifecycleCmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("invalid path value. must point to '%s' file. path: ", GopkgLockFilename))
}

func TestLifecycleCommandPathInvalidFile(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = types.Configuration{Path: GopkgLockFilename, SkipUpdateCheck: true}
	err := doLifecycle(lifecycleCmd, []string{})

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "could not find project"), err.Error())
}

func setupLifecycleConfigFile(t *testing.T, tempDir string) {
	cfgDirIQ := path.Join(tempDir, localossindex.IQServerDirName)
	assert.Nil(t, os.Mkdir(cfgDirIQ, 0700))

	cfgFileIQ = localossindex.GetIQServerConfigFile(tempDir)
}

func resetLifecycleConfigFile() {
	cfgFileIQ = ""
}

func TestInitLifecycleConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := t.TempDir()

	setupTestOSSIConfigFileValues(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()

	setupLifecycleConfigFile(t, tempDir)
	defer func() {
		resetLifecycleConfigFile()
	}()

	credentials := fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n",
		viperKeyLifecycleUsername, "iqUsernameValue",
		viperKeyLifecycleToken, "iqTokenValue",
		viperKeyLifecycleServer, "iqServerValue")
	assert.Nil(t, os.WriteFile(cfgFileIQ, []byte(credentials), 0644))

	// init order is not guaranteed
	initLifecycleConfig()
	initConfig()

	// verify the OSSI stuff, since we will call both OSSI and IQ
	assert.Equal(t, "ossiUsernameValue", viper.GetString(viperKeyOssiUsername))
	assert.Equal(t, "ossiTokenValue", viper.GetString(viperKeyOssiToken))
	// verify the Lifecycle stuff
	assert.Equal(t, "iqUsernameValue", viper.GetString(viperKeyLifecycleUsername))
	assert.Equal(t, "iqTokenValue", viper.GetString(viperKeyLifecycleToken))
	assert.Equal(t, "iqServerValue", viper.GetString(viperKeyLifecycleServer))
}

func TestInitLifecycleConfigWithNoConfigFile(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := t.TempDir()

	setupTestOSSIConfigFileValues(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()

	setupLifecycleConfigFile(t, tempDir)
	defer func() {
		resetLifecycleConfigFile()
	}()
	credentials := fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n",
		viperKeyLifecycleUsername, "iqUsernameValue",
		viperKeyLifecycleToken, "iqTokenValue",
		viperKeyLifecycleServer, "iqServerValue")
	assert.Nil(t, os.WriteFile(cfgFileIQ, []byte(credentials), 0644))

	// delete the config files
	assert.NoError(t, os.Remove(cfgFile))
	assert.NoError(t, os.Remove(cfgFileIQ))

	// init order is not guaranteed
	initLifecycleConfig()
	initConfig()

	// verify the OSSI stuff, since we will call both OSSI and IQ
	assert.Equal(t, "", viper.GetString(viperKeyOssiUsername))
	assert.Equal(t, "", viper.GetString(viperKeyOssiToken))
	// verify the Lifecycle stuff
	assert.Equal(t, "", viper.GetString(viperKeyLifecycleUsername))
	assert.Equal(t, "", viper.GetString(viperKeyLifecycleToken))
	assert.Equal(t, "", viper.GetString(viperKeyLifecycleServer))
}

type iqFactoryMock struct {
	mockIqServer internaliq.IServer
}

func (f iqFactoryMock) create() internaliq.IServer {
	return f.mockIqServer
}

type mockIqServer struct {
	auditPackagesStatusURLResult internaliq.StatusURLResult
	auditPackagesErr             error
}

// noinspection GoUnusedParameter
func (s mockIqServer) AuditPackages(purls []string) (internaliq.StatusURLResult, error) {
	return s.auditPackagesStatusURLResult, s.auditPackagesErr
}

// use compiler to ensure interface is implemented by mock
var _ internaliq.IServer = (*mockIqServer)(nil)

func TestAuditWithLifecycleServerAuditPackagesError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	expectedErr := fmt.Errorf("forced error")
	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesErr: expectedErr}}

	err := auditWithLifecycleServer(testPurls)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAuditWithLifecycleServerResponseError(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: internaliq.StatusURLResult{IsError: true, ErrorMessage: "resErrMsg"}}}

	err := auditWithLifecycleServer(testPurls)

	assert.Error(t, err)
	assert.Equal(t, errors.New("resErrMsg"), err)
}

func TestAuditWithLifecycleServerPolicyActionNotFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: internaliq.StatusURLResult{}}}

	err := auditWithLifecycleServer(testPurls)

	assert.Nil(t, err)
}

func TestAuditWithLifecycleServerPolicyActionFailure(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: internaliq.StatusURLResult{PolicyAction: "Failure"}}}

	err := auditWithLifecycleServer(testPurls)

	typedError, ok := err.(customerrors.ErrorExit)
	assert.True(t, ok)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1}, typedError)
}

func TestAuditWithLifecycleServerPolicyActionWarning(t *testing.T) {
	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	logLady, _ = test.NewNullLogger()

	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesStatusURLResult: internaliq.StatusURLResult{PolicyAction: "Warning"}}}

	err := auditWithLifecycleServer(testPurls)

	assert.Nil(t, err)
}

func TestDoLifecycleInvalidStdIn(t *testing.T) {
	err := doLifecycle(lifecycleCmd, []string{})
	assert.Equal(t, customerrors.ErrorShowLogPath{Err: errStdInInvalid}, err)
}

func TestDoLifecycleParseGoListError(t *testing.T) {
	oldStdIn, tmpFile := createFakeStdInWithString(t, "!   ")
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	err := doLifecycle(lifecycleCmd, []string{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "index out of range")
}

func TestDoLifecycleWithLifecycleServerError(t *testing.T) {
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

	origIqCreator := iqCreator
	defer func() {
		iqCreator = origIqCreator
	}()
	forcedError := customerrors.ErrorShowLogPath{
		Err: fmt.Errorf("error communicating with Lifecycle server to get your internal application ID")}
	iqCreator = &iqFactoryMock{mockIqServer: mockIqServer{auditPackagesErr: forcedError}}

	bindViperLifecycle(lifecycleCmd)

	err := doLifecycle(lifecycleCmd, []string{})
	assert.NotNil(t, err)

	typedError, ok := err.(customerrors.ErrorShowLogPath)
	assert.True(t, ok)

	assert.Contains(t, typedError.Error(), forcedError.Error())
}

func TestDoLifecycleHappyPath(t *testing.T) {
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

	err := doLifecycle(lifecycleCmd, []string{})
	assert.Nil(t, err)
}

func TestLifecycleCreatorDefaultOptions(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := t.TempDir()

	// setup empty config files
	setupTestOSSIConfigFile(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()
	setupLifecycleConfigFile(t, tempDir)
	defer func() {
		resetLifecycleConfigFile()
	}()

	logLady, _ = test.NewNullLogger()

	origConfigIqApplication := configIQ.IQApplication
	defer func() {
		configIQ.IQApplication = origConfigIqApplication
	}()
	configIQ.IQApplication = "testapp"

	bindViperLifecycle(lifecycleCmd)

	iqServer := iqCreator.create()
	assert.NotNil(t, iqServer)

	lifecycleServer, ok := iqServer.(*internaliq.Server)
	assert.True(t, ok)
	assert.Equal(t, "admin", lifecycleServer.Options.User)
	assert.Equal(t, "admin123", lifecycleServer.Options.Token)
	assert.Equal(t, "http://localhost:8070", lifecycleServer.Options.Server)
}

func TestLifecycleCreatorOptionsLogging(t *testing.T) {
	origConfigIqApplication := configIQ.IQApplication
	defer func() {
		configIQ.IQApplication = origConfigIqApplication
	}()
	configIQ.IQApplication = "testapp"

	bindViperLifecycle(lifecycleCmd)

	logLady, _ = test.NewNullLogger()
	logLady.Level = logrus.DebugLevel
	assert.NotNil(t, iqCreator.create())
}

func Test_showPolicyActionMessage(t *testing.T) {
	verifyReportURL(t, "anythingElse") //default policy action
	verifyReportURL(t, internaliq.PolicyActionWarning)
	verifyReportURL(t, internaliq.PolicyActionFailure)
}

func verifyReportURL(t *testing.T, policyAction string) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	theURL := "someURL"
	showPolicyActionMessage(internaliq.StatusURLResult{AbsoluteReportHTMLURL: theURL, PolicyAction: policyAction}, bufWriter)
	assert.NoError(t, bufWriter.Flush())
	assert.True(t, strings.Contains(buf.String(), "Report URL:  "+theURL), buf.String())
}

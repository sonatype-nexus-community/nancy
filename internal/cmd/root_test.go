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
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sonatype-nexus-community/go-sona-types/configuration"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex"
	ossIndexTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/sonatype-nexus-community/nancy/types"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sonatype-nexus-community/nancy/internal/audit"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()

	return buf.String(), err
}

func TestRootCommandNoArgs(t *testing.T) {
	_, err := executeCommand(rootCmd, "")
	assert.Nil(t, err)
}

func TestRootCommandUnknownCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "one", "two")
	assert.Contains(t, output, "Error: unknown command \"one\" for \"nancy\"")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unknown command \"one\" for \"nancy\"")
}

func TestRootCommandCleanCache(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	output, err := executeCommand(rootCmd, "-c")
	assert.Equal(t, output, "")
	assert.Nil(t, err)
}

func TestProcessConfigInvalidStdIn(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = types.Configuration{}
	logLady, _ = test.NewNullLogger()

	err := processConfig()
	assert.Equal(t, errStdInInvalid, err)
}

func TestDoRootCleanCacheError(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = types.Configuration{CleanCache: true}

	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	expectedError := fmt.Errorf("forced clean cache error")
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{auditPackagesErr: expectedError}}

	err := doRoot(nil, nil)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), expectedError.Error()), err.Error())
}

func TestProcessConfigPath(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = types.Configuration{Path: "../../packages/testdata/" + GopkgLockFilename}

	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	ossiCreator = &ossiFactoryMock{}

	err := processConfig()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), " are not within any known GOPATH"), err.Error())
}

func TestGetIsQuiet(t *testing.T) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()

	// all false defaults to quiet
	configOssi = types.Configuration{}
	assert.Equal(t, true, getIsQuiet())

	configOssi = types.Configuration{Quiet: true}
	assert.Equal(t, true, getIsQuiet())

	configOssi = types.Configuration{Loud: true}
	assert.Equal(t, false, getIsQuiet())

	// loud overrides quiet - feel the noise
	configOssi = types.Configuration{Quiet: true, Loud: true}
	assert.Equal(t, false, getIsQuiet())
}

func TestProcessConfigWithVolumeEnabledFormatters(t *testing.T) {
	// cobra default - can't depend on state of configOssi during concurrent tests
	//validateFormatterVolume(t, configOssi, audit.AuditLogTextFormatter{Quiet: true})

	origOutputFormat := outputFormat
	defer func() {
		outputFormat = origOutputFormat
	}()

	outputFormat = "" // default format
	// empty config
	validateFormatterVolume(t, types.Configuration{}, audit.AuditLogTextFormatter{Quiet: true})
	// not quiet, will not be loud - gotta want the volume baby. e.g. --quiet=false
	validateFormatterVolume(t, types.Configuration{Quiet: false}, audit.AuditLogTextFormatter{Quiet: true})
	// loud overrides quiet - feel the noise
	validateFormatterVolume(t, types.Configuration{Quiet: true, Loud: true}, audit.AuditLogTextFormatter{Quiet: false})
	// loud is loud
	validateFormatterVolume(t, types.Configuration{Loud: true}, audit.AuditLogTextFormatter{Quiet: false})
	// not loud is quiet
	validateFormatterVolume(t, types.Configuration{Loud: false}, audit.AuditLogTextFormatter{Quiet: true})

	outputFormat = "text" // explicit text format
	// empty config
	validateFormatterVolume(t, types.Configuration{}, audit.AuditLogTextFormatter{Quiet: true})
	// not quiet, will not be loud - gotta want the volume baby. e.g. --quiet=false
	validateFormatterVolume(t, types.Configuration{Quiet: false}, audit.AuditLogTextFormatter{Quiet: true})
	// loud overrides quiet - feel the noise
	validateFormatterVolume(t, types.Configuration{Quiet: true, Loud: true}, audit.AuditLogTextFormatter{Quiet: false})
	// loud is loud
	validateFormatterVolume(t, types.Configuration{Loud: true}, audit.AuditLogTextFormatter{Quiet: false})
	// not loud is quiet
	validateFormatterVolume(t, types.Configuration{Loud: false}, audit.AuditLogTextFormatter{Quiet: true})

	outputFormat = "csv" // csv format
	// empty config
	validateFormatterVolume(t, types.Configuration{}, audit.CsvFormatter{Quiet: true})
	// not quiet, will not be loud - gotta want the volume baby. e.g. --quiet=false
	validateFormatterVolume(t, types.Configuration{Quiet: false}, audit.CsvFormatter{Quiet: true})
	// loud overrides quiet - feel the noise
	validateFormatterVolume(t, types.Configuration{Quiet: true, Loud: true}, audit.CsvFormatter{Quiet: false})
	// loud is loud
	validateFormatterVolume(t, types.Configuration{Loud: true}, audit.CsvFormatter{Quiet: false})
	// not loud is quiet
	validateFormatterVolume(t, types.Configuration{Loud: false}, audit.CsvFormatter{Quiet: true})
}

func validateFormatterVolume(t *testing.T, testConfig types.Configuration, expectedFormatter logrus.Formatter) {
	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = testConfig

	logLady, _ = test.NewNullLogger()

	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	ossiCreator = &ossiFactoryMock{}

	err := processConfig()
	assert.Equal(t, errStdInInvalid, err)
	assert.Equal(t, expectedFormatter, configOssi.Formatter)
}

func TestDoDepAndParseInvalidPath(t *testing.T) {
	logLady, _ = test.NewNullLogger()
	err := doDepAndParse(ossiFactoryMock{}.create(), GopkgLockFilename)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "could not find project"))
}

func createFakeStdIn(t *testing.T) (oldStdIn *os.File, tmpFile *os.File) {
	return createFakeStdInWithString(t, "Testing")
}
func createFakeStdInWithString(t *testing.T, inputString string) (oldStdIn *os.File, tmpFile *os.File) {
	content := []byte(inputString)
	tmpFile, err := os.CreateTemp("", "tempfile")
	if err != nil {
		t.Error(err)
	}

	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}

	oldStdIn = os.Stdin

	os.Stdin = tmpFile
	return oldStdIn, tmpFile
}

func validateConfigOssi(t *testing.T, expectedConfig types.Configuration, args ...string) {
	oldStdIn, tmpFile := createFakeStdIn(t)
	defer func() {
		os.Stdin = oldStdIn
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	// @todo Special case for empty args tests. maybe submit bug and/or patch to Cobra about it
	// this issue only occurs when running tests individually
	if len(args) == 0 {
		// cobra command adds os arg[0] if command has empty testArgs. see: cobra.Command.go -> line: 914
		origFirstOsArg := os.Args[0]
		os.Args[0] = "cobra.test"
		defer func() {
			os.Args[0] = origFirstOsArg
		}()
	}

	origConfig := configOssi
	defer func() {
		configOssi = origConfig
	}()
	configOssi = types.Configuration{}

	defer func() {
		additionalExcludeVulnerabilityFilePaths = []string{}
	}()

	_, err := executeCommand(rootCmd, args...)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, configOssi)
}

func TestRootCommandLogVerbosity(t *testing.T) {
	logLady, _ = test.NewNullLogger()

	validateConfigOssi(t, types.Configuration{}, "")
	validateConfigOssi(t, types.Configuration{LogLevel: 1}, "-v")
	validateConfigOssi(t, types.Configuration{LogLevel: 2}, "-vv")
	validateConfigOssi(t, types.Configuration{LogLevel: 3}, "-vvv")
}

func TestConfigOssi_defaults(t *testing.T) {
	validateConfigOssi(t, types.Configuration{}, []string{}...)
}

func TestConfigOssi_version(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Version: true, Formatter: logrus.Formatter(nil)}, []string{"--version"}...)
}

func TestConfigOssi_log_level_of_info(t *testing.T) {
	validateConfigOssi(t, types.Configuration{LogLevel: 1}, []string{"-v"}...)
}

func TestConfigOssi_log_level_of_debug(t *testing.T) {
	validateConfigOssi(t, types.Configuration{LogLevel: 2}, []string{"-vv"}...)
}

func TestConfigOssi_log_level_of_trace(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

	validateConfigOssi(t, types.Configuration{LogLevel: 3}, []string{"-vvv"}...)
}

func TestConfigOssi_cleanCache(t *testing.T) {
	validateConfigOssi(t, types.Configuration{CleanCache: true}, []string{"--clean-cache"}...)
}

func TestConfigOssi_skip_update_check(t *testing.T) {
	validateConfigOssi(t, types.Configuration{SkipUpdateCheck: true}, []string{"--skip-update-check"}...)
}

func setupConfig(t *testing.T) (tempDir string) {
	tempDir, err := os.MkdirTemp("", "config-test")
	assert.NoError(t, err)
	return tempDir
}

func resetConfig(t *testing.T, tempDir string) {
	var err error
	assert.NoError(t, err)
	_ = os.RemoveAll(tempDir)
}

func TestInitConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	setupTestOSSIConfigFileValues(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()

	initConfig()

	assert.Equal(t, "ossiUsernameValue", viper.GetString(configuration.ViperKeyUsername))
	assert.Equal(t, "ossiTokenValue", viper.GetString(configuration.ViperKeyToken))
}

func TestInitConfigWithNoConfigFile(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	setupTestOSSIConfigFileValues(t, tempDir)
	defer func() {
		resetOSSIConfigFile()
	}()
	// delete the config file
	assert.NoError(t, os.Remove(cfgFile))

	initConfig()

	assert.Equal(t, "", viper.GetString(configuration.ViperKeyUsername))
	assert.Equal(t, "", viper.GetString(configuration.ViperKeyToken))
}

func setupTestOSSIConfigFile(t *testing.T, tempDir string) {
	cfgDir := path.Join(tempDir, ossIndexTypes.OssIndexDirName)
	assert.Nil(t, os.Mkdir(cfgDir, 0700))

	cfgFile = ossIndexTypes.GetOssIndexConfigFile(tempDir)
}

func resetOSSIConfigFile() {
	cfgFile = ""
}

func setupTestOSSIConfigFileValues(t *testing.T, tempDir string) {
	setupTestOSSIConfigFile(t, tempDir)

	const credentials = configuration.ViperKeyUsername + ": ossiUsernameValue\n" +
		configuration.ViperKeyToken + ": ossiTokenValue"
	assert.Nil(t, os.WriteFile(cfgFile, []byte(credentials), 0644))
}

type ossiFactoryMock struct {
	mockOssiServer ossindex.IServer
}

func (f ossiFactoryMock) create() ossindex.IServer {
	return f.mockOssiServer
}

type mockOssiServer struct {
	auditPackagesResults []ossIndexTypes.Coordinate
	auditPackagesErr     error
}

// noinspection GoUnusedParameter
func (s mockOssiServer) AuditPackages(purls []string) ([]ossIndexTypes.Coordinate, error) {
	return s.auditPackagesResults, s.auditPackagesErr
}

// noinspection GoUnusedParameter
func (s mockOssiServer) Audit(purls []string) (results map[string]ossIndexTypes.Coordinate, err error) {
	results = make(map[string]ossIndexTypes.Coordinate)

	return
}

func (s mockOssiServer) NoCacheNoProblems() error {
	return s.auditPackagesErr
}

// use compiler to ensure interface is implemented by mock
var _ ossindex.IServer = (*mockOssiServer)(nil)

func TestCheckOSSIndexAuditPackagesError(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()

	expectedError := fmt.Errorf("forced error")
	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{auditPackagesErr: expectedError}}

	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Equal(t, expectedError, err)
}

func TestCheckOSSIndexNoVulnerabilities(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{}}

	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Nil(t, err)
}

func TestCheckOSSIndexOneVulnerability(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{auditPackagesResults: []ossIndexTypes.Coordinate{
		{Coordinates: "coord1"},
		{Coordinates: "coord2", Vulnerabilities: []ossIndexTypes.Vulnerability{{}}}}}}

	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1}, err)
}

func TestCheckOSSIndexTwoVulnerabilities(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{auditPackagesResults: []ossIndexTypes.Coordinate{
		{Coordinates: "coord1", Vulnerabilities: []ossIndexTypes.Vulnerability{{}}},
		{Coordinates: "coord2", Vulnerabilities: []ossIndexTypes.Vulnerability{{}}}}}}

	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 2}, err)
}

func TestCheckOSSIndexTwoVulnerabilitiesOnOneCoordinate(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{auditPackagesResults: []ossIndexTypes.Coordinate{
		{Coordinates: "coord1", Vulnerabilities: []ossIndexTypes.Vulnerability{{}, {}}},
		{Coordinates: "coord2"}}}}

	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Equal(t, customerrors.ErrorExit{ExitCode: 1}, err)
}

func TestCheckOSSIndexWithInvalidPurl(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()
	configOssi.Formatter = &logrus.TextFormatter{}

	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{}}

	err := checkOSSIndex(ossiCreator.create(), testPurls, []string{"bogusPurl"})
	assert.Nil(t, err)
}

func TestOssiCreatorOptions(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	logLady, _ = test.NewNullLogger()
	ossIndex := ossiCreator.create()

	ossIndexServer, ok := ossIndex.(*ossindex.Server)
	assert.True(t, ok)
	assert.Equal(t, "", ossIndexServer.Options.Username)
	assert.Equal(t, "", ossIndexServer.Options.Token)
}

func TestOssiCreatorOptionsLogging(t *testing.T) {
	logLady, _ = test.NewNullLogger()
	logLady.Level = logrus.DebugLevel
	ossiCreator.create()
}

func TestCleanUserName(t *testing.T) {
	assert.Equal(t, "***hidden***", cleanUserName(""))
	assert.Equal(t, "1***hidden***1", cleanUserName("1"))
	assert.Equal(t, "1***hidden***2", cleanUserName("12"))
}

func TestViperKeyNameReplacer(t *testing.T) {
	envVarName := viperKeyReplacer.Replace(configuration.ViperKeyUsername)
	assert.Equal(t, "ossi_Username", envVarName)
}

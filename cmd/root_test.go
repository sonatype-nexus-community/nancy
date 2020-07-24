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
	"github.com/sonatype-nexus-community/go-sona-types/ossindex"
	ossIndexTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/customerrors"
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

func createFakeStdIn(t *testing.T) (oldStdIn *os.File, tmpFile *os.File) {
	return createFakeStdInWithString(t, "Testing")
}
func createFakeStdInWithString(t *testing.T, inputString string) (oldStdIn *os.File, tmpFile *os.File) {
	content := []byte(inputString)
	tmpFile, err := ioutil.TempFile("", "tempfile")
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

	configOssi = types.Configuration{}

	_, err := executeCommand(rootCmd, args...)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, configOssi)
}

var defaultAuditLogFormatter = audit.AuditLogTextFormatter{}

func TestRootCommandLogVerbosity(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter}, "")
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 1}, "-v")
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 2}, "-vv")
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 3}, "-vvv")
}

func TestConfigOssi_defaults(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter}, []string{}...)
}

func TestConfigOssi_no_color(t *testing.T) {
	validateConfigOssi(t, types.Configuration{NoColor: true, Formatter: audit.AuditLogTextFormatter{NoColor: true}}, []string{"--no-color"}...)
}

func TestConfigOssi_quiet(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Quiet: true, Formatter: audit.AuditLogTextFormatter{Quiet: true}}, []string{"--quiet"}...)
}

func TestConfigOssi_version(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Version: true, Formatter: logrus.Formatter(nil)}, []string{"--version"}...)
}

func TestConfigOssi_exclude_vulnerabilities(t *testing.T) {
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability=CVE123,CVE988"}...)
}

func TestConfigOssi_stdIn_as_input(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter}, []string{}...)
}

const testdataDir = "../configuration/testdata"

func TestConfigOssi_exclude_vulnerabilities_with_sane_file(t *testing.T) {
	file, _ := os.Open(testdataDir + "/normalIgnore")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999"}}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + file.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_file_empty(t *testing.T) {
	emptyFile, _ := os.Open(testdataDir + "/emptyFile")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + emptyFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_tons_of_newlines(t *testing.T) {
	lotsOfRandomNewlinesFile, _ := os.Open(testdataDir + "/lotsOfRandomWhitespace")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_are_combined_with_file_and_args_values(t *testing.T) {
	lotsOfRandomNewlinesFile, _ := os.Open(testdataDir + "/lotsOfRandomWhitespace")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988", "CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability=CVE123,CVE988", "--exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_file_not_found_does_not_matter(t *testing.T) {
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=/blah-blah-doesnt-exists"}...)
}

func TestConfigOssi_exclude_vulnerabilities_passed_as_directory_does_not_matter(t *testing.T) {
	dir, _ := ioutil.TempDir("", "prefix")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + dir}...)
}

func TestConfigOssi_exclude_vulnerabilities_does_not_need_to_be_passed_if_default_value_is_used(t *testing.T) {
	defaultFileName := ".nancy-ignore"
	err := ioutil.WriteFile(defaultFileName, []byte("DEF-111\nDEF-222"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(defaultFileName)
	}()

	// reset exclude file path, is changed by prior tests
	origExcludeVulnerabilityFilePath := excludeVulnerabilityFilePath
	defer func() {
		excludeVulnerabilityFilePath = origExcludeVulnerabilityFilePath
	}()
	excludeVulnerabilityFilePath = defaultExcludeFilePath

	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"DEF-111", "DEF-222"}}, Formatter: defaultAuditLogFormatter}, []string{}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_comments(t *testing.T) {
	commentedFile, _ := os.Open(testdataDir + "/commented")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + commentedFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_untils(t *testing.T) {
	untilsFile, _ := os.Open(testdataDir + "/untilsAndComments")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"NO-UNTIL-888", "MUST-BE-IGNORED-999", "MUST-BE-IGNORED-1999"}}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + untilsFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_invalid_value_in_untils(t *testing.T) {
	invalidUntilsFile, _ := os.Open(testdataDir + "/untilsInvaild")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + invalidUntilsFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_invalid_date_in_untils(t *testing.T) {
	invalidDateUntilsFile, _ := os.Open(testdataDir + "/untilsBadDateFormat")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, []string{"--exclude-vulnerability-file=" + invalidDateUntilsFile.Name()}...)
}

func TestConfigOssi_output_of_json(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: audit.JsonFormatter{}}, []string{"--output=json"}...)
}

func TestConfigOssi_output_of_json_pretty_print(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: audit.JsonFormatter{PrettyPrint: true}}, []string{"--output=json-pretty"}...)
}

func TestConfigOssi_output_of_csv(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: audit.CsvFormatter{}}, []string{"--output=csv"}...)
}

func TestConfigOssi_output_of_text(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter}, []string{"--output=text"}...)
}

func TestConfigOssi_output_of_bad_value(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter}, []string{"--output=aintgonnadoit"}...)
}

func TestConfigOssi_log_level_of_info(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 1}, []string{"-v"}...)
}

func TestConfigOssi_log_level_of_debug(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 2}, []string{"-vv"}...)
}

func TestConfigOssi_log_level_of_trace(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 3}, []string{"-vvv"}...)
}

func TestConfigOssi_cleanCache(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter, CleanCache: true}, []string{"--clean-cache"}...)
}

func setupConfig(t *testing.T) (tempDir string) {
	tempDir, err := ioutil.TempDir("", "config-test")
	assert.NoError(t, err)
	configuration.HomeDir = tempDir
	return tempDir
}

func resetConfig(t *testing.T, tempDir string) {
	var err error
	configuration.HomeDir, err = os.UserHomeDir()
	assert.NoError(t, err)
	_ = os.RemoveAll(tempDir)
}

func TestInitConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	cfgDir := path.Join(tempDir, types.OssIndexDirName)
	assert.Nil(t, os.Mkdir(cfgDir, 0700))

	cfgFile = path.Join(tempDir, types.OssIndexDirName, types.OssIndexConfigFileName)

	const credentials = "Username: ossiUsername\n" +
		"Token: ossiToken"
	assert.Nil(t, ioutil.WriteFile(cfgFile, []byte(credentials), 0644))

	initConfig()

	assert.Equal(t, "ossiUsername", viper.GetString("Username"))
	assert.Equal(t, "ossiToken", viper.GetString("Token"))
}

type ossiFactoryMock struct {
	mockOssiServer ossindex.IServer
}

func (f ossiFactoryMock) create() ossindex.IServer {
	return f.mockOssiServer
}

type mockOssiServer struct {
	apResults []ossIndexTypes.Coordinate
	apErr     error
}

//noinspection GoUnusedParameter
func (s mockOssiServer) AuditPackages(purls []string) ([]ossIndexTypes.Coordinate, error) {
	return s.apResults, s.apErr
}
func (s mockOssiServer) NoCacheNoProblems() error {
	return s.apErr
}

// use compiler to ensure interface is implemented by mock
var _ ossindex.IServer = (*mockOssiServer)(nil)

func TestCheckOSSIndexAuditPackagesError(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	expectedError := fmt.Errorf("forced error")
	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{apErr: expectedError}}

	logLady = logger.GetLogger("", configOssi.LogLevel)

	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Equal(t, expectedError, err)
}

func TestCheckOSSIndexNoVulnerabilities(t *testing.T) {
	origCreator := ossiCreator
	defer func() {
		ossiCreator = origCreator
	}()
	expectedError := fmt.Errorf("forced error")
	ossiCreator = &ossiFactoryMock{mockOssiServer: mockOssiServer{apResults: []ossIndexTypes.Coordinate{
		{Coordinates: "coord1", Vulnerabilities: []ossIndexTypes.Vulnerability{}},
		{Coordinates: "coord2", Vulnerabilities: []ossIndexTypes.Vulnerability{}},
	}}}

	logLady = logger.GetLogger("", configOssi.LogLevel)
	/*	outputFormat = "text"
		assert.Nil(t, processConfig())
	*/
	err := checkOSSIndex(ossiCreator.create(), testPurls, nil)
	assert.Equal(t, expectedError, err)
}

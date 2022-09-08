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
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sonatype-nexus-community/go-sona-types/configuration"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex"
	"github.com/sonatype-nexus-community/nancy/internal/audit"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestSleuthCommandNoArgs(t *testing.T) {
	_, err := executeCommand(rootCmd, sleuthCmd.Use)
	assert.NotNil(t, err)
	assert.Equal(t, customerrors.ErrorShowLogPath{Err: errStdInInvalid}, err)
}

func TestSleuthCommandPathInvalidName(t *testing.T) {
	_, err := executeCommand(rootCmd, sleuthCmd.Use, "--path", "invalidPath")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("invalid path value. must point to '%s' file. path: ", GopkgLockFilename))
}

func TestSleuthCommandPathInvalidFile(t *testing.T) {
	_, err := executeCommand(rootCmd, sleuthCmd.Use, "--path", GopkgLockFilename)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "could not find project"))
}

func TestConfigOssi_no_color(t *testing.T) {
	validateConfigOssi(t, types.Configuration{NoColor: true, Formatter: audit.AuditLogTextFormatter{NoColor: true, Quiet: true}}, []string{sleuthCmd.Use, "--no-color"}...)
}

func TestConfigOssi_quiet(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Quiet: true, Formatter: audit.AuditLogTextFormatter{Quiet: true}},
		[]string{sleuthCmd.Use, "--quiet"}...)
}

func TestConfigOssi_loud(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Loud: true, Formatter: audit.AuditLogTextFormatter{Quiet: false}},
		[]string{sleuthCmd.Use, "--loud"}...)
}

var defaultAuditLogFormatter = audit.AuditLogTextFormatter{Quiet: true}

func TestConfigOssi_exclude_vulnerabilities(t *testing.T) {
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability=CVE123,CVE988"}...)
}

const testdataDir = "testdata/sleuth"

func TestConfigOssi_exclude_vulnerabilities_with_sane_file(t *testing.T) {
	file, _ := os.Open(testdataDir + "/normalIgnore")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + file.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_file_empty(t *testing.T) {
	emptyFile, _ := os.Open(testdataDir + "/emptyFile")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + emptyFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_tons_of_newlines(t *testing.T) {
	lotsOfRandomNewlinesFile, _ := os.Open(testdataDir + "/lotsOfRandomWhitespace")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_are_combined_with_file_and_args_values(t *testing.T) {
	lotsOfRandomNewlinesFile, _ := os.Open(testdataDir + "/lotsOfRandomWhitespace")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988", "CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability=CVE123,CVE988", "--exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_file_not_found_does_not_matter(t *testing.T) {
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=/blah-blah-doesnt-exists"}...)
}

func TestConfigOssi_exclude_vulnerabilities_passed_as_directory_does_not_matter(t *testing.T) {
	dir, _ := os.MkdirTemp("", "prefix")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + dir}...)
}

func TestConfigOssi_additional_exclude_vulnerabilities(t *testing.T) {
	rootFile, _ := os.Open(testdataDir + "/normalIgnore")
	normalAdditionalIgnore, _ := os.Open(testdataDir + "/normalAdditionalIgnore")
	messyAdditionalIgnore, _ := os.Open(testdataDir + "/messyAdditionalIgnore")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999", "CVX-1016", "CVX-1032", "CVX-1512", "CVX-2064"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + rootFile.Name(), "--additional-exclude-vulnerability-files=" + normalAdditionalIgnore.Name() + "," + messyAdditionalIgnore.Name()}...)
}

func TestConfigOssi_additional_exclude_vulnerabilities_with_empty_base(t *testing.T) {
	rootFile, _ := os.Open(testdataDir + "/emptyFile")
	normalAdditionalIgnore, _ := os.Open(testdataDir + "/normalAdditionalIgnore")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVX-1016", "CVX-1032"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + rootFile.Name(), "--additional-exclude-vulnerability-files=" + normalAdditionalIgnore.Name()}...)
}

func TestConfigOssi_additional_exclude_vulnerabilities_with_empty_additional(t *testing.T) {
	rootFile, _ := os.Open(testdataDir + "/commented")
	normalAdditionalIgnore, _ := os.Open(testdataDir + "/emptyFile")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + rootFile.Name(), "--additional-exclude-vulnerability-files=" + normalAdditionalIgnore.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_does_not_need_to_be_passed_if_default_value_is_used(t *testing.T) {
	defaultFileName := ".nancy-ignore"
	err := os.WriteFile(defaultFileName, []byte("DEF-111\nDEF-222"), 0644)
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

	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"DEF-111", "DEF-222"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_comments(t *testing.T) {
	commentedFile, _ := os.Open(testdataDir + "/commented")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + commentedFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_untils(t *testing.T) {
	untilsFile, _ := os.Open(testdataDir + "/untilsAndComments")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{Cves: []string{"NO-UNTIL-888", "MUST-BE-IGNORED-999", "MUST-BE-IGNORED-1999"}}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + untilsFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_invalid_value_in_untils(t *testing.T) {
	invalidUntilsFile, _ := os.Open(testdataDir + "/untilsInvaild")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + invalidUntilsFile.Name()}...)
}

func TestConfigOssi_exclude_vulnerabilities_when_has_invalid_date_in_untils(t *testing.T) {
	invalidDateUntilsFile, _ := os.Open(testdataDir + "/untilsBadDateFormat")
	validateConfigOssi(t, types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--exclude-vulnerability-file=" + invalidDateUntilsFile.Name()}...)
}

func TestConfigOssi_output_of_json(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: audit.JsonFormatter{}}, []string{sleuthCmd.Use, "--output=json"}...)
}

func TestConfigOssi_output_of_json_pretty_print(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: audit.JsonFormatter{PrettyPrint: true}},
		[]string{sleuthCmd.Use, "--output=json-pretty"}...)
}

func TestConfigOssi_output_of_csv(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: audit.CsvFormatter{Quiet: true}},
		[]string{sleuthCmd.Use, "--output=csv"}...)
}

func TestConfigOssi_output_of_text(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--output=text"}...)
}

func TestConfigOssi_output_of_bad_value(t *testing.T) {
	validateConfigOssi(t, types.Configuration{Formatter: defaultAuditLogFormatter},
		[]string{sleuthCmd.Use, "--output=aintgonnadoit"}...)
}

const testEnvVarValue = "myUnlikelyTestEnvVarValue"

func createServerViaEnv(t *testing.T, viperKey, expectedEnvVarName string) *ossindex.Server {
	envVarName := strings.ToUpper(viperKeyReplacer.Replace(viperKey))
	assert.Equal(t, expectedEnvVarName, envVarName)

	origEnvVarValue := os.Getenv(envVarName)
	defer func() {
		if origEnvVarValue == "" {
			assert.Nil(t, os.Unsetenv(envVarName))
		} else {
			assert.Nil(t, os.Setenv(envVarName, origEnvVarValue))
		}
	}()

	assert.Nil(t, os.Setenv(envVarName, testEnvVarValue))

	logLady, _ = test.NewNullLogger()
	// force call to enable automatic environment variable feature in viper
	setupViperAutomaticEnv()
	ossIndex := ossiCreator.create()
	server := ossIndex.(*ossindex.Server)
	return server
}

func TestConfigOssiUserViaEnv(t *testing.T) {
	assert.Equal(t, testEnvVarValue, createServerViaEnv(t, configuration.ViperKeyUsername, "OSSI_USERNAME").Options.Username)
	assert.Equal(t, testEnvVarValue, createServerViaEnv(t, configuration.ViperKeyToken, "OSSI_TOKEN").Options.Token)
}

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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/cobra"
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

func validateConfigOssi(t *testing.T, expectedError error, expectedConfig types.Configuration, args ...string) {
	// @todo fix hack below!!!!!, maybe submit bug and/or patch to Cobra about it
	// if len(args) == 0 {
	// 	// cobra command adds os arg[0] if command has empty args. see: cobra.Command.go -> line: 914
	// 	origOsArg1 := os.Args[0]
	// 	os.Args[0] = "cobra.test"
	// 	defer func() {
	// 		os.Args[0] = origOsArg1
	// 	}()
	// }

	_, err := executeCommand(rootCmd, args...)

	var ee customerrors.ErrorExit
	if errors.As(expectedError, &ee) && errors.As(err, &ee) {
		// special case comparison for ErrorExit type where errCause may be of type we can't duplicate
		// compare string of errCause
		compareErrorExit(t, expectedError, err)
	} else {
		assert.Equal(t, expectedError, err)
	}
	assert.Equal(t, expectedConfig, configOssi)
}

func compareErrorExit(t *testing.T, expectedErrExit error, actualErrExit error) {
	var eExpected customerrors.ErrorExit
	assert.True(t, errors.As(expectedErrExit, &eExpected))

	var eActual customerrors.ErrorExit
	assert.True(t, errors.As(actualErrExit, &eActual))

	assert.Equal(t, eExpected.ExitCode, eActual.ExitCode)
	assert.Equal(t, eExpected.Message, eActual.Message)

	if eExpected.Err == nil {
		assert.Nil(t, eActual.Err)
	} else {
		// special case comparison for ErrorExit type where errCause may be of a type we can't duplicate, so we
		// compare string representation of errCause
		assert.Equal(t, eExpected.Err.Error(), eActual.Err.Error())
	}
}

var noColor = false
var quiet = false
var testDefaultFormatter = audit.AuditLogTextFormatter{Quiet: quiet, NoColor: noColor}

func TestRootCommandLogVerbosity(t *testing.T) {
	validateConfigOssi(t, stdInInvalid, types.Configuration{Formatter: testDefaultFormatter})
	validateConfigOssi(t, stdInInvalid, types.Configuration{Formatter: testDefaultFormatter, LogLevel: 1}, "-v")
	validateConfigOssi(t, stdInInvalid, types.Configuration{Formatter: testDefaultFormatter, LogLevel: 2}, "-vv")
	validateConfigOssi(t, stdInInvalid, types.Configuration{Formatter: testDefaultFormatter, LogLevel: 3}, "-vvv")
}

func setup() {
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
}

func TestConfigOssi(t *testing.T) {
	const testdataDir = "../configuration/testdata"
	file, _ := os.Open(testdataDir + "/normalIgnore")
	emptyFile, _ := os.Open(testdataDir + "/emptyFile")
	lotsOfRandomNewlinesFile, _ := os.Open(testdataDir + "/lotsOfRandomWhitespace")
	commentedFile, _ := os.Open(testdataDir + "/commented")
	untilsFile, _ := os.Open(testdataDir + "/untilsAndComments")
	invalidUntilsFile, _ := os.Open(testdataDir + "/untilsInvaild")
	invalidUntilLine, _ := bufio.NewReader(invalidUntilsFile).ReadString('\n')
	invalidUntilLine = strings.TrimSpace(invalidUntilLine)

	invalidDateUntilsFile, _ := os.Open(testdataDir + "/untilsBadDateFormat")
	invalidDateUntilLine, _ := bufio.NewReader(invalidDateUntilsFile).ReadString('\n')
	invalidDateUntilLine = strings.TrimSpace(invalidDateUntilLine)

	dir, _ := ioutil.TempDir("", "prefix")

	boolFalse := false
	boolTrue := true

	defaultAuditLogFormatter := audit.AuditLogTextFormatter{Quiet: boolFalse, NoColor: boolFalse}
	quietDefaultFormatter := audit.AuditLogTextFormatter{Quiet: boolTrue, NoColor: boolFalse}

	tests := map[string]struct {
		args           []string
		expectedConfig types.Configuration
		expectedErr    error
	}{
		// TODO: likely not needed
		// "defaults modfile": {args: []string{"/tmp/go.sum"}, expectedConfig: types.Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go.sum", Formatter: defaultAuditLogFormatter}, expectedErr: customerrors.ErrorExit{ExitCode: 3, Err: &os.PathError{Op: "stat", Path: "/tmp/go.sum", Err: fmt.Errorf("no such file or directory")}, Message: "No go.sum found at path: /tmp/go.sum"}},
		// todo Fix help test
		//"help":                                   {args: []string{"--help"}, expectedConfig: configuration.Configuration{UseStdIn: true, Help: true, NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		// todo Fix help test
		//"help modilfe":                           {args: []string{"--help", "/tmp/go2.sum"}, expectedConfig: configuration.Configuration{Help: true, NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum", Formatter: defaultAuditLogFormatter}, expectedErr: createCustomErrorInvalidPathArg("/tmp/go2.sum")},
		// "no color modfile":                        {args: []string{"--no-color", "/tmp/go2.sum"}, expectedConfig: types.Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum", Formatter: audit.AuditLogTextFormatter{Quiet: &boolFalse, NoColor: &boolTrue}}, expectedErr: createCustomErrorInvalidPathArg("/tmp/go2.sum")},
		// "quiet modfile":                           {args: []string{"--quiet", "/tmp/go3.sum"}, expectedConfig: types.Configuration{NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go3.sum", Formatter: quietDefaultFormatter}, expectedErr: createCustomErrorInvalidPathArg("/tmp/go3.sum")},
		// "exclude vulnerabilities modfile":        {args: []string{"--exclude-vulnerability=CVE123,CVE988", "/tmp/go5.sum"}, expectedConfig: types.Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Path: "/tmp/go5.sum", Formatter: defaultAuditLogFormatter}, expectedErr: createCustomErrorInvalidPathArg("/tmp/go5.sum")},
		// "version modfile":                        {args: []string{"--version", "/tmp/go4.sum"}, expectedConfig: types.Configuration{NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Path: "/tmp/go4.sum", Formatter: defaultAuditLogFormatter}, expectedErr: customerrors.ErrorExit{ExitCode: 0}},
		// 		"path but invalid arg":                   {args: []string{"--invalid", "/tmp/go6.sum"}, expectedConfig: types.Configuration{}, expectedErr: errors.New("unknown flag: --invalid")},
		// 		"multiple paths":                         {args: []string{"/tmp/go6.sum", "/tmp/another"}, expectedConfig: types.Configuration{}, expectedErr: customerrors.ErrorExit{ExitCode: 1, Err: errors.New("wrong number of modfile paths: [/tmp/go6.sum /tmp/another]"), Message: "wrong number of modfile paths: [/tmp/go6.sum /tmp/another]"}},
		// 		"output of json modfile":      {args: []string{"--output=json", "/tmp/go14.sum"}, expectedConfig: types.Configuration{Formatter: audit.JsonFormatter{}, Path: "/tmp/go14.sum"}, expectedErr: createCustomErrorInvalidPathArg("/tmp/go14.sum")},
		"defaults":                               {args: []string{}, expectedConfig: types.Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"no color":                               {args: []string{"--no-color"}, expectedConfig: types.Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "", Formatter: audit.AuditLogTextFormatter{Quiet: boolFalse, NoColor: boolTrue}}, expectedErr: nil},
		"quiet":                                  {args: []string{"--quiet"}, expectedConfig: types.Configuration{NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Formatter: quietDefaultFormatter}, expectedErr: nil},
		"version":                                {args: []string{"--version"}, expectedConfig: types.Configuration{NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, expectedErr: customerrors.ErrorExit{ExitCode: 0}},
		"exclude vulnerabilities":                {args: []string{"--exclude-vulnerability=CVE123,CVE988"}, expectedConfig: types.Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"std in as input":                        {args: []string{}, expectedConfig: types.Configuration{Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities with sane file": {args: []string{"--exclude-vulnerability-file=" + file.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999"}}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities when file empty":                                    {args: []string{"--exclude-vulnerability-file=" + emptyFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities when file has tons of newlines":                     {args: []string{"--exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities are combined with file and args values":             {args: []string{"--exclude-vulnerability=CVE123,CVE988", "--exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988", "CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities file not found doesn't matter":                      {args: []string{"--exclude-vulnerability-file=/blah-blah-doesnt-exists"}, expectedConfig: types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities passed as directory doesn't matter":                 {args: []string{"--exclude-vulnerability-file=" + dir}, expectedConfig: types.Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities doesn't need to be passed if default value is used": {args: []string{}, expectedConfig: types.Configuration{CveList: types.CveListFlag{Cves: []string{"DEF-111", "DEF-222"}}, Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"exclude vulnerabilities when has comments":                                  {args: []string{"--exclude-vulnerability-file=" + commentedFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Path: "", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities when has untils":                                    {args: []string{"--exclude-vulnerability-file=" + untilsFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{Cves: []string{"NO-UNTIL-888", "MUST-BE-IGNORED-999", "MUST-BE-IGNORED-1999"}}, Path: "", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities when has invalid value in untils":                   {args: []string{"--exclude-vulnerability-file=" + invalidUntilsFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{}, Path: "", Formatter: defaultAuditLogFormatter}, expectedErr: createCustomErrorWithErrMsg(1, errors.New("failed to parse until at line \""+invalidUntilLine+"\". Expected format is 'until=yyyy-MM-dd'"))},
		"exclude vulnerabilities when has invalid date in untils":                    {args: []string{"--exclude-vulnerability-file=" + invalidDateUntilsFile.Name()}, expectedConfig: types.Configuration{CveList: types.CveListFlag{}, Path: "", Formatter: defaultAuditLogFormatter}, expectedErr: createCustomErrorWithErrMsg(1, errors.New("failed to parse until at line \""+invalidDateUntilLine+"\". Expected format is 'until=yyyy-MM-dd'"))},
		"output of json":              {args: []string{"--output=json"}, expectedConfig: types.Configuration{Formatter: audit.JsonFormatter{}}, expectedErr: nil},
		"output of json pretty print": {args: []string{"--output=json-pretty"}, expectedConfig: types.Configuration{Formatter: audit.JsonFormatter{PrettyPrint: true}, Path: ""}, expectedErr: nil},
		"output of csv":               {args: []string{"--output=csv"}, expectedConfig: types.Configuration{Formatter: audit.CsvFormatter{Quiet: boolFalse}, Path: ""}, expectedErr: nil},
		"output of text":              {args: []string{"--output=text"}, expectedConfig: types.Configuration{Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"output of bad value":         {args: []string{"--output=aintgonnadoit"}, expectedConfig: types.Configuration{Formatter: defaultAuditLogFormatter, Path: ""}, expectedErr: nil},
		"log level of info":           {args: []string{"-v"}, expectedConfig: types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 1}, expectedErr: nil},
		"log level of debug":          {args: []string{"-vv"}, expectedConfig: types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 2}, expectedErr: nil},
		"log level of trace":          {args: []string{"-vvv"}, expectedConfig: types.Configuration{Formatter: defaultAuditLogFormatter, LogLevel: 3}, expectedErr: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			configOssi = types.Configuration{}

			content := []byte("Testing")
			tmpFile, err := ioutil.TempFile("", "tempfile")
			if err != nil {
				t.Error(err)
			}

			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(content); err != nil {
				t.Error(err)
			}

			if _, err := tmpFile.Seek(0, 0); err != nil {
				t.Error(err)
			}

			oldStdIn := os.Stdin

			defer func() {
				os.Stdin = oldStdIn
			}()

			os.Stdin = tmpFile

			setup()

			if name == "exclude vulnerabilities doesn't need to be passed if default value is used" {
				defaultFileName := ".nancy-ignore"
				err := ioutil.WriteFile(defaultFileName, []byte("DEF-111\nDEF-222"), 0644)
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(defaultFileName)
			}

			validateConfigOssi(t, test.expectedErr, test.expectedConfig, test.args...)

			if err := tmpFile.Close(); err != nil {
				t.Error(err)
			}
		})
	}
}

func createCustomErrorInvalidPathArg(path string) customerrors.ErrorExit {
	return customerrors.ErrorExit{ExitCode: 3, Message: "invalid path arg: " + path}
}

func createCustomErrorWithErrMsg(exitCode int, err error) customerrors.ErrorExit {
	return customerrors.ErrorExit{ExitCode: exitCode, Err: err, Message: err.Error()}
}

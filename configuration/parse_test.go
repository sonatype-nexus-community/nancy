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

package configuration

import (
	"bufio"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
)

func TestConfigParse(t *testing.T) {
	file, _ := os.Open("testdata/normalIgnore")
	emptyFile, _ := os.Open("testdata/emptyFile")
	lotsOfRandomNewlinesFile, _ := os.Open("testdata/lotsOfRandomWhitespace")
	commentedFile, _ := os.Open("testdata/commented")
	untilsFile, _ := os.Open("testdata/untilsAndComments")
	invalidUntilsFile, _ := os.Open("testdata/untilsInvaild")
	invalidUntilLine, _ := bufio.NewReader(invalidUntilsFile).ReadString('\n')
	invalidUntilLine = strings.TrimSpace(invalidUntilLine)

	invalidDateUntilsFile, _ := os.Open("testdata/untilsBadDateFormat")
	invalidDateUntilLine, _ := bufio.NewReader(invalidDateUntilsFile).ReadString('\n')
	invalidDateUntilLine = strings.TrimSpace(invalidDateUntilLine)

	dir, _ := ioutil.TempDir("", "prefix")

	boolFalse := false
	boolTrue := true

	defaultAuditLogFormatter := &audit.AuditLogTextFormatter{Quiet: &boolFalse, NoColor: &boolFalse}
	quietDefaultFormatter := &audit.AuditLogTextFormatter{Quiet: &boolTrue, NoColor: &boolFalse}

	tests := map[string]struct {
		args           []string
		expectedConfig Configuration
		expectedErr    error
	}{
		"defaults":                               {args: []string{}, expectedConfig: Configuration{UseStdIn: true, NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"help":                                   {args: []string{"-help"}, expectedConfig: Configuration{UseStdIn: true, Help: true, NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"no color":                               {args: []string{"-no-color"}, expectedConfig: Configuration{UseStdIn: true, NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "", Formatter: &audit.AuditLogTextFormatter{Quiet: &boolFalse, NoColor: &boolTrue}}, expectedErr: nil},
		"no color pkglock":                       {args: []string{"-no-color", "/tmp/Gopkg2.lock"}, expectedConfig: Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/Gopkg2.lock", Formatter: &audit.AuditLogTextFormatter{Quiet: &boolFalse, NoColor: &boolTrue}}, expectedErr: nil},
		"quiet":                                  {args: []string{"-quiet"}, expectedConfig: Configuration{UseStdIn: true, NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Formatter: quietDefaultFormatter}, expectedErr: nil},
		"quiet pkglock":                          {args: []string{"-quiet", "/tmp/Gopkg3.lock"}, expectedConfig: Configuration{NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/Gopkg3.lock", Formatter: quietDefaultFormatter}, expectedErr: nil},
		"version":                                {args: []string{"-version"}, expectedConfig: Configuration{UseStdIn: true, NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"version pkglock":                        {args: []string{"-version", "/tmp/Gopkg4.lock"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Path: "/tmp/Gopkg4.lock", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities":                {args: []string{"-exclude-vulnerability=CVE123,CVE988"}, expectedConfig: Configuration{UseStdIn: true, NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities pkglock":        {args: []string{"-exclude-vulnerability=CVE123,CVE988", "/tmp/Gopkg5.lock"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Path: "/tmp/Gopkg5.lock", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"std in as input":                        {args: []string{}, expectedConfig: Configuration{UseStdIn: true, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"path but invalid arg":                   {args: []string{"-invalid", "/tmp/Gopkg6.lock"}, expectedConfig: Configuration{}, expectedErr: errors.New("flag provided but not defined: -invalid")},
		"multiple paths":                         {args: []string{"/tmp/Gopkg6.lock", "/tmp/another"}, expectedConfig: Configuration{}, expectedErr: errors.New("wrong number of paths: [/tmp/Gopkg6.lock /tmp/another]. only expected 1.")},
		"exclude vulnerabilities with sane file": {args: []string{"-exclude-vulnerability-file=" + file.Name(), "/tmp/Gopkg7.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg7.lock"}, expectedErr: nil},
		"exclude vulnerabilities when file empty":                                    {args: []string{"-exclude-vulnerability-file=" + emptyFile.Name(), "/tmp/Gopkg8.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg8.lock"}, expectedErr: nil},
		"exclude vulnerabilities when file has tons of newlines":                     {args: []string{"-exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name(), "/tmp/Gopkg9.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg9.lock"}, expectedErr: nil},
		"exclude vulnerabilities are combined with file and args values":             {args: []string{"-exclude-vulnerability=CVE123,CVE988", "-exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name(), "/tmp/Gopkg10.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988", "CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg10.lock"}, expectedErr: nil},
		"exclude vulnerabilities file not found doesn't matter":                      {args: []string{"-exclude-vulnerability-file=/blah-blah-doesnt-exists", "/tmp/Gopkg11.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg11.lock"}, expectedErr: nil},
		"exclude vulnerabilities passed as directory doesn't matter":                 {args: []string{"-exclude-vulnerability-file=" + dir, "/tmp/Gopkg12.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg12.lock"}, expectedErr: nil},
		"exclude vulnerabilities doesn't need to be passed if default value is used": {args: []string{"/tmp/Gopkg13.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"DEF-111", "DEF-222"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg13.lock"}, expectedErr: nil},
		"exclude vulnerabilities when has comments":                                  {args: []string{"-exclude-vulnerability-file=" + commentedFile.Name(), "/tmp/Gopkg14.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Path: "/tmp/Gopkg14.lock", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities when has untils":                                    {args: []string{"-exclude-vulnerability-file=" + untilsFile.Name(), "/tmp/Gopkg15.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"NO-UNTIL-888", "MUST-BE-IGNORED-999", "MUST-BE-IGNORED-1999"}}, Path: "/tmp/Gopkg15.lock", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities when has invalid value in untils":                   {args: []string{"-exclude-vulnerability-file=" + invalidUntilsFile.Name(), "/tmp/Gopkg16.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/Gopkg16.lock", Formatter: defaultAuditLogFormatter}, expectedErr: errors.New("failed to parse until at line \"" + invalidUntilLine + "\". Expected format is 'until=yyyy-MM-dd'")},
		"exclude vulnerabilities when has invalid date in untils":                    {args: []string{"-exclude-vulnerability-file=" + invalidDateUntilsFile.Name(), "/tmp/Gopkg17.lock"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/Gopkg17.lock", Formatter: defaultAuditLogFormatter}, expectedErr: errors.New("failed to parse until at line \"" + invalidDateUntilLine + "\". Expected format is 'until=yyyy-MM-dd'")},
		"output of json":              {args: []string{"-output=json"}, expectedConfig: Configuration{UseStdIn: true, Formatter: &audit.JsonFormatter{}}, expectedErr: nil},
		"output of json pkglock":      {args: []string{"-output=json", "/tmp/Gopkg14.lock"}, expectedConfig: Configuration{Formatter: &audit.JsonFormatter{}, Path: "/tmp/Gopkg14.lock"}, expectedErr: nil},
		"output of json pretty print": {args: []string{"-output=json-pretty", "/tmp/Gopkg15.lock"}, expectedConfig: Configuration{Formatter: &audit.JsonFormatter{PrettyPrint: true}, Path: "/tmp/Gopkg15.lock"}, expectedErr: nil},
		"output of csv":               {args: []string{"-output=csv", "/tmp/Gopkg16.lock"}, expectedConfig: Configuration{Formatter: &audit.CsvFormatter{Quiet: &boolFalse}, Path: "/tmp/Gopkg16.lock"}, expectedErr: nil},
		"output of text":              {args: []string{"-output=text", "/tmp/Gopkg17.lock"}, expectedConfig: Configuration{Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg17.lock"}, expectedErr: nil},
		"output of bad value":         {args: []string{"-output=aintgonnadoit", "/tmp/Gopkg18.lock"}, expectedConfig: Configuration{Formatter: defaultAuditLogFormatter, Path: "/tmp/Gopkg18.lock"}, expectedErr: nil},
		"log level of info":           {args: []string{"-v"}, expectedConfig: Configuration{UseStdIn: true, Formatter: defaultAuditLogFormatter, Info: true}, expectedErr: nil},
		"log level of debug":          {args: []string{"-vv"}, expectedConfig: Configuration{UseStdIn: true, Formatter: defaultAuditLogFormatter, Debug: true}, expectedErr: nil},
		"log level of trace":          {args: []string{"-vvv"}, expectedConfig: Configuration{UseStdIn: true, Formatter: defaultAuditLogFormatter, Trace: true}, expectedErr: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			setup()

			if name == "exclude vulnerabilities doesn't need to be passed if default value is used" {
				defaultFileName := ".nancy-ignore"
				err := ioutil.WriteFile(defaultFileName, []byte("DEF-111\nDEF-222"), 0644)
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(defaultFileName)
			}

			actualConfig, actualErr := Parse(test.args)
			assert.Equal(t, test.expectedErr, actualErr)
			assert.EqualValues(t, test.expectedConfig, actualConfig)
		})
	}
}

func TestConfigParseIQ(t *testing.T) {
	tests := map[string]struct {
		args           []string
		expectedConfig IqConfiguration
		expectedErr    error
	}{
		"defaults":                {args: []string{"iq"}, expectedConfig: IqConfiguration{Help: false, Version: false, User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 300}, expectedErr: nil},
		"user token non defaults": {args: []string{"iq", "-user", "nonadmin", "-token", "admin1234"}, expectedConfig: IqConfiguration{Help: false, Version: false, User: "nonadmin", Token: "admin1234", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 300}, expectedErr: nil},
		"server-url non default":  {args: []string{"iq", "-server-url", "http://localhost:8090"}, expectedConfig: IqConfiguration{Help: false, Version: false, User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8090", MaxRetries: 300}, expectedErr: nil},
		"max-retries non default": {args: []string{"iq", "-max-retries", "200"}, expectedConfig: IqConfiguration{Help: false, Version: false, User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 200}, expectedErr: nil},
		"stage non default":       {args: []string{"iq", "-stage", "build"}, expectedConfig: IqConfiguration{Help: false, Version: false, User: "admin", Token: "admin123", Stage: "build", Server: "http://localhost:8070", MaxRetries: 300}, expectedErr: nil},
		"specify application":     {args: []string{"iq", "-application", "testapp"}, expectedConfig: IqConfiguration{Help: false, Version: false, User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 300, Application: "testapp"}, expectedErr: nil},
		"log level of info":       {args: []string{"iq", "-v"}, expectedConfig: IqConfiguration{User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 300, Info: true}, expectedErr: nil},
		"log level of debug":      {args: []string{"iq", "-vv"}, expectedConfig: IqConfiguration{User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 300, Debug: true}, expectedErr: nil},
		"log level of trace":      {args: []string{"iq", "-vvv"}, expectedConfig: IqConfiguration{User: "admin", Token: "admin123", Stage: "develop", Server: "http://localhost:8070", MaxRetries: 300, Trace: true}, expectedErr: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			setup()

			actualConfig, actualErr := ParseIQ(test.args[1:])
			assert.Equal(t, test.expectedErr, actualErr)
			assert.Equal(t, test.expectedConfig, actualConfig)
		})
	}
}

func setup() {
	// Set HomeDir to a nonsensical location to avoid loading file based config
	HomeDir = "/doesnt/exist"
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
}

func TestParseUsage(t *testing.T) {
	setup()
	_, err := Parse([]string{""})
	assert.NoError(t, err)
	// should NOT call os.Exit
	flag.Usage()
}

func TestParseIQUsage(t *testing.T) {
	setup()
	_, err := ParseIQ([]string{""})
	assert.NoError(t, err)
	// should NOT call os.Exit
	flag.Usage()
}

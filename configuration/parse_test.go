// Copyright 2020 Sonatype Inc.
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
	file, _ := os.Open("../testdata/nancyignores/normalIgnore")
	emptyFile, _ := os.Open("../testdata/nancyignores/emptyFile")
	lotsOfRandomNewlinesFile, _ := os.Open("../testdata/nancyignores/lotsOfRandomWhitespace")
	commentedFile, _ := os.Open("../testdata/nancyignores/commented")
	untilsFile, _ := os.Open("../testdata/nancyignores/untilsAndComments")
	invalidUntilsFile, _ := os.Open("../testdata/nancyignores/untilsInvaild")
	invalidUntilLine, _ := bufio.NewReader(invalidUntilsFile).ReadString('\n')
	invalidUntilLine = strings.TrimSpace(invalidUntilLine)

	invalidDateUntilsFile, _ := os.Open("../testdata/nancyignores/untilsBadDateFormat")
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
		"defaults":                               {args: []string{"/tmp/go.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go.sum", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"help":                                   {args: []string{"-help", "/tmp/go2.sum"}, expectedConfig: Configuration{Help: true, NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"no color":                               {args: []string{"-no-color", "/tmp/go2.sum"}, expectedConfig: Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum", Formatter: &audit.AuditLogTextFormatter{Quiet: &boolFalse, NoColor: &boolTrue}}, expectedErr: nil},
		"quiet":                                  {args: []string{"-quiet", "/tmp/go3.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go3.sum", Formatter: quietDefaultFormatter}, expectedErr: nil},
		"quiet stdIn":                            {args: []string{"-quiet"}, expectedConfig: Configuration{UseStdIn: true, NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Formatter: quietDefaultFormatter}, expectedErr: nil},
		"version":                                {args: []string{"-version", "/tmp/go4.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Path: "/tmp/go4.sum", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities":                {args: []string{"-exclude-vulnerability=CVE123,CVE988", "/tmp/go5.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Path: "/tmp/go5.sum", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"std in as input":                        {args: []string{}, expectedConfig: Configuration{UseStdIn: true, Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"path but invalid arg":                   {args: []string{"-invalid", "/tmp/go6.sum"}, expectedConfig: Configuration{}, expectedErr: errors.New("flag provided but not defined: -invalid")},
		"exclude vulnerabilities with sane file": {args: []string{"-exclude-vulnerability-file=" + file.Name(), "/tmp/go7.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go7.sum"}, expectedErr: nil},
		"exclude vulnerabilities when file empty":                                    {args: []string{"-exclude-vulnerability-file=" + emptyFile.Name(), "/tmp/go8.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go8.sum"}, expectedErr: nil},
		"exclude vulnerabilities when file has tons of newlines":                     {args: []string{"-exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name(), "/tmp/go9.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go9.sum"}, expectedErr: nil},
		"exclude vulnerabilities are combined with file and args values":             {args: []string{"-exclude-vulnerability=CVE123,CVE988", "-exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name(), "/tmp/go10.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988", "CVN-111", "CVN-123", "CVN-543"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go10.sum"}, expectedErr: nil},
		"exclude vulnerabilities file not found doesn't matter":                      {args: []string{"-exclude-vulnerability-file=/blah-blah-doesnt-exists", "/tmp/go11.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go11.sum"}, expectedErr: nil},
		"exclude vulnerabilities passed as directory doesn't matter":                 {args: []string{"-exclude-vulnerability-file=" + dir, "/tmp/go12.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go12.sum"}, expectedErr: nil},
		"exclude vulnerabilities doesn't need to be passed if default value is used": {args: []string{"/tmp/go13.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"DEF-111", "DEF-222"}}, Formatter: defaultAuditLogFormatter, Path: "/tmp/go13.sum"}, expectedErr: nil},
		"exclude vulnerabilities when has comments":                                  {args: []string{"-exclude-vulnerability-file=" + commentedFile.Name(), "/tmp/go14.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Path: "/tmp/go14.sum", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities when has untils":                                    {args: []string{"-exclude-vulnerability-file=" + untilsFile.Name(), "/tmp/go15.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"NO-UNTIL-888", "MUST-BE-IGNORED-999", "MUST-BE-IGNORED-1999"}}, Path: "/tmp/go15.sum", Formatter: defaultAuditLogFormatter}, expectedErr: nil},
		"exclude vulnerabilities when has invalid value in untils":                   {args: []string{"-exclude-vulnerability-file=" + invalidUntilsFile.Name(), "/tmp/go16.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/go16.sum", Formatter: defaultAuditLogFormatter}, expectedErr: errors.New("failed to parse until at line '" + invalidUntilLine + "'. Expected format is 'until=yyyy-MM-dd'")},
		"exclude vulnerabilities when has invalid date in untils":                    {args: []string{"-exclude-vulnerability-file=" + invalidDateUntilsFile.Name(), "/tmp/go17.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/go17.sum", Formatter: defaultAuditLogFormatter}, expectedErr: errors.New("failed to parse until at line '" + invalidDateUntilLine + "'. Expected format is 'until=yyyy-MM-dd'")},
		"output of json":              {args: []string{"-output=json", "/tmp/go14.sum"}, expectedConfig: Configuration{Formatter: &audit.JsonFormatter{}, Path: "/tmp/go14.sum"}, expectedErr: nil},
		"output of json pretty print": {args: []string{"-output=json-pretty", "/tmp/go15.sum"}, expectedConfig: Configuration{Formatter: &audit.JsonFormatter{PrettyPrint: true}, Path: "/tmp/go15.sum"}, expectedErr: nil},
		"output of csv":               {args: []string{"-output=csv", "/tmp/go16.sum"}, expectedConfig: Configuration{Formatter: &audit.CsvFormatter{Quiet: &boolFalse}, Path: "/tmp/go16.sum"}, expectedErr: nil},
		"output of text":              {args: []string{"-output=text", "/tmp/go17.sum"}, expectedConfig: Configuration{Formatter: defaultAuditLogFormatter, Path: "/tmp/go17.sum"}, expectedErr: nil},
		"output of bad value":         {args: []string{"-output=aintgonnadoit", "/tmp/go18.sum"}, expectedConfig: Configuration{Formatter: defaultAuditLogFormatter, Path: "/tmp/go18.sum"}, expectedErr: nil},
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
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
}

package configuration

import (
	"bufio"
	"errors"
	"flag"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestConfigParse(t *testing.T) {
	file := setupCVEExcludeFile(t, "CVF-000\nCVF-123\nCVF-9999")
	emptyFile := setupCVEExcludeFile(t, "")
	lotsOfRandomNewlinesFile := setupCVEExcludeFile(t, "\n\nCVN-111\n\n\nCVN-123\nCVN-543\n\n")
	dir, _ := ioutil.TempDir("", "prefix")

	defer os.Remove(file.Name())
	defer os.Remove(emptyFile.Name())
	defer os.Remove(lotsOfRandomNewlinesFile.Name())
	defer os.Remove(dir)

	tests := map[string]struct {
		args           []string
		expectedConfig Configuration
		expectedErr    error
	}{
		"defaults":                               {args: []string{"/tmp/go.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go.sum"}, expectedErr: nil},
		"no color":                               {args: []string{"-no-color", "/tmp/go2.sum"}, expectedConfig: Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum"}, expectedErr: nil},
		"no color (deprecated)":                  {args: []string{"-noColor", "/tmp/go2.sum"}, expectedConfig: Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum"}, expectedErr: nil},
		"quiet":                                  {args: []string{"-quiet", "/tmp/go3.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go3.sum"}, expectedErr: nil},
		"version":                                {args: []string{"-version", "/tmp/go4.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Path: "/tmp/go4.sum"}, expectedErr: nil},
		"exclude vulnerabilities":                {args: []string{"-exclude-vulnerability=CVE123,CVE988", "/tmp/go5.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988"}}, Path: "/tmp/go5.sum"}, expectedErr: nil},
		"no args":                                {args: []string{}, expectedConfig: Configuration{}, expectedErr: errors.New("no arguments passed")},
		"path but invalid arg":                   {args: []string{"-invalid", "/tmp/go6.sum"}, expectedConfig: Configuration{}, expectedErr: errors.New("flag provided but not defined: -invalid")},
		"exclude vulnerabilities with sane file": {args: []string{"-exclude-vulnerability-file=" + file.Name(), "/tmp/go7.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVF-000", "CVF-123", "CVF-9999"}}, Path: "/tmp/go7.sum"}, expectedErr: nil},
		"exclude vulnerabilities when file empty":                                    {args: []string{"-exclude-vulnerability-file=" + emptyFile.Name(), "/tmp/go8.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/go8.sum"}, expectedErr: nil},
		"exclude vulnerabilities when file has tons of newlines":                     {args: []string{"-exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name(), "/tmp/go9.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVN-111", "CVN-123", "CVN-543"}}, Path: "/tmp/go9.sum"}, expectedErr: nil},
		"exclude vulnerabilities are combined with file and args values":             {args: []string{"-exclude-vulnerability=CVE123,CVE988", "-exclude-vulnerability-file=" + lotsOfRandomNewlinesFile.Name(), "/tmp/go10.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"CVE123", "CVE988", "CVN-111", "CVN-123", "CVN-543"}}, Path: "/tmp/go10.sum"}, expectedErr: nil},
		"exclude vulnerabilities file not found doesn't matter":                      {args: []string{"-exclude-vulnerability-file=/blah-blah-doesnt-exists", "/tmp/go11.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/go11.sum"}, expectedErr: nil},
		"exclude vulnerabilities passed as directory doesn't matter":                 {args: []string{"-exclude-vulnerability-file=" + dir, "/tmp/go12.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{}, Path: "/tmp/go12.sum"}, expectedErr: nil},
		"exclude vulnerabilities doesn't need to be passed if default value is used": {args: []string{"/tmp/go13.sum"}, expectedConfig: Configuration{CveList: types.CveListFlag{Cves: []string{"DEF-111","DEF-222"}}, Path: "/tmp/go13.sum"}, expectedErr: nil},
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
			assert.Equal(t, test.expectedConfig, actualConfig)
		})
	}
}

func setupCVEExcludeFile(t *testing.T, fileContents string) (file *os.File) {
	file, err := ioutil.TempFile("", "prefix")
	if err != nil {
		t.Fatal(err)
	}
	w := bufio.NewWriter(file)
	_, err = w.WriteString(fileContents)
	if err != nil {
		t.Fatal(err)
	}
	err = w.Flush()
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func setup() {
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
}

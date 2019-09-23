package configuration

import (
	"errors"
	"flag"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigParse(t *testing.T) {
	tests := map[string]struct {
		args           []string
		expectedConfig Configuration
		expectedErr    error
	}{
		"defaults":                {args: []string{"/tmp/go.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go.sum"}, expectedErr: nil},
		"no color":                {args: []string{"-noColor", "/tmp/go2.sum"}, expectedConfig: Configuration{NoColor: true, Quiet: false, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go2.sum"}, expectedErr: nil},
		"quiet":                   {args: []string{"-quiet", "/tmp/go3.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: true, Version: false, CveList: types.CveListFlag{}, Path: "/tmp/go3.sum"}, expectedErr: nil},
		"version":                 {args: []string{"-version", "/tmp/go4.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: true, CveList: types.CveListFlag{}, Path: "/tmp/go4.sum"}, expectedErr: nil},
		"exclude vulnerabilities": {args: []string{"-exclude-vulnerability=CVE123,CVE988", "/tmp/go5.sum"}, expectedConfig: Configuration{NoColor: false, Quiet: false, Version: false, CveList: types.CveListFlag{Cves:[]string{"CVE123", "CVE988"}}, Path: "/tmp/go5.sum"}, expectedErr: nil},
		"no args": {args: []string{}, expectedConfig: Configuration{}, expectedErr: errors.New("no arguments passed")},
		"path but invalid arg": {args: []string{"-invalid", "/tmp/go6.sum"}, expectedConfig: Configuration{}, expectedErr: errors.New("flag provided but not defined: -invalid")},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			setup()

			actualConfig, actualErr := Parse(test.args)
			assert.Equal(t, test.expectedErr, actualErr)
			assert.Equal(t, test.expectedConfig, actualConfig)
		})
	}
}

func setup() {
	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
}
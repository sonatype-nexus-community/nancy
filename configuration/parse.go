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

// Package configuration is for getting and setting OSS Index or Nexus IQ Server configuration
package configuration

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/audit"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"
)

// Config is a struct for loading OSS Index or Nexus IQ Server config
type Config struct {
	Type           ConfigType
	Help           bool
	Version        bool
	Info           bool
	Debug          bool
	Trace          bool
	UseStdIn       bool
	NoColor        bool
	Quiet          bool
	CleanCache     bool
	Path           string
	IQConfig       IQConfig
	OSSIndexConfig OSSIndexConfig
}

// ConfigType is a struct for valid Nancy config types
type ConfigType int

func (t ConfigType) String() string {
	types := [...]string{"OSSIndex", "IQServer"}
	if t < OSSIndex || t > IQServer {
		return "Unsupported"
	}
	return types[t]
}

const (
	OSSIndex ConfigType = 0
	IQServer ConfigType = 1
)

var unixComments = regexp.MustCompile(`#.*$`)
var untilComment = regexp.MustCompile(`(until=)(.*)`)

// OssIndexOutputFormats represents the potential formats that can be used to output results
var OssIndexOutputFormats = map[string]logrus.Formatter{
	"json":        &audit.JsonFormatter{},
	"json-pretty": &audit.JsonFormatter{PrettyPrint: true},
	"text":        &audit.AuditLogTextFormatter{},
	"csv":         &audit.CsvFormatter{},
}

func NewConfig(typeOfConfig ConfigType) *Config {
	config := Config{}
	config.Type = typeOfConfig
	return &config
}

// Parse is used to parse command line args, and populate the Config struct with the values necessary
func (c *Config) Parse(args []string) (err error) {
	switch c.Type {
	case OSSIndex:
		return c.doParseOssIndex(args)
	case IQServer:
		return c.doParseIqServer(args)
	default:
		return c.doParseOssIndex(args)
	}
}

func (c *Config) doParseOssIndex(args []string) (err error) {
	var excludeVulnerabilityFilePath string
	var outputFormat string

	flag.BoolVar(&c.Help, "help", false, "provides help text on how to use nancy")
	flag.BoolVar(&c.NoColor, "no-color", false, "indicate output should not be colorized")
	flag.BoolVar(&c.Quiet, "quiet", false, "indicate output should contain only packages with vulnerabilities")
	flag.BoolVar(&c.Version, "version", false, "prints current nancy version")
	flag.BoolVar(&c.CleanCache, "clean-cache", false, "Deletes local cache directory")
	flag.BoolVar(&c.Info, "v", false, "Set log level to Info")
	flag.BoolVar(&c.Debug, "vv", false, "Set log level to Debug")
	flag.BoolVar(&c.Trace, "vvv", false, "Set log level to Trace")
	flag.Var(&c.OSSIndexConfig.CveList, "exclude-vulnerability", "Comma separated list of CVEs to exclude")
	flag.StringVar(&c.OSSIndexConfig.Username, "user", "", "Specify OSS Index username for request")
	flag.StringVar(&c.OSSIndexConfig.Token, "token", "", "Specify OSS Index API token for request")
	flag.StringVar(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "./.nancy-ignore", "Path to a file containing newline separated CVEs to be excluded")
	flag.StringVar(&outputFormat, "output", "text", "Styling for output format. "+fmt.Sprintf("%+q", reflect.ValueOf(OssIndexOutputFormats).MapKeys()))

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `Usage:
	go list -m all | nancy [options]
	go list -m all | nancy iq [options]
	nancy config
	nancy [options] </path/to/Gopkg.lock>
	nancy [options] </path/to/go.sum>
			
Options:
`)
		flag.PrintDefaults()
		os.Exit(2)
	}

	ConfigLocation = filepath.Join(HomeDir, types.OssIndexDirName, types.OssIndexConfigFileName)

	err = c.loadConfigFromFile(ConfigLocation, OSSIndex)
	if err != nil {
		fmt.Println(err)
		LogLady.Info("Unable to load OSS Index config from file")
	}

	err = flag.CommandLine.Parse(args)
	if err != nil {
		return
	}

	modfilePath, err := getModfilePath()
	if err != nil {
		return
	}
	if len(modfilePath) > 0 {
		c.Path = modfilePath
	} else {
		c.UseStdIn = true
	}

	err = c.handleOssIndexOutputFormat(outputFormat)
	if err != nil {
		return
	}

	err = c.getCVEExcludesFromFile(excludeVulnerabilityFilePath)
	if err != nil {
		return
	}

	return
}

func (c *Config) doParseIqServer(args []string) (err error) {
	iqCommand := flag.NewFlagSet("iq", flag.ExitOnError)
	iqCommand.BoolVar(&c.Info, "v", false, "Set log level to Info")
	iqCommand.BoolVar(&c.Debug, "vv", false, "Set log level to Debug")
	iqCommand.BoolVar(&c.Trace, "vvv", false, "Set log level to Trace")
	iqCommand.StringVar(&c.IQConfig.Username, "user", "admin", "Specify Nexus IQ username for request")
	iqCommand.StringVar(&c.IQConfig.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	iqCommand.StringVar(&c.IQConfig.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL/port")
	iqCommand.StringVar(&c.IQConfig.Application, "application", "", "Specify application ID for request")
	iqCommand.StringVar(&c.IQConfig.Stage, "stage", "develop", "Specify stage for application")
	iqCommand.IntVar(&c.IQConfig.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `Usage:
	go list -m all | nancy iq [options]
			
Options:
`)
		iqCommand.PrintDefaults()
		os.Exit(2)
	}

	ConfigLocation = filepath.Join(HomeDir, types.IQServerDirName, types.IQServerConfigFileName)
	err = c.loadConfigFromFile(ConfigLocation, IQServer)
	if err != nil {
		LogLady.Info("Unable to load IQ Server config from file")
	}

	ConfigLocation = filepath.Join(HomeDir, types.OssIndexDirName, types.OssIndexConfigFileName)
	err = c.loadConfigFromFile(ConfigLocation, OSSIndex)
	if err != nil {
		LogLady.Info("Unable to load OSS Index config from file")
	}

	err = iqCommand.Parse(args)
	if err != nil {
		return
	}

	return
}

func (c *Config) loadConfigFromFile(configLocation string, configType ConfigType) error {
	b, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	if configType == OSSIndex {
		err = yaml.Unmarshal(b, &c.OSSIndexConfig)
	} else {
		err = yaml.Unmarshal(b, &c.IQConfig)
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) getCVEExcludesFromFile(excludeVulnerabilityFilePath string) error {
	fi, err := os.Stat(excludeVulnerabilityFilePath)
	if (fi != nil && fi.IsDir()) || (err != nil && os.IsNotExist(err)) {
		return nil
	}
	file, err := os.Open(excludeVulnerabilityFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ogLine := scanner.Text()
		err := c.determineIfLineIsExclusion(ogLine)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *Config) determineIfLineIsExclusion(ogLine string) error {
	line := unixComments.ReplaceAllString(ogLine, "")
	until := untilComment.FindStringSubmatch(line)
	line = untilComment.ReplaceAllString(line, "")
	cveOnly := strings.TrimSpace(line)
	if len(cveOnly) > 0 {
		if until != nil {
			parseDate, err := time.Parse("2006-01-02", strings.TrimSpace(until[2]))
			if err != nil {
				return fmt.Errorf("failed to parse until at line '%s'. Expected format is 'until=yyyy-MM-dd'", ogLine)
			}
			if parseDate.After(time.Now()) {
				c.OSSIndexConfig.CveList.Cves = append(c.OSSIndexConfig.CveList.Cves, cveOnly)
			}
		} else {
			c.OSSIndexConfig.CveList.Cves = append(c.OSSIndexConfig.CveList.Cves, cveOnly)
		}
	}
	return nil
}

func (c *Config) handleOssIndexOutputFormat(outputFormat string) error {
	setTextOutputFormat := func() {
		OssIndexOutputFormats["text"].(*audit.AuditLogTextFormatter).Quiet = &c.Quiet
		OssIndexOutputFormats["text"].(*audit.AuditLogTextFormatter).NoColor = &c.NoColor
	}

	if OssIndexOutputFormats[outputFormat] != nil {
		switch outputFormat {
		case "text":
			setTextOutputFormat()
		case "csv":
			OssIndexOutputFormats[outputFormat].(*audit.CsvFormatter).Quiet = &c.Quiet
		}
		c.OSSIndexConfig.Formatter = OssIndexOutputFormats[outputFormat]
	} else {
		width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}
		fmt.Println(strings.Repeat("!", width))
		fmt.Println("!!! Output format of", strings.TrimSpace(outputFormat), "is not valid. Defaulting to text output")
		fmt.Println(strings.Repeat("!", width))
		setTextOutputFormat()
		c.OSSIndexConfig.Formatter = OssIndexOutputFormats["text"]
	}
	return nil
}

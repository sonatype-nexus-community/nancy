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

// Configuration is a struct used for loading OSS Index configuration
//
// Deprecated: please use Config instead
type Configuration struct {
	UseStdIn   bool
	Help       bool
	NoColor    bool
	Quiet      bool
	Version    bool
	CleanCache bool
	CveList    types.CveListFlag
	Path       string
	Formatter  logrus.Formatter
	Info       bool
	Debug      bool
	Trace      bool
	Username   string `yaml:"Username"`
	Token      string `yaml:"Token"`
}

// IqConfiguration is a struct used for loading Nexus IQ configuration
//
// Deprecated: please use Config instead
type IqConfiguration struct {
	Help        bool
	Version     bool
	User        string
	Token       string
	Stage       string
	Application string
	Server      string
	MaxRetries  int
	Info        bool
	Debug       bool
	Trace       bool
}

// Config is a struct for loading OSS Index or Nexus IQ Server config
type Config struct {
	Type        string
	Help        bool
	Version     bool
	Stage       string
	Application string
	Server      string
	MaxRetries  int
	Info        bool
	Debug       bool
	Trace       bool
	UseStdIn    bool
	NoColor     bool
	Quiet       bool
	CleanCache  bool
	CveList     types.CveListFlag
	Path        string
	Formatter   logrus.Formatter
	Username    string `yaml:"Username"`
	Token       string `yaml:"Token"`
}

var unixComments = regexp.MustCompile(`#.*$`)
var untilComment = regexp.MustCompile(`(until=)(.*)`)

// OssIndexOutputFormats represents the potential formats that can be used to output results
var OssIndexOutputFormats = map[string]logrus.Formatter{
	"json":        &audit.JsonFormatter{},
	"json-pretty": &audit.JsonFormatter{PrettyPrint: true},
	"text":        &audit.AuditLogTextFormatter{},
	"csv":         &audit.CsvFormatter{},
}

// Parse is used to parse command line args, and populate the Config struct with the values necessary
func (c *Config) Parse(args []string) (err error) {
	switch c.Type {
	case "ossindex":
		return c.doParseOssIndex(args)
	case "iq":
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
	flag.Var(&c.CveList, "exclude-vulnerability", "Comma separated list of CVEs to exclude")
	flag.StringVar(&c.Username, "user", "", "Specify OSS Index username for request")
	flag.StringVar(&c.Token, "token", "", "Specify OSS Index API token for request")
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

	err = c.loadConfigFromFile(ConfigLocation)
	if err != nil {
		fmt.Println(err)
		LogLady.Info("Unable to load config from file")
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
	iqCommand.StringVar(&c.Username, "user", "admin", "Specify Nexus IQ username for request")
	iqCommand.StringVar(&c.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	iqCommand.StringVar(&c.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL/port")
	iqCommand.StringVar(&c.Application, "application", "", "Specify application ID for request")
	iqCommand.StringVar(&c.Stage, "stage", "develop", "Specify stage for application")
	iqCommand.IntVar(&c.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `Usage:
	go list -m all | nancy iq [options]
			
Options:
`)
		iqCommand.PrintDefaults()
		os.Exit(2)
	}

	ConfigLocation = filepath.Join(HomeDir, types.IQServerDirName, types.IQServerConfigFileName)

	err = c.loadConfigFromFile(ConfigLocation)
	if err != nil {
		fmt.Println(err)
		LogLady.Info("Unable to load config from file")
	}

	err = iqCommand.Parse(args)
	if err != nil {
		return
	}

	return
}

func (c *Config) loadConfigFromFile(configLocation string) error {
	b, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, c)
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
				c.CveList.Cves = append(c.CveList.Cves, cveOnly)
			}
		} else {
			c.CveList.Cves = append(c.CveList.Cves, cveOnly)
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
		default:
			setTextOutputFormat()
		}
		c.Formatter = OssIndexOutputFormats[outputFormat]
	} else {
		width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}
		fmt.Println(strings.Repeat("!", width))
		fmt.Println("!!! Output format of", strings.TrimSpace(outputFormat), "is not valid. Defaulting to text output")
		fmt.Println(strings.Repeat("!", width))
		setTextOutputFormat()
		c.Formatter = OssIndexOutputFormats["text"]
	}
	return nil
}

// DEPRECATED ZONE: Only bad kids below this line

// ParseIQ is a method to parse a Nexus IQ Server configuration
//
// Deprecated: please instantiate a new Config struct and use the Parse method on it
func ParseIQ(args []string) (config IqConfiguration, err error) {
	iqCommand := flag.NewFlagSet("iq", flag.ExitOnError)
	iqCommand.BoolVar(&config.Info, "v", false, "Set log level to Info")
	iqCommand.BoolVar(&config.Debug, "vv", false, "Set log level to Debug")
	iqCommand.BoolVar(&config.Trace, "vvv", false, "Set log level to Trace")
	iqCommand.StringVar(&config.User, "user", "admin", "Specify Nexus IQ username for request")
	iqCommand.StringVar(&config.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	iqCommand.StringVar(&config.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL/port")
	iqCommand.StringVar(&config.Application, "application", "", "Specify application ID for request")
	iqCommand.StringVar(&config.Stage, "stage", "develop", "Specify stage for application")
	iqCommand.IntVar(&config.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `Usage:
	go list -m all | nancy iq [options]
			
Options:
`)
		iqCommand.PrintDefaults()
		os.Exit(2)
	}

	ConfigLocation = filepath.Join(HomeDir, types.IQServerDirName, types.IQServerConfigFileName)

	err = loadIQConfigFromFile(ConfigLocation, &config)
	if err != nil {
		fmt.Println(err)
		LogLady.Info("Unable to load config from file")
	}

	err = iqCommand.Parse(args)
	if err != nil {
		return config, err
	}

	return config, nil
}

func loadConfigFromFile(configLocation string, config *Configuration) error {
	b, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return err
	}

	return nil
}

func loadIQConfigFromFile(configLocation string, config *IqConfiguration) error {
	b, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return err
	}

	return nil
}

// Parse is a method to parse a OSS Index configuration
//
// Deprecated: please instantiate a new Config struct and use the Parse method on it
func Parse(args []string) (Configuration, error) {
	config := Configuration{}
	var excludeVulnerabilityFilePath string
	var outputFormat string

	flag.BoolVar(&config.Help, "help", false, "provides help text on how to use nancy")
	flag.BoolVar(&config.NoColor, "no-color", false, "indicate output should not be colorized")
	flag.BoolVar(&config.Quiet, "quiet", false, "indicate output should contain only packages with vulnerabilities")
	flag.BoolVar(&config.Version, "version", false, "prints current nancy version")
	flag.BoolVar(&config.CleanCache, "clean-cache", false, "Deletes local cache directory")
	flag.BoolVar(&config.Info, "v", false, "Set log level to Info")
	flag.BoolVar(&config.Debug, "vv", false, "Set log level to Debug")
	flag.BoolVar(&config.Trace, "vvv", false, "Set log level to Trace")
	flag.Var(&config.CveList, "exclude-vulnerability", "Comma separated list of CVEs to exclude")
	flag.StringVar(&config.Username, "user", "", "Specify OSS Index username for request")
	flag.StringVar(&config.Token, "token", "", "Specify OSS Index API token for request")
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

	err := loadConfigFromFile(ConfigLocation, &config)
	if err != nil {
		fmt.Println(err)
		LogLady.Info("Unable to load config from file")
	}

	err = flag.CommandLine.Parse(args)
	if err != nil {
		return config, err
	}

	modfilePath, err := getModfilePath()
	if err != nil {
		return config, err
	}
	if len(modfilePath) > 0 {
		config.Path = modfilePath
	} else {
		config.UseStdIn = true
	}

	err = handleOssIndexOutputFormat(&config, outputFormat)
	if err != nil {
		return config, err
	}

	err = getCVEExcludesFromFile(&config, excludeVulnerabilityFilePath)
	if err != nil {
		return config, err
	}

	return config, nil
}

func handleOssIndexOutputFormat(config *Configuration, outputFormat string) error {
	setTextOutputFormat := func() {
		OssIndexOutputFormats["text"].(*audit.AuditLogTextFormatter).Quiet = &config.Quiet
		OssIndexOutputFormats["text"].(*audit.AuditLogTextFormatter).NoColor = &config.NoColor
	}

	if OssIndexOutputFormats[outputFormat] != nil {
		switch outputFormat {
		case "text":
			setTextOutputFormat()
		case "csv":
			OssIndexOutputFormats[outputFormat].(*audit.CsvFormatter).Quiet = &config.Quiet
		default:
			setTextOutputFormat()
		}
		config.Formatter = OssIndexOutputFormats[outputFormat]
	} else {
		width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}
		fmt.Println(strings.Repeat("!", width))
		fmt.Println("!!! Output format of", strings.TrimSpace(outputFormat), "is not valid. Defaulting to text output")
		fmt.Println(strings.Repeat("!", width))
		setTextOutputFormat()
		config.Formatter = OssIndexOutputFormats["text"]
	}
	return nil
}

func getModfilePath() (modfilepath string, err error) {
	if flag.CommandLine.NArg() > 0 {
		nonFlagArgs := flag.CommandLine.Args()
		if len(nonFlagArgs) != 1 {
			return modfilepath, fmt.Errorf("wrong number of modfile paths: %s", nonFlagArgs)
		}
		for _, nonFlagArg := range nonFlagArgs {
			return nonFlagArg, err
		}
	}
	return modfilepath, err
}

func getCVEExcludesFromFile(config *Configuration, excludeVulnerabilityFilePath string) error {
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
		err := determineIfLineIsExclusion(ogLine, config)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func determineIfLineIsExclusion(ogLine string, config *Configuration) error {
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
				config.CveList.Cves = append(config.CveList.Cves, cveOnly)
			}
		} else {
			config.CveList.Cves = append(config.CveList.Cves, cveOnly)
		}
	}
	return nil
}

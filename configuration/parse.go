package configuration

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/types"
)

type Configuration struct {
	UseStdIn bool
	Help bool
	NoColor bool
	Quiet bool
	Version bool
	CleanCache bool
	CveList types.CveListFlag
	Path    string
	Formatter logrus.Formatter
}

var unixComments = regexp.MustCompile(`#.*$`)

func Parse(args []string) (Configuration, error) {
	config := Configuration{}
	var excludeVulnerabilityFilePath string
	var outputFormat string

	var outputFormats = map[string]logrus.Formatter{
		"json":        &audit.JsonFormatter{},
		"json-pretty": &audit.JsonFormatter{PrettyPrint: true},
		"text":        &audit.AuditLogTextFormatter{Quiet: &config.Quiet, NoColor: &config.NoColor},
		"csv":         &audit.CsvFormatter{Quiet: &config.Quiet},
	}

	flag.BoolVar(&config.Help, "help", false, "provides help text on how to use nancy")
	flag.BoolVar(&config.NoColor, "no-color", false, "indicate output should not be colorized")
	flag.BoolVar(&config.Quiet, "quiet", false, "indicate output should contain only packages with vulnerabilities")
	flag.BoolVar(&config.Version, "version", false, "prints current nancy version")
	flag.BoolVar(&config.CleanCache, "clean-cache", false, "Deletes local cache directory")
	flag.Var(&config.CveList, "exclude-vulnerability", "Comma separated list of CVEs to exclude")
	flag.StringVar(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "./.nancy-ignore", "Path to a file containing newline separated CVEs to be excluded")
	flag.StringVar(&outputFormat, "output", "text", "Styling for output format. "+fmt.Sprintf("%+q", reflect.ValueOf(outputFormats).MapKeys()))

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: \ngo list -m all | nancy [options]\nnancy [options] </path/to/Gopkg.lock>\nnancy [options] </path/to/go.sum>\n\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	// Parse config from the command line output
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return config, err
	}

	if len(flag.Args()) == 0 {
		config.UseStdIn = true
	} else {
		config.Path = args[len(args)-1]
	}

	if outputFormats[outputFormat] != nil {
		config.Formatter = outputFormats[outputFormat]
	} else {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!! Output format of", strings.TrimSpace(outputFormat), "is not valid. Defaulting to text output")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		config.Formatter = outputFormats["text"]
	}

	err = getCVEExcludesFromFile(&config, excludeVulnerabilityFilePath)
	if err != nil {
		return config, err
	}

	return config, nil
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
		line := scanner.Text()
		line = unixComments.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)

		if len(line) > 0 {
			config.CveList.Cves = append(config.CveList.Cves, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

package configuration

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/sonatype-nexus-community/nancy/types"
	"os"
	"regexp"
	"strings"
)

type Configuration struct {
	UseStdIn bool
	Help bool
	NoColor bool
	Quiet bool
	Version bool
	CveList types.CveListFlag
	Path    string
}

var unixComments = regexp.MustCompile(`#.*$`)

func Parse(args []string) (Configuration, error) {
	config := Configuration{}
	var excludeVulnerabilityFilePath string
	var noColorDeprecated bool

	flag.BoolVar(&config.Help, "help", false, "provides help text on how to use nancy")
	flag.BoolVar(&config.NoColor, "no-color", false, "indicate output should not be colorized")
	flag.BoolVar(&noColorDeprecated, "noColor", false, "indicate output should not be colorized (deprecated: please use no-color)")
	flag.BoolVar(&config.Quiet, "quiet", false, "indicate output should contain only packages with vulnerabilities")
	flag.BoolVar(&config.Version, "version", false, "prints current nancy version")
	flag.Var(&config.CveList, "exclude-vulnerability", "Comma separated list of CVEs to exclude")
	flag.StringVar(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "./.nancy-ignore", "Path to a file containing newline separated CVEs to be excluded")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: \nnancy [options] </path/to/Gopkg.lock>\nnancy [options] </path/to/go.sum>\n\nOptions:\n")
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
	}else{
		config.Path = args[len(args)-1]
	}

	if noColorDeprecated == true {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!!! DEPRECATION WARNING : Please change 'noColor' param to be 'no-color'. This one will be removed in a future release. !!!!")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		config.NoColor = noColorDeprecated
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

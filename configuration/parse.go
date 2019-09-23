package configuration

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sonatype-nexus-community/nancy/types"
	"os"
)

type Configuration struct {
	NoColor bool
	NoColorDeprecated bool
	Quiet bool
	Version bool
	CveList types.CveListFlag
	Path string
}


func Parse(args []string) (Configuration, error){
	config := Configuration{}

	flag.BoolVar(&config.NoColor, "no-color", false, "indicate output should not be colorized")
	flag.BoolVar(&config.NoColorDeprecated, "noColor", false, "indicate output should not be colorized (deprecated: please use no-color)")
	flag.BoolVar(&config.Quiet,"quiet", false, "indicate output should contain only packages with vulnerabilities")
	flag.BoolVar(&config.Version,"version", false, "prints current nancy version")
	flag.Var(&config.CveList, "exclude-vulnerability", "Comma seperated list of CVEs to exclude")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: \nnancy [options] </path/to/Gopkg.lock>\nnancy [options] </path/to/go.sum>\n\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if len(args) < 1 {
		return config, errors.New("no arguments passed")
	}

	// Parse config from the command line output
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return config, err
	}
	config.Path = args[len(args)-1]

	return config, nil
}
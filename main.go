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

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"

	"github.com/common-nighthawk/go-figure"
	"github.com/golang/dep"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/iq"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
)

func main() {
	LogLady.Info("Starting Nancy")

	var err error
	if len(os.Args) > 1 && os.Args[1] == "iq" {
		err = doIq(os.Args[2:])
	} else if len(os.Args) > 1 && os.Args[1] == "config" {
		err = doConfig(os.Stdin)
	} else {
		err = doOssi(os.Args[1:])
	}

	if err != nil {
		if exiterr, ok := err.(customerrors.ErrorExit); ok {
			os.Exit(exiterr.ExitCode)
		} else {
			// really don't expect this
			LogLady.WithError(err).Error("unexpected error in main")
			os.Exit(4)
		}
	}
}

func doOssi(ossiArgs []string) (err error) {
	LogLady.Info("Nancy parsing config for OSS Index")
	ossIndexConfig, err := configuration.Parse(ossiArgs)
	if err != nil {
		flag.Usage()
		err = customerrors.ErrorExit{Err: err, ExitCode: 1}
		return
	}
	if err = processConfig(ossIndexConfig); err != nil {
		LogLady.Info("Nancy finished parsing config for OSS Index, vulnerability found")
		return
	}
	LogLady.Info("Nancy finished parsing config for OSS Index")
	return
}

func doConfig(stdin io.Reader) (err error) {
	LogLady.Info("Nancy setting config via the command line")
	if err = configuration.GetConfigFromCommandLine(stdin); err != nil {
		err = customerrors.NewErrorExitPrintHelp(err, "Unable to set config for Nancy")
	}
	return
}

func doIq(iqArgs []string) (err error) {
	LogLady.Info("Nancy parsing config for IQ")
	config, err := configuration.ParseIQ(iqArgs)
	if err != nil {
		flag.Usage()
		err = customerrors.ErrorExit{Err: err, ExitCode: 2} // flag.Usage used to exit with code 2
		return
	}
	LogLady.WithField("config", config).Info("Obtained IQ config")
	if err = processIQConfig(config); err != nil {
		LogLady.Info("Nancy finished parsing config for IQ, vulnerability found")
		return
	}
	LogLady.Info("Nancy finished parsing config for IQ")
	return
}

func printHeader(print bool) {
	if print {
		LogLady.Info("Attempting to print header")
		figure.NewFigure("Nancy", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		LogLady.WithFields(logrus.Fields{
			"build_time":       buildversion.BuildTime,
			"build_commit":     buildversion.BuildCommit,
			"version":          buildversion.BuildVersion,
			"operating_system": runtime.GOOS,
			"architecture":     runtime.GOARCH,
		}).Info("Printing Nancy version")
		fmt.Println("Nancy version: " + buildversion.BuildVersion)
		LogLady.Info("Finished printing header")
	}
}

func processConfig(config configuration.Configuration) (err error) {
	if config.Help {
		LogLady.Info("Printing usage and exiting clean")
		flag.Usage()
		err = customerrors.ErrorExit{ExitCode: 0}
		return
	}

	if config.Version {
		LogLady.WithFields(logrus.Fields{
			"build_time":       buildversion.BuildTime,
			"build_commit":     buildversion.BuildCommit,
			"version":          buildversion.BuildVersion,
			"operating_system": runtime.GOOS,
			"architecture":     runtime.GOARCH,
		}).Info("Printing version information and exiting clean")

		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		err = customerrors.ErrorExit{ExitCode: 0}
		return
	}

	if config.Info {
		LogLady.Level = logrus.InfoLevel
	}
	if config.Debug {
		LogLady.Level = logrus.DebugLevel
	}
	if config.Trace {
		LogLady.Level = logrus.TraceLevel
	}

	if config.CleanCache {
		LogLady.Info("Attempting to clean cache")
		if err = ossindex.RemoveCacheDirectory(); err != nil {
			LogLady.WithField("error", err).Error("Error cleaning cache")
			fmt.Printf("ERROR: cleaning cache: %v\n", err)
			err = customerrors.ErrorExit{Err: err, ExitCode: 1}
			return
		}
		LogLady.Info("Cache cleaned")
		return
	}

	printHeader(!config.Quiet && reflect.TypeOf(config.Formatter).String() == "*audit.AuditLogTextFormatter")

	if config.UseStdIn {
		LogLady.Info("Parsing config for StdIn")
		if err = doStdInAndParse(config); err != nil {
			return
		}
	}
	if !config.UseStdIn {
		LogLady.Info("Parsing config for file based scan")
		if err = doCheckExistenceAndParse(config); err != nil {
			return
		}
	}
	return
}

func processIQConfig(config configuration.IqConfiguration) (err error) {
	// TODO: a lot of this code is a duplication of the OSS Index config, probably should extract some of it
	if config.Help {
		LogLady.Info("Printing usage and exiting clean")
		flag.Usage()
		err = customerrors.ErrorExit{ExitCode: 0}
		return
	}

	if config.Version {
		LogLady.WithFields(logrus.Fields{
			"build_time":   buildversion.BuildTime,
			"build_commit": buildversion.BuildCommit,
			"version":      buildversion.BuildVersion,
		}).Info("Printing version information and exiting clean")

		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		err = customerrors.ErrorExit{ExitCode: 0}
		return
	}

	if config.Info {
		LogLady.Level = logrus.InfoLevel
	}
	if config.Debug {
		LogLady.Level = logrus.DebugLevel
	}
	if config.Trace {
		LogLady.Level = logrus.TraceLevel
	}

	if config.Application == "" {
		LogLady.Info("No application specified, printing usage and exiting with error")
		flag.Usage()
		err = customerrors.NewErrorExitPrintHelp(fmt.Errorf("no IQ application id specified"), "Missing IQ application ID")
		return
	}

	printHeader(true)

	LogLady.Info("Parsing IQ config for StdIn")
	if err = doStdInAndParseForIQ(config); err != nil {
		return
	}
	return
}

func doStdInAndParse(config configuration.Configuration) (err error) {
	LogLady.Info("Beginning StdIn parse for OSS Index")
	if err = checkStdIn(); err != nil {
		return
	}
	LogLady.Info("Instantiating go.mod package")

	mod := packages.Mod{}
	scanner := bufio.NewScanner(os.Stdin)

	LogLady.Info("Beginning to parse StdIn")
	mod.ProjectList, _ = parse.GoList(scanner)
	LogLady.WithFields(logrus.Fields{
		"projectList": mod.ProjectList,
	}).Debug("Obtained project list")

	var purls = mod.ExtractPurlsFromManifest()
	LogLady.WithFields(logrus.Fields{
		"purls": purls,
	}).Debug("Extracted purls")

	LogLady.Info("Auditing purls with OSS Index")
	err = checkOSSIndex(purls, nil, config)
	return
}

func checkStdIn() (err error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		LogLady.Info("StdIn is valid")
	} else {
		LogLady.Error("StdIn is invalid, either empty or another reason")
		flag.Usage()
		err = customerrors.ErrorExit{ExitCode: 1}
		return
	}
	return
}

func doStdInAndParseForIQ(config configuration.IqConfiguration) (err error) {
	LogLady.Debug("Beginning StdIn parse for IQ")
	if err = checkStdIn(); err != nil {
		return
	}
	LogLady.Info("Instantiating go.mod package")

	mod := packages.Mod{}
	scanner := bufio.NewScanner(os.Stdin)

	LogLady.Info("Beginning to parse StdIn")
	mod.ProjectList, _ = parse.GoList(scanner)
	LogLady.WithFields(logrus.Fields{
		"projectList": mod.ProjectList,
	}).Debug("Obtained project list")

	var purls = mod.ExtractPurlsFromManifest()
	LogLady.WithFields(logrus.Fields{
		"purls": purls,
	}).Debug("Extracted purls")

	LogLady.Info("Auditing purls with IQ Server")
	err = auditWithIQServer(purls, config.Application, config)
	return
}

func doCheckExistenceAndParse(config configuration.Configuration) error {
	switch {
	case strings.Contains(config.Path, "Gopkg.lock"):
		workingDir := filepath.Dir(config.Path)
		if workingDir == "." {
			workingDir, _ = os.Getwd()
		}
		getenv := os.Getenv("GOPATH")
		ctx := dep.Ctx{
			WorkingDir: workingDir,
			GOPATHs:    []string{getenv},
		}
		project, err := ctx.LoadProject()
		if err != nil {
			return customerrors.NewErrorExitPrintHelp(err, fmt.Sprintf("could not read lock at path %s", config.Path))
		}
		if project.Lock == nil {
			return customerrors.NewErrorExitPrintHelp(errors.New("dep failed to parse lock file and returned nil"), "nancy could not continue due to dep failure")
		}

		purls, invalidPurls := packages.ExtractPurlsUsingDep(project)

		return checkOSSIndex(purls, invalidPurls, config)
	case strings.Contains(config.Path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = config.Path
		manifestExists, err := mod.CheckExistenceOfManifest()
		if err != nil {
			return err
		}
		if manifestExists {
			mod.ProjectList, _ = parse.GoSum(config.Path)
			var purls = mod.ExtractPurlsFromManifest()

			return checkOSSIndex(purls, nil, config)
		}
	default:
		return customerrors.ErrorExit{ExitCode: 3}
	}
	return nil
}

func checkOSSIndex(purls []string, invalidpurls []string, config configuration.Configuration) error {
	var packageCount = len(purls)
	coordinates, err := ossindex.AuditPackagesWithOSSIndex(purls, &config)
	if err != nil {
		return customerrors.NewErrorExitPrintHelp(err, "Error auditing packages")
	}

	var invalidCoordinates []types.Coordinate
	for _, invalidpurl := range invalidpurls {
		invalidCoordinates = append(invalidCoordinates, types.Coordinate{Coordinates: invalidpurl, InvalidSemVer: true})
	}

	if count := audit.LogResults(config.Formatter, packageCount, coordinates, invalidCoordinates, config.CveList.Cves); count > 0 {
		return customerrors.ErrorExit{ExitCode: count}
	}
	return nil
}

func auditWithIQServer(purls []string, applicationID string, config configuration.IqConfiguration) error {
	LogLady.Debug("Sending purls to be Audited by IQ Server")
	res, err := iq.AuditPackages(purls, applicationID, config)
	if err != nil {
		return customerrors.NewErrorExitPrintHelp(err, "Uh oh! There was an error with your request to Nexus IQ Server")
	}

	fmt.Println()
	if res.IsError {
		LogLady.WithField("res", res).Error("An error occurred with the request to IQ Server")
		return customerrors.NewErrorExitPrintHelp(errors.New(res.ErrorMessage), "Uh oh! There was an error with your request to Nexus IQ Server")
	}

	if res.PolicyAction != "Failure" {
		LogLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return nil
	} else {
		LogLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return customerrors.ErrorExit{ExitCode: 1}
	}
}

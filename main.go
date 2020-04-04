// Copyright 2018 Sonatype Inc.
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
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"

	figure "github.com/common-nighthawk/go-figure"
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

	if len(os.Args) > 1 && os.Args[1] == "iq" {
		LogLady.Info("Nancy parsing config for IQ")
		config, err := configuration.ParseIQ(os.Args[2:])
		if err != nil {
			flag.Usage()
			os.Exit(1)
		}
		LogLady.WithField("config", config).Info("Obtained IQ config")
		processIQConfig(config)
		LogLady.Info("Nancy finished parsing config for IQ")
	} else if len(os.Args) > 1 && os.Args[1] == "config" {
		LogLady.Info("Nancy setting config via the command line")
		configuration.GetConfigFromCommandLine()
	} else {
		LogLady.Info("Nancy parsing config for OSS Index")
		ossIndexConfig, err := configuration.Parse(os.Args[1:])
		if err != nil {
			flag.Usage()
			os.Exit(1)
		}
		processConfig(ossIndexConfig)
		LogLady.Info("Nancy finished parsing config for OSS Index")
	}
}

func printHeader(print bool) {
	if print {
		LogLady.Info("Attempting to print header")
		figure.NewFigure("Nancy", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		LogLady.WithField("version", buildversion.BuildVersion).Info("Printing Nancy version")
		fmt.Println("Nancy version: " + buildversion.BuildVersion)
		LogLady.Info("Finished printing header")
	}
}

func processConfig(config configuration.Configuration) {
	if config.Help {
		LogLady.Info("Printing usage and exiting clean")
		flag.Usage()
		os.Exit(0)
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
		os.Exit(0)
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
		if err := ossindex.RemoveCacheDirectory(); err != nil {
			LogLady.WithField("error", err).Error("Error cleaning cache")
			fmt.Printf("ERROR: cleaning cache: %v\n", err)
			os.Exit(1)
		}
		LogLady.Info("Cache cleaned")
		return
	}

	printHeader((!config.Quiet && reflect.TypeOf(config.Formatter).String() == "*audit.AuditLogTextFormatter"))

	if config.UseStdIn {
		LogLady.Info("Parsing config for StdIn")
		doStdInAndParse(config)
	}
	if !config.UseStdIn {
		LogLady.Info("Parsing config for file based scan")
		doCheckExistenceAndParse(config)
	}
}

func processIQConfig(config configuration.IqConfiguration) {
	// TODO: a lot of this code is a duplication of the OSS Index config, probably should extract some of it
	if config.Help {
		LogLady.Info("Printing usage and exiting clean")
		flag.Usage()
		os.Exit(0)
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
		os.Exit(0)
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
		LogLady.Info("No application specified, printing usage and exiting clean")
		flag.Usage()
		os.Exit(0)
	}

	printHeader(true)

	LogLady.Info("Parsing IQ config for StdIn")
	doStdInAndParseForIQ(config)
}

func doStdInAndParse(config configuration.Configuration) {
	LogLady.Info("Beginning StdIn parse for OSS Index")
	checkStdIn()
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
	checkOSSIndex(purls, nil, config)
}

func checkStdIn() {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		LogLady.Info("StdIn is valid")
	} else {
		LogLady.Error("StdIn is invalid, either empty or another reason")
		flag.Usage()
		os.Exit(1)
	}
}

func doStdInAndParseForIQ(config configuration.IqConfiguration) {
	LogLady.Debug("Beginning StdIn parse for IQ")
	checkStdIn()
	LogLady.Info("Instantiating go.mod package")

	mod := packages.Mod{}
	scanner := bufio.NewScanner(os.Stdin)

	LogLady.Info("Beginning to parse StdIn")
	mod.ProjectList, _ = parse.GoList(scanner)
	LogLady.WithFields(logrus.Fields{
		"projectList": mod.ProjectList,
	}).Debug("Obtained project list")

	var purls = mod.ExtractPurlsFromManifestForIQ()
	LogLady.WithFields(logrus.Fields{
		"purls": purls,
	}).Debug("Extracted purls")

	LogLady.Info("Auditing purls with IQ Server")
	auditWithIQServer(purls, config.Application, config)
}

func doCheckExistenceAndParse(config configuration.Configuration) {
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
			customerrors.Check(err, fmt.Sprintf("could not read lock at path %s", config.Path))
		}
		if project.Lock == nil {
			customerrors.Check(errors.New("dep failed to parse lock file and returned nil"), "nancy could not continue due to dep failure")
		}

		purls, invalidPurls := packages.ExtractPurlsUsingDep(project)

		checkOSSIndex(purls, invalidPurls, config)
	case strings.Contains(config.Path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = config.Path
		if mod.CheckExistenceOfManifest() {
			mod.ProjectList, _ = parse.GoSum(config.Path)
			var purls = mod.ExtractPurlsFromManifest()

			checkOSSIndex(purls, nil, config)
		}
	default:
		os.Exit(3)
	}
}

func checkOSSIndex(purls []string, invalidpurls []string, config configuration.Configuration) {
	var packageCount = len(purls)
	coordinates, err := ossindex.AuditPackagesWithOSSIndex(purls, &config)
	customerrors.Check(err, "Error auditing packages")

	var invalidCoordinates []types.Coordinate
	for _, invalidpurl := range invalidpurls {
		invalidCoordinates = append(invalidCoordinates, types.Coordinate{Coordinates: invalidpurl, InvalidSemVer: true})
	}

	if count := audit.LogResults(config.Formatter, packageCount, coordinates, invalidCoordinates, config.CveList.Cves); count > 0 {
		os.Exit(count)
	}
}

func auditWithIQServer(purls []string, applicationID string, config configuration.IqConfiguration) {
	LogLady.Debug("Sending purls to be Audited by IQ Server")
	res, err := iq.AuditPackages(purls, applicationID, config)
	customerrors.Check(err, "Uh oh! There was an error with your request to Nexus IQ Server")

	fmt.Println()
	if res.IsError {
		LogLady.WithField("res", res).Error("An error occurred with the request to IQ Server")
		customerrors.Check(errors.New(res.ErrorMessage), "Uh oh! There was an error with your request to Nexus IQ Server")
	}

	if res.PolicyAction != "Failure" {
		LogLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		os.Exit(0)
	} else {
		LogLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		os.Exit(1)
	}
}

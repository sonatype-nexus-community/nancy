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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/dep"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
)

var config configuration.Configuration

func main() {
	var err error
	config, err = configuration.Parse(os.Args[1:])
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	if config.Help {
		flag.Usage()
		os.Exit(0)
	}

	if config.Version {
		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		os.Exit(0)
	}

	if !config.Quiet {
		fmt.Println("Nancy version: " + buildversion.BuildVersion)
	}

	if config.UseStdIn == true {
		doStdInAndParse()
	} else {
		doCheckExistenceAndParse()
	}
}

func doStdInAndParse() {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if (fi.Mode() & os.ModeNamedPipe) == 0 {
		flag.Usage()
		os.Exit(1)
	} else {
		mod := packages.Mod{}
		scanner := bufio.NewScanner(os.Stdin)
		mod.ProjectList, _ = parse.GoList(scanner)
		var purls = mod.ExtractPurlsFromManifest()
		var packageCount = len(purls)
		checkOSSIndex(purls, packageCount)
	}
}

func doCheckExistenceAndParse() {
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
			customerrors.Check(err, fmt.Sprint("could not read lock at path "+config.Path))
		}

		purls, invalidPurls := packages.ExtractPurlsUsingDep(*project)
		if len(invalidPurls) > 0 {
			audit.LogInvalidSemVerWarning(config.NoColor, config.Quiet, invalidPurls)
		}

		var packageCount = len(purls)
		checkOSSIndex(purls, packageCount)
	case strings.Contains(config.Path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = config.Path
		if mod.CheckExistenceOfManifest() {
			mod.ProjectList, _ = parse.GoSum(config.Path)
			var purls = mod.ExtractPurlsFromManifest()
			var packageCount = len(purls)

			checkOSSIndex(purls, packageCount)
		}
	default:
		os.Exit(3)
	}
}

func checkOSSIndex(purls []string, packageCount int) {
	coordinates, err := ossindex.AuditPackages(purls)
	customerrors.Check(err, "Error auditing packages")

	if count := audit.LogResults(config.NoColor, config.Quiet, packageCount, coordinates, config.CveList.Cves); count > 0 {
		os.Exit(count)
	}
}

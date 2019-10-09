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
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/golang/dep"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
	"path/filepath"
)

var noColorPtr *bool
var quietPtr *bool
var path string
var cveList types.CveListFlag

func main() {
	args := os.Args[1:]

	noColorDepPtr := flag.Bool("noColor", false, "indicate output should not be colorized (deprecated: please use no-color)")
	noColorPtr = flag.Bool("no-color", false, "indicate output should not be colorized")
	quietPtr = flag.Bool("quiet", false, "indicate output should contain only packages with vulnerabilities")
	version := flag.Bool("version", false, "prints current nancy version")
	flag.Var(&cveList, "exclude-vulnerability", "Comma seperated list of CVEs to exclude")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: \nnancy [options] </path/to/Gopkg.lock>\nnancy [options] </path/to/go.sum>\n\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Parse flags from the command line output
	flag.Parse()

	if *noColorDepPtr == true {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!!! DEPRECATION WARNING : Please change 'noColor' param to be 'no-color'. This one will be removed in a future release. !!!!")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		noColorPtr = noColorDepPtr
	}

	if *version {
		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		os.Exit(0)
	}

	path = args[len(args)-1]

	// Currently only checks Dep, can eventually check for go mod, etc...
	doCheckExistenceAndParse()
}

func doCheckExistenceAndParse() {
	switch {
	case strings.Contains(path, "Gopkg.lock"):
		workingDir := filepath.Dir(path)
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
			customerrors.Check(err, fmt.Sprint("could not read lock at path "+path))
		}

		purls, invalidPurls := packages.ExtractPurlsUsingDep(*project)
		if len(invalidPurls) > 0 {
			audit.LogInvalidSemVerWarning(*noColorPtr, *quietPtr, invalidPurls)
		}

		var packageCount = len(purls)
		checkOSSIndex(purls, packageCount)
	case strings.Contains(path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = path
		if mod.CheckExistenceOfManifest() {
			mod.ProjectList, _ = parse.GoSum(path)
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

	if count := audit.LogResults(*noColorPtr, *quietPtr, packageCount, coordinates, cveList.Cves); count > 0 {
		os.Exit(count)
	}
}

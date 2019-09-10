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

	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
)

var noColorPtr *bool
var quietPtr *bool
var path string
var cveList types.CveListFlag

func main() {
	args := os.Args[1:]

	noColorPtr = flag.Bool("noColor", false, "indicate output should not be colorized")
	quietPtr = flag.Bool("quiet", false, "indicate output should contain only packages with vulnerabilities")
	version := flag.Bool("version", false, "prints current nancy version")
	flag.Var(&cveList, "exclude", "Comma seperated list of CVEs to exclude")

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
		dep := packages.Dep{}
		dep.GopkgPath = path
		if dep.CheckExistenceOfManifest() {
			dep.ProjectList, _ = parse.GopkgLock(path)
			var purls = processPackages(dep)
			var packageCount = len(purls)

			checkOSSIndex(purls, packageCount)
		}
	case strings.Contains(path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = path
		if mod.CheckExistenceOfManifest() {
			mod.ProjectList, _ = parse.GoSum(path)
			var purls = processPackages(mod)
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

	if count := audit.LogResults(*noColorPtr, *quietPtr, packageCount, coordinates); count > 0 {
		os.Exit(count)
	}
}

func processPackages(p packages.Packages) []string {
	return p.ExtractPurlsFromManifest()
}

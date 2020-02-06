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
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/golang/dep"
	"github.com/mbndr/figlet4go"
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
	if len(os.Args) > 1 && os.Args[1] == "iq" {
		config, err := configuration.ParseIQ(os.Args[2:])
		if err != nil {
			flag.Usage()
			os.Exit(1)
		}
		processIQConfig(config)
	} else {
		ossIndexConfig, err := configuration.Parse(os.Args[1:])
		if err != nil {
			flag.Usage()
			os.Exit(1)
		}
		processConfig(ossIndexConfig)
	}
}

func printHeader() {
	ascii := figlet4go.NewAsciiRender()
	options := figlet4go.NewRenderOptions()
	options.FontColor = []figlet4go.Color{
		figlet4go.ColorGreen,
	}
	fontsPath, _ := os.Getwd()
	fontsPath = path.Join(fontsPath, "fonts")
	ascii.LoadFont(fontsPath)
	options.FontName = "3D-ASCII"
	renderStr, _ := ascii.RenderOpts("Nancy", options)
	fmt.Print(renderStr)
	options.FontName = "Small Slant"
	renderStr, _ = ascii.RenderOpts("By Sonatype", options)
	fmt.Print(renderStr)
}

func processConfig(config configuration.Configuration) {
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

	if config.Quiet {
		log.SetOutput(ioutil.Discard)
	}

	printHeader()

	log.Println("Nancy version: " + buildversion.BuildVersion)

	if config.UseStdIn {
		doStdInAndParse(config)
	}
	if !config.UseStdIn {
		doCheckExistenceAndParse(config)
	}
}

func processIQConfig(config configuration.IqConfiguration) {
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

	if config.Application == "" {
		flag.Usage()
	}

	log.Println("Nancy version: " + buildversion.BuildVersion)

	doStdInAndParseForIQ(config)
}

func doStdInAndParse(config configuration.Configuration) {
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
		checkOSSIndex(purls, packageCount, config)
	}
}

func doStdInAndParseForIQ(config configuration.IqConfiguration) {
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
		var purls = mod.ExtractPurlsFromManifestForIQ()
		auditWithIQServer(purls, config.Application, config)
	}
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
			customerrors.Check(err, fmt.Sprint("could not read lock at path "+config.Path))
		}
		if project.Lock == nil {
			customerrors.Check(errors.New("dep failed to parse lock file and returned nil"), "nancy could not continue due to dep failure")
		}

		purls, invalidPurls := packages.ExtractPurlsUsingDep(*project)
		if len(invalidPurls) > 0 {
			audit.LogInvalidSemVerWarning(config.NoColor, config.Quiet, invalidPurls)
		}

		var packageCount = len(purls)
		checkOSSIndex(purls, packageCount, config)
	case strings.Contains(config.Path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = config.Path
		if mod.CheckExistenceOfManifest() {
			mod.ProjectList, _ = parse.GoSum(config.Path)
			var purls = mod.ExtractPurlsFromManifest()
			var packageCount = len(purls)

			checkOSSIndex(purls, packageCount, config)
		}
	default:
		os.Exit(3)
	}
}

func checkOSSIndex(purls []string, packageCount int, config configuration.Configuration) {
	coordinates, err := ossindex.AuditPackages(purls)
	customerrors.Check(err, "Error auditing packages")

	if count := audit.LogResults(config.NoColor, packageCount, coordinates, config.CveList.Cves); count > 0 {
		os.Exit(count)
	}
}

func auditWithIQServer(purls []string, applicationID string, config configuration.IqConfiguration) {
	res := iq.AuditPackages(purls, applicationID, config)

	if res.IsError {
		customerrors.Check(errors.New(res.ErrorMessage), "Uh oh! There was an error with your request to Nexus IQ Server")
	}

	if res.PolicyAction != "Failure" {
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		os.Exit(0)
	} else {
		fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		os.Exit(1)
	}
}

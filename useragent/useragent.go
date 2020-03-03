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

// Package useragent has functions for setting a user agent with helpful information
package useragent

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sonatype-nexus-community/nancy/buildversion"
)

var GOOS = runtime.GOOS
var GOARCH = runtime.GOARCH

func GetUserAgent() (useragent string) {
	useragent = fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion)
	if checkForCIEnvironment() {
		callerInfo := getCallerInfo()
		if callerInfo == "" {
			useragent = checkCIEnvironments(useragent)
			return
		}
		useragent = useragent + fmt.Sprintf(" (%s; %s %s)", callerInfo, GOOS, GOARCH)
		return
	}
	useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "non ci usage", GOOS, GOARCH)

	return
}

func checkCIEnvironments(useragent string) string {
	if checkForCISystem("CIRCLECI") {
		useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "circleci", GOOS, GOARCH)
	}
	if checkForCISystem("BITBUCKET_BUILD_NUMBER") {
		useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "bitbucket", GOOS, GOARCH)
	}
	if checkForCISystem("TRAVIS") {
		useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "travis-ci", GOOS, GOARCH)
	}
	if checkIfJenkins() {
		useragent = useragent + fmt.Sprintf(" (%s; %s %s)", "jenkins", GOOS, GOARCH)
	}
	if checkIfGitHub() {
		id := getGitHubActionID()
		useragent = useragent + fmt.Sprintf(" (%s %s; %s %s)", "github-action", id, GOOS, GOARCH)
	}
	return useragent
}

func checkForCIEnvironment() bool {
	s := os.Getenv("CI")
	if s != "" {
		return true
	}
	return checkIfJenkins() || checkIfGitHub()
}

func checkIfJenkins() bool {
	s := os.Getenv("JENKINS_HOME")
	if s != "" {
		return true
	}
	return false
}

func checkIfGitHub() bool {
	s := os.Getenv("GITHUB_ACTIONS")
	if s != "" {
		return true
	}
	return false
}

// Returns info from SC_CALLER_INFO, example: bitbucket-nancy-pipe-0.1.9
func getCallerInfo() string {
	s := os.Getenv("SC_CALLER_INFO")
	return s
}

func getGitHubActionID() string {
	s := os.Getenv("GITHUB_ACTION")
	return s
}

func checkForCISystem(system string) bool {
	s := os.Getenv(system)
	if s != "" {
		return true
	}
	return false
}

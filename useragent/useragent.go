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

// Variables that can be overriden (primarily for tests), or for consumers
var (
	GOOS       = runtime.GOOS
	GOARCH     = runtime.GOARCH
	CLIENTTOOL = "nancy-client"
)

// GetUserAgent provides a user-agent to nancy that provides info on what version of nancy
// (or upstream consumers like ahab or cheque) is running, and if the process is being run in
// CI. If so, it looks for what CI system, and other information such as SC_CALLER_INFO which
// can be used to tell if nancy is being ran inside an orb, bitbucket pipeline, etc... that
// we authored
func GetUserAgent() string {
	if checkForCIEnvironment() {
		callerInfo := getCallerInfo()
		if callerInfo == "" {
			return checkCIEnvironments()
		}
		return getCIBasedUserAgent(callerInfo)
	}
	return getCIBasedUserAgent("non ci usage")
}

func getUserAgentBaseAndVersion() string {
	return fmt.Sprintf("%s/%s", CLIENTTOOL, buildversion.BuildVersion)
}

func checkCIEnvironments() string {
	if checkForCISystem("CIRCLECI") {
		return getCIBasedUserAgent("circleci")
	}
	if checkForCISystem("BITBUCKET_BUILD_NUMBER") {
		return getCIBasedUserAgent("bitbucket")
	}
	if checkForCISystem("TRAVIS") {
		return getCIBasedUserAgent("travis-ci")
	}
	if checkForCISystem("GITLAB_CI") {
		return getCIBasedUserAgent("gitlab-ci")
	}
	if checkIfJenkins() {
		return getCIBasedUserAgent("jenkins")
	}
	if checkIfGitHub() {
		id := getGitHubActionID()
		return getCIBasedUserAgent(fmt.Sprintf("github-action %s", id))
	}

	return getCIBasedUserAgent("ci usage")
}

func getCIBasedUserAgent(agent string) string {
	return fmt.Sprintf("%s (%s; %s %s)", getUserAgentBaseAndVersion(), agent, GOOS, GOARCH)
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

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
	callTree := os.Getenv("SC_CALLER_INFO")
	if checkForCIEnvironment() {
		return checkCIEnvironments(callTree)
	}
	return getUserAgent("non ci usage", callTree)
}

func getUserAgentBaseAndVersion() string {
	return fmt.Sprintf("%s/%s", CLIENTTOOL, buildversion.BuildVersion)
}

func checkCIEnvironments(callTree string) string {
	if checkForCISystem("CIRCLECI") {
		return getUserAgent("circleci", callTree)
	}
	if checkForCISystem("BITBUCKET_BUILD_NUMBER") {
		return getUserAgent("bitbucket", callTree)
	}
	if checkForCISystem("TRAVIS") {
		return getUserAgent("travis-ci", callTree)
	}
	if checkForCISystem("GITLAB_CI") {
		return getUserAgent("gitlab-ci", callTree)
	}
	if checkIfJenkins() {
		return getUserAgent("jenkins", callTree)
	}
	if checkIfGitHub() {
		id := getGitHubActionID()
		return getUserAgent(fmt.Sprintf("github-action %s", id), callTree)
	}

	return getUserAgent("ci usage", callTree)
}

func getUserAgent(agent string, callTree string) string {
	return fmt.Sprintf("%s (%s; %s %s; %s)", getUserAgentBaseAndVersion(), agent, GOOS, GOARCH, callTree)
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

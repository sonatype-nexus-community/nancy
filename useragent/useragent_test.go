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
	"os"
	"testing"
)

func TestGetUserAgentNonCI(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (non-ci-usage 0; linux amd64; )"

	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentCircleCI(t *testing.T) {
	expected := "nancy-client/development (circleci 21; linux amd64; )"

	os.Setenv("CI", "true")
	os.Setenv("CIRCLECI", "true")
	os.Setenv("CIRCLE_BUILD_NUM", "21")

	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("CIRCLECI")
	os.Unsetenv("CIRCLE_BUILD_NUM")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentJenkins(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (jenkins 22; linux amd64; )"

	os.Setenv("JENKINS_HOME", "/a/place/under/the/sun")
	os.Setenv("BUILD_NUMBER", "22")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("JENKINS_HOME")
	os.Unsetenv("BUILD_NUMBER")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentTravisCI(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (travis-ci 23; linux amd64; )"

	os.Setenv("CI", "true")
	os.Setenv("TRAVIS", "true")
	os.Setenv("TRAVIS_BUILD_NUMBER", "23")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("TRAVIS")
	os.Unsetenv("TRAVIS_BUILD_NUMBER")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentGitLabCI(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (gitlab-ci 24; linux amd64; )"

	os.Setenv("CI", "true")
	os.Setenv("GITLAB_CI", "true")
	os.Setenv("CI_RUNNER_ID", "24") // is this the same as a "build number" in GitLab?
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("GITLAB_CI")
	os.Unsetenv("CI_RUNNER_ID")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentGitHubAction(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (github-action 25; linux amd64; )"

	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("GITHUB_ACTION", "25")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITHUB_ACTION")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentBitBucket(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (bitbucket 26; linux amd64; )"

	os.Setenv("CI", "true")
	os.Setenv("BITBUCKET_BUILD_NUMBER", "26")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("BITBUCKET_BUILD_NUMBER")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentScCallerInfo(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (ci-unknown 0; linux amd64; bitbucket-nancy-pipe-0.1.9)"

	os.Setenv("CI", "true")
	os.Setenv("SC_CALLER_INFO", "bitbucket-nancy-pipe-0.1.9")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("SC_CALLER_INFO")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

// HDS User-Agent string matching regex shown below.
// vs: Nexus_IQ_Visual_Studio/1.1.0 (.NET 4.0.30319.42000; Microsoft Windows NT 10.0.17134.0; Visual Studio 16.0)
//private static final Pattern PATTERN_IQ_IDE_USER_AGENT = Pattern
//.compile("([^/]+)[/](\\S+)\\s[(](\\S+)\\s(\\S+);\\s(\\S+)\\s(.+);\\s(.+)[)].*");
//
// The regex groups map to:
// Product/Product Version (Environment EnvironmentVersion; OS OSVersion; Other)
// We can map this to:
// CLIENTTOOL/CLIENTTOOLVersion (CIName CIBuildNumber; GOOS GOARCH; DansCrazyCallTree)
//
// where DansCrazyCallTree is:
// toolName__toolVersion___subToolName__subToolVersion___subSubToolName__subSubToolVersion
//
// double underscore "__" delimits Name/Version
// triple underscore "___" delimits currentCaller/priorCaller/priorPriorCaller
//

func TestGetUserAgentScCallerInfoTwoDeep(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (bitbucket 27; linux amd64; nancy-pipe__0.1.8___subSubToolName__subSubToolVersion)"

	os.Setenv("CI", "true")
	os.Setenv("BITBUCKET_BUILD_NUMBER", "27")
	os.Setenv("SC_CALLER_INFO", "nancy-pipe__0.1.8___subSubToolName__subSubToolVersion")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("BITBUCKET_BUILD_NUMBER")
	os.Unsetenv("SC_CALLER_INFO")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentScCallerInfoEmpty(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (bitbucket 28; linux amd64; )"

	os.Setenv("CI", "true")
	os.Setenv("BITBUCKET_BUILD_NUMBER", "28")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")
	os.Unsetenv("BITBUCKET_BUILD_NUMBER")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func TestGetUserAgentCINoClueWhatSystem(t *testing.T) {
	clearCircleCIVariables()
	expected := "nancy-client/development (ci-unknown 0; linux amd64; )"

	os.Setenv("CI", "true")
	GOOS = "linux"
	GOARCH = "amd64"

	agent := GetUserAgent()

	os.Unsetenv("CI")

	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}

func clearCircleCIVariables() {
	os.Unsetenv("CI")
	os.Unsetenv("CIRCLECI")
}

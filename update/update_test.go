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

package update

import (
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckForUpdatesPackageManagerUnknown(t *testing.T) {
	currentSemver := "0.1.2"
	check, err := CheckForUpdates("", "", currentSemver, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	expectedSemver := semver.Version{Minor: 1, Patch: 2}
	expectedCheck := &Options{Current: expectedSemver}
	assert.Equal(t, expectedCheck, check)
}

func TestCheckForUpdatesPackageManagerBrew(t *testing.T) {
	// will fail if "brew" not installed
	if _, err := findBrew(); err != nil {
		t.Skipf("brew package manager not found: %+v", err)
	}

	currentSemver := "0.1.2"
	packageManager := "homebrew"
	check, err := CheckForUpdates("", "", currentSemver, packageManager)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	expectedSemver := semver.Version{Minor: 1, Patch: 2}
	expectedCheck := &Options{Current: expectedSemver, PackageManager: packageManager}
	assert.Equal(t, expectedCheck, check)
}

func TestCheckForUpdatesPackageManagerSourceWithInvalidSlug(t *testing.T) {
	currentSemver := "0.1.2"
	packageManager := "source"
	check, err := CheckForUpdates("", "", currentSemver, packageManager)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to query the GitHub API for updates.")

	expectedSemver := semver.Version{Minor: 1, Patch: 2}
	expectedUpdater, err := selfupdate.NewUpdater(selfupdate.Config{})
	assert.Nil(t, err)
	expectedCheck := &Options{Current: expectedSemver, PackageManager: packageManager, updater: expectedUpdater}
	assert.Equal(t, expectedCheck, check)
}

func TestCheckForUpdatesPackageManagerSourceEmptyGithubAPI(t *testing.T) {
	currentSemver := "0.1.2"
	packageManager := "source"
	check, err := CheckForUpdates("", NancySlug, currentSemver, packageManager)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	expectedSemver := semver.Version{Minor: 1, Patch: 2}
	expectedCheck := &Options{Current: expectedSemver, PackageManager: packageManager, updater: check.updater, slug: NancySlug}
	assert.Equal(t, expectedCheck, check)
}

func TestCheckForUpdatesPackageManagerSource(t *testing.T) {
	currentSemver := "0.1.2"
	packageManager := "source"
	check, err := CheckForUpdates("", NancySlug, currentSemver, packageManager)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	assert.True(t, check.Found)
	assert.NotNil(t, check.Latest)
	assert.NotNil(t, check.Latest.AssetURL)
	assert.Equal(t, check.Latest.RepoName, NancyAppName)
}

func TestCheckForUpdatesPackageManagerRelease(t *testing.T) {
	currentSemver := "0.1.2"
	packageManager := "release"
	check, err := CheckForUpdates("", NancySlug, currentSemver, packageManager)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	assert.True(t, check.Found)
	assert.NotNil(t, check.Latest)
	assert.NotNil(t, check.Latest.AssetURL)
	assert.Equal(t, check.Latest.RepoName, NancyAppName)
}

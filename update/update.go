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
	"encoding/json"
	"fmt"
	"github.com/sonatype-nexus-community/nancy/settings"
	"os/exec"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const (
	//DefaultGithubEnterpriseAPI = "https://api.github.com/"
	NancyAppName = "nancy"
	NancySlug    = "sonatype-nexus-community/" + NancyAppName
)

// hoursBeforeCheck is used to configure the delay between auto-update checks
var hoursBeforeCheck = 28

// ShouldCheckForUpdates tell us if the last update check was more than a day ago
func ShouldCheckForUpdates(upd *settings.UpdateCheck) bool {
	diff := time.Since(upd.LastUpdateCheck)
	return diff.Hours() >= float64(hoursBeforeCheck)
}

// CheckForUpdates will check for updates given the proper package manager
func CheckForUpdates(githubAPI, slug, current, packageManager string) (*Options, error) {
	var (
		err   error
		check *Options
	)

	check = &Options{
		Current:        semver.MustParse(current),
		PackageManager: packageManager,
		githubAPI:      githubAPI,
		slug:           slug,
	}

	switch check.PackageManager {
	case "release":
		err = checkFromSource(check)
	case "source":
		err = checkFromSource(check)
	case "homebrew":
		err = checkFromHomebrew(check)
	}

	return check, err
}

func checkFromSource(check *Options) error {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		EnterpriseBaseURL: check.githubAPI,
	})
	if err != nil {
		return err
	}

	check.updater = updater

	err = latestRelease(check)

	return err
}

func checkFromHomebrew(check *Options) error {
	brew, err := findBrew()
	if err != nil {
		return errors.Wrap(err, "Expected to find `brew` in your $PATH but wasn't able to find it")
	}

	command := exec.Command(brew, "outdated", "--json=v2") // #nosec
	out, err := command.Output()
	if err != nil {
		return errors.Wrap(err, "failed to check for updates. `brew outdated --json=v2` returned an error")
	}

	var outdated HomebrewOutdated

	err = json.Unmarshal(out, &outdated)
	if err != nil {
		return errors.Wrap(err, "failed to parse output of `brew outdated --json=v2`")
	}

	for _, o := range outdated.Formulae {
		if o.Name == NancyAppName {
			if len(o.InstalledVersions) > 0 {
				check.Current = semver.MustParse(o.InstalledVersions[0])
			}

			check.Latest = &selfupdate.Release{
				Version: semver.MustParse(o.CurrentVersion),
			}

			// We found a release so update state of updates check
			check.Found = true
		}
	}

	return nil
}

func findBrew() (brew string, err error) {
	return exec.LookPath("brew")
}

// HomebrewOutdated wraps the JSON output from running `brew outdated --json=v2`
// We're specifically looking for this kind of structured data from the command:
//
//   {
//     "formulae": [
//       {
//         "name": "nancy",
//         "installed_versions": [
//           "0.1.1248"
//         ],
//         "current_version": "0.1.3923",
//         "pinned": false,
//         "pinned_version": null
//       }
//     ],
//     "casks": []
//   }
type HomebrewOutdated struct {
	Formulae []struct {
		Name              string   `json:"name"`
		InstalledVersions []string `json:"installed_versions"`
		CurrentVersion    string   `json:"current_version"`
		Pinned            bool     `json:"pinned"`
		PinnedVersion     string   `json:"pinned_version"`
	} `json:"formulae"`
}

// Options contains everything we need to check for or perform updates of the CLI.
type Options struct {
	Current        semver.Version
	Found          bool
	Latest         *selfupdate.Release
	PackageManager string

	updater   *selfupdate.Updater
	githubAPI string
	slug      string
}

// latestRelease will set the last known release as a member on the Options instance.
// We also update options if any releases were found or not.
func latestRelease(opts *Options) error {
	latest, found, err := opts.updater.DetectLatest(opts.slug)
	opts.Latest = latest
	opts.Found = found

	if err != nil {
		return errors.Wrap(err, `Failed to query the GitHub API for updates.

This is most likely due to GitHub rate-limiting on unauthenticated requests.

To make authenticated requests please:

  1. Generate a token at https://github.com/settings/tokens
  2. Set the token by either adding it to your ~/.gitconfig or
     setting the GITHUB_TOKEN environment variable.

Instructions for generating a token can be found at:
https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/

We call the GitHub releases API to look for new releases.
More information about that API can be found here: https://developer.github.com/v3/repos/releases/

`)
	}

	return nil
}

// IsLatestVersion will tell us if the current version is the latest version available
func IsLatestVersion(opts *Options) bool {
	if opts.Current.String() == "" || opts.Latest == nil {
		return true
	}

	return opts.Latest.Version.Equals(opts.Current)
}

// InstallLatest will execute the updater and replace the current CLI with the latest version available.
func InstallLatest(opts *Options) (string, error) {
	release, err := opts.updater.UpdateSelf(opts.Current, opts.slug)
	if err != nil {
		return "", errors.Wrap(err, "failed to install update")
	}

	return fmt.Sprintf("Updated to %s", release.Version), nil
}

// DebugVersion returns a nicely formatted string representing the state of the current version.
// Intended to be printed to standard error for developers.
func DebugVersion(opts *Options) string {
	return strings.Join([]string{
		fmt.Sprintf("Latest version: %s", opts.Latest.Version),
		fmt.Sprintf("Published: %s", opts.Latest.PublishedAt),
		fmt.Sprintf("Current Version: %s", opts.Current),
	}, "\n")
}

// ReportVersion returns a nicely formatted string representing the state of the current version.
// Intended to be printed to the user.
func ReportVersion(opts *Options) string {
	return strings.Join([]string{
		fmt.Sprintf("You are running %s", opts.Current),
		fmt.Sprintf("A new release is available (%s)", opts.Latest.Version),
	}, "\n")
}

// HowToUpdate returns a message teaching the user how to update to the latest version.
func HowToUpdate(opts *Options) string {
	switch opts.PackageManager {
	case "homebrew":
		return "You can update with `brew upgrade " + NancyAppName + "`"
	case "release":
		return "You can update with `" + NancyAppName + " update install`"
	case "source":
		return strings.Join([]string{
			"You can visit the Github releases page for the CLI to manually download and install:",
			"https://github.com/" + NancySlug + "/releases",
		}, "\n")
	}

	// Do nothing if we don't expect one of the supported package managers above
	return ""
}

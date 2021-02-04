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

package cmd

import (
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/update"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = newUpdateCommand()

func newUpdateCommand() *cobra.Command {

	updateCmd := &cobra.Command{
		Use: "update",
		//Short: "Update the tool to the latest version",
		Short: "Check if there are any updates available",
		RunE: func(_ *cobra.Command, _ []string) error {
			//return updateCLI("", true)
			return updateCLI("", false)
		},
	}

	return updateCmd
}

func updateCLI(gitHubAPI string, performUpdate bool) error {
	logAndShowMessage("Checking for updates...")
	latest, found, err := selfupdate.DetectLatest(update.NancySlug)
	if err != nil {
		return err
	}
	if !found {
		logLady.Info("did not find latest release for " + update.NancyAppName)
	} else {
		logLady.WithFields(logrus.Fields{
			"latest release": latest,
		}).Debug()
	}

	check, err := update.CheckForUpdates(gitHubAPI, update.NancySlug, buildversion.BuildVersion, buildversion.PackageManager())
	if err != nil {
		return err
	}
	logLady.WithField("check results", check).Debug("")

	if !check.Found {
		logAndShowMessage("No updates found.")
		return nil
	}

	if update.IsLatestVersion(check) {
		logAndShowMessage("Already up-to-date.")
		return nil
	}

	logLady.Debug(update.DebugVersion(check))
	logAndShowMessage(update.ReportVersion(check))

	if !performUpdate {
		logAndShowMessage(update.HowToUpdate(check))
		return nil
	}

	logAndShowMessage("Installing update...")
	message, err := update.InstallLatest(check)
	if err != nil {
		return err
	}

	logAndShowMessage(message)

	return nil
}

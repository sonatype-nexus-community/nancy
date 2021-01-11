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

	check, err := update.CheckForUpdates(gitHubAPI, update.NancySlug, getVersionNumberSemver(), buildversion.PackageManager())
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

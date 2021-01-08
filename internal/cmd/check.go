package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/settings"
	"github.com/sonatype-nexus-community/nancy/update"
	"time"
)

// For use in checking for newer release version during app startup (not during explicit command to check for update)
func checkForUpdates(gitHubAPI string) error {
	updateCheck := &settings.UpdateCheck{
		LastUpdateCheck: time.Time{},
	}

	err := updateCheck.Load()
	if err != nil {
		return err
	}
	logLady.WithFields(logrus.Fields{
		"LastUpdateCheck": updateCheck.LastUpdateCheck,
		"FileUsed":        updateCheck.FileUsed,
	}).Trace("updateCheck")

	if update.ShouldCheckForUpdates(updateCheck) {
		slug := "sonatype-community/nancy"

		logAndShowMessage("Checking for updates...")

		logLady.WithFields(logrus.Fields{
			"BuildVersion":    buildversion.BuildVersion,
			"current version": getVersionNumberSemver(),
			"PackageManager":  buildversion.PackageManager(),
		}).Debug("before CheckForUpdates")

		check, err := update.CheckForUpdates(gitHubAPI, slug, getVersionNumberSemver(), buildversion.PackageManager())

		if err != nil {
			logLady.Error("error checking for updates: " + err.Error())
			return err
		}

		if !check.Found {
			logAndShowMessage("No updates found.")

			updateCheck.LastUpdateCheck = time.Now()
			err = updateCheck.WriteToDisk()
			return err
		}

		if update.IsLatestVersion(check) {
			logAndShowMessage("Already up-to-date.")

			updateCheck.LastUpdateCheck = time.Now()
			err = updateCheck.WriteToDisk()
			return err
		}
		logLady.Debug(update.DebugVersion(check))

		logAndShowMessage(update.ReportVersion(check))
		logAndShowMessage(update.HowToUpdate(check))
		logAndShowMessage("\n") // Print a new-line after all of that

		updateCheck.LastUpdateCheck = time.Now()
		err = updateCheck.WriteToDisk()
		if err != nil {
			return err
		}
	}

	return nil
}

func getVersionNumberSemver() (currentVersion string) {
	// this value will be overridden during release, but for dev, we need a semver compliant value
	if //goland:noinspection GoBoolExpressions
	buildversion.BuildVersion == "development" {
		currentVersion = "0.0.0"
	} else {
		currentVersion = buildversion.BuildVersion
	}
	return currentVersion
}

func logAndShowMessage(message string) {
	logLady.Info(message)
	fmt.Println(message)
}

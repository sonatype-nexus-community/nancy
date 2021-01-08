package cmd

import (
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/settings"
	"github.com/sonatype-nexus-community/nancy/update"
	"time"
)

// For use in checking for updated version during app startup (not during explicit command to check for update)
func checkForUpdates(gitHubAPI string) error {
	updateCheck := &settings.UpdateCheck{
		LastUpdateCheck: time.Time{},
	}

	err := updateCheck.Load()
	if err != nil {
		return err
	}

	if update.ShouldCheckForUpdates(updateCheck) {
		slug := "sonatype-community/nancy"

		logLady.Info("Checking for updates...")

		check, err := update.CheckForUpdates(gitHubAPI, slug, getCleanVersionNumber(buildversion.BuildVersion), buildversion.PackageManager())

		if err != nil {
			logLady.Error("error checking for updates: " + err.Error())
			return err
		}

		if !check.Found {
			logLady.Info("No updates found.")

			updateCheck.LastUpdateCheck = time.Now()
			err = updateCheck.WriteToDisk()
			return err
		}

		if update.IsLatestVersion(check) {
			logLady.Info("Already up-to-date.")

			updateCheck.LastUpdateCheck = time.Now()
			err = updateCheck.WriteToDisk()
			return err
		}
		logLady.Debug(update.DebugVersion(check))

		logLady.Info(update.ReportVersion(check))
		logLady.Info(update.HowToUpdate(check))
		logLady.Info("\n") // Print a new-line after all of that

		updateCheck.LastUpdateCheck = time.Now()
		err = updateCheck.WriteToDisk()
		if err != nil {
			return err
		}
	}

	return nil
}

func getCleanVersionNumber(currentVersion string) string {
	// this value will be overridden during release, but for dev, we need a semver compliant value
	if //goland:noinspection GoBoolExpressions
	buildversion.BuildVersion == "development" {
		currentVersion = "0.0.0"
	} else {
		currentVersion = buildversion.BuildVersion
	}
	return currentVersion
}

package cmd

import (
	"fmt"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
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

	/*	updateCmd.AddCommand(&cobra.Command{
			Use:    "check",
			Hidden: true,
			Short:  "Check if there are any updates available",
			RunE: func(_ *cobra.Command, _ []string) error {
				return updateCLI("", false)
			},
		})
	*/
	/*	updateCmd.AddCommand(&cobra.Command{
			Use:    "install",
			Hidden: true,
			Short:  "Update the tool to the latest version",
			PersistentPreRun: func(_ *cobra.Command, _ []string) {
				opts.cfg.SkipUpdateCheck = true
			},
			PreRun: func(cmd *cobra.Command, args []string) {
				opts.args = args
			},
			RunE: func(_ *cobra.Command, _ []string) error {
				return updateCLI(opts)
			},
		})
	*/
	//updateCmd.PersistentFlags().BoolVar(&opts.dryRun, "check", false, "Check if there are any updates available without installing")

	return updateCmd
}

func logAndShowMessage(message string) {
	logLady.Info(message)
	fmt.Println(message)
}

func updateCLI(gitHubAPI string, performUpdate bool) error {
	logAndShowMessage("Checking for updates...")
	latest, found, err := selfupdate.DetectLatest(update.NancySlug)
	if err != nil {
		return err
	}
	if !found {
		logAndShowMessage("did not find latest release for " + update.NancyAppName)
	} else {
		logAndShowMessage(fmt.Sprintf("latest release for %s: %v", update.NancyAppName, latest))
	}

	check, err := update.CheckForUpdates(gitHubAPI, update.NancySlug, getCleanVersionNumber(buildversion.BuildVersion), buildversion.PackageManager())
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

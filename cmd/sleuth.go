package cmd

import (
	"fmt"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(sleuthCmd)

	sleuthCmd.Flags().BoolVarP(&configOssi.NoColor, "no-color", "n", false, "indicate output should not be colorized")
	sleuthCmd.Flags().VarP(&configOssi.CveList, "exclude-vulnerability", "e", "Comma separated list of CVEs to exclude")
	sleuthCmd.Flags().StringVarP(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "x", defaultExcludeFilePath, "Path to a file containing newline separated CVEs to be excluded")
	sleuthCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Styling for output format. json, json-pretty, text, csv")
}

var sleuthCmd = &cobra.Command{
	Use:     "sleuth",
	Example: `  go list -json -m all | nancy sleuth --` + flagNameOssiUsername + ` your_user --` + flagNameOssiToken + ` your_token`,
	Short:   "Check for vulnerabilities in your Golang dependencies using Sonatype's OSS Index",
	Long:    `'nancy sleuth' is a command to check for vulnerabilities in your Golang dependencies, powered by the 'Sonatype OSS Index'.`,
	PreRun:  func(cmd *cobra.Command, args []string) { bindViper(rootCmd) },
	RunE:    doOSSI,
}

//noinspection GoUnusedParameter
func doOSSI(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			err = customerrors.ErrorShowLogPath{Err: err}
		}
	}()

	logLady = logger.GetLogger("", configOssi.LogLevel)
	logLady.Info("Nancy parsing config for OSS Index")

	err = processConfig()
	if err != nil {
		if errExit, ok := err.(customerrors.ErrorExit); ok {
			logLady.Info(fmt.Sprintf("Nancy finished parsing config for OSS Index, vulnerability found. exit code: %d", errExit.ExitCode))
			os.Exit(errExit.ExitCode)
		} else {
			logLady.WithError(err).Error("unexpected error in root cmd")
			panic(err)
		}
	}

	logLady.Info("Nancy finished parsing config for OSS Index")
	return
}

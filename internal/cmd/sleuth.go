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
	"fmt"
	"os"

	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sleuthCmd)

	sleuthCmd.Flags().BoolVarP(&configOssi.NoColor, "no-color", "n", false, "indicate output should not be colorized")
	sleuthCmd.Flags().VarP(&configOssi.CveList, "exclude-vulnerability", "e", "Comma separated list of CVEs or OSS Index IDs to exclude")
	sleuthCmd.Flags().StringVarP(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "x", defaultExcludeFilePath, "Path to a file containing newline separated CVEs or OSS Index IDs to be excluded")
	sleuthCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Styling for output format. json, json-pretty, text, csv, sarif")
}

var sleuthCmd = &cobra.Command{
	Use: "sleuth",
	Example: `  go list -json -m all | nancy sleuth --` + flagNameOssiUsername + ` your_user --` + flagNameOssiToken + ` your_token
  nancy sleuth -p Gopkg.lock --` + flagNameOssiUsername + ` your_user --` + flagNameOssiToken + ` your_token`,
	Short:  "Check for vulnerabilities in your Golang dependencies using Sonatype's OSS Index",
	Long:   `'nancy sleuth' is a command to check for vulnerabilities in your Golang dependencies, powered by the 'Sonatype OSS Index'.`,
	PreRun: func(cmd *cobra.Command, args []string) { bindViperRootCmd() },
	RunE:   doOSSI,
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

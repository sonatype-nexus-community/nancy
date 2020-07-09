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
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configIQ types.Configuration
	iqServer *iq.Server
)

// iqCmd represents the iq command
var iqCmd = &cobra.Command{
	Use:   "iq",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("pkg: %v", r)
				}

				logger.PrintErrorAndLogLocation(err)
			}
		}()

		logLady = logger.GetLogger("", configOssi.LogLevel)

		if err = checkStdIn(); err != nil {
			panic(err)
		}

		mod := packages.Mod{}

		mod.ProjectList, err = parse.GoListAgnostic(os.Stdin)
		if err != nil {
			panic(err)
		}

		var purls = mod.ExtractPurlsFromManifest()

		err = auditWithIQServer(purls, configIQ.Application)
		if err != nil {
			panic(err)
		}

		return
	},
}

func init() {
	cobra.OnInitialize(initIQConfig)

	iqCmd.Flags().StringVarP(&configIQ.User, "user", "u", "", "Specify Nexus IQ username for request")
	iqCmd.Flags().StringVarP(&configIQ.Token, "token", "t", "", "Specify Nexus IQ token for request")
	iqCmd.Flags().StringVarP(&configIQ.Stage, "stage", "s", "", "Specify Nexus IQ stage for request")
	iqCmd.Flags().StringVarP(&configIQ.Application, "application", "a", "", "Specify Nexus IQ public application ID for request")
	iqCmd.Flags().StringVarP(&configIQ.Server, "host", "x", "", "Specify Nexus IQ server url for request")

	// Bind viper to the flags passed in via the command line, so it will override config from file
	viper.BindPFlag("username", iqCmd.Flags().Lookup("user"))
	viper.BindPFlag("token", iqCmd.Flags().Lookup("token"))
	viper.BindPFlag("server", iqCmd.Flags().Lookup("host"))

	rootCmd.AddCommand(iqCmd)
}

func initIQConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		configPath := path.Join(home, types.IQServerDirName)

		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName(types.IQServerConfigFileName)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// TODO: Add log statements for config
	}
}

func auditWithIQServer(purls []string, applicationID string) error {
	iqServer = iq.New(logLady, iq.Options{
		User:        configIQ.User,
		Token:       configIQ.Token,
		Application: configIQ.Application,
		Stage:       configIQ.Stage,
		Server:      configIQ.Server,
		Tool:        "nancy-client",
		DBCacheName: "nancy-cache",
		MaxRetries:  300,
	})

	logLady.Debug("Sending purls to be Audited by IQ Server")
	res, err := iqServer.AuditPackages(purls, applicationID)
	if err != nil {
		return customerrors.NewErrorExitPrintHelp(err, "Uh oh! There was an error with your request to Nexus IQ Server")
	}

	fmt.Println()
	if res.IsError {
		logLady.WithField("res", res).Error("An error occurred with the request to IQ Server")
		return customerrors.NewErrorExitPrintHelp(errors.New(res.ErrorMessage), "Uh oh! There was an error with your request to Nexus IQ Server")
	}

	if res.PolicyAction != "Failure" {
		logLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return nil
	} else {
		logLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return customerrors.ErrorExit{ExitCode: 1}
	}
}

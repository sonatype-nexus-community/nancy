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

type iqServerFactory interface {
	create() iq.IServer
}

type iqFactory struct{}

func (iqFactory) create() iq.IServer {
	iqServer := iq.New(logLady, iq.Options{
		User:        configIQ.IQUsername,
		Token:       configIQ.IQToken,
		Application: configIQ.IQApplication,
		Stage:       configIQ.IQStage,
		Server:      configIQ.IQServer,
		Tool:        "nancy-client",
		DBCacheName: "nancy-cache",
		MaxRetries:  300,
	})
	return iqServer
}

var (
	configIQ  types.Configuration
	iqCreator iqServerFactory = iqFactory{}
)

var iqCmd = &cobra.Command{
	Use:     "iq",
	Example: `  go list -m -json all | nancy iq --iqapplication your_public_application_id --iqserver http://your_iq_server_url:port --iqusername your_user --iqtoken your_token --iqstage develop`,
	Short:   "Check for vulnerabilities in your Golang dependencies using 'Sonatype's Nexus IQ IQServer'",
	Long:    `'nancy iq' is a command to check for vulnerabilities in your Golang dependencies, powered by 'Sonatype's Nexus IQ IQServer', allowing you a smooth experience as a Golang developer, using the best tools in the market!`,
	PreRun:  func(cmd *cobra.Command, args []string) { bindViperIq(cmd) },
	RunE:    doIQ,
}

//noinspection GoUnusedParameter
func doIQ(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	printHeader(!configOssi.Quiet)

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

	err = auditWithIQServer(purls, configIQ.IQApplication)
	if err != nil {
		panic(err)
	}

	return
}

func init() {
	cobra.OnInitialize(initIQConfig)

	iqCmd.Flags().StringVarP(&configIQ.IQUsername, "iqusername", "u", "admin", "Specify Nexus IQ username for request")
	iqCmd.Flags().StringVarP(&configIQ.IQToken, "iqtoken", "t", "admin123", "Specify Nexus IQ token for request")
	iqCmd.Flags().StringVarP(&configIQ.IQStage, "iqstage", "s", "develop", "Specify Nexus IQ stage for request")

	iqCmd.Flags().StringVarP(&configIQ.IQApplication, "iqapplication", "a", "", "Specify Nexus IQ public application ID for request")
	if err := iqCmd.MarkFlagRequired("iqapplication"); err != nil {
		panic(err)
	}

	iqCmd.Flags().StringVarP(&configIQ.IQServer, "iqserver-url", "x", "http://localhost:8070", "Specify Nexus IQ server url for request")

	rootCmd.AddCommand(iqCmd)
}

func bindViperIq(cmd *cobra.Command) {
	// need to defer bind call until command is run. see: https://github.com/spf13/viper/issues/233

	// Bind viper to the flags passed in via the command line, so it will override config from file
	_ = viper.BindPFlag("iqusername", cmd.Flags().Lookup("iqusername"))
	_ = viper.BindPFlag("iqtoken", cmd.Flags().Lookup("iqtoken"))
	_ = viper.BindPFlag("iqserver", cmd.Flags().Lookup("iqserver"))
}

func initIQConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		viper.SetConfigType(configTypeYaml)
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

	if err := viper.ReadInConfig(); err == nil {
		// TODO: Add log statements for config
		fmt.Printf("Todo: Add log statement for IQ config\n")
	}
}

func auditWithIQServer(purls []string, applicationID string) error {
	iqServer := iqCreator.create()

	logLady.Debug("Sending purls to be Audited by IQ Server")
	res, err := iqServer.AuditPackages(purls, applicationID)
	if err != nil {
		return customerrors.ErrorExit{ExitCode: 3, Err: err}
	}

	fmt.Println()
	if res.IsError {
		logLady.WithField("res", res).Error("An error occurred with the request to IQ Server")
		return customerrors.ErrorExit{ExitCode: 3, Err: errors.New(res.ErrorMessage)}
	}

	if res.PolicyAction != "Failure" {
		logLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return nil
	}
	logLady.WithField("res", res).Debug("Successful in communicating with IQ Server")
	fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
	fmt.Println("Report URL: ", res.ReportHTMLURL)
	return customerrors.ErrorExit{ExitCode: 1}
}

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
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/logger"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/sonatype-nexus-community/go-sona-types/iq"
	"github.com/sonatype-nexus-community/nancy/customerrors"
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
		User:          viper.GetString(configuration.YamlKeyIQUsername),
		Token:         viper.GetString(configuration.YamlKeyIQToken),
		Application:   configIQ.IQApplication,
		Stage:         configIQ.IQStage,
		Server:        viper.GetString(configuration.YamlKeyIQServer),
		OSSIndexUser:  viper.GetString(configuration.YamlKeyUsername),
		OSSIndexToken: viper.GetString(configuration.YamlKeyToken),
		Tool:          "nancy-client",
		DBCacheName:   "nancy-cache",
		MaxRetries:    300,
	})

	logLady.WithField("iqServer", iq.Options{
		User:          cleanUserName(iqServer.Options.User),
		Token:         "***hidden***",
		Application:   iqServer.Options.Application,
		Stage:         iqServer.Options.Application,
		OSSIndexUser:  cleanUserName(iqServer.Options.OSSIndexUser),
		OSSIndexToken: "***hidden***",
		Tool:          iqServer.Options.Tool,
		DBCacheName:   iqServer.Options.DBCacheName,
		MaxRetries:    iqServer.Options.MaxRetries,
	}).Debug("Created iqServer server")

	return iqServer
}

var (
	cfgFileIQ string
	configIQ  types.Configuration
	iqCreator iqServerFactory = iqFactory{}
)

var iqCmd = &cobra.Command{
	Use:     "iq",
	Example: `  go list -m -json all | nancy iq --` + flagNameIqApplication + ` your_public_application_id --` + flagNameIqServerUrl + ` http://your_iq_server_url:port --` + flagNameIqUsername + ` your_user --` + flagNameIqToken + ` your_token --` + flagNameIqStage + ` develop`,
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
			err = customerrors.ErrorShowLogPath{Err: err}
		}
	}()

	logLady = logger.GetLogger("", configOssi.LogLevel)
	logLady.Info("Nancy parsing config for IQ")

	printHeader(!configOssi.Quiet)

	if err = checkStdIn(); err != nil {
		logLady.WithError(err).Error("unexpected error in iq cmd")
		panic(err)
	}

	mod := packages.Mod{}

	mod.ProjectList, err = parse.GoListAgnostic(os.Stdin)
	if err != nil {
		logLady.WithError(err).Error("unexpected error in iq cmd")
		panic(err)
	}

	var purls = mod.ExtractPurlsFromManifest()

	err = auditWithIQServer(purls, configIQ.IQApplication)
	if err != nil {
		if errExit, ok := err.(customerrors.ErrorExit); ok {
			os.Exit(errExit.ExitCode)
		} else {
			logLady.WithError(err).Error("unexpected error in iq cmd")
			panic(err)
		}
	}

	return
}

const (
	flagNameIqUsername    = "iq-username"
	flagNameIqToken       = "iq-token"
	flagNameIqStage       = "iq-stage"
	flagNameIqApplication = "iq-application"
	flagNameIqServerUrl   = "iq-server-url"
)

func init() {
	cobra.OnInitialize(initIQConfig)

	iqCmd.Flags().StringVarP(&configIQ.IQUsername, flagNameIqUsername, "l", "admin", "Specify Nexus IQ username for request")
	iqCmd.Flags().StringVarP(&configIQ.IQToken, flagNameIqToken, "p", "admin123", "Specify Nexus IQ token for request")
	iqCmd.Flags().StringVarP(&configIQ.IQStage, flagNameIqStage, "s", "develop", "Specify Nexus IQ stage for request")

	iqCmd.Flags().StringVarP(&configIQ.IQApplication, flagNameIqApplication, "a", "", "Specify Nexus IQ public application ID for request")
	if err := iqCmd.MarkFlagRequired(flagNameIqApplication); err != nil {
		panic(err)
	}

	iqCmd.Flags().StringVarP(&configIQ.IQServer, flagNameIqServerUrl, "x", "http://localhost:8070", "Specify Nexus IQ server url for request")

	rootCmd.AddCommand(iqCmd)
}

func bindViperIq(cmd *cobra.Command) {
	// need to defer bind call until command is run. see: https://github.com/spf13/viper/issues/233

	// need to ensure ossi CLI flags will override ossi config file values when running IQ command
	bindViper(cmd)

	// Bind viper to the flags passed in via the command line, so it will override config from file
	if err := viper.BindPFlag(configuration.YamlKeyIQUsername, cmd.Flags().Lookup(flagNameIqUsername)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(configuration.YamlKeyIQToken, cmd.Flags().Lookup(flagNameIqToken)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(configuration.YamlKeyIQServer, cmd.Flags().Lookup(flagNameIqServerUrl)); err != nil {
		panic(err)
	}
}

func initIQConfig() {
	var cfgFileToCheck string
	if cfgFileIQ != "" {
		viper.SetConfigFile(cfgFileIQ)
		viper.SetConfigType(configTypeYaml)
		cfgFileToCheck = cfgFileIQ
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		configPath := path.Join(home, types.IQServerDirName)

		viper.AddConfigPath(configPath)
		viper.SetConfigType(configTypeYaml)
		viper.SetConfigName(types.IQServerConfigFileName)

		cfgFileToCheck = path.Join(configPath, types.IQServerConfigFileName)
	}

	if fileExists(cfgFileToCheck) {
		// 'merge' IQ config here, since we also need OSSI config, and load order is not guaranteed
		if err := viper.MergeInConfig(); err != nil {
			panic(err)
		}
	}
}

func auditWithIQServer(purls []string, applicationID string) error {
	iqServer := iqCreator.create()

	logLady.Debug("Sending purls to be Audited by IQ Server")
	// go-sona-types library now takes care of querying both ossi and iq with reformatted purls as needed (to v or not to v).
	res, err := iqServer.AuditPackages(purls, applicationID)
	if err != nil {
		return err
	}

	fmt.Println()
	if res.IsError {
		logLady.WithField("res", res).Error("An error occurred with the request to IQ Server")
		return errors.New(res.ErrorMessage)
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

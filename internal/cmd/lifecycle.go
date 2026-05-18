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
	"io"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	internaliq "github.com/sonatype-nexus-community/nancy/internal/iq"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
	localossindex "github.com/sonatype-nexus-community/nancy/internal/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// Viper keys for Lifecycle/IQ configuration — matches config file YAML tags.
	viperKeyLifecycleServer   = "iq.Server"
	viperKeyLifecycleUsername = "iq.Username"
	viperKeyLifecycleToken    = "iq.Token"
)

const (
	unexpectedLifecycleErr = "unexpected error in lifecycle cmd"
	reportURLLabel         = "Report URL: "
)

type lifecycleCreator interface {
	create() internaliq.IServer
}

type iqFactory struct{}

func (iqFactory) create() internaliq.IServer {
	server, err := internaliq.New(logLady, internaliq.Options{
		User:        viper.GetString(viperKeyLifecycleUsername),
		Token:       viper.GetString(viperKeyLifecycleToken),
		Application: configIQ.IQApplication,
		Stage:       configIQ.IQStage,
		Server:      viper.GetString(viperKeyLifecycleServer),
		MaxRetries:  300,
	})
	if err != nil {
		logLady.WithError(err).Error("unexpected error creating lifecycle server")
		panic(err)
	}
	return server
}

var (
	cfgFileIQ string
	configIQ  types.Configuration
	iqCreator lifecycleCreator = iqFactory{}
)

const (
	flagNameLifecycleUsername    = "lifecycle-username"
	flagNameLifecycleToken       = "lifecycle-token"
	flagNameLifecycleStage       = "lifecycle-stage"
	flagNameLifecycleApplication = "lifecycle-application"
	flagNameLifecycleServerUrl   = "lifecycle-server-url"

	// Deprecated aliases (v2 only — removed in v3)
	flagNameIqUsername    = "iq-username"
	flagNameIqToken       = "iq-token"
	flagNameIqStage       = "iq-stage"
	flagNameIqApplication = "iq-application"
	flagNameIqServerUrl   = "iq-server-url"
)

var lifecycleCmd = &cobra.Command{
	Use: "lifecycle",
	Example: `  go list -json -deps ./... | nancy lifecycle --` + flagNameLifecycleApplication + ` your_public_application_id --` + flagNameLifecycleServerUrl + ` http://your_lifecycle_url:port --` + flagNameLifecycleUsername + ` your_user --` + flagNameLifecycleToken + ` your_token --` + flagNameLifecycleStage + ` develop`,
	Short: "Check for vulnerabilities in your Golang dependencies using Sonatype Lifecycle",
	Long:  `'nancy lifecycle' checks for vulnerabilities in your Golang dependencies, powered by Sonatype Lifecycle (previously Nexus IQ Server).`,
	PreRun: func(cmd *cobra.Command, args []string) {
		migrateIqFlagsToLifecycle(cmd)
		bindViperLifecycle(cmd)
	},
	RunE: doLifecycle,
}

// migrateIqFlagsToLifecycle copies --iq-* flag values to --lifecycle-* if the lifecycle flag was not explicitly set.
func migrateIqFlagsToLifecycle(cmd *cobra.Command) {
	migrations := map[string]string{
		flagNameIqUsername:    flagNameLifecycleUsername,
		flagNameIqToken:       flagNameLifecycleToken,
		flagNameIqStage:       flagNameLifecycleStage,
		flagNameIqApplication: flagNameLifecycleApplication,
		flagNameIqServerUrl:   flagNameLifecycleServerUrl,
	}
	for iqFlag, lcFlag := range migrations {
		iqF := cmd.Flags().Lookup(iqFlag)
		lcF := cmd.Flags().Lookup(lcFlag)
		if iqF != nil && lcF != nil && iqF.Changed && !lcF.Changed {
			_ = lcF.Value.Set(iqF.Value.String())
		}
	}
}

// noinspection GoUnusedParameter
func doLifecycle(cmd *cobra.Command, args []string) (err error) {
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
	logLady.Info("Nancy parsing config for Lifecycle")

	err = checkForUpdates("", true)
	if err != nil {
		return
	}

	printHeader(!configOssi.Quiet)

	var purls []string
	purls, err = getPurls()

	err = auditWithLifecycleServer(purls)
	if err != nil {
		if errExit, ok := err.(customerrors.ErrorExit); ok {
			os.Exit(errExit.ExitCode)
		} else {
			logLady.WithError(err).Error(unexpectedLifecycleErr)
			panic(err)
		}
	}

	return
}

func getPurls() (purls []string, err error) {
	var reader io.Reader
	if stdinHasData() {
		reader = os.Stdin
	} else {
		logLady.Info("No stdin detected; auto-running go list")
		reader, err = autoRunGoList()
		if err != nil {
			logLady.WithError(err).Error(unexpectedLifecycleErr)
			panic(err)
		}
	}

	mod := packages.Mod{}
	mod.ProjectList, err = parse.GoListAgnostic(reader)
	if err != nil {
		logLady.WithError(err).Error(unexpectedLifecycleErr)
		panic(err)
	}
	purls = mod.ExtractPurlsFromManifest()
	return purls, err
}

func init() {
	cobra.OnInitialize(initLifecycleConfig)

	flags := lifecycleCmd.Flags()

	// Primary lifecycle flags
	flags.StringVarP(&configIQ.IQUsername, flagNameLifecycleUsername, "l", "admin", "Specify Lifecycle username for request")
	flags.StringVarP(&configIQ.IQToken, flagNameLifecycleToken, "k", "admin123", "Specify Lifecycle token for request")
	flags.StringVarP(&configIQ.IQStage, flagNameLifecycleStage, "s", "develop", "Specify Lifecycle stage for request")
	flags.StringVarP(&configIQ.IQApplication, flagNameLifecycleApplication, "a", "", "Specify Lifecycle public application ID for request")
	if err := lifecycleCmd.MarkFlagRequired(flagNameLifecycleApplication); err != nil {
		panic(err)
	}
	flags.StringVarP(&configIQ.IQServer, flagNameLifecycleServerUrl, "x", "http://localhost:8070", "Specify Lifecycle server url for request")
	flags.BoolVar(&configOssi.NoFail, "no-fail", false,
		"Exit 0 even when vulnerabilities are found (useful for informational CI steps)")

	// Deprecated --iq-* aliases (v2 — removed in v3)
	flags.StringP(flagNameIqUsername, "", "admin", "Deprecated: use --lifecycle-username")
	flags.StringP(flagNameIqToken, "", "admin123", "Deprecated: use --lifecycle-token")
	flags.StringP(flagNameIqStage, "", "develop", "Deprecated: use --lifecycle-stage")
	flags.StringP(flagNameIqApplication, "", "", "Deprecated: use --lifecycle-application")
	flags.StringP(flagNameIqServerUrl, "", "http://localhost:8070", "Deprecated: use --lifecycle-server-url")

	_ = flags.MarkDeprecated(flagNameIqUsername, "use --lifecycle-username instead (will be removed in v3)")
	_ = flags.MarkDeprecated(flagNameIqToken, "use --lifecycle-token instead (will be removed in v3)")
	_ = flags.MarkDeprecated(flagNameIqStage, "use --lifecycle-stage instead (will be removed in v3)")
	_ = flags.MarkDeprecated(flagNameIqApplication, "use --lifecycle-application instead (will be removed in v3)")
	_ = flags.MarkDeprecated(flagNameIqServerUrl, "use --lifecycle-server-url instead (will be removed in v3)")

	rootCmd.AddCommand(lifecycleCmd)
}

func bindViperLifecycle(cmd *cobra.Command) {
	bindViperRootCmd()

	if err := viper.BindPFlag(viperKeyLifecycleUsername, lookupFlagNotNil(flagNameLifecycleUsername, cmd)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(viperKeyLifecycleToken, lookupFlagNotNil(flagNameLifecycleToken, cmd)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(viperKeyLifecycleServer, lookupFlagNotNil(flagNameLifecycleServerUrl, cmd)); err != nil {
		panic(err)
	}
}

func lookupFlagNotNil(flagName string, cmd *cobra.Command) *pflag.Flag {
	foundFlag := cmd.Flags().Lookup(flagName)
	if foundFlag == nil {
		panic(fmt.Errorf("flag lookup for name: '%s' returned nil", flagName))
	}
	return foundFlag
}

func initLifecycleConfig() {
	viper.SetConfigType("yaml")
	var cfgFileToCheck string
	if cfgFileIQ != "" {
		viper.SetConfigFile(cfgFileIQ)
		cfgFileToCheck = cfgFileIQ
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(localossindex.GetIQServerDirectory(home))
		viper.SetConfigName(localossindex.IQServerConfigFileName)
		cfgFileToCheck = localossindex.GetIQServerConfigFile(home)
	}

	if fileExists(cfgFileToCheck) {
		if err := viper.MergeInConfig(); err != nil {
			panic(err)
		}
	}
}

func auditWithLifecycleServer(purls []string) error {
	iqServer := iqCreator.create()

	logLady.Debug("Sending purls to be Audited by Lifecycle")
	res, err := iqServer.AuditPackages(purls)
	if err != nil {
		return err
	}

	fmt.Println()
	if res.IsError {
		logLady.WithField("res", res).Error("An error occurred with the request to Lifecycle")
		return errors.New(res.ErrorMessage)
	}

	logLady.WithField("res", res).Debug("Successful in communicating with Lifecycle")
	showPolicyActionMessage(res, os.Stdout)
	switch res.PolicyAction {
	case internaliq.PolicyActionFailure:
		return customerrors.ErrorExit{ExitCode: 1}
	}
	return nil
}

func showPolicyActionMessage(res internaliq.StatusURLResult, writer io.Writer) {
	switch res.PolicyAction {
	case internaliq.PolicyActionFailure:
		_, _ = fmt.Fprintln(writer, "Hi, Nancy here, you have some policy violations to clean up!")
		_, _ = fmt.Fprintln(writer, reportURLLabel, res.AbsoluteReportHTMLURL)
	case internaliq.PolicyActionWarning:
		_, _ = fmt.Fprintln(writer, "Read, read, read. That's all I can say. There are policy warnings to investigate!")
		_, _ = fmt.Fprintln(writer, reportURLLabel, res.AbsoluteReportHTMLURL)
	default:
		_, _ = fmt.Fprintln(writer, "Wonderbar! No policy violations reported for this audit!")
		_, _ = fmt.Fprintln(writer, reportURLLabel, res.AbsoluteReportHTMLURL)
	}
}

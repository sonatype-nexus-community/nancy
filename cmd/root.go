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
	"bufio"
	"flag"
	"fmt"
	"path"
	"github.com/common-nighthawk/go-figure"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/cobra"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

var (
	cfgFile string
	configOssi configuration.Configuration
	excludeVulnerabilityFilePath string
	outputFormat string
)

var outputFormats = map[string]logrus.Formatter{
	"json":        &audit.JsonFormatter{},
	"json-pretty": &audit.JsonFormatter{PrettyPrint: true},
	"text":        &audit.AuditLogTextFormatter{Quiet: &configOssi.Quiet, NoColor: &configOssi.NoColor},
	"csv":         &audit.CsvFormatter{Quiet: &configOssi.Quiet},
}

var rootCmd = &cobra.Command{
	Use:   "nancy",
	Short: "Check for vulnerabilities in your Golang dependencies",
	Long: `nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by the 'Sonatype OSS Index', and as well, works with Nexus IQ Server, allowing you
a smooth experience as a Golang developer, using the best tools in the market!`,
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

		err = processConfig()
		if err != nil {
			panic(err)
		}

		return
	},
}

func Execute() (err error) {
	if err = rootCmd.Execute(); err != nil {
		return
	}
	return
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().CountVarP(&configOssi.LogLevel, "", "v", "Set log level, multiple v's is more verbose")
	rootCmd.PersistentFlags().BoolVarP(&configOssi.Quiet, "quiet", "q", false, "indicate output should contain only packages with vulnerabilities")
	rootCmd.Flags().BoolVarP(&configOssi.NoColor, "no-color", "n", false, "indicate output should not be colorized")
	rootCmd.Flags().BoolVarP(&configOssi.CleanCache, "clean-cache", "c", false, "Deletes local cache directory")
	rootCmd.Flags().VarP(&configOssi.CveList, "exclude-vulnerability", "e", "Comma separated list of CVEs to exclude")
	rootCmd.Flags().StringVarP(&configOssi.Username, "username", "u", "", "Specify OSS Index username for request")
	rootCmd.Flags().StringVarP(&configOssi.Token, "token", "t", "", "Specify OSS Index API token for request")
	rootCmd.Flags().StringVarP(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "x", "./.nancy-ignore", "Path to a file containing newline separated CVEs to be excluded")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Styling for output format. "+fmt.Sprintf("%+q", reflect.ValueOf(outputFormats).MapKeys()))

	// Bind viper to the flags passed in via the command line, so it will override config from file
	viper.BindPFlag("username", rootCmd.Flags().Lookup("username"))
	viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		configPath := path.Join(home, types.OssIndexDirName)

		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName(types.OssIndexConfigFileName)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// TODO: Add log statements for config
	}
}

func processConfig() (err error) {
	if outputFormats[outputFormat] != nil {
		configOssi.Formatter = outputFormats[outputFormat]
	} else {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!! Output format of", strings.TrimSpace(outputFormat), "is not valid. Defaulting to text output")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		configOssi.Formatter = outputFormats["text"]
	}

	// @todo Change to use a switch statement
	if configOssi.LogLevel == 1 || configOssi.Info {
		// LogLady.Level = logrus.InfoLevel
	}
	if configOssi.LogLevel == 2 || configOssi.Debug {
		// LogLady.Level = logrus.DebugLevel
	}
	if configOssi.LogLevel == 3 || configOssi.Trace {
		// LogLady.Level = logrus.TraceLevel
	}

	if configOssi.CleanCache {
		if err := ossindex.RemoveCacheDirectory(); err != nil {
			fmt.Printf("ERROR: cleaning cache: %v\n", err)
			os.Exit(1)
		}
		return
	}

	printHeader(!configOssi.Quiet && reflect.TypeOf(configOssi.Formatter).String() == "*audit.AuditLogTextFormatter")

	if err = doStdInAndParse(configOssi); err != nil {
		return
	}

	return
}

func printHeader(print bool) {
	if print {
		figure.NewFigure("Nancy", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		fmt.Println("Nancy version: " + buildversion.BuildVersion)
	}
}

func doStdInAndParse(config configuration.Configuration) (err error) {
	if err = checkStdIn(); err != nil {
		return err
	}

	mod := packages.Mod{}
	scanner := bufio.NewScanner(os.Stdin)

	mod.ProjectList, _ = parse.GoList(scanner)

	var purls = mod.ExtractPurlsFromManifest()

	err = checkOSSIndex(purls, nil, config)

	return err
}

func checkOSSIndex(purls []string, invalidpurls []string, config configuration.Configuration) error {
	var packageCount = len(purls)
	coordinates, err := ossindex.AuditPackagesWithOSSIndex(purls, &config)
	if err != nil {
		return customerrors.NewErrorExitPrintHelp(err, "Error auditing packages")
	}

	var invalidCoordinates []types.Coordinate
	for _, invalidpurl := range invalidpurls {
		invalidCoordinates = append(invalidCoordinates, types.Coordinate{Coordinates: invalidpurl, InvalidSemVer: true})
	}

	if count := audit.LogResults(config.Formatter, packageCount, coordinates, invalidCoordinates, config.CveList.Cves); count > 0 {
		os.Exit(count)
	}
	return nil
}

var stdInInvalid = customerrors.ErrorExit{ExitCode: 1, Message: "StdIn is invalid, either empty or another reason"}

func checkStdIn() (err error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
	} else {
		flag.Usage()
		err = stdInInvalid
	}
	return
}

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
	"errors"
	"flag"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/golang/dep"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	. "github.com/sonatype-nexus-community/nancy/logger"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nancy",
	Short: "Check for vulnerabilities in your Golang dependencies",
	Long: `nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by Sonatype OSS Index, and as well, works with Nexus IQ Server, allowing you
a smooth experience as a Golang developer, using the best tools in the market!`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		LogLady.Info("Nancy parsing config for OSS Index")
		ossIndexConfig, err := configuration.Parse(os.Args[1:])
		if err != nil {
			flag.Usage()
			err = customerrors.Exit(1)
			return
		}
		if err = processConfig(ossIndexConfig); err != nil {
			return
		}
		LogLady.Info("Nancy finished parsing config for OSS Index")
		return
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nancy.yaml)")

	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "less loud")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".nancy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".nancy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func processConfig(config configuration.Configuration) (err error) {
	if config.Help {
		LogLady.Info("Printing usage and exiting clean")
		flag.Usage()
		os.Exit(0)
	}

	if config.Version {
		LogLady.WithFields(logrus.Fields{
			"build_time":   buildversion.BuildTime,
			"build_commit": buildversion.BuildCommit,
			"version":      buildversion.BuildVersion,
		}).Info("Printing version information and exiting clean")

		fmt.Println(buildversion.BuildVersion)
		_, _ = fmt.Printf("build time: %s\n", buildversion.BuildTime)
		_, _ = fmt.Printf("build commit: %s\n", buildversion.BuildCommit)
		os.Exit(0)
	}

	if config.Info {
		LogLady.Level = logrus.InfoLevel
	}
	if config.Debug {
		LogLady.Level = logrus.DebugLevel
	}
	if config.Trace {
		LogLady.Level = logrus.TraceLevel
	}

	if config.CleanCache {
		LogLady.Info("Attempting to clean cache")
		if err := ossindex.RemoveCacheDirectory(); err != nil {
			LogLady.WithField("error", err).Error("Error cleaning cache")
			fmt.Printf("ERROR: cleaning cache: %v\n", err)
			os.Exit(1)
		}
		LogLady.Info("Cache cleaned")
		return
	}

	printHeader(!config.Quiet && reflect.TypeOf(config.Formatter).String() == "*audit.AuditLogTextFormatter")

	if config.UseStdIn {
		LogLady.Info("Parsing config for StdIn")
		if err = doStdInAndParse(config); err != nil {
			return
		}
	}
	if !config.UseStdIn {
		LogLady.Info("Parsing config for file based scan")
		doCheckExistenceAndParse(config)
	}

	return
}

func printHeader(print bool) {
	if print {
		LogLady.Info("Attempting to print header")
		figure.NewFigure("Nancy", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		LogLady.WithField("version", buildversion.BuildVersion).Info("Printing Nancy version")
		fmt.Println("Nancy version: " + buildversion.BuildVersion)
		LogLady.Info("Finished printing header")
	}
}

func doStdInAndParse(config configuration.Configuration) (err error) {
	LogLady.Info("Beginning StdIn parse for OSS Index")
	if err = checkStdIn(); err != nil {
		return err
	}
	LogLady.Info("Instantiating go.mod package")

	mod := packages.Mod{}
	scanner := bufio.NewScanner(os.Stdin)

	LogLady.Info("Beginning to parse StdIn")
	mod.ProjectList, _ = parse.GoList(scanner)
	LogLady.WithFields(logrus.Fields{
		"projectList": mod.ProjectList,
	}).Debug("Obtained project list")

	var purls = mod.ExtractPurlsFromManifest()
	LogLady.WithFields(logrus.Fields{
		"purls": purls,
	}).Debug("Extracted purls")

	LogLady.Info("Auditing purls with OSS Index")
	checkOSSIndex(purls, nil, config)

	return err
}

func doCheckExistenceAndParse(config configuration.Configuration) {
	switch {
	case strings.Contains(config.Path, "Gopkg.lock"):
		workingDir := filepath.Dir(config.Path)
		if workingDir == "." {
			workingDir, _ = os.Getwd()
		}
		getenv := os.Getenv("GOPATH")
		ctx := dep.Ctx{
			WorkingDir: workingDir,
			GOPATHs:    []string{getenv},
		}
		project, err := ctx.LoadProject()
		if err != nil {
			customerrors.Check(err, fmt.Sprintf("could not read lock at path %s", config.Path))
		}
		if project.Lock == nil {
			customerrors.Check(errors.New("dep failed to parse lock file and returned nil"), "nancy could not continue due to dep failure")
		}

		purls, invalidPurls := packages.ExtractPurlsUsingDep(project)

		checkOSSIndex(purls, invalidPurls, config)
	case strings.Contains(config.Path, "go.sum"):
		mod := packages.Mod{}
		mod.GoSumPath = config.Path
		if mod.CheckExistenceOfManifest() {
			mod.ProjectList, _ = parse.GoSum(config.Path)
			var purls = mod.ExtractPurlsFromManifest()

			checkOSSIndex(purls, nil, config)
		}
	default:
		os.Exit(3)
	}
}

func checkOSSIndex(purls []string, invalidpurls []string, config configuration.Configuration) {
	var packageCount = len(purls)
	coordinates, err := ossindex.AuditPackagesWithOSSIndex(purls, &config)
	customerrors.Check(err, "Error auditing packages")

	var invalidCoordinates []types.Coordinate
	for _, invalidpurl := range invalidpurls {
		invalidCoordinates = append(invalidCoordinates, types.Coordinate{Coordinates: invalidpurl, InvalidSemVer: true})
	}

	if count := audit.LogResults(config.Formatter, packageCount, coordinates, invalidCoordinates, config.CveList.Cves); count > 0 {
		os.Exit(count)
	}
}

func checkStdIn() (err error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		LogLady.Info("StdIn is valid")
	} else {
		LogLady.Error("StdIn is invalid, either empty or another reason")
		flag.Usage()
		err = customerrors.Exit(2) // same exit code as used in Usage() function. will remove later
	}
	return
}

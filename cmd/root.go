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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/golang/dep"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex"
	ossIndexTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/sonatype-nexus-community/nancy/audit"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/parse"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

type ossiServerFactory interface {
	create() ossindex.IServer
}

type ossiFactory struct{}

func (ossiFactory) create() ossindex.IServer {
	server := ossindex.New(logLady, ossIndexTypes.Options{
		Username:    viper.GetString("username"),
		Token:       viper.GetString("token"),
		Tool:        "nancy-client",
		Version:     buildversion.BuildVersion,
		DBCacheName: "nancy-cache",
		TTL:         time.Now().Local().Add(time.Hour * 12),
	})
	return server
}

var (
	cfgFile                      string
	configOssi                   types.Configuration
	excludeVulnerabilityFilePath string
	outputFormat                 string
	logLady                      *logrus.Logger
	ossiCreator                  ossiServerFactory = ossiFactory{}
	unixComments                                   = regexp.MustCompile(`#.*$`)
	untilComment                                   = regexp.MustCompile(`(until=)(.*)`)
	stdInInvalid                                   = customerrors.ErrorExit{ExitCode: 1, Message: "StdIn is invalid, either empty or another reason"}
)

var rootCmd = &cobra.Command{
	Version: buildversion.BuildVersion,
	Use:     "nancy",
	Example: `  Typical usage will pipe the output of 'go list -m all' to 'nancy':
  go list -m all | nancy [flags]
  go list -m all | nancy iq [flags]
  go list -json -m all | nancy [flags]
  go list -json -m all | nancy iq [flags]`,
	Short: "Check for vulnerabilities in your Golang dependencies using Sonatype's OSS Index",
	Long: `nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by the 'Sonatype OSS Index', and as well, works with Nexus IQ Server, allowing you
a smooth experience as a Golang developer, using the best tools in the market!`,
	PreRun: func(cmd *cobra.Command, args []string) { bindViper(cmd) },
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
		}
	}()

	logLady = logger.GetLogger("", configOssi.LogLevel)

	err = processConfig()
	if err != nil {
		if errExit, ok := err.(customerrors.ErrorExit); ok {
			os.Exit(errExit.ExitCode)
		} else {
			panic(err)
		}
	}

	return
}

func Execute() (err error) {
	if err = rootCmd.Execute(); err != nil {
		if errExit, ok := err.(customerrors.ErrorExit); ok {
			os.Exit(errExit.ExitCode)
		} else {
			os.Exit(1)
		}
	}
	return
}

const defaultExcludeFilePath = "./.nancy-ignore"

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().CountVarP(&configOssi.LogLevel, "", "v", "Set log level, multiple v's is more verbose")
	rootCmd.PersistentFlags().BoolVar(&configOssi.Version, "version", false, "Get the version")
	rootCmd.PersistentFlags().BoolVarP(&configOssi.Quiet, "quiet", "q", true, "indicate output should contain only packages with vulnerabilities")
	rootCmd.Flags().BoolVarP(&configOssi.NoColor, "no-color", "n", false, "indicate output should not be colorized")
	rootCmd.Flags().BoolVarP(&configOssi.CleanCache, "clean-cache", "c", false, "Deletes local cache directory")
	rootCmd.Flags().VarP(&configOssi.CveList, "exclude-vulnerability", "e", "Comma separated list of CVEs to exclude")
	rootCmd.Flags().StringVarP(&configOssi.Path, "path", "p", "", "Specify a path to a dep Gopkg.lock file for scanning")
	rootCmd.Flags().StringVarP(&configOssi.Username, "username", "u", "", "Specify OSS Index username for request")
	rootCmd.Flags().StringVarP(&configOssi.Token, "token", "t", "", "Specify OSS Index API token for request")
	rootCmd.Flags().StringVarP(&excludeVulnerabilityFilePath, "exclude-vulnerability-file", "x", defaultExcludeFilePath, "Path to a file containing newline separated CVEs to be excluded")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Styling for output format. json, json-pretty, text, csv")
}

func bindViper(cmd *cobra.Command) {
	// need to defer bind call until command is run. see: https://github.com/spf13/viper/issues/233

	// Bind viper to the flags passed in via the command line, so it will override config from file
	_ = viper.BindPFlag("username", cmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("token", cmd.Flags().Lookup("token"))
}

const configTypeYaml = "yaml"

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		viper.SetConfigType(configTypeYaml)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		configPath := path.Join(home, types.OssIndexDirName)

		viper.AddConfigPath(configPath)
		viper.SetConfigType(configTypeYaml)
		viper.SetConfigName(types.OssIndexConfigFileName)
	}

	if err := viper.ReadInConfig(); err == nil {
		// TODO: Add log statements for config
		fmt.Printf("Todo: Add log statement for OSSI config\n")
	}
}

func processConfig() (err error) {
	switch format := outputFormat; format {
	case "text":
		configOssi.Formatter = audit.AuditLogTextFormatter{Quiet: configOssi.Quiet, NoColor: configOssi.NoColor}
	case "json":
		configOssi.Formatter = audit.JsonFormatter{}
	case "json-pretty":
		configOssi.Formatter = audit.JsonFormatter{PrettyPrint: true}
	case "csv":
		configOssi.Formatter = audit.CsvFormatter{Quiet: configOssi.Quiet}
	default:
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!! Output format of", strings.TrimSpace(format), "is not valid. Defaulting to text output")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		configOssi.Formatter = audit.AuditLogTextFormatter{Quiet: configOssi.Quiet, NoColor: configOssi.NoColor}
	}

	// @todo Change to use a switch statement
	if configOssi.LogLevel == 1 || configOssi.Info {
		logLady.Level = logrus.InfoLevel
	}
	if configOssi.LogLevel == 2 || configOssi.Debug {
		logLady.Level = logrus.DebugLevel
	}
	if configOssi.LogLevel == 3 || configOssi.Trace {
		logLady.Level = logrus.TraceLevel
	}

	ossIndex := ossiCreator.create()

	if configOssi.CleanCache {
		if err = ossIndex.NoCacheNoProblems(); err != nil {
			fmt.Printf("ERROR: cleaning cache: %v\n", err)
			return
		}
		return
	}

	printHeader(!configOssi.Quiet && reflect.TypeOf(configOssi.Formatter).String() == "audit.AuditLogTextFormatter")

	// todo: should errors from this call be ignored
	_ = getCVEExcludesFromFile(excludeVulnerabilityFilePath)
	/*	if err = getCVEExcludesFromFile(excludeVulnerabilityFilePath); err != nil {
			return
		}
	*/

	if configOssi.Path != "" {
		if !strings.Contains(configOssi.Path, "Gopkg.lock") {
			err = fmt.Errorf("invalid path value. must point to 'Gopkg.lock' file. path: %s", configOssi.Path)
			return
		}
		if err = doDepAndParse(ossIndex, configOssi.Path); err != nil {
			return
		}
	} else {
		if err = doStdInAndParse(ossIndex); err != nil {
			return
		}
	}

	return
}

func doDepAndParse(ossIndex ossindex.IServer, path string) (err error) {
	workingDir := filepath.Dir(path)
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
		return
	}

	if project.Lock == nil {
		err = fmt.Errorf("dep failed to parse lock file and returned nil, nancy could not continue due to dep failure")
		return
	}

	purls, invalidPurls := packages.ExtractPurlsUsingDep(project)

	err = checkOSSIndex(ossIndex, purls, invalidPurls)
	if err != nil {
		return
	}

	return
}

func getCVEExcludesFromFile(excludeVulnerabilityFilePath string) error {
	fi, err := os.Stat(excludeVulnerabilityFilePath)
	if (fi != nil && fi.IsDir()) || (err != nil && os.IsNotExist(err)) {
		return nil
	}
	file, err := os.Open(excludeVulnerabilityFilePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ogLine := scanner.Text()
		err := determineIfLineIsExclusion(ogLine)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func determineIfLineIsExclusion(ogLine string) error {
	line := unixComments.ReplaceAllString(ogLine, "")
	until := untilComment.FindStringSubmatch(line)
	line = untilComment.ReplaceAllString(line, "")
	cveOnly := strings.TrimSpace(line)

	if len(cveOnly) > 0 {
		if until != nil {
			parseDate, err := time.Parse("2006-01-02", strings.TrimSpace(until[2]))
			if err != nil {
				return fmt.Errorf("failed to parse until at line %q. Expected format is 'until=yyyy-MM-dd'", ogLine)
			}
			if parseDate.After(time.Now()) {
				configOssi.CveList.Cves = append(configOssi.CveList.Cves, cveOnly)
			}
		} else {
			configOssi.CveList.Cves = append(configOssi.CveList.Cves, cveOnly)
		}
	}

	return nil
}

func printHeader(print bool) {
	if print {
		figure.NewFigure("Nancy", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		fmt.Println("Nancy version: " + buildversion.BuildVersion)
	}
}

func doStdInAndParse(ossIndex ossindex.IServer) (err error) {
	if err = checkStdIn(); err != nil {
		return err
	}

	mod := packages.Mod{}

	mod.ProjectList, _ = parse.GoListAgnostic(os.Stdin)

	var purls = mod.ExtractPurlsFromManifest()

	err = checkOSSIndex(ossIndex, purls, nil)

	return err
}

func checkOSSIndex(ossIndex ossindex.IServer, purls []string, invalidpurls []string) (err error) {
	var packageCount = len(purls)
	coordinates, err := ossIndex.AuditPackages(purls)
	if err != nil {
		return
	}

	var invalidCoordinates []ossIndexTypes.Coordinate
	for _, invalidpurl := range invalidpurls {
		invalidCoordinates = append(invalidCoordinates, ossIndexTypes.Coordinate{Coordinates: invalidpurl, InvalidSemVer: true})
	}

	if count := audit.LogResults(configOssi.Formatter, packageCount, coordinates, invalidCoordinates, configOssi.CveList.Cves); count > 0 {
		err = customerrors.ErrorExit{ExitCode: count}
		return
	}
	return
}

func checkStdIn() (err error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
	} else {
		err = stdInInvalid
	}
	return
}

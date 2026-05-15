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
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/common-nighthawk/go-figure"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/configuration"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex"
	ossIndexTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/internal/audit"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
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
	// Get OSS Index URL from viper (which handles env vars, config file, and CLI flags)
	ossIndexURL := viper.GetString(viperKeyOSSIndexURL)

	// Fallback to legacy OSSIndexURL env var for backwards compatibility
	if ossIndexURL == "" {
		ossIndexURL = os.Getenv("OSSIndexURL")
	}

	if ossIndexURL != "" {
		logLady.WithField("ossIndexURL", ossIndexURL).Debug("Using custom OSS Index URL")
	}

	server := ossindex.New(logLady, ossIndexTypes.Options{
		Username:    viper.GetString(configuration.ViperKeyUsername),
		Token:       viper.GetString(configuration.ViperKeyToken),
		Tool:        "nancy-client",
		OSSIndexURL: ossIndexURL,
		Version:     buildversion.BuildVersion,
		DBCachePath: configOssi.DBCachePath,
		DBCacheName: "nancy-cache",
		TTL:         time.Now().Local().Add(time.Hour * 12),
	})

	logLady.WithField("ossiServer", ossIndexTypes.Options{
		Username:    cleanUserName(server.Options.Username),
		Token:       "***hidden***",
		Tool:        server.Options.Tool,
		Version:     server.Options.Version,
		DBCacheName: server.Options.DBCacheName,
		TTL:         server.Options.TTL,
	}).Debug("Created ossiIndex server")

	return server
}

func cleanUserName(origUsername string) string {
	runes := []rune(origUsername)
	cleanUsername := "***hidden***"
	if len(runes) > 0 {
		first := string(runes[0])
		last := string(runes[len(runes)-1])
		cleanUsername = first + "***hidden***" + last
	}
	return cleanUsername
}

//goland:noinspection GoErrorStringFormat
var (
	cfgFile                                 string
	configOssi                              types.Configuration
	excludeVulnerabilityFilePath            string
	additionalExcludeVulnerabilityFilePaths []string
	outputFormat                            string
	logLady                                 *logrus.Logger
	ossiCreator                             ossiServerFactory = ossiFactory{}
	unixComments                                              = regexp.MustCompile(`#.*$`)
	untilComment                                              = regexp.MustCompile(`(until=)(.*)`)
)

//Substitute the _ to .
var viperKeyReplacer = strings.NewReplacer(".", "_")

func setupViperAutomaticEnv() {
	viper.AutomaticEnv()
	//Substitute the _ to .
	viper.SetEnvKeyReplacer(viperKeyReplacer)
}

var rootCmd = &cobra.Command{
	Version: buildversion.BuildVersion,
	Use:     "nancy",
	Example: `  go list -json -deps ./... | nancy sleuth [flags]
  go list -json -deps ./... | nancy lifecycle [flags]

  Or simply (auto-detects and runs go list):
  nancy sleuth [flags]
`,
	Short: "Check for vulnerabilities in your Golang dependencies using Sonatype Guide",
	Long: `nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by Sonatype Guide (previously OSS Index), and works with Sonatype Lifecycle (previously Nexus IQ Server),
allowing you a smooth experience as a Golang developer, using the best tools in the market!

Note: OSS Index credentials are deprecated. Please migrate to Sonatype Guide credentials.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		setupViperAutomaticEnv()
		logLady = logger.GetLogger("", configOssi.LogLevel)
		return nil
	},
	RunE: doRoot,
}

//goland:noinspection GoUnusedParameter
func doRoot(cmd *cobra.Command, args []string) (err error) {
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

	logLady.Info("Nancy parsing config for root command")

	if configOssi.CleanCache {
		ossIndex := ossiCreator.create()
		if err = doCleanCache(ossIndex); err != nil {
			panic(err)
		}
	} else {
		_ = cmd.Usage()
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
const (
	flagNameOssiUsername = "username"
	flagNameOssiToken    = "token"
	flagNameOssiURL      = "ossindex-url"

	// viperKeyOSSIndexURL is the key for OSS Index URL in viper config
	// Following the same pattern as configuration.ViperKeyUsername and configuration.ViperKeyToken
	viperKeyOSSIndexURL = "ossi.OSSIndexURL"
)

func init() {
	cobra.OnInitialize(initConfig)

	persistentFlags := rootCmd.PersistentFlags()
	persistentFlags.CountVarP(&configOssi.LogLevel, "", "v", "Set log level, multiple v's is more verbose")
	persistentFlags.BoolVarP(&configOssi.Version, "version", "V", false, "Get the version")
	persistentFlags.BoolVarP(&configOssi.Quiet, "quiet", "q", true, "indicate output should contain only packages with vulnerabilities")
	persistentFlags.BoolVar(&configOssi.Loud, "loud", false, "indicate output should include non-vulnerable packages")
	rootCmd.Flags().BoolVarP(&configOssi.CleanCache, "clean-cache", "c", false, "Deletes local cache directory")
	persistentFlags.StringVarP(&configOssi.Username, flagNameOssiUsername, "u", "", "Specify OSS Index username for request")
	persistentFlags.StringVarP(&configOssi.Token, flagNameOssiToken, "t", "", "Specify OSS Index API token for request")
	persistentFlags.StringVar(&configOssi.OSSIndexURL, flagNameOssiURL, "", "Specify an alternate OSS Index URL/host")
	persistentFlags.StringVarP(&configOssi.DBCachePath, "db-cache-path", "d", "", "Specify an alternate path for caching responses from OSS Inde, example: /tmp")
	persistentFlags.BoolVar(&configOssi.SkipUpdateCheck, "skip-update-check", configuration.SkipUpdateByDefault(), "Skip the check for updates.")
}

func bindViperRootCmd() {
	// need to defer bind call until command is run. see: https://github.com/spf13/viper/issues/233

	// Bind viper to the flags passed in via the command line, so it will override config from file
	if err := viper.BindPFlag(configuration.ViperKeyUsername, lookupPersistentFlagNotNil(flagNameOssiUsername, rootCmd)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(configuration.ViperKeyToken, lookupPersistentFlagNotNil(flagNameOssiToken, rootCmd)); err != nil {
		panic(err)
	}
	if err := viper.BindPFlag(viperKeyOSSIndexURL, lookupPersistentFlagNotNil(flagNameOssiURL, rootCmd)); err != nil {
		panic(err)
	}
}

func lookupPersistentFlagNotNil(flagName string, cmd *cobra.Command) *pflag.Flag {
	// see: https://github.com/spf13/viper/pull/949
	foundFlag := cmd.PersistentFlags().Lookup(flagName)
	if foundFlag == nil {
		panic(fmt.Errorf("persisent flag lookup for name: '%s' returned nil", flagName))
	}
	return foundFlag
}

func initConfig() {
	viper.SetConfigType(configuration.ConfigTypeYaml)
	var cfgFileToCheck string
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		cfgFileToCheck = cfgFile
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(ossIndexTypes.GetOssIndexDirectory(home))
		viper.SetConfigName(ossIndexTypes.OssIndexConfigFileName)

		cfgFileToCheck = ossIndexTypes.GetOssIndexConfigFile(home)
	}

	if fileExists(cfgFileToCheck) {
		// 'merge' OSSI config here, since IQ cmd also need OSSI config, and init order is not guaranteed
		if err := viper.MergeInConfig(); err != nil {
			panic(err)
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func processConfig() (err error) {
	isQuiet := getIsQuiet()

	switch format := outputFormat; format {
	case "text":
		configOssi.Formatter = audit.AuditLogTextFormatter{Quiet: isQuiet, NoColor: configOssi.NoColor}
	case "json":
		configOssi.Formatter = audit.JsonFormatter{}
	case "json-pretty":
		configOssi.Formatter = audit.JsonFormatter{PrettyPrint: true}
	case "csv":
		configOssi.Formatter = audit.CsvFormatter{Quiet: isQuiet}
	default:
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!! Output format of", strings.TrimSpace(format), "is not valid. Defaulting to text output")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		configOssi.Formatter = audit.AuditLogTextFormatter{Quiet: isQuiet, NoColor: configOssi.NoColor}
	}

	var isTotallyQuiet = isQuiet && outputFormat != "text"

	err = checkForUpdates("", isTotallyQuiet)
	if err != nil {
		return
	}

	ossIndex := ossiCreator.create()

	printHeader(!getIsQuiet() && reflect.TypeOf(configOssi.Formatter).String() == "audit.AuditLogTextFormatter")

	// todo: should errors from getCVEExcludesFromFile() calls be ignored?
	_ = getCVEExcludesFromFile(excludeVulnerabilityFilePath)

	for _, additionalExcludeVulnerabilityFilePath := range additionalExcludeVulnerabilityFilePaths {
		_ = getCVEExcludesFromFile(additionalExcludeVulnerabilityFilePath)
	}

	logLady.Info("Parsing config for StdIn")
	if err = doStdInAndParse(ossIndex); err != nil {
		return
	}

	deduplicateCveList()

	return
}

func doCleanCache(ossIndex ossindex.IServer) (err error) {
	logLady.Info("Attempting to clean cache")
	if err = ossIndex.NoCacheNoProblems(); err != nil {
		logLady.WithField("error", err).Error("Error cleaning cache")
		fmt.Printf("ERROR: cleaning cache: %v\n", err)
		return
	}
	logLady.Info("Cache cleaned")
	return
}

func getIsQuiet() bool {
	return !configOssi.Loud
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

func deduplicateCveList() {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range configOssi.CveList.Cves {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}

	configOssi.CveList.Cves = list
}

func printHeader(print bool) {
	if print {
		figure.NewFigure("Nancy", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		fmt.Println("Nancy version: " + buildversion.BuildVersion)
	}

	logLady.WithFields(logrus.Fields{
		"build_time":       buildversion.BuildTime,
		"build_commit":     buildversion.BuildCommit,
		"version":          buildversion.BuildVersion,
		"operating_system": runtime.GOOS,
		"architecture":     runtime.GOARCH,
	}).Info("Printing Nancy version")
}

func stdinHasData() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func autoRunGoList() (io.Reader, error) {
	fmt.Fprintln(os.Stderr, "No input detected. Running: go list -json -deps ./...")
	cmd := exec.Command("go", "list", "-json", "-deps", "./...")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("auto go list failed: %w\nTip: run manually and pipe output: go list -json -deps ./... | nancy sleuth", err)
	}
	return bytes.NewReader(out), nil
}

func doStdInAndParse(ossIndex ossindex.IServer) (err error) {
	var reader io.Reader
	if stdinHasData() {
		reader = os.Stdin
	} else {
		logLady.Info("No stdin detected; auto-running go list")
		reader, err = autoRunGoList()
		if err != nil {
			return
		}
	}

	mod := packages.Mod{}

	mod.ProjectList, err = parse.GoListAgnostic(reader)
	if err != nil {
		logLady.Error(err)
		return
	}
	logLady.WithFields(logrus.Fields{
		"projectList": mod.ProjectList,
	}).Debug("Obtained project list")

	var purls = mod.ExtractPurlsFromManifest()
	logLady.WithFields(logrus.Fields{
		"purls": purls,
	}).Debug("Extracted purls")

	logLady.Info("Auditing purls with OSS Index")
	err = checkOSSIndex(ossIndex, purls, nil)

	return err
}

func checkOSSIndex(ossIndex ossindex.IServer, purls []string, invalidpurls []string) (err error) {
	var packageCount = len(purls)
	coordinates, err := ossIndex.AuditPackages(purls)
	if err != nil {
		return
	}

	invalidCoordinates := convertInvalidPurlsToCoordinates(invalidpurls)

	if count := audit.LogResults(configOssi.Formatter, packageCount, coordinates, invalidCoordinates, configOssi.CveList.Cves); count > 0 {
		if !configOssi.NoFail {
			err = customerrors.ErrorExit{ExitCode: count}
		}
		return
	}
	return
}

func convertInvalidPurlsToCoordinates(invalidPurls []string) []ossIndexTypes.Coordinate {
	var invalidCoordinates []ossIndexTypes.Coordinate
	for _, invalidPurl := range invalidPurls {
		invalidCoordinates = append(invalidCoordinates, ossIndexTypes.Coordinate{Coordinates: invalidPurl, InvalidSemVer: true})
	}
	return invalidCoordinates
}


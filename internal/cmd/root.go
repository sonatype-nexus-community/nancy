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
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/common-nighthawk/go-figure"
	"github.com/golang/dep"
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
	server := ossindex.New(logLady, ossIndexTypes.Options{
		Username:    viper.GetString(configuration.ViperKeyUsername),
		Token:       viper.GetString(configuration.ViperKeyToken),
		Tool:        "nancy-client",
		Version:     buildversion.BuildVersion,
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
	cfgFile                      string
	configOssi                   types.Configuration
	excludeVulnerabilityFilePath string
	outputFormat                 string
	logLady                      *logrus.Logger
	ossiCreator                  ossiServerFactory = ossiFactory{}
	unixComments                                   = regexp.MustCompile(`#.*$`)
	untilComment                                   = regexp.MustCompile(`(until=)(.*)`)
	stdInInvalid                                   = fmt.Errorf("StdIn is invalid or empty. Did you forget to pipe 'go list' to nancy?")
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
	Example: `  Typical usage will pipe the output of 'go list -json -m all' to 'nancy':
  go list -json -m all | nancy sleuth [flags]
  go list -json -m all | nancy iq [flags]

  If using dep typical usage is as follows :
  nancy sleuth -p Gopkg.lock [flags]
  nancy iq -p Gopkg.lock [flags]
`,
	Short: "Check for vulnerabilities in your Golang dependencies using Sonatype's OSS Index",
	Long: `nancy is a tool to check for vulnerabilities in your Golang dependencies,
powered by the 'Sonatype OSS Index', and as well, works with Nexus IQ Server, allowing you
a smooth experience as a Golang developer, using the best tools in the market!`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		setupViperAutomaticEnv()
		logLady = logger.GetLogger("", configOssi.LogLevel)
		return checkForUpdates("")
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

	GopkgLockFilename = "Gopkg.lock"
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
	persistentFlags.StringVarP(&configOssi.Path, "path", "p", "", "Specify a path to a dep "+GopkgLockFilename+" file for scanning")
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
	case "sarif":
		configOssi.Formatter = audit.SarifFormatter{UsingDep: configOssi.Path != ""}
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

	ossIndex := ossiCreator.create()

	printHeader(!getIsQuiet() && reflect.TypeOf(configOssi.Formatter).String() == "audit.AuditLogTextFormatter")

	// todo: should errors from this call be ignored
	_ = getCVEExcludesFromFile(excludeVulnerabilityFilePath)
	/*	if err = getCVEExcludesFromFile(excludeVulnerabilityFilePath); err != nil {
			return
		}
	*/

	if configOssi.Path != "" {
		if err = doDepAndParse(ossIndex, configOssi.Path); err != nil {
			logLady.WithField("error", err).Error("Error in file based scan")
			return
		}
	} else {
		logLady.Info("Parsing config for StdIn")
		if err = doStdInAndParse(ossIndex); err != nil {
			return
		}
	}

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

func getPurlsFromPath(path string) (purls []string, invalidPurls []string, err error) {
	logLady.Info("Parsing config for file based scan")
	if !strings.Contains(path, GopkgLockFilename) {
		err = fmt.Errorf("invalid path value. must point to '%s' file. path: %s", GopkgLockFilename, path)
		logLady.WithField("error", err).Error("Path error in file based scan")
		return
	}

	workingDir := filepath.Dir(path)
	if workingDir == "." {
		workingDir, _ = os.Getwd()
	}
	getenv := os.Getenv("GOPATH")
	ctx := dep.Ctx{
		WorkingDir: workingDir,
		GOPATHs:    []string{getenv},
	}

	var project *dep.Project
	project, err = ctx.LoadProject()
	if err != nil {
		return
	}

	if project.Lock == nil {
		err = fmt.Errorf("dep failed to parse lock file and returned nil, nancy could not continue due to dep failure")
		return
	}

	purls, invalidPurls = packages.ExtractPurlsUsingDep(project)
	return
}

func doDepAndParse(ossIndex ossindex.IServer, path string) (err error) {
	var purls, invalidPurls []string
	if purls, invalidPurls, err = getPurlsFromPath(path); err != nil {
		return
	}

	if err = checkOSSIndex(ossIndex, purls, invalidPurls); err != nil {
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

	logLady.WithFields(logrus.Fields{
		"build_time":       buildversion.BuildTime,
		"build_commit":     buildversion.BuildCommit,
		"version":          buildversion.BuildVersion,
		"operating_system": runtime.GOOS,
		"architecture":     runtime.GOARCH,
	}).Info("Printing Nancy version")
}

func doStdInAndParse(ossIndex ossindex.IServer) (err error) {
	if err = checkStdIn(); err != nil {
		return err
	}

	mod := packages.Mod{}

	mod.ProjectList, err = parse.GoListAgnostic(os.Stdin)
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
		err = customerrors.ErrorExit{ExitCode: count}
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

func checkStdIn() (err error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		logLady.Info("StdIn is valid")
	} else {
		err = stdInInvalid
		logLady.Error(err)
	}
	return
}

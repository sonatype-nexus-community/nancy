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

package configuration

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"gopkg.in/yaml.v2"
)

// these consts must match their associated yaml tag below. for use where tag name matters, like viper
const YamlKeyIQServer = "IQServer"
const YamlKeyIQUsername = "IQUsername"
const YamlKeyIQToken = "IQToken"

// IQConfig is a struct for holding IQ Configuration, and for writing it to yaml
type IQConfig struct {
	IQServer   string `yaml:"IQServer"`
	IQUsername string `yaml:"IQUsername"`
	IQToken    string `yaml:"IQToken"`
}

// these consts must match their associated yaml tag below. for use where tag name matters, like viper
const YamlKeyUsername = "Username"
const YamlKeyToken = "Token"

// OSSIndexConfig is a struct for holding OSS Index Configuration, and for writing it to yaml
type OSSIndexConfig struct {
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

var (
	// HomeDir is exported so that in testing it can be set to a location like /tmp
	HomeDir string
	// ConfigLocation is exported so that in testing it can be used to test if the file has been written properly
	ConfigLocation string
	logLady        *logrus.Logger
)

func init() {
	HomeDir, _ = os.UserHomeDir()
	logLady = logger.GetLogger("", 3)
}

// GetConfigFromCommandLine is a method to obtain IQ or OSS Index config from the command line,
// and then write it to disk.
func GetConfigFromCommandLine(stdin io.Reader) (err error) {
	logLady.Info("Starting process to obtain config from user")
	reader := bufio.NewReader(stdin)
	fmt.Print("Hi! What config can I help you set, IQ or OSS Index (values: iq, ossindex, enter for exit)? ")
	configType, _ := reader.ReadString('\n')

	switch str := strings.TrimSpace(configType); str {
	case "iq":
		logLady.Info("User chose to set IQ Config, moving forward")
		ConfigLocation = filepath.Join(HomeDir, types.IQServerDirName, types.IQServerConfigFileName)
		err = getAndSetIQConfig(reader)
	case "ossindex":
		logLady.Info("User chose to set OSS Index config, moving forward")
		ConfigLocation = filepath.Join(HomeDir, types.OssIndexDirName, types.OssIndexConfigFileName)
		err = getAndSetOSSIndexConfig(reader)
	case "":
		// TODO should this return an error, because it means config setup was not completed?
		return
	default:
		logLady.Infof("User chose invalid config type: %s, will retry", str)
		fmt.Println("Invalid value, 'iq' and 'ossindex' are accepted values, try again!")
		err = GetConfigFromCommandLine(stdin)
	}

	if err != nil {
		logLady.Error(err)
		return
	}
	return
}

func getAndSetIQConfig(reader *bufio.Reader) (err error) {
	logLady.Info("Getting config for IQ Server from user")

	iqConfig := IQConfig{IQServer: "http://localhost:8070", IQUsername: "admin", IQToken: "admin123"}

	fmt.Print("What is the address of your Nexus IQ Server (default: http://localhost:8070)? ")
	server, _ := reader.ReadString('\n')
	iqConfig.IQServer = emptyOrDefault(server, iqConfig.IQServer)

	fmt.Print("What username do you want to authenticate as (default: admin)? ")
	username, _ := reader.ReadString('\n')
	iqConfig.IQUsername = emptyOrDefault(username, iqConfig.IQUsername)

	fmt.Print("What token do you want to use (default: admin123)? ")
	token, _ := reader.ReadString('\n')
	iqConfig.IQToken = emptyOrDefault(token, iqConfig.IQToken)

	if iqConfig.IQUsername == "admin" || iqConfig.IQToken == "admin123" {
		logLady.Info("Warning user of bad life choices, using default values for IQ Server username or token")
		warnUserOfBadLifeChoices()
		fmt.Print("[y/N]? ")
		theChoice, _ := reader.ReadString('\n')
		theChoice = emptyOrDefault(theChoice, "y")
		if theChoice == "y" {
			logLady.Info("User chose to rectify their bad life choices, asking for config again")
			err = getAndSetIQConfig(reader)
		} else {
			logLady.Info("Successfully got IQ Server config from user, attempting to save to disk")
			err = marshallAndWriteToDisk(iqConfig)
		}
	} else {
		logLady.Info("Successfully got IQ Server config from user, attempting to save to disk")
		err = marshallAndWriteToDisk(iqConfig)
	}

	if err != nil {
		return
	}
	return
}

func emptyOrDefault(value string, defaultValue string) string {
	str := strings.Trim(strings.TrimSpace(value), "\n")
	if str == "" {
		return defaultValue
	}
	return str
}

func getAndSetOSSIndexConfig(reader *bufio.Reader) (err error) {
	logLady.Info("Getting config for OSS Index from user")

	ossIndexConfig := OSSIndexConfig{}

	fmt.Print("What username do you want to authenticate as (ex: admin)? ")
	ossIndexConfig.Username, _ = reader.ReadString('\n')
	ossIndexConfig.Username = strings.Trim(strings.TrimSpace(ossIndexConfig.Username), "\n")

	fmt.Print("What token do you want to use? ")
	ossIndexConfig.Token, _ = reader.ReadString('\n')
	ossIndexConfig.Token = strings.Trim(strings.TrimSpace(ossIndexConfig.Token), "\n")

	logLady.Info("Successfully got OSS Index config from user, attempting to save to disk")
	err = marshallAndWriteToDisk(ossIndexConfig)
	if err != nil {
		logLady.Error(err)
		return
	}

	return
}

func marshallAndWriteToDisk(config interface{}) (err error) {
	d, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	base := filepath.Dir(ConfigLocation)

	if _, err = os.Stat(base); os.IsNotExist(err) {
		err = os.Mkdir(base, os.ModePerm)
		if err != nil {
			return
		}
	}

	err = ioutil.WriteFile(ConfigLocation, d, 0644)
	if err != nil {
		return
	}

	logLady.WithField("config_location", ConfigLocation).Info("Successfully wrote config to disk")
	fmt.Printf("Successfully wrote config to: %s\n", ConfigLocation)
	return
}

func warnUserOfBadLifeChoices() {
	fmt.Println()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!! WARNING : You are using the default username and/or password for Nexus IQ. !!!!")
	fmt.Println("!!!! You are strongly encouraged to change these, and use a token.              !!!!")
	fmt.Println("!!!! Would you like to change them and try again?                               !!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println()
}

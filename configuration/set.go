package configuration

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"gopkg.in/yaml.v2"
)

// IQConfig is a struct for holding IQ Configuration, and for writing it to yaml
type IQConfig struct {
	Server   string `yaml:"Server"`
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

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
)

func init() {
	HomeDir, _ = os.UserHomeDir()
}

// GetConfigFromCommandLine is a method to obtain IQ or OSS Index config from the command line,
// and then write it to disk.
func GetConfigFromCommandLine(stdin io.Reader) (err error) {
	LogLady.Info("Starting process to obtain config from user")
	reader := bufio.NewReader(stdin)
	fmt.Print("Hi! What config can I help you set, IQ or OSS Index (values: iq, ossindex, enter for exit)? ")
	configType, _ := reader.ReadString('\n')

	switch str := strings.TrimSpace(configType); str {
	case "iq":
		LogLady.Info("User chose to set IQ Config, moving forward")
		ConfigLocation = filepath.Join(HomeDir, types.IQServerDirName, types.IQServerConfigFileName)
		err = getAndSetIQConfig(reader)
	case "ossindex":
		LogLady.Info("User chose to set OSS Index config, moving forward")
		ConfigLocation = filepath.Join(HomeDir, types.OssIndexDirName, types.OssIndexConfigFileName)
		err = getAndSetOSSIndexConfig(reader)
	case "":
		return
	default:
		LogLady.Info("User chose to set OSS Index config, moving forward")
		fmt.Println("Invalid value, 'iq' and 'ossindex' are accepted values, try again!")
		err = GetConfigFromCommandLine(stdin)
	}

	if err != nil {
		LogLady.Error(err)
		return
	}
	return
}

func getAndSetIQConfig(reader *bufio.Reader) (err error) {
	LogLady.Info("Getting config for IQ Server from user")

	iqConfig := IQConfig{Server: "http://localhost:8070", Username: "admin", Token: "admin123"}

	fmt.Print("What is the address of your Nexus IQ Server (default: http://localhost:8070)? ")
	server, _ := reader.ReadString('\n')
	iqConfig.Server = emptyOrDefault(server, iqConfig.Server)

	fmt.Print("What username do you want to authenticate as (default: admin)? ")
	username, _ := reader.ReadString('\n')
	iqConfig.Username = emptyOrDefault(username, iqConfig.Username)

	fmt.Print("What token do you want to use (default: admin123)? ")
	token, _ := reader.ReadString('\n')
	iqConfig.Token = emptyOrDefault(token, iqConfig.Token)

	if iqConfig.Username == "admin" || iqConfig.Token == "admin123" {
		LogLady.Info("Warning user of bad life choices, using default values for IQ Server username or token")
		warnUserOfBadLifeChoices()
		fmt.Print("[y/N]? ")
		theChoice, _ := reader.ReadString('\n')
		theChoice = emptyOrDefault(theChoice, "y")
		if theChoice == "y" {
			LogLady.Info("User chose to rectify their bad life choices, asking for config again")
			err = getAndSetIQConfig(reader)
		} else {
			LogLady.Info("Successfully got IQ Server config from user, attempting to save to disk")
			err = marshallAndWriteToDisk(iqConfig)
		}
	} else {
		LogLady.Info("Successfully got IQ Server config from user, attempting to save to disk")
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
	LogLady.Info("Getting config for OSS Index from user")

	ossIndexConfig := OSSIndexConfig{}

	fmt.Print("What username do you want to authenticate as (ex: admin)? ")
	ossIndexConfig.Username, _ = reader.ReadString('\n')
	ossIndexConfig.Username = strings.Trim(strings.TrimSpace(ossIndexConfig.Username), "\n")

	fmt.Print("What token do you want to use? ")
	ossIndexConfig.Token, _ = reader.ReadString('\n')
	ossIndexConfig.Token = strings.Trim(strings.TrimSpace(ossIndexConfig.Token), "\n")

	LogLady.Info("Successfully got OSS Index config from user, attempting to save to disk")
	err = marshallAndWriteToDisk(ossIndexConfig)
	if err != nil {
		LogLady.Error(err)
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

	LogLady.WithField("config_location", ConfigLocation).Info("Successfully wrote config to disk")
	fmt.Printf("Successfully wrote config to: %s", ConfigLocation)
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

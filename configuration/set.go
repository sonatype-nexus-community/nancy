package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"gopkg.in/yaml.v2"
)

type IQConfig struct {
	Server   string `yaml:"Server"`
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

type OSSIndexConfig struct {
	Username string `yaml:"Username"`
	Token    string `yaml:"Token"`
}

var homeDir string

func GetConfigFromCommandLine() {
	homeDir, _ = os.UserHomeDir()
	fmt.Print("Hi! What config can I help you set, IQ or OSS Index (values: iq, ossindex, enter for exit)? ")
	var configType string
	fmt.Scanln(&configType)
	switch configType {
	case "iq":
		getAndSetIQConfig()
	case "ossindex":
		getAndSetOSSIndexConfig()
	case "":
		os.Exit(0)
	default:
		fmt.Println("Invalid value, iq and ossindex are accepted values, try again!")
		GetConfigFromCommandLine()
	}
}

func getAndSetIQConfig() {
	iqConfig := IQConfig{Server: "http://localhost:8070", Username: "admin", Token: "admin123"}
	fmt.Print("What is the address of your Nexus IQ Server (default: http://localhost:8070)? ")
	fmt.Scanln(&iqConfig.Server)
	fmt.Print("What username do you want to authenticate as (default: admin)? ")
	fmt.Scanln(&iqConfig.Username)
	fmt.Print("What token do you want to use (default: admin123)? ")
	fmt.Scanln(&iqConfig.Token)

	if iqConfig.Username == "admin" && iqConfig.Token == "admin123" {
		var theChoice string
		warnUserOfBadLifeChoices()
		fmt.Print("[y/N]? ")
		fmt.Scanln(&theChoice)
		if theChoice == "y" {
			getAndSetIQConfig()
		}
	}

	err := marshallAndWriteToDisk(iqConfig, filepath.Join(homeDir, types.IQServerDirName, types.IQServerConfigFileName))
	if err != nil {
		LogLady.Error(err)
	}
}

func getAndSetOSSIndexConfig() {
	ossIndexConfig := OSSIndexConfig{}
	fmt.Print("What username do you want to authenticate as (ex: admin)? ")
	fmt.Scanln(&ossIndexConfig.Username)
	fmt.Print("What token do you want to use? ")
	fmt.Scanln(&ossIndexConfig.Token)

	err := marshallAndWriteToDisk(ossIndexConfig, filepath.Join(homeDir, types.OssIndexDirName, types.OssIndexConfigFileName))
	if err != nil {
		LogLady.Error(err)
	}
}

func marshallAndWriteToDisk(config interface{}, configLocation string) error {
	d, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configLocation, d, 0644)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Successfully wrote config to: %s", configLocation))
	return nil
}

func warnUserOfBadLifeChoices() {
	fmt.Println()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!! WARNING : You are using the default username and password for Nexus IQ. !!!!")
	fmt.Println("!!!! You are strongly encouraged to change these, and use a token.           !!!!")
	fmt.Println("!!!! Would you like to change them and try again?                            !!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println()
}

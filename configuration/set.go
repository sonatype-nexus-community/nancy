package configuration

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	reader := bufio.NewReader(stdin)
	fmt.Print("Hi! What config can I help you set, IQ or OSS Index (values: iq, ossindex, enter for exit)? ")
	configType, _ := reader.ReadString('\n')

	switch str := strings.TrimSpace(configType); str {
	case "iq":
		ConfigLocation = filepath.Join(HomeDir, types.IQServerDirName, types.IQServerConfigFileName)
		err = getAndSetIQConfig(reader)
	case "ossindex":
		ConfigLocation = filepath.Join(HomeDir, types.OssIndexDirName, types.OssIndexConfigFileName)
		err = getAndSetOSSIndexConfig(reader)
	case "":
		return errors.New("Uh oh")
	default:
		fmt.Println("Invalid value, 'iq' and 'ossindex' are accepted values, try again!")
		GetConfigFromCommandLine(stdin)
	}

	if err != nil {
		return
	}
	return nil
}

func getAndSetIQConfig(reader *bufio.Reader) (err error) {
	iqConfig := IQConfig{Server: "http://localhost:8070", Username: "admin", Token: "admin123"}
	fmt.Print("What is the address of your Nexus IQ Server (default: http://localhost:8070)? ")
	iqConfig.Server, _ = reader.ReadString('\n')
	iqConfig.Server = strings.TrimSpace(iqConfig.Server)
	fmt.Print("What username do you want to authenticate as (default: admin)? ")
	iqConfig.Username, _ = reader.ReadString('\n')
	iqConfig.Username = strings.TrimSpace(iqConfig.Username)
	fmt.Print("What token do you want to use (default: admin123)? ")
	iqConfig.Token, _ = reader.ReadString('\n')
	iqConfig.Token = strings.TrimSpace(iqConfig.Token)

	if iqConfig.Username == "admin" || iqConfig.Token == "admin123" {
		warnUserOfBadLifeChoices()
		fmt.Print("[y/N]? ")
		theChoice, _ := reader.ReadString('\n')
		theChoice = strings.TrimSpace(theChoice)
		if theChoice == "y" {
			getAndSetIQConfig(reader)
		}
	}

	err = marshallAndWriteToDisk(iqConfig)
	if err != nil {
		LogLady.Error(err)
		return
	}
	return
}

func getAndSetOSSIndexConfig(reader *bufio.Reader) (err error) {
	ossIndexConfig := OSSIndexConfig{}
	fmt.Print("What username do you want to authenticate as (ex: admin)? ")
	ossIndexConfig.Username, _ = reader.ReadString('\n')
	ossIndexConfig.Username = strings.TrimSpace(ossIndexConfig.Username)
	fmt.Print("What token do you want to use? ")
	ossIndexConfig.Token, _ = reader.ReadString('\n')
	ossIndexConfig.Token = strings.TrimSpace(ossIndexConfig.Token)

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

	log.Print(base)

	if _, err := os.Stat(base); os.IsNotExist(err) {
		os.Mkdir(base, os.ModePerm)
	}

	err = ioutil.WriteFile(ConfigLocation, d, 0644)
	if err != nil {
		return
	}

	fmt.Println(fmt.Sprintf("Successfully wrote config to: %s", ConfigLocation))
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

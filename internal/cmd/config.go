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
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sonatype-nexus-community/nancy/v2/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/v2/internal/logger"
	localossindex "github.com/sonatype-nexus-community/nancy/v2/internal/ossindex"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Setup credentials to use when connecting to services",
	Long: `Save credentials for reuse in connecting to various backend services.
The config command will prompt for the type of credentials to save.`,
	RunE: doConfig,
}

//noinspection GoUnusedParameter
func doConfig(cmd *cobra.Command, args []string) (err error) {
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

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("What credentials would you like to configure?")
	fmt.Println("  1) Sonatype Guide credentials (recommended)")
	fmt.Println("  2) Sonatype Lifecycle (IQ Server) credentials")
	fmt.Print("Enter selection: ")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())

	home, err := homedir.Dir()
	if err != nil {
		return
	}

	switch choice {
	case "1", "":
		err = configureGuide(scanner, home)
	case "2":
		err = configureLifecycle(scanner, home)
	default:
		fmt.Println("Unknown selection. Exiting.")
	}
	return
}

func configureGuide(scanner *bufio.Scanner, home string) error {
	fmt.Println("\nSonatype Guide Configuration")

	fmt.Print("Sonatype Guide Bearer token (recommended, leave blank to skip): ")
	scanner.Scan()
	guideToken := strings.TrimSpace(scanner.Text())

	var username, ossiToken string
	fmt.Print("Configure OSS Index credentials (deprecated, press Enter to skip)? [y/N]: ")
	scanner.Scan()
	if strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
		fmt.Println("  OSS Index credentials are deprecated and will be removed in v3.x.")
		fmt.Println("  Obtain a Sonatype Guide Bearer token at https://guide.sonatype.com")

		fmt.Print("  Username: ")
		scanner.Scan()
		username = strings.TrimSpace(scanner.Text())

		fmt.Print("  OSS Index API token: ")
		scanner.Scan()
		ossiToken = strings.TrimSpace(scanner.Text())
	}

	type guideConf struct {
		Guide struct {
			Token string `yaml:"token"`
		} `yaml:"guide"`
		Ossi struct {
			Username string `yaml:"Username"`
			Token    string `yaml:"Token"`
		} `yaml:"ossi"`
	}

	conf := guideConf{}
	conf.Guide.Token = guideToken
	conf.Ossi.Username = username
	conf.Ossi.Token = ossiToken

	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}

	dir := localossindex.GetOssIndexDirectory(home)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	cfgFile := filepath.Join(dir, localossindex.OssIndexConfigFileName)
	if err = os.WriteFile(cfgFile, data, 0600); err != nil {
		return err
	}

	fmt.Println("Credentials saved to", cfgFile)
	return nil
}

func configureLifecycle(scanner *bufio.Scanner, home string) error {
	fmt.Println("\nSonatype Lifecycle Configuration")

	fmt.Print("Server URL [http://localhost:8070]: ")
	scanner.Scan()
	server := strings.TrimSpace(scanner.Text())
	if server == "" {
		server = "http://localhost:8070"
	}

	fmt.Print("Username [admin]: ")
	scanner.Scan()
	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		username = "admin"
	}

	fmt.Print("Token [admin123]: ")
	scanner.Scan()
	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		token = "admin123"
	}

	type iqConf struct {
		Iq struct {
			Server   string `yaml:"Server"`
			Username string `yaml:"Username"`
			Token    string `yaml:"Token"`
		} `yaml:"iq"`
	}

	conf := iqConf{}
	conf.Iq.Server = server
	conf.Iq.Username = username
	conf.Iq.Token = token

	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}

	dir := localossindex.GetIQServerDirectory(home)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	cfgFile := filepath.Join(dir, localossindex.IQServerConfigFileName)
	if err = os.WriteFile(cfgFile, data, 0600); err != nil {
		return err
	}

	fmt.Println("Credentials saved to", cfgFile)
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
}

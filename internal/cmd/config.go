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
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
	localossindex "github.com/sonatype-nexus-community/nancy/internal/ossindex"
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
	fmt.Println("  For Bearer token auth: leave username empty and enter your Guide token.")
	fmt.Println("  For OSS Index compatibility: enter username (email) and API token.")

	fmt.Print("Username (leave empty for Bearer token): ")
	scanner.Scan()
	username := strings.TrimSpace(scanner.Text())

	fmt.Print("Token (Guide Bearer token or OSS Index API token): ")
	scanner.Scan()
	token := strings.TrimSpace(scanner.Text())

	type ossiConf struct {
		Ossi struct {
			Username string `yaml:"Username"`
			Token    string `yaml:"Token"`
		} `yaml:"ossi"`
	}

	conf := ossiConf{}
	conf.Ossi.Username = username
	conf.Ossi.Token = token

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

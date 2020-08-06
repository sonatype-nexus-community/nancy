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
	"fmt"
	"os"

	"github.com/sonatype-nexus-community/nancy/internal/customerrors"

	"github.com/sonatype-nexus-community/nancy/internal/configuration"
	"github.com/spf13/cobra"
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

	if err = configuration.GetConfigFromCommandLine(os.Stdin); err != nil {
		panic(err)
	}
	return
}

func init() {
	rootCmd.AddCommand(configCmd)
}

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

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/go-sona-types/cyclonedx"
	"github.com/sonatype-nexus-community/nancy/internal/customerrors"
	"github.com/sonatype-nexus-community/nancy/internal/logger"
	"github.com/spf13/cobra"
)

var sbomCmd = &cobra.Command{
	Use:     "sbom",
	Example: `  go list -json -m all | nancy sbom`,
	Short:   "Output a CycloneDX Software Bill Of Materials",
	Long:    `'nancy sbom' is a command to output a CycloneDX Software Bill Of Materials`,
	RunE:    doSbom,
}

//noinspection GoUnusedParameter
func doSbom(cmd *cobra.Command, args []string) (err error) {
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
	logLady.Info("Nancy parsing config for generating SBOM")

	var purls []string
	purls, err = getPurls()

	var packageUrls []packageurl.PackageURL

	for _, v := range purls {
		purl, err := packageurl.FromString(v)
		if err != nil {
			logLady.WithError(err).Error("unexpected error in sbom cmd")
		}
		packageUrls = append(packageUrls, purl)
	}

	cyclonedx := cyclonedx.Default(logLady)

	sbom := cyclonedx.FromPackageURLs(packageUrls)

	fmt.Print(sbom)

	return
}

func init() {
	rootCmd.AddCommand(sbomCmd)
}

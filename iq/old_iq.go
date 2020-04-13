// Copyright 2020 Sonatype Inc.
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

// Package iq has definitions and functions for processing golang purls with Nexus IQ Server
package iq

import (
	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/types"
)

func purlsToPackageURL(purls []string) (newPurls []packageurl.PackageURL) {
	for _, v := range purls {
		newPurl, _ := packageurl.FromString(v)
		newPurls = append(newPurls, newPurl)
	}
	return
}

// AuditPackages accepts a slice of purls, public application ID, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
//
// Deprecated: please use Audit instead
func AuditPackages(purls []string, applicationID string, config configuration.IqConfiguration) (types.StatusURLResult, error) {
	return doAudit(
		purlsToPackageURL(purls),
		configuration.Config{
			Application: config.Application,
			Stage:       config.Stage,
			IQConfig: configuration.IQConfig{
				Username: config.User,
				Token:    config.Token,
				Server:   config.Server,
			},
		})
}

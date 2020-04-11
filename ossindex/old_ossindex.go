// Copyright 2018 Sonatype Inc.
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

// Package ossindex has definitions and functions for processing the OSS Index Feed
package ossindex

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

// AuditPackages will given a list of Package URLs, run an OSS Index audit.
//
// Deprecated: AuditPackages is old and being maintained for upstream compatibility at the moment.
// It will be removed when we go to a major version release. Use AuditPackagesWithOSSIndex instead.
func AuditPackages(purls []string) ([]types.Coordinate, error) {
	return doAudit(purlsToPackageURL(purls), nil)
}

// AuditPackagesWithOSSIndex will given a list of Package URLs, run an OSS Index audit, and takes OSS Index configuration
//
// Deprecated: AuditPackagesWithOSSIndex is old and being maintained for upstream compatibility at the moment.
// It will be removed when we go to a major version release. Use Audit instead.
func AuditPackagesWithOSSIndex(purls []string, config *configuration.Configuration) ([]types.Coordinate, error) {
	updatedConfig := configuration.Config{Username: config.Username, Token: config.Token}
	return doAudit(purlsToPackageURL(purls), &updatedConfig)
}

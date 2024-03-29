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

package audit

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/sonatype-nexus-community/nancy/buildversion"
)

func isEntryValid(params ...interface{}) bool {
	for _, v := range params {
		if v == nil {
			return false
		}
	}
	return true
}

// LogResults will given a number of expected results and the results themselves, log the
// results.
func LogResults(formatter log.Formatter, packageCount int, coordinates []types.Coordinate, invalidCoordinates []types.Coordinate, exclusions []string) int {
	vulnerableCount := 0

	for _, c := range coordinates {
		c.ExcludeVulnerabilities(exclusions)
	}

	var auditedCoordinates []types.Coordinate
	var vulnerableCoordinates []types.Coordinate
	var excludedVulnerabilities []types.Vulnerability

	for i := 0; i < len(coordinates); i++ {
		coordinate := coordinates[i]
		for _, v := range coordinate.Vulnerabilities {
			if v.Excluded {
				excludedVulnerabilities = append(excludedVulnerabilities, v)
			}
		}
		if coordinate.IsVulnerable() {
			vulnerableCount++
			vulnerableCoordinates = append(vulnerableCoordinates, coordinate)
		}
		auditedCoordinates = append(auditedCoordinates, coordinate)
	}

	if invalidCoordinates == nil {
		invalidCoordinates = make([]types.Coordinate, 0)
	}
	if exclusions == nil {
		exclusions = make([]string, 0)
	}
	if vulnerableCoordinates == nil {
		vulnerableCoordinates = make([]types.Coordinate, 0)
	}

	exclusionCount := len(excludedVulnerabilities)

	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.WithFields(log.Fields{
		"exclusions":     exclusions,
		"num_audited":    packageCount,
		"num_vulnerable": vulnerableCount,
		"num_exclusions": exclusionCount,
		"audited":        auditedCoordinates,
		"vulnerable":     vulnerableCoordinates,
		"excluded":       excludedVulnerabilities,
		"invalid":        invalidCoordinates,
		"version":        buildversion.BuildVersion,
	}).Info("")

	return vulnerableCount
}

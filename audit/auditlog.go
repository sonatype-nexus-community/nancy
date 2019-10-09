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
package audit

import (
	"fmt"
	"strconv"

	"github.com/logrusorgru/aurora"
	"github.com/sonatype-nexus-community/nancy/types"
)

func LogInvalidSemVerWarning(noColor bool, quiet bool, invalidPurls []string) {
	if !quiet {
		packageCount := len(invalidPurls)
		warningMessage := "!!!!! WARNING !!!!!\nScanning cannot be completed on the following package(s) since they do not use semver."
		au := aurora.NewAurora(!noColor)
		fmt.Println(au.Red(warningMessage))

		for i := 0; i < len(invalidPurls); i++ {
			idx := i + 1
			purl := invalidPurls[i]
			fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]", au.Bold(purl))
		}
		fmt.Println()
	}
}

func logPackage(noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	if noColor {
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			coordinate.Coordinates,
			"   No known vulnerabilities against package/version")
	} else {
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			aurora.Bold(coordinate.Coordinates),
			aurora.Gray(20-1,"   No known vulnerabilities against package/version"))
	}
}

func logVulnerablePackage(noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	if noColor {
		fmt.Println("------------------------------------------------------------")
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			coordinate.Coordinates+"  [Vulnerable]",
			"   "+strconv.Itoa(len(coordinate.Vulnerabilities)),
			"known vulnerabilities affecting installed version")

		for j := 0; j < len(coordinate.Vulnerabilities); j++ {
			if !coordinate.Vulnerabilities[j].Excluded {
				fmt.Printf("\n%s\n%s\n\nID:%s\nDetails:%s",
					coordinate.Vulnerabilities[j].Title,
					coordinate.Vulnerabilities[j].Description,
					coordinate.Vulnerabilities[j].Id,
					coordinate.Vulnerabilities[j].Reference)
			}
		}
	} else {
		fmt.Println("------------------------------------------------------------")
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			aurora.Bold(aurora.Red(coordinate.Coordinates+"  [Vulnerable]")),
			"   "+strconv.Itoa(len(coordinate.Vulnerabilities)),
			"known vulnerabilities affecting installed version")

		for j := 0; j < len(coordinate.Vulnerabilities); j++ {
			if !coordinate.Vulnerabilities[j].Excluded {
				fmt.Printf("\n%s\n%s\n\nID:%s\nDetails:%s",
					coordinate.Vulnerabilities[j].Title,
					coordinate.Vulnerabilities[j].Description,
					coordinate.Vulnerabilities[j].Id,
					coordinate.Vulnerabilities[j].Reference)
			}
		}
	}

	return
}

// LogResults will given a number of expected results and the results themselves, log the
// results.
func LogResults(noColor bool, quiet bool, packageCount int, coordinates []types.Coordinate, exclusions []string) int {
	vulnerableCount := 0

	for _, c := range coordinates {
		c.ExcludeVulnerabilities(exclusions)
	}

	for i := 0; i < len(coordinates); i++ {
		coordinate := coordinates[i]
		idx := i + 1

		if !coordinate.IsVulnerable() {
			if !quiet {
				logPackage(noColor, idx, packageCount, coordinate)
			}
		} else {
			logVulnerablePackage(noColor, idx, packageCount, coordinate)
			vulnerableCount++
		}
	}

	fmt.Println()
	if noColor {
		fmt.Println("Audited dependencies:", strconv.Itoa(packageCount)+",",
			"Vulnerable:", strconv.Itoa(vulnerableCount))
	} else {
		fmt.Println("Audited dependencies:", strconv.Itoa(packageCount)+",",
			"Vulnerable:", aurora.Bold(aurora.Red(strconv.Itoa(vulnerableCount))))
	}

	return vulnerableCount
}

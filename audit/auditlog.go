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
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/sonatype-nexus-community/nancy/types"
)

func logPackage(noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	if noColor {
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			coordinate.Coordinates,
			"   No known vulnerabilities against package/version")
	} else {
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			aurora.Bold(coordinate.Coordinates),
			aurora.Gray("   No known vulnerabilities against package/version"))
	}
}

func logVulnerablePackage(noColor bool, idx int, packageCount int, coordinate types.Coordinate) (vulnerableCount bool) {
	if noColor {
		fmt.Println("------------------------------------------------------------")
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			coordinate.Coordinates+"  [Vulnerable]",
			"   "+strconv.Itoa(len(coordinate.Vulnerabilities)),
			"known vulnerabilities affecting installed version")

		for j := 0; j < len(coordinate.Vulnerabilities); j++ {
			if !coordinate.Vulnerabilities[j].Excluded {
				fmt.Println()
				vulnerability := coordinate.Vulnerabilities[j]
				fmt.Println(vulnerability.Title)
				fmt.Println(vulnerability.Description)
				fmt.Println()
				fmt.Println("ID:", vulnerability.Id)
				fmt.Println("Details:", vulnerability.Reference)
				vulnerableCount = true
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
				fmt.Println()
				vulnerability := coordinate.Vulnerabilities[j]
				fmt.Println(aurora.Bold(aurora.Red(vulnerability.Title)))
				fmt.Println(vulnerability.Description)
				fmt.Println()
				fmt.Println(aurora.Bold("ID:"), vulnerability.Id)
				fmt.Println(aurora.Bold("Details:"), vulnerability.Reference)
				vulnerableCount = true
			}
		}
	}

	return
}

// LogResults will given a number of expected results and the results themselves, log the
// results.
func LogResults(noColor bool, quiet bool, packageCount int, coordinates []types.Coordinate, exclusions []string) int {
	vulnerableCount := 0

	list := removeVulnerabilitiesIfExcluded(exclusions, coordinates)

	for i := 0; i < len(list); i++ {
		coordinate := list[i]
		idx := i + 1

		if !coordinate.Vulnerable {
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

func removeVulnerabilitiesIfExcluded(exclusions []string, coordinates []types.Coordinate) (list []types.Coordinate) {
	if len(exclusions) == 0 {
		return
	}

	for i, val := range coordinates {
		list = append(list, val)
		count := 0
		list[i].Vulnerabilities, count = markVulnerabilitesAsExcluded(exclusions, val.Vulnerabilities)
		if count > 0 {
			list[i].Vulnerable = true
		}
	}

	return
}

func markVulnerabilitesAsExcluded(exclusions []string, vulnerabilities []types.Vulnerability) (list []types.Vulnerability, vulnerableCount int) {
	for i, vuln := range vulnerabilities {
		list = append(list, vuln)
		list[i].Excluded = false
		vulnerableCount++
		for _, exclusion := range exclusions {
			if strings.Contains(vuln.Title, exclusion) {
				list[i].Excluded = true
				vulnerableCount--
			}
		}
	}

	return
}

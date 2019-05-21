//
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
//

package audit

import (
	"fmt"
	aurora "github.com/logrusorgru/aurora"
	"github.com/sonatype-nexus-community/nancy/types"
	"strconv"
	"strings"
)

func logPackage(noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	if noColor {
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			strings.Replace(coordinate.Coordinates, "pkg:", "", 1),
			"   No known vulnerabilities against package/version...")
	} else {
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			aurora.Bold(strings.Replace(coordinate.Coordinates, "pkg:", "", 1)),
			aurora.Gray("   No known vulnerabilities against package/version..."))
	}
}

func logVulnerablePackage(noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	if noColor {
		fmt.Println("------------------------------------------------------------")
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			strings.Replace(coordinate.Coordinates, "pkg:", "", 1)+"  [Vulnerable]",
			"   "+strconv.Itoa(len(coordinate.Vulnerabilities)),
			"known vulnerabilities affecting installed version")

		for j := 0; j < len(coordinate.Vulnerabilities); j++ {
			fmt.Println()
			vulnerability := coordinate.Vulnerabilities[j]
			fmt.Println(vulnerability.Title)
			fmt.Println(vulnerability.Description)
			fmt.Println()
			fmt.Println("ID:", vulnerability.Id)
			fmt.Println("Details:", vulnerability.Reference)
		}
	} else {
		fmt.Println("------------------------------------------------------------")
		fmt.Println("["+strconv.Itoa(idx)+"/"+strconv.Itoa(packageCount)+"]",
			aurora.Bold(aurora.Red(strings.Replace(coordinate.Coordinates, "pkg:", "", 1)+"  [Vulnerable]")),
			"   "+strconv.Itoa(len(coordinate.Vulnerabilities)),
			"known vulnerabilities affecting installed version")

		for j := 0; j < len(coordinate.Vulnerabilities); j++ {
			fmt.Println()
			vulnerability := coordinate.Vulnerabilities[j]
			fmt.Println(aurora.Bold(aurora.Red(vulnerability.Title)))
			fmt.Println(vulnerability.Description)
			fmt.Println()
			fmt.Println(aurora.Bold("ID:"), vulnerability.Id)
			fmt.Println(aurora.Bold("Details:"), vulnerability.Reference)
		}
	}
}

// LogResults will given a number of expected results and the results themselves, log the
// results.
func LogResults(noColor bool, packageCount int, coordinates []types.Coordinate) int {
	vulnerableCount := 0

	for i := 0; i < len(coordinates); i++ {
		coordinate := coordinates[i]
		idx := i + 1

		if len(coordinate.Vulnerabilities) == 0 {
			logPackage(noColor, idx, packageCount, coordinate)
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

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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
)

type AuditLogTextFormatter struct {
	Quiet   *bool
	NoColor *bool
}

func logPackage(sb *strings.Builder, noColor bool, quiet bool, idx int, packageCount int, coordinate types.Coordinate) {
	if !quiet {
		au := aurora.NewAurora(!noColor)
		packageLog := "[" + strconv.Itoa(idx) + "/" + strconv.Itoa(packageCount) + "]" +
			au.Bold(coordinate.Coordinates).String() +
			au.Gray(20-1, "   No known vulnerabilities against package/version\n").String()
		sb.WriteString(packageLog)
	}
}

func logInvalidSemVerWarning(sb *strings.Builder, noColor bool, quiet bool, invalidPurls []types.Coordinate) {
	if !quiet {
		packageCount := len(invalidPurls)
		if packageCount > 0 {
			warningMessage := "!!!!! WARNING !!!!!\nScanning cannot be completed on the following package(s) since they do not use semver.\n"
			au := aurora.NewAurora(!noColor)
			sb.WriteString(au.Red(warningMessage).String())

			for i := 0; i < len(invalidPurls); i++ {
				idx := i + 1
				purl := invalidPurls[i].Coordinates
				sb.WriteString("[" + strconv.Itoa(idx) + "/" + strconv.Itoa(packageCount) + "]" + au.Bold(purl).String() + "\n")
			}
			sb.WriteString("\n")
		}
	}
}

func logVulnerablePackage(sb *strings.Builder, noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	au := aurora.NewAurora(!noColor)
	sb.WriteString("------------------------------------------------------------\n")

	vulnLog := "[" + strconv.Itoa(idx) + "/" + strconv.Itoa(packageCount) + "]" +
		au.Bold(au.Red(coordinate.Coordinates+"  [Vulnerable]")).String() +
		"   " + strconv.Itoa(len(coordinate.Vulnerabilities)) +
		" known vulnerabilities affecting installed version\n"
	sb.WriteString(vulnLog)

	for j := 0; j < len(coordinate.Vulnerabilities); j++ {
		if !coordinate.Vulnerabilities[j].Excluded {
			sb.WriteString(fmt.Sprintf("\n%s\n%s\n\nID:%s\nDetails:%s\n",
				coordinate.Vulnerabilities[j].Title,
				coordinate.Vulnerabilities[j].Description,
				coordinate.Vulnerabilities[j].Id,
				coordinate.Vulnerabilities[j].Reference))
		}
	}
}

func (f *AuditLogTextFormatter) Format(entry *Entry) ([]byte, error) {
	auditedEntries := entry.Data["audited"]
	invalidEntries := entry.Data["invalid"]
	packageCount := entry.Data["num_audited"]
	numVulnerable := entry.Data["num_vulnerable"]
	buildVersion := entry.Data["version"]
	if auditedEntries != nil && invalidEntries != nil && packageCount != nil && numVulnerable != nil && buildVersion != nil {
		auditedEntries := entry.Data["audited"].([]types.Coordinate)
		invalidEntries := entry.Data["invalid"].([]types.Coordinate)
		packageCount := entry.Data["num_audited"].(int)
		numVulnerable := entry.Data["num_vulnerable"].(int)

		var sb strings.Builder

		logInvalidSemVerWarning(&sb, *f.NoColor, *f.Quiet, invalidEntries)
		for idx := 0; idx < len(auditedEntries); idx++ {
			coordinate := auditedEntries[idx]
			if !coordinate.IsVulnerable() && !coordinate.InvalidSemVer {
				logPackage(&sb, *f.NoColor, *f.Quiet, idx+1, packageCount, coordinate)
			}
			if coordinate.IsVulnerable() {
				logVulnerablePackage(&sb, *f.NoColor, idx+1, packageCount, coordinate)
			}
		}

		if !*f.Quiet {
			sb.WriteString("\n")
		}
		au := aurora.NewAurora(!*f.NoColor)
		sb.WriteString("Audited dependencies:" + strconv.Itoa(packageCount) + "," +
			"Vulnerable:" + au.Bold(au.Red(strconv.Itoa(numVulnerable))).String() + "\n")

		return []byte(sb.String()), nil
	} else {
		return nil, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else")
	}
}

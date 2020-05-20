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
	"sort"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/logrusorgru/aurora"
	"github.com/shopspring/decimal"
	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
)

var (
	nine, seven, four decimal.Decimal
)

func init() {
	nine, _ = decimal.NewFromString("9")
	seven, _ = decimal.NewFromString("7")
	four, _ = decimal.NewFromString("4")
}

type AuditLogTextFormatter struct {
	Quiet   *bool
	NoColor *bool
}

func logPackage(sb *strings.Builder, noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	au := aurora.NewAurora(!noColor)
	packageLog := "[" + strconv.Itoa(idx) + "/" + strconv.Itoa(packageCount) + "] " +
		au.Bold(au.Green(coordinate.Coordinates)).String() + "\n"
	sb.WriteString(packageLog)
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
				sb.WriteString("[" + strconv.Itoa(idx) + "/" + strconv.Itoa(packageCount) + "] " + au.Bold(purl).String() + "\n")
			}
			sb.WriteString("\n")
		}
	}
}

func logVulnerablePackage(sb *strings.Builder, noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	au := aurora.NewAurora(!noColor)

	sb.WriteString(fmt.Sprintf(
		"[%s/%s] %s\n%s \n",
		strconv.Itoa(idx),
		strconv.Itoa(packageCount),
		au.Bold(au.Red(coordinate.Coordinates)).String(),
		au.Red(strconv.Itoa(len(coordinate.Vulnerabilities))+" known vulnerabilities affecting installed version").String(),
	))

	sort.Slice(coordinate.Vulnerabilities, func(i, j int) bool {
		return coordinate.Vulnerabilities[i].CvssScore.GreaterThan(coordinate.Vulnerabilities[j].CvssScore)
	})

	for _, v := range coordinate.Vulnerabilities {
		if !v.Excluded {
			t := table.NewWriter()
			t.SetStyle(table.StyleBold)
			t.SetTitle(printColorBasedOnCvssScore(v.CvssScore, v.Title, noColor))
			t.AppendRow([]interface{}{"Description", text.WrapSoft(v.Description, 75)})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"OSS Index ID", v.Id})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"CVSS Score", fmt.Sprintf("%s/10 (%s)", v.CvssScore, scoreAssessment(v.CvssScore))})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"CVSS Vector", v.CvssVector})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"Link for more info", v.Reference})
			sb.WriteString(t.Render() + "\n")
		}
	}
}

func printColorBasedOnCvssScore(score decimal.Decimal, text string, noColor bool) string {
	au := aurora.NewAurora(!noColor)
	if score.GreaterThanOrEqual(nine) {
		return au.Red(au.Bold(text)).String()
	}
	if score.GreaterThanOrEqual(seven) {
		return au.Red(text).String()
	}
	if score.GreaterThanOrEqual(four) {
		return au.Yellow(text).String()
	}
	return au.Green(text).String()
}

func scoreAssessment(score decimal.Decimal) string {
	if score.GreaterThanOrEqual(nine) {
		return "Critical"
	}
	if score.GreaterThanOrEqual(seven) {
		return "High"
	}
	if score.GreaterThanOrEqual(four) {
		return "Medium"
	}
	return "Low"
}

func groupAndPrint(vulnerable []types.Coordinate, nonVulnerable []types.Coordinate, quiet bool, noColor bool, sb *strings.Builder) {
	if !quiet {
		sb.WriteString("\nNon Vulnerable Packages\n\n")
		for k, v := range nonVulnerable {
			logPackage(sb, noColor, k+1, len(nonVulnerable), v)
		}
	}
	if len(vulnerable) > 0 {
		sb.WriteString("\nVulnerable Packages\n\n")
		for k, v := range vulnerable {
			logVulnerablePackage(sb, noColor, k+1, len(vulnerable), v)
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
		nonVulnerablePackages, vulnerablePackages := splitPackages(auditedEntries)

		groupAndPrint(vulnerablePackages, nonVulnerablePackages, *f.Quiet, *f.NoColor, &sb)

		au := aurora.NewAurora(!*f.NoColor)
		t := table.NewWriter()
		t.SetStyle(table.StyleBold)
		t.SetTitle("Summary")
		t.AppendRow([]interface{}{"Audited Dependencies", strconv.Itoa(packageCount)})
		t.AppendSeparator()
		t.AppendRow([]interface{}{"Vulnerable Dependencies", au.Bold(au.Red(strconv.Itoa(numVulnerable)))})
		sb.WriteString(t.Render())

		return []byte(sb.String()), nil
	}
	return nil, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else")
}

func splitPackages(entries []types.Coordinate) (nonVulnerable []types.Coordinate, vulnerable []types.Coordinate) {
	for _, v := range entries {
		if v.IsVulnerable() {
			vulnerable = append(vulnerable, v)
		} else {
			nonVulnerable = append(nonVulnerable, v)
		}
	}
	return
}

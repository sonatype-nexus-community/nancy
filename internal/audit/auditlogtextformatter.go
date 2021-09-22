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
	"text/tabwriter"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/logrusorgru/aurora"
	"github.com/shopspring/decimal"
	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
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
	Quiet   bool
	NoColor bool
}

func logPackage(sb *strings.Builder, noColor bool, coordinate types.Coordinate) {
	au := aurora.NewAurora(!noColor)

	sb.WriteString(
		fmt.Sprintf("%s\n",
			au.Bold(au.Green(coordinate.Coordinates)).String(),
		),
	)
}

func logInvalidSemVerWarning(sb *strings.Builder, noColor bool, quiet bool, invalidPurls []types.Coordinate) {
	if !quiet {
		if len(invalidPurls) > 0 {
			au := aurora.NewAurora(!noColor)
			sb.WriteString(au.Red("!!!!! WARNING !!!!!\nScanning cannot be completed on the following package(s) since they do not use semver.\n").String())

			for _, v := range invalidPurls {
				sb.WriteString(
					fmt.Sprintf("%s\n",
						au.Bold(v.Coordinates).String(),
					),
				)
			}

			sb.WriteString("\n")
		}
	}
}

func logVulnerablePackage(sb *strings.Builder, noColor bool, coordinate types.Coordinate) {
	au := aurora.NewAurora(!noColor)
	sb.WriteString(fmt.Sprintf(
		"%s\n%s \n",
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
			t.AppendRow([]interface{}{"OSS Index ID", v.ID})
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
		sb.WriteString("\n")
		for _, v := range nonVulnerable {
			logPackage(sb, noColor, v)
		}
		sb.WriteString(fmt.Sprintf("\n%d Non Vulnerable Packages\n\n", len(nonVulnerable)))
	}
	if len(vulnerable) > 0 {
		for _, v := range vulnerable {
			logVulnerablePackage(sb, noColor, v)
		}
		sb.WriteString(fmt.Sprintf("\n%d Vulnerable Packages\n\n", len(vulnerable)))
	}
}

func (f AuditLogTextFormatter) Format(entry *Entry) ([]byte, error) {
	auditedEntries := entry.Data["audited"]
	invalidEntries := entry.Data["invalid"]
	excludedEntries := entry.Data["excluded"]
	packageCount := entry.Data["num_audited"]
	numVulnerable := entry.Data["num_vulnerable"]
	numExcluded := entry.Data["num_exclusions"]
	buildVersion := entry.Data["version"]
	if auditedEntries != nil && invalidEntries != nil && excludedEntries != nil && packageCount != nil && numVulnerable != nil && numExcluded != nil && buildVersion != nil {
		auditedEntries := entry.Data["audited"].([]types.Coordinate)
		invalidEntries := entry.Data["invalid"].([]types.Coordinate)
		packageCount := entry.Data["num_audited"].(int)
		numVulnerable := entry.Data["num_vulnerable"].(int)
		numExcluded := entry.Data["num_exclusions"].(int)

		var sb strings.Builder

		w := tabwriter.NewWriter(&sb, 9, 3, 0, '\t', 0)
		_ = w.Flush()

		logInvalidSemVerWarning(&sb, f.NoColor, f.Quiet, invalidEntries)
		nonVulnerablePackages, vulnerablePackages := splitPackages(auditedEntries)

		groupAndPrint(vulnerablePackages, nonVulnerablePackages, f.Quiet, f.NoColor, &sb)

		au := aurora.NewAurora(!f.NoColor)
		t := table.NewWriter()
		t.SetStyle(table.StyleBold)
		t.SetTitle("Summary")
		t.AppendRow([]interface{}{"Audited Dependencies", strconv.Itoa(packageCount)})
		t.AppendSeparator()
		t.AppendRow([]interface{}{"Vulnerable Dependencies", au.Bold(au.Red(strconv.Itoa(numVulnerable)))})
    if numExcluded > 0 {
      t.AppendSeparator()
      t.AppendRow([]interface{}{"Ignored Vulnerabilities", au.Bold(au.Yellow(strconv.Itoa(numExcluded)))})
    }
		sb.WriteString(t.Render())
		sb.WriteString("\n")

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

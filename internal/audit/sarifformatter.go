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
	"bytes"
	"strings"
	"text/template"

	"github.com/owenrumney/go-sarif/sarif"
	"github.com/shopspring/decimal"
	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
)

type SarifFormatter struct {
	UsingDep bool
}

func (f SarifFormatter) Format(entry *Entry) ([]byte, error) {
	auditedEntries := entry.Data["audited"].([]types.Coordinate)
	invalidEntries := entry.Data["invalid"].([]types.Coordinate)
	buildVersion := entry.Data["version"].(string)

	report, err := sarif.New(sarif.Version210)
	if err != nil {
		return nil, err
	}
	run := sarif.NewRun("nancy", "https://ossindex.sonatype.org/integration/nancy")
	run.Tool.Driver.WithVersion(buildVersion)
	report.AddRun(run)

	buildSarifForAuditedEntries(f.UsingDep, auditedEntries, run)
	buildSarifForInvalidEntries(f.UsingDep, invalidEntries, run)

	var buff bytes.Buffer
	err = report.PrettyWrite(&buff)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

const invalidCoordinateId = "InvalidCoordinate"

func buildSarifForInvalidEntries(usingDep bool, invalidEntries []types.Coordinate, run *sarif.Run) {
	for _, coordinate := range invalidEntries {
		rule := run.AddRule(invalidCoordinateId).
			WithDescription("Scanning cannot be completed on the following package(s) since they do not use semver.")

		message := sarif.NewMessage().WithText(convertToGoModSyntax(coordinate.Coordinates))

		ruleResult := run.AddResult(rule.ID)

		artifactLocation := determineLocation(usingDep)
		fingerPrints := map[string]interface{}{
			"sonatypeId": invalidCoordinateId,
			"coordinate": coordinate.Coordinates,
		}
		ruleResult.
			WithLocation(sarif.NewLocation().
				WithPhysicalLocation(sarif.NewPhysicalLocation().
					WithArtifactLocation(artifactLocation).
					WithRegion(sarif.NewRegion().WithStartLine(1))),
			).
			WithPartialFingerPrints(fingerPrints).
			WithLevel("note").WithMessage(message)
	}
}

func determineLocation(dep bool) *sarif.ArtifactLocation {
	artifactLocation := sarif.NewArtifactLocation().WithUri("go.mod")
	if dep == true {
		artifactLocation = sarif.NewArtifactLocation().WithUri("Gopkg.lock")
	}
	return artifactLocation
}

func buildSarifForAuditedEntries(usingDep bool, auditedEntries []types.Coordinate, run *sarif.Run) {
	for _, coordinate := range auditedEntries {
		if coordinate.IsVulnerable() {
			for _, vuln := range coordinate.Vulnerabilities {
				data := map[string]interface{}{
					"Score":            vuln.CvssScore,
					"SonatypeSeverity": scoreAssessment(vuln.CvssScore),
					"URL":              vuln.Reference,
				}
				var ruleHelpMarkdown bytes.Buffer
				tmpl, _ := template.New("test").Parse(`
CVSS Score of **{{.Score}} ({{.SonatypeSeverity}})**
Find more details here: 
{{.URL}}`)
				tmpl.Execute(&ruleHelpMarkdown, data)
				var ruleHelpText = vuln.Cve + " " + vuln.Description + " " + vuln.Reference

				rule := run.AddRule(vuln.ID).
					WithDescription(vuln.Title).
					WithFullDescription(sarif.NewMultiformatMessageString(vuln.Description))

				rule.Help = sarif.NewMultiformatMessageString(ruleHelpText).WithMarkdown(ruleHelpMarkdown.String())

				message := sarif.NewMessage().WithText(convertToGoModSyntax(coordinate.Coordinates))
				level := sarifScoreAssessment(vuln.CvssScore)

				ruleResult := run.AddResult(rule.ID)

				artifactLocation := determineLocation(usingDep)
				fingerPrints := map[string]interface{}{
					"sonatypeId": vuln.ID,
					"coordinate": coordinate.Coordinates,
				}
				ruleResult.
					WithLocation(sarif.NewLocation().
						WithPhysicalLocation(sarif.NewPhysicalLocation().
							WithArtifactLocation(artifactLocation).
							WithRegion(sarif.NewRegion().WithStartLine(1))),
					).
					WithPartialFingerPrints(fingerPrints).
					WithLevel(level).WithMessage(message)
			}
		}
	}
}

func convertToGoModSyntax(coordinate string) string{
	gomodCoor := strings.ReplaceAll(coordinate, "pkg:golang/", "")
	gomodCoor = strings.ReplaceAll(gomodCoor, "@", " ")
	return gomodCoor
}

func sarifScoreAssessment(score decimal.Decimal) string {
	if score.GreaterThanOrEqual(seven) {
		return "error"
	}
	return "warning"
}

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
	"testing"

	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/stretchr/testify/assert"
)

func TestSarifOutput(t *testing.T) {
	data := map[string]interface{}{
		"audited": []types.Coordinate{
			{Coordinates: "good1"},
			{Coordinates: "pkg:golang/vuln1@1.0.0", Vulnerabilities: createVulnerabilities(1)},
		},
		"invalid": []types.Coordinate{},
		"num_audited":    2,
		"num_vulnerable": 1,
		"version":        "development",
	}
	entry := Entry{Data: data}

	formatter := SarifFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)
	actual := string(logMessage)
	assert.Equal(t, `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "nancy",
          "version": "development",
          "informationUri": "https://ossindex.sonatype.org/integration/nancy",
          "rules": [
            {
              "id": "123",
              "shortDescription": {
                "text": "Vulnerability"
              },
              "fullDescription": {
                "text": "Description"
              },
              "help": {
                "text": "CVE-123 Description Reference",
                "markdown": "\nCVSS Score of **7.88 (High)**\nFind more details here: \nReference"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "123",
          "level": "error",
          "message": {
            "text": "vuln1 1.0.0"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "go.mod"
                },
                "region": {
                  "startLine": 1
                }
              }
            }
          ],
          "partialFingerprints": {
            "coordinate": "pkg:golang/vuln1@1.0.0",
            "sonatypeId": "123"
          }
        }
      ]
    }
  ]
}`, actual)
}

func TestSarifOutputWhenDep(t *testing.T) {
	t.Fail()
}

func TestSerifOutputWithWarningsLevelVulns(t *testing.T) {
	t.Fail()
}

func TestSerifOutputWithInvalidEntries(t *testing.T) {
	t.Fail()
}

func TestSerifOutputWithMultipleVulnerabilities(t *testing.T) {
	t.Fail()
}
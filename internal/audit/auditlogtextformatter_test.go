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
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/stretchr/testify/assert"
)

func TestFormatterErrorsIfEntryNotValid(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := AuditLogTextFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, logMessage)
	assert.NotNil(t, e)
	assert.Equal(t, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else"), e)
}

func verifyFormatterSummaryLoudness(t *testing.T, quiet bool) {
	data := map[string]interface{}{
		"audited":        []types.Coordinate{},
		"invalid":        []types.Coordinate{},
		"num_audited":    0,
		"num_vulnerable": 0,
		"version":        0,
	}
	entry := Entry{Data: data}

	formatter := AuditLogTextFormatter{Quiet: quiet, NoColor: false}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)

	expectedSummary := "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n┃ Summary                     ┃\n┣━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━┫\n┃ Audited Dependencies    ┃ 0 ┃\n┣━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━┫\n┃ Vulnerable Dependencies ┃ \x1b[1;31m0\x1b[0m ┃\n┗━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━┛\n"
	if !quiet {
		expectedSummary = "\n\n0 Non Vulnerable Packages\n\n" + expectedSummary
	}
	assert.Equal(t, expectedSummary, string(logMessage))
}

func TestFormatterSummary(t *testing.T) {
	verifyFormatterSummaryLoudness(t, false)
	verifyFormatterSummaryLoudness(t, true)
}

func TestFormatterLogInvalidSemVerWarning(t *testing.T) {
	entry := Entry{Data: map[string]interface{}{
		"audited": []types.Coordinate{},
		"invalid": []types.Coordinate{
			{
				Coordinates: "MyInvalidCoords",
			},
		},
		"num_audited":    0,
		"num_vulnerable": 0,
		"version":        0,
	}}
	formatter := AuditLogTextFormatter{NoColor: true}
	logMessage, e := formatter.Format(&entry)
	assert.Nil(t, e)

	expectedMessagePrefix := `!!!!! WARNING !!!!!
Scanning cannot be completed on the following package(s) since they do not use semver.
[1/1]	MyInvalidCoords

`
	assert.True(t, strings.HasPrefix(string(logMessage), expectedMessagePrefix))
}

func TestPrintColorBasedOnCvssScore(t *testing.T) {
	assert.Equal(t, "\x1b[1;31m\x1b[0m", printColorBasedOnCvssScore(decimal.New(9, 0), "", false))
	assert.Equal(t, "", printColorBasedOnCvssScore(decimal.New(9, 0), "", true))

	assert.Equal(t, "\x1b[31m\x1b[0m", printColorBasedOnCvssScore(decimal.New(7, 0), "", false))
	assert.Equal(t, "", printColorBasedOnCvssScore(decimal.New(7, 0), "", true))

	assert.Equal(t, "\x1b[33m\x1b[0m", printColorBasedOnCvssScore(decimal.New(4, 0), "", false))
	assert.Equal(t, "", printColorBasedOnCvssScore(decimal.New(4, 0), "", true))

	assert.Equal(t, "\x1b[32m\x1b[0m", printColorBasedOnCvssScore(decimal.New(0, 0), "", false))
	assert.Equal(t, "", printColorBasedOnCvssScore(decimal.New(0, 0), "", true))
}

func TestScoreAssessment(t *testing.T) {
	assert.Equal(t, "Critical", scoreAssessment(decimal.New(9, 0)))
	assert.Equal(t, "High", scoreAssessment(decimal.New(7, 0)))
	assert.Equal(t, "Medium", scoreAssessment(decimal.New(4, 0)))
	assert.Equal(t, "Low", scoreAssessment(decimal.New(0, 0)))
}

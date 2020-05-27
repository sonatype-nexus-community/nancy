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
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/stretchr/testify/assert"
)

func TestFormatterErrorsIfEntryNotValid(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := logrus.Entry{Data: data}

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
	entry := logrus.Entry{Data: data}

	formatter := AuditLogTextFormatter{Quiet: &quiet, NoColor: new(bool)}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)

	expectedSummary := "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓\n┃ Summary                     ┃\n┣━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━┫\n┃ Audited Dependencies    ┃ 0 ┃\n┣━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━┫\n┃ Vulnerable Dependencies ┃ \x1b[1;31m0\x1b[0m ┃\n┗━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━┛"
	if !quiet {
		expectedSummary = "\nNon Vulnerable Packages\n\n" + expectedSummary
	}
	assert.Equal(t, expectedSummary, string(logMessage))
}

func TestFormatterSummary(t *testing.T) {
	verifyFormatterSummaryLoudness(t, false)
	verifyFormatterSummaryLoudness(t, true)
}

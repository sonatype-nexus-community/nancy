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
	"encoding/csv"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/sonatype-nexus-community/nancy/internal/customerrors"

	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
)

type CsvFormatter struct {
	Quiet bool
}

func (f CsvFormatter) Format(entry *Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	auditedEntries := entry.Data["audited"]
	invalidEntries := entry.Data["invalid"]
	packageCount := entry.Data["num_audited"]
	numVulnerable := entry.Data["num_vulnerable"]
	excludedCount := entry.Data["num_exclusions"]
	buildVersion := entry.Data["version"]

	if auditedEntries != nil && invalidEntries != nil && packageCount != nil && numVulnerable != nil && excludedCount != nil && buildVersion != nil {
		auditedEntries := entry.Data["audited"].([]types.Coordinate)
		invalidEntries := entry.Data["invalid"].([]types.Coordinate)
		packageCount := entry.Data["num_audited"].(int)
		numVulnerable := entry.Data["num_vulnerable"].(int)
		excludedCount := entry.Data["num_exclusions"].(int)
		buildVersion := entry.Data["version"].(string)

		var summaryHeader = []string{"Audited Count", "Vulnerable Count", "Ignored Vulnerabilities", "Build Version"}
		var invalidHeader = []string{"Count", "Package", "Reason"}
		var auditedHeader = []string{"Count", "Package", "Is Vulnerable", "Num Vulnerabilities", "Vulnerabilities"}
		var summaryRow = []string{strconv.Itoa(packageCount), strconv.Itoa(numVulnerable), strconv.Itoa(excludedCount), buildVersion}

		var buf bytes.Buffer
		w := csv.NewWriter(&buf)

		var err error
		if err = f.write(w, []string{"Summary"}); err != nil {
			return nil, err
		}
		if err = f.write(w, summaryHeader); err != nil {
			return nil, err
		}
		if err = f.write(w, summaryRow); err != nil {
			return nil, err
		}

		if !f.Quiet {
			invalidCount := len(invalidEntries)
			if invalidCount > 0 {
				if err = f.write(w, []string{""}); err != nil {
					return nil, err
				}
				if err = f.write(w, []string{"Invalid Package(s)"}); err != nil {
					return nil, err
				}
				if err = f.write(w, invalidHeader); err != nil {
					return nil, err
				}
				for i := 1; i <= invalidCount; i++ {
					invalidEntry := invalidEntries[i-1]
					if err = f.write(w, []string{"[" + strconv.Itoa(i) + "/" + strconv.Itoa(invalidCount) + "]", invalidEntry.Coordinates, "Does not use SemVer"}); err != nil {
						return nil, err
					}
				}
			}
		}

		if !f.Quiet || numVulnerable > 0 {
			if err = f.write(w, []string{""}); err != nil {
				return nil, err
			}
			if err = f.write(w, []string{"Audited Package(s)"}); err != nil {
				return nil, err
			}
			if err = f.write(w, auditedHeader); err != nil {
				return nil, err
			}
		}
		for i := 1; i <= len(auditedEntries); i++ {
			auditEntry := auditedEntries[i-1]
			if auditEntry.IsVulnerable() || !f.Quiet {
				jsonVulns, _ := json.Marshal(auditEntry.Vulnerabilities)
				if err = f.write(w, []string{"[" + strconv.Itoa(i) + "/" + strconv.Itoa(packageCount) + "]", auditEntry.Coordinates, strconv.FormatBool(auditEntry.IsVulnerable()), strconv.Itoa(len(auditEntry.Vulnerabilities)), string(jsonVulns)}); err != nil {
					return nil, err
				}
			}
		}

		w.Flush()

		return buf.Bytes(), nil
	}
	return nil, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else")
}

func (f CsvFormatter) write(w *csv.Writer, line []string) error {
	if err := w.Write(line); err != nil {
		return customerrors.NewErrorExitPrintHelp(err, "Failed to write data to csv")
	}
	return nil
}

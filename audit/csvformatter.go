// Copyright 2020 Sonatype Inc.
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
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strconv"

	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
)

type CsvFormatter struct {
	Quiet *bool
}

func (f *CsvFormatter) Format(entry *Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
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
		buildVersion := entry.Data["version"].(string)

		var summaryHeader = []string{"Audited Count", "Vulnerable Count", "Build Version"}
		var invalidHeader = []string{"Count", "Package", "Reason"}
		var auditedHeader = []string{"Count", "Package", "Is Vulnerable", "Num Vulnerabilities", "Vulnerabilities"}
		var summaryRow = []string{strconv.Itoa(packageCount), strconv.Itoa(numVulnerable), buildVersion}

		var buf bytes.Buffer
		w := csv.NewWriter(&buf)

		f.write(w, []string{"Summary"})
		f.write(w, summaryHeader)
		f.write(w, summaryRow)

		if !*f.Quiet {
			invalidCount := len(invalidEntries)
			if invalidCount > 0 {
				f.write(w, []string{""})
				f.write(w, []string{"Invalid Package(s)"})
				f.write(w, invalidHeader)
				for i := 1; i <= invalidCount; i++ {
					invalidEntry := invalidEntries[i-1]
					f.write(w, []string{"[" + strconv.Itoa(i) + "/" + strconv.Itoa(invalidCount) + "]", invalidEntry.Coordinates, "Does not use SemVer"})
				}
			}
		}

		if !*f.Quiet || numVulnerable > 0 {
			f.write(w, []string{""})
			f.write(w, []string{"Audited Package(s)"})
			f.write(w, auditedHeader)
		}
		for i := 1; i <= len(auditedEntries); i++ {
			auditEntry := auditedEntries[i-1]
			if auditEntry.IsVulnerable() || !*f.Quiet {
				jsonVulns, _ := json.Marshal(auditEntry.Vulnerabilities)
				f.write(w, []string{"[" + strconv.Itoa(i) + "/" + strconv.Itoa(packageCount) + "]", auditEntry.Coordinates, strconv.FormatBool(auditEntry.IsVulnerable()), strconv.Itoa(len(auditEntry.Vulnerabilities)), string(jsonVulns)})
			}
		}

		w.Flush()

		return buf.Bytes(), nil
	} else {
		return nil, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else")
	}

}

func (f *CsvFormatter) write(w *csv.Writer, line []string) {
	err := w.Write(line)
	customerrors.Check(err, "Failed to write data to csv")
}

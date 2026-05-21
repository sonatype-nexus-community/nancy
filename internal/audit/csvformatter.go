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

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/internal/ossindex"
)

type CsvFormatter struct {
	Quiet bool
}

func (f CsvFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	auditedEntries := entry.Data["audited"]
	invalidEntries := entry.Data["invalid"]
	packageCount := entry.Data["num_audited"]
	numVulnerable := entry.Data["num_vulnerable"]
	excludedCount := entry.Data["num_exclusions"]
	buildVersion := entry.Data["version"]

	if !isEntryValid(auditedEntries, invalidEntries, packageCount, numVulnerable, excludedCount, buildVersion) {
		return nil, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else")
	}

	audited := entry.Data["audited"].([]ossindex.Coordinate)
	invalid := entry.Data["invalid"].([]ossindex.Coordinate)
	pkgCount := entry.Data["num_audited"].(int)
	numVuln := entry.Data["num_vulnerable"].(int)
	excCount := entry.Data["num_exclusions"].(int)
	version := entry.Data["version"].(string)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := f.writeSummarySection(w, pkgCount, numVuln, excCount, version); err != nil {
		return nil, err
	}
	if err := f.writeInvalidSection(w, invalid); err != nil {
		return nil, err
	}
	if err := f.writeAuditedSection(w, audited, pkgCount, numVuln); err != nil {
		return nil, err
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (f CsvFormatter) writeSummarySection(w *csv.Writer, pkgCount, numVuln, excCount int, version string) error {
	summaryHeader := []string{"Audited Count", "Vulnerable Count", "Ignored Vulnerabilities", "Build Version"}
	summaryRow := []string{strconv.Itoa(pkgCount), strconv.Itoa(numVuln), strconv.Itoa(excCount), version}
	if err := f.write(w, []string{"Summary"}); err != nil {
		return err
	}
	if err := f.write(w, summaryHeader); err != nil {
		return err
	}
	return f.write(w, summaryRow)
}

func (f CsvFormatter) writeInvalidSection(w *csv.Writer, invalid []ossindex.Coordinate) error {
	if f.Quiet || len(invalid) == 0 {
		return nil
	}
	count := len(invalid)
	if err := f.write(w, []string{""}); err != nil {
		return err
	}
	if err := f.write(w, []string{"Invalid Package(s)"}); err != nil {
		return err
	}
	if err := f.write(w, []string{"Count", "Package", "Reason"}); err != nil {
		return err
	}
	for i, e := range invalid {
		row := []string{"[" + strconv.Itoa(i+1) + "/" + strconv.Itoa(count) + "]", e.Coordinates, "Does not use SemVer"}
		if err := f.write(w, row); err != nil {
			return err
		}
	}
	return nil
}

func (f CsvFormatter) writeAuditedSection(w *csv.Writer, audited []ossindex.Coordinate, pkgCount, numVuln int) error {
	if !f.Quiet || numVuln > 0 {
		if err := f.write(w, []string{""}); err != nil {
			return err
		}
		if err := f.write(w, []string{"Audited Package(s)"}); err != nil {
			return err
		}
		if err := f.write(w, []string{"Count", "Package", "Is Vulnerable", "Num Vulnerabilities", "Vulnerabilities"}); err != nil {
			return err
		}
	}
	for i, e := range audited {
		if !e.IsVulnerable() && f.Quiet {
			continue
		}
		jsonVulns, _ := json.Marshal(e.Vulnerabilities)
		row := []string{
			"[" + strconv.Itoa(i+1) + "/" + strconv.Itoa(pkgCount) + "]",
			e.Coordinates,
			strconv.FormatBool(e.IsVulnerable()),
			strconv.Itoa(len(e.Vulnerabilities)),
			string(jsonVulns),
		}
		if err := f.write(w, row); err != nil {
			return err
		}
	}
	return nil
}

func (f CsvFormatter) write(w *csv.Writer, line []string) error {
	if err := w.Write(line); err != nil {
		return customerrors.NewErrorExitPrintHelp(err, "Failed to write data to csv")
	}
	return nil
}

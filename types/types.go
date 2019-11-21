// Copyright 2018 Sonatype Inc.
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
package types

import (
	"fmt"
	"strings"

	decimal "github.com/shopspring/decimal"
)

type Vulnerability struct {
	Id          string
	Title       string
	Description string
	CvssScore   decimal.Decimal
	CvssVector  string
	Cve         string
	Reference   string
	Excluded    bool
}

//Mark the given vulnerability as excluded if it appears in the exclusion list
func (v *Vulnerability) maybeExcludeVulnerability(exclusions []string) {
	for _, ex := range exclusions {
		if v.Cve == ex || v.Id == ex {
			v.Excluded = true
		}
	}
}

type Coordinate struct {
	Coordinates     string
	Reference       string
	Vulnerabilities []Vulnerability
	InvalidSemVer    bool
}

func (c Coordinate) IsVulnerable() bool {
	for _, v := range c.Vulnerabilities {
		if !v.Excluded {
			return true
		}
	}
	return false
}

//Mark Excluded=true for all Vulnerabilities of the given Coordinate if their Title is in the list of exclusions
func (c *Coordinate) ExcludeVulnerabilities(exclusions []string) {
	for i, _ := range c.Vulnerabilities {
		c.Vulnerabilities[i].maybeExcludeVulnerability(exclusions)
	}
}

type AuditRequest struct {
	Coordinates []string `json:"coordinates"`
}

type Projects struct {
	Name    string
	Version string
}
type ProjectList struct {
	Projects []Projects
}

type CveListFlag struct {
	Cves []string
}

func (cve *CveListFlag) String() string {
	return fmt.Sprint(cve.Cves)
}

func (cve *CveListFlag) Set(value string) error {
	if len(cve.Cves) > 0 {
		return fmt.Errorf("The CVE Exclude Flag is already set")
	}
	cve.Cves = strings.Split(strings.ReplaceAll(value, " ", ""), ",")

	return nil
}

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
	"encoding/xml"
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
	InvalidSemVer   bool
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
	for i := range c.Vulnerabilities {
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

// IQ Types

// StatusURLResult is a struct to let the consumer know what the response from Nexus IQ Server was
type StatusURLResult struct {
	PolicyAction  string `json:"policyAction"`
	ReportHTMLURL string `json:"reportHtmlUrl"`
	IsError       bool   `json:"isError"`
	ErrorMessage  string `json:"errorMessage"`
}

// CycloneDX Types

// Sbom is a struct to begin assembling a minimal SBOM
type Sbom struct {
	XMLName    xml.Name   `xml:"bom"`
	Xmlns      string     `xml:"xmlns,attr"`
	XMLNSV     string     `xml:"xmlns:v,attr"`
	Version    string     `xml:"version,attr"`
	Components Components `xml:"components"`
}

// Components is a struct to list the components in a SBOM
type Components struct {
	Component []Component `xml:"component"`
}

// Component is a struct to list the properties of a component in a SBOM
type Component struct {
	Type            string          `xml:"type,attr"`
	BomRef          string          `xml:"bom-ref,attr"`
	Name            string          `xml:"name"`
	Version         string          `xml:"version"`
	Purl            string          `xml:"purl"`
	Vulnerabilities Vulnerabilities `xml:"v:vulnerabilities,omitempty"`
}

type Vulnerabilities struct {
	Vulnerability []SbomVulnerability `xml:"v:vulnerability,omitempty"`
}

type SbomVulnerability struct {
	Ref         string    `xml:"ref,attr,omitempty"`
	ID          string    `xml:"v:id"`
	Source      Source    `xml:"v:source"`
	Ratings     []Ratings `xml:"v:ratings"`
	Description string    `xml:"v:description"`
}

type Ratings struct {
	Rating Rating `xml:"v:rating"`
}

type Rating struct {
	Score    Score  `xml:"v:score,omitempty"`
	Severity string `xml:"v:severity,omitempty"`
	Method   string `xml:"v:method,omitempty"`
	Vector   string `xml:"v:vector,omitempty"`
}

type Score struct {
	Base           decimal.Decimal `xml:"v:base,omitempty"`
	Impact         string          `xml:"v:impact,omitempty"`
	Exploitability string          `xml:"v:exploitability,omitempty"`
}

type Source struct {
	Name string `xml:"name,attr"`
	URL  string `xml:"v:url"`
}

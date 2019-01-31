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
}

type Coordinate struct {
	Coordinates     string
	Reference       string
	Vulnerabilities []Vulnerability
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

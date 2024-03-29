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

package types

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Configuration struct {
	DBCachePath     string
	Version         bool
	NoColor         bool
	Quiet           bool
	Loud            bool
	CleanCache      bool
	CveList         CveListFlag
	Path            string
	Formatter       logrus.Formatter
	LogLevel        int
	Username        string
	Token           string
	Help            bool
	IQUsername      string
	IQToken         string
	IQStage         string
	IQApplication   string
	IQServer        string
	MaxRetries      int
	SkipUpdateCheck bool
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
		//goland:noinspection GoErrorStringFormat
		return fmt.Errorf("the CVE Exclude Flag is already set")
	}
	cve.Cves = strings.Split(strings.ReplaceAll(value, " ", ""), ",")

	return nil
}

func (cve *CveListFlag) Type() string { return "CveListFlag" }

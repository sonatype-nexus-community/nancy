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
	"time"

	"github.com/sirupsen/logrus"
)

// Helpful constants to pull strings we use more than once out of code
const (
	OssIndexDirName        = ".ossindex"
	OssIndexConfigFileName = ".oss-index-config"
	IQServerDirName        = ".iqserver"
	IQServerConfigFileName = ".iq-server-config"
)

type GoListModule struct {
	Path      string        // module path
	Version   string        // module version
	Versions  []string      // available module versions (with -versions)
	Replace   *GoListModule // replaced by this module
	Time      *time.Time    // time version was created
	Update    *GoListModule // available update, if any (with -u)
	Main      bool          // is this the main module?
	Indirect  bool          // is this module only an indirect dependency of main module?
	Dir       string        // directory holding files for this module, if any
	GoMod     string        // path to go.mod file for this module, if any
	GoVersion string        // go version used in module
}

type Configuration struct {
	Version       bool
	NoColor       bool
	Quiet         bool
	Loud          bool
	CleanCache    bool
	CveList       CveListFlag
	Path          string
	Formatter     logrus.Formatter
	LogLevel      int
	Username      string
	Token         string
	Help          bool
	IQUsername    string
	IQToken       string
	IQStage       string
	IQApplication string
	IQServer      string
	MaxRetries    int
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
		return fmt.Errorf("The CVE Exclude Flag is already set")
	}
	cve.Cves = strings.Split(strings.ReplaceAll(value, " ", ""), ",")

	return nil
}

func (cve *CveListFlag) Type() string { return "CveListFlag" }

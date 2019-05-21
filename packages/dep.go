//
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
//

package packages

import (
	"fmt"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"os"
	"strings"
)

// Dep is an implementation of Packages interface
type Dep struct {
	ProjectList types.ProjectList
	GopkgPath   string
}

// ExtractPurlsFromManifest will convert Gopkg projects to Package URLs
func (d Dep) ExtractPurlsFromManifest() []string {
	var purls []string
	for _, s := range d.ProjectList.Projects {
		var version string
		version = strings.Replace(s.Version, "v", "", -1)

		if len(version) > 0 { // There must be a version we can use
			var purl = "pkg:" + convertGopkgNameToPurl(s.Name) + "@" + version
			purls = append(purls, purl)
		}
	}
	return purls
}

// CheckExistenceOfManifest will see if a Gopkg exists at the given path
func (d Dep) CheckExistenceOfManifest() bool {
	if _, err := os.Stat(d.GopkgPath); os.IsNotExist(err) {
		customerrors.Check(err, fmt.Sprint("No Gopkg found at path: "+d.GopkgPath))
	}
	return true
}

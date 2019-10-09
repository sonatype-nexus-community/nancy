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
package packages

import (
	"github.com/Masterminds/semver"
	"github.com/golang/dep"
	"strings"
)

func ExtractPurlsUsingDep(project dep.Project) ([]string, []string) {
	lockedProjects := project.Lock.P;
	var purls []string
	var invalidPurls []string
	for _, lockedProject := range lockedProjects {
		var version string
		i := lockedProject.Version().String()

		version = strings.Replace(i, "v", "", -1)

		if len(version) > 0 { // There must be a version we can use
			name := lockedProject.Ident().String()
			packageName := convertGopkgNameToPurl(string(name))
			var purl = "pkg:" + packageName + "@" + version

			_, err := semver.NewVersion(version)
			if err != nil {
				invalidPurls = append(invalidPurls, purl)
			}else{
				purls = append(purls, purl)
			}
		}
	}
	return purls, invalidPurls
}

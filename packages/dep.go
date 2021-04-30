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

package packages

import (
	"strings"

	"github.com/Masterminds/semver"
	"github.com/golang/dep"
	"github.com/sonatype-nexus-community/nancy/types"
)

func ExtractPurlsUsingDep(project *dep.Project) (deps map[string]types.Dependency) {
	deps = make(map[string]types.Dependency)
	lockedProjects := project.Lock.P

	for _, lockedProject := range lockedProjects {
		var version string
		i := lockedProject.Version().String()

		version = strings.Replace(i, "v", "", -1)

		// There must be a version we can use
		if len(version) > 0 {
			name := lockedProject.Ident().String()
			packageName := convertGopkgNameToPurl(name)
			var purl = "pkg:" + packageName + "@" + version

			_, err := semver.NewVersion(version)
			if err != nil {
				dep := types.Dependency{PackageManager: "dep", Name: packageName, Version: version, Valid: false}

				deps[purl] = dep
			} else {
				dep := types.Dependency{PackageManager: "dep", Name: packageName, Version: version, Valid: true}

				deps[purl] = dep
			}
		}
	}

	return
}

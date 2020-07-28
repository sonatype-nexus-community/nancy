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

	"github.com/sonatype-nexus-community/nancy/types"
)

type Mod struct {
	ProjectList types.ProjectList
	GoSumPath   string
}

func (m Mod) ExtractPurlsFromManifest() (purls []string) {
	for _, s := range m.ProjectList.Projects {
		if len(s.Version) > 0 { // There must be a version we can use
			// OSS Index no likey v before version, IQ does though, comment left so I will never forget.
			// go-sona-types library now takes care of querying both ossi and iq with reformatted purls as needed (to v or not to v).
			version := strings.Replace(s.Version, "v", "", -1)
			version = strings.Replace(version, "+incompatible", "", -1)
			var purl = "pkg:" + convertGopkgNameToPurl(s.Name) + "@" + version
			purls = append(purls, purl)
		}
	}

	purls = removeDuplicates(purls)

	return
}

func removeDuplicates(purls []string) (dedupedPurls []string) {
	encountered := map[string]bool{}

	for _, v := range purls {
		if encountered[v] {
			// LogLady.WithField("dep", v).Debug("Found duplicate dependency, eliminating it")
		} else {
			// LogLady.WithField("dep", v).Debug("Unique dependency, adding it")
			encountered[v] = true
			dedupedPurls = append(dedupedPurls, v)
		}
	}

	return
}

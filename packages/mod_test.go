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
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/sonatype-nexus-community/nancy/types"
)

const versionFormatDateHash = "v0.0.0-20201221181555-eec23a3978ad"
const versionFormatIncompatible = "v2.0.3+incompatible"

// Simulate calling parse.GopkgLock()
func getProjectList() (projectList types.ProjectList) {
	appendProject("github.com/AndreasBriese/bbloom", "", &projectList)
	appendProject("gopkg.in/BurntSushi/toml", "v0.3.1", &projectList)
	appendProject("github.com/dgraph-io/badger", "v1.5.4", &projectList)
	appendProject("github.com/dgryski/go-farm", "", &projectList)
	appendProject("github.com/golang/protobuf", "v1.2.0", &projectList)
	appendProject("github.com/logrusorgru/aurora", "", &projectList)
	appendProject("github.com/pkg/errors", "v0.8.0", &projectList)
	appendProject("github.com/shopspring/decimal", "1.1.0", &projectList)
	appendProject("golang.org/x/net", "", &projectList)
	appendProject("golang.org/x/sys", "", &projectList)
	appendProject("golang/golang.org/x/crypto", versionFormatDateHash, &projectList)
	appendProject("github.com/logrusorgru/aurora", versionFormatIncompatible, &projectList)

	return projectList
}

func getProjectListDuplicates() (projectList types.ProjectList) {
	appendProject("github.com/AndreasBriese/bbloom", "", &projectList)
	appendProject("gopkg.in/BurntSushi/toml", "v0.3.1", &projectList)
	appendProject("github.com/dgraph-io/badger", "v1.5.4", &projectList)
	appendProject("github.com/dgraph-io/badger", "v1.5.4", &projectList)
	appendProject("gopkg.in/dgraph-io/badger", "v1.5.4", &projectList)
	appendProject("github.com/dgryski/go-farm", "", &projectList)
	appendProject("github.com/golang/protobuf", "v1.2.0", &projectList)
	appendProject("github.com/golang/protobuf", "v1.2.0", &projectList)
	appendProject("github.com/golang/protobuf", "v1.2.0", &projectList)
	appendProject("github.com/logrusorgru/aurora", "", &projectList)
	appendProject("github.com/pkg/errors", "v0.8.0", &projectList)
	appendProject("github.com/shopspring/decimal", "1.1.0", &projectList)
	appendProject("golang.org/x/net", "", &projectList)
	appendProject("golang.org/x/sys", "", &projectList)

	return projectList
}

func appendProject(name string, version string, projectList *types.ProjectList) {
	projectList.Projects = append(projectList.Projects, types.Projects{Name: name, Version: version})
}

func TestModExtractPurlsFromManifest(t *testing.T) {
	mod := Mod{}
	mod.ProjectList = getProjectList()

	result := mod.ExtractPurlsFromManifest()
	if len(result) != 7 {
		t.Error(result)
	}

	// verify version format with date and hashcode is not altered
	assert.Equal(t, "pkg:golang/golang/golang.org/x/crypto@"+versionFormatDateHash, result[5])

	// verify version format with '+incompatible' has that string removed
	assert.Equal(t, "pkg:golang/github.com/logrusorgru/aurora@v2.0.3", result[6])
}

func TestModExtractPurlsFromManifestDuplicates(t *testing.T) {
	mod := Mod{}
	mod.ProjectList = getProjectListDuplicates()

	result := mod.ExtractPurlsFromManifest()
	if len(result) != 5 {
		t.Error(result)
	}
}

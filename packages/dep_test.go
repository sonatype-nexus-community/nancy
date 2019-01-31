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
	"github.com/sonatype-nexus-community/nancy/types"
	"testing"
)

var testPkgName = "Gopkg.lock"

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

	return projectList
}

func appendProject(name string, version string, projectList *types.ProjectList) {
	projectList.Projects = append(projectList.Projects, types.Projects{Name: name, Version: version})
}

func TestCheckExistenceOfManifestExists(t *testing.T) {
	dep := Dep{}
	dep.GopkgPath = testPkgName
	exists := dep.CheckExistenceOfManifest()

	if !exists {
		t.Errorf("Expected existence of %s", testPkgName)
	}
}

func TestExtractPurlsFromManifest(t *testing.T) {
	var err error
	dep := Dep{}
	dep.GopkgPath = testPkgName
	dep.ProjectList = getProjectList()
	if err != nil {
		t.Error(err)
	}

	result := dep.ExtractPurlsFromManifest()
	if len(result) != 5 {
		t.Error(result)
	}
}

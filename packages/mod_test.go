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
	"testing"

	"github.com/sonatype-nexus-community/nancy/types"
)

type dependency struct {
	name    string
	version string
}

// Simulate calling parse.GopkgLock()
func getDependencyList() (dependencies []types.Dependency) {
	deps := []dependency{
		{name: "github.com/AndreasBriese/bbloom", version: ""},
		{name: "gopkg.in/BurntSushi/toml", version: "v0.3.1"},
		{name: "github.com/dgraph-io/badger", version: "v1.5.4"},
		{name: "github.com/dgryski/go-farm", version: ""},
		{name: "github.com/golang/protobuf", version: "v1.2.0"},
		{name: "github.com/logrusorgru/aurora", version: ""},
		{name: "github.com/pkg/errors", version: "0.8.0"},
		{name: "github.com/shopspring/decimal", version: "1.1.0"},
		{name: "golang.org/x/net", version: ""},
		{name: "golang.org/x/sys", version: ""},
	}

	return appendDependencies(deps)
}

func appendDependencies(deps []dependency) (dependencies []types.Dependency) {
	for _, v := range deps {
		dependencies = append(dependencies, types.Dependency{Name: v.name, Version: v.version})
	}

	return
}

func getProjectListDuplicates() (dependencies []types.Dependency) {
	deps := []dependency{
		{name: "github.com/AndreasBriese/bbloom", version: ""},
		{name: "gopkg.in/BurntSushi/toml", version: "v0.3.1"},
		{name: "github.com/dgraph-io/badger", version: "v1.5.4"},
		{name: "github.com/dgraph-io/badger", version: "v1.5.4"},
		{name: "gopkg.in/dgraph-io/badger", version: "v1.5.4"},
		{name: "github.com/dgryski/go-farm", version: ""},
		{name: "github.com/golang/protobuf", version: "v1.2.0"},
		{name: "github.com/golang/protobuf", version: "v1.2.0"},
		{name: "github.com/golang/protobuf", version: "v1.2.0"},
		{name: "github.com/logrusorgru/aurora", version: ""},
		{name: "github.com/pkg/errors", version: "v0.8.0"},
		{name: "github.com/shopspring/decimal", version: "1.1.0"},
		{name: "golang.org/x/net", version: ""},
		{name: "golang.org/x/sys", version: ""},
	}

	return appendDependencies(deps)
}

func TestModExtractGoModPurls(t *testing.T) {
	deps := getDependencyList()

	result := ExtractGoModPurls(deps)
	if len(result) != 5 {
		t.Error(result)
	}
}

func TestModExtractPurlsFromManifestDuplicates(t *testing.T) {
	deps := getProjectListDuplicates()

	result := ExtractGoModPurls(deps)
	if len(result) != 5 {
		t.Error(result)
	}
}

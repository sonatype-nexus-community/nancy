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

package parse

import (
	"os"
	"strings"
	"testing"
)

func TestGoListJson(t *testing.T) {
	goListJSONFile, err := os.Open("testdata/golistjson.out")
	if err != nil {
		t.Error(err)
	}

	deps, err := GoListAgnostic(goListJSONFile)
	if err != nil {
		t.Error(err)
	}
	if len(deps) != 48 {
		t.Errorf("Unsuccessfully parsed go list -json -m all output, 48 dependencies were expected, but %d encountered", len(deps))
	}
}

func TestGoListAgnostic(t *testing.T) {
	goListFile, err := os.Open("testdata/golist.out")
	if err != nil {
		t.Error(err)
	}

	deps, err := GoListAgnostic(goListFile)
	if err != nil {
		t.Error(err)
	}
	if len(deps) != 48 {
		t.Errorf("Unsuccessfully parsed go list -m all output, 48 dependencies were expected, but %d encountered", len(deps))
	}
}

func TestGoListJsonReplace(t *testing.T) {
	goListJSONReplaceFile, err := os.Open("testdata/golistjsonreplace.out")
	if err != nil {
		t.Error(err)
	}

	deps, err := GoListAgnostic(goListJSONReplaceFile)
	if err != nil {
		t.Error(err)
	}
	if len(deps) != 134 {
		t.Errorf("Unsuccessfully parsed go list -m all output, 134 dependencies were expected, but %d encountered", len(deps))
	}

	purl := "pkg:golang/github.com/gorilla/websocket@1.4.2"

	if val, ok := deps[purl]; ok {
		if val.Version != "v1.4.2" {
			t.Errorf("Version expected to be v1.4.2, but encountered %s", val.Version)
		}
	} else {
		t.Error(deps)
		t.Errorf("Did not find my purl, where is my purl?! %+v", val)
	}
}

func TestGoListReplace(t *testing.T) {
	goListReplaceFile, err := os.Open("testdata/golistreplace.out")
	if err != nil {
		t.Error(err)
	}

	deps, err := GoListAgnostic(goListReplaceFile)
	if err != nil {
		t.Error(err)
	}
	if len(deps) != 1 {
		t.Errorf("Unsuccessfully parsed go list -m all output, 1 dependency was expected, but %d encountered", len(deps))
	}

	purl := "pkg:golang/github.com/gorilla/websocket@1.4.2"

	if val, ok := deps[purl]; ok {
		if val.Version != "v1.4.2" {
			t.Errorf("Version expected to be v1.4.2, but encountered %s", val.Version)
		}
	} else {
		t.Error(deps)
		t.Errorf("Did not find my purl, where is my purl?! %+v", val)
	}
}

func TestGoListAllWithSelfReference(t *testing.T) {
	goListSelfReferenceOutput, err := os.Open("testdata/self-reference.out")
	if err != nil {
		t.Error(err)
	}

	deps, err := GoListAgnostic(goListSelfReferenceOutput)
	if err != nil {
		t.Error(err)
	}

	if len(deps) != 517 {
		t.Error(deps)
	}

	kratosClientPurl := "pkg:golang/github.com/ory/kratos-client-go@0.5.4-alpha.1"
	kratosClientCorpPurl := "pkg:golang/github.com/ory/kratos/corp@0.0.0-00010101000000-000000000000"

	if _, ok := deps[kratosClientPurl]; ok {
		t.Error("Project with name github.com/ory/kratos-client-go should be ignored b/c it references a submodule")
	}

	if _, ok := deps[kratosClientCorpPurl]; ok {
		t.Error("Project with name github.com/ory/kratos/corp should be ignored b/c it references a submodule")
	}
}

func TestGoListAll(t *testing.T) {
	goListMAllOutput := `github.com/sonatype-nexus-community/nancy
github.com/AndreasBriese/bbloom v0.0.0-20180913140656-343706a395b7
github.com/BurntSushi/toml v0.3.1
github.com/davecgh/go-spew v1.1.0
github.com/dgraph-io/badger v1.5.5-0.20181004181505-439fd464b155
github.com/dgryski/go-farm v0.0.0-20180109070241-2de33835d102
github.com/dustin/go-humanize v1.0.0
github.com/golang/protobuf v1.2.0
github.com/logrusorgru/aurora v0.0.0-20181002194514-a7b3b318ed4e
github.com/pkg/errors v0.8.0
github.com/pmezard/go-difflib v1.0.0
github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
github.com/stretchr/objx v0.1.0
github.com/stretchr/testify v1.3.0
golang.org/x/net v0.0.0-20181220203305-927f97764cc3
golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
golang.org/x/sys v0.0.0-20181228144115-9a3f9b0469bb`

	deps, err := GoListAgnostic(strings.NewReader(goListMAllOutput))
	if err != nil {
		t.Error(err)
	}

	if len(deps) != 16 {
		t.Error(deps)
	}
}

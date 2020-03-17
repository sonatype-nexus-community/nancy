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
package parse

import (
	"bufio"
	"strings"
	"testing"
)

func TestGoSum(t *testing.T) {
	deps, err := GoSum("testdata/go.sum")
	if err != nil {
		t.Error(err)
	}

	if len(deps.Projects) != 10 {
		t.Error(deps)
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
	scanner := bufio.NewScanner(strings.NewReader(goListMAllOutput))

	deps, err := GoList(scanner)
	if err != nil {
		t.Error(err)
	}

	if len(deps.Projects) != 16 {
		t.Error(deps)
	}
}

func TestGoSumError(t *testing.T) {
	_, err := GoSum("../testdata/parse/go.notsum")
	if err == nil {
		t.Error(err)
	}
}

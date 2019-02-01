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
	"github.com/BurntSushi/toml"
	"github.com/sonatype-nexus-community/nancy/types"
	"os"
	"strings"
)

// GopkgLock parses the Gopkg file and returns an error if unsuccessful
func GopkgLock(path string) (deps types.ProjectList, err error) {
	// Load the dependency data
	_, err = toml.DecodeFile(path, &deps)
	if err != nil {
		return deps, err
	}
	return deps, nil
}

// GoSum parses the go.sum file and returns an error if unsuccessful
func GoSum(path string) (deps types.ProjectList, err error) {
	file, err := os.Open(path)
	if err != nil {
		return deps, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), " ")
		if !strings.HasSuffix(s[1], "/go.mod") {
			deps.Projects = append(deps.Projects, types.Projects{Name: s[0], Version: s[1]})
		}
	}
	return deps, nil
}

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
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/sonatype-nexus-community/nancy/types"
)

var goListDependencyCriteria = func(s []string) bool {
	return len(s) > 1
}

func GoList(stdIn *bufio.Scanner) (deps types.ProjectList, err error) {
	for stdIn.Scan() {
		parseSpaceSeparatedDependency(stdIn, &deps, goListDependencyCriteria)
	}

	return deps, nil
}

// GoListAgnostic will take an io.Reader that is likely the os.StdIn, and parse it as
// a map of string keys to interfaces. If a "Module" key exists, I know I'm in
// 'go list -deps' town. If it doesn't, I look for a "Path" key, which is 'go list -m all' town.
// It returns either an error, or a deps of types.ProjectList
func GoListAgnostic(stdIn io.Reader) (deps types.ProjectList, err error) {
	// stdIn should never be massive, so taking this approach over reading from a stream
	// multiple times
	johnnyFiveNeedInput, err := ioutil.ReadAll(stdIn)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(johnnyFiveNeedInput)))

	for {
		var mod map[string]interface{}

		decodeErr := decoder.Decode(&mod)

		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			err = decodeErr
			break
		}

		if module, ok := mod["Module"].(map[string]interface{}); ok {
			if version, ok := module["Version"].(string); ok {
				deps.Projects = append(deps.Projects, types.Projects{Version: version, Name: module["Path"].(string)})
			}

			continue
		}

		if path, ok := mod["Path"].(string); ok {
			if replace, ok := mod["Replace"].(map[string]interface{}); ok {
				if version, ok := replace["Version"].(string); ok {
					deps.Projects = append(deps.Projects, types.Projects{Version: version, Name: path})

					continue
				}

				// No version found in replace block
				continue
			}

			if version, ok := mod["Version"].(string); ok {
				deps.Projects = append(deps.Projects, types.Projects{Version: version, Name: path})

				continue
			}

			continue
		}
	}

	if err != nil {
		err = nil
		scanner := bufio.NewScanner(strings.NewReader(string(johnnyFiveNeedInput)))
		deps, err = GoList(scanner)
		if err != nil {
			return
		}
	}

	return
}

func parseSpaceSeparatedDependency(scanner *bufio.Scanner, deps *types.ProjectList, criteria func(s []string) bool) {
	text := scanner.Text()
	rewrite := strings.Split(text, "=>")

	if len(rewrite) == 2 {
		v2 := strings.Split(strings.TrimSpace(rewrite[1]), " ")
		addProjectDep(criteria, v2, deps)
	} else {
		s := strings.Split(text, " ")
		addProjectDep(criteria, s, deps)
	}
}

func addProjectDep(criteria func(s []string) bool, s []string, deps *types.ProjectList) {
	if criteria(s) {
		if len(s) > 3 {
			deps.Projects = append(deps.Projects, types.Projects{Name: s[0], Version: s[4]})
		} else {
			deps.Projects = append(deps.Projects, types.Projects{Name: s[0], Version: s[1]})
		}
	}
}

type NoVersionError struct {
	err error
}

func (n *NoVersionError) Error() string {
	return n.err.Error()
}

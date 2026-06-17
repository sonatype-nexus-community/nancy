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
	"fmt"
	"io"
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

// DefaultMaxStdInBytes is the default cap on the amount of 'go list' output
// read from stdin when no explicit limit is provided.
const DefaultMaxStdInBytes int64 = 100 * 1024 * 1024 // 100 MB

// GoListAgnostic will take an io.Reader that is likely the os.StdIn, and parse it as
// a map of string keys to interfaces. If a "Module" key exists, I know I'm in
// 'go list -json -deps' town. If it doesn't, I look for a "Path" key, which is 'go list -json -m all' town.
// If I find nothing, then I try and parse it like I'm parsing a non json output
// It returns either an error, or a deps of types.ProjectList
func GoListAgnostic(stdIn io.Reader, maxStdinBytes int64) (deps types.ProjectList, err error) {
	if maxStdinBytes <= 0 {
		maxStdinBytes = DefaultMaxStdInBytes
	}

	// stdIn should never be massive, so taking this approach over reading from a stream
	// multiple times

	limited := io.LimitReader(stdIn, maxStdinBytes+1)
	johnnyFiveNeedInput, err := io.ReadAll(limited)
	if err != nil {
		return
	}
	if int64(len(johnnyFiveNeedInput)) > maxStdinBytes {
		err = fmt.Errorf("stdin input exceeded %d MB limit; ensure you are piping valid 'go list' output",
			maxStdinBytes/(1024*1024))
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(johnnyFiveNeedInput)))

	for {
		var mod map[string]interface{}
		if decodeErr := decoder.Decode(&mod); decodeErr == io.EOF {
			break
		} else if decodeErr != nil {
			err = decodeErr
			break
		}

		if _, hasModule := mod["Module"]; hasModule {
			if p, ok := extractModuleDep(mod); ok {
				deps.Projects = append(deps.Projects, p)
			}
			continue
		}
		if p, ok := extractListDep(mod); ok {
			deps.Projects = append(deps.Projects, p)
		}
	}

	if err != nil {
		scanner := bufio.NewScanner(strings.NewReader(string(johnnyFiveNeedInput)))
		deps, err = GoList(scanner)
	}

	return
}

// extractModuleDep parses a dep from `go list -json -deps` output (has "Module" key).
func extractModuleDep(mod map[string]interface{}) (types.Projects, bool) {
	module, ok := mod["Module"].(map[string]interface{})
	if !ok {
		return types.Projects{}, false
	}
	if replace, ok := module["Replace"].(map[string]interface{}); ok {
		if version, ok := replace["Version"].(string); ok {
			return types.Projects{Version: version, Name: replace["Path"].(string)}, true
		}
		// Replace block exists but has no version — fall through to module version
	}
	if version, ok := module["Version"].(string); ok {
		return types.Projects{Version: version, Name: module["Path"].(string)}, true
	}
	return types.Projects{}, false
}

// extractListDep parses a dep from `go list -json -m all` output (has "Path" key, no "Module").
func extractListDep(mod map[string]interface{}) (types.Projects, bool) {
	path, ok := mod["Path"].(string)
	if !ok {
		return types.Projects{}, false
	}
	if replace, ok := mod["Replace"].(map[string]interface{}); ok {
		replacePath, pathOk := replace["Path"].(string)
		version, versionOk := replace["Version"].(string)
		if pathOk && versionOk {
			return types.Projects{Version: version, Name: replacePath}, true
		}
		return types.Projects{}, false
	}
	if version, ok := mod["Version"].(string); ok {
		return types.Projects{Version: version, Name: path}, true
	}
	return types.Projects{}, false
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

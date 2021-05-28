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

// GoListAgnostic will take a io.Reader that is likely the os.StdIn, try to parse it as
// if `go list -json -m all` was ran, and then try to reparse it as if `go list -m all`
// was ran instead. It returns either an error, or a deps of types.ProjectList
func GoListAgnostic(stdIn io.Reader) (deps types.ProjectList, err error) {
	// stdIn should never be massive, so taking this approach over reading from a stream
	// multiple times
	johnnyFiveNeedInput, err := ioutil.ReadAll(stdIn)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(johnnyFiveNeedInput)))

	for {
		project, err := decodeModToProjectList(decoder)

		if err == io.EOF {
			err = nil
			break
		}

		if _, ok := err.(*NoVersionError); ok {

			// w didn't find a modulse, check for depenecies (i.e. a "go list -deps")
			project, err := decodeDepToProjectList(decoder)

			if err == io.EOF {
				err = nil
				break
			}

			if _, ok := err.(*NoVersionError); ok {
				continue
			}

			deps.Projects = append(deps.Projects, project)

			continue
		}

		deps.Projects = append(deps.Projects, project)
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

func decodeModToProjectList(decoder *json.Decoder) (project types.Projects, err error) {
	var mod types.GoListModule
	err = decoder.Decode(&mod)

	if err == io.EOF {
		return
	}
	if err != nil {
		return
	}

	project, err = modToProjectList(mod)

	return project, err
}

func modToProjectList(mod types.GoListModule) (dep types.Projects, err error) {
	if mod.Replace != nil {
		if mod.Replace.Version == "" {
			err = &NoVersionError{err: fmt.Errorf("no version found for mod")}
			return
		}
		dep.Name = mod.Replace.Path
		dep.Version = mod.Replace.Version
		return
	}
	if mod.Version == "" {
		err = &NoVersionError{err: fmt.Errorf("no version found for mod")}
		return
	}
	dep.Name = mod.Path
	dep.Version = mod.Version
	return
}

func decodeDepToProjectList(decoder *json.Decoder) (project types.Projects, err error) {
	var mod types.GoListDependecy
	err = decoder.Decode(&mod)

	if err == io.EOF {
		return
	}
	if err != nil {
		return
	}

	project, err = depToProjectList(mod)

	return project, err
}

func depToProjectList(dep types.GoListDependecy) (project types.Projects, err error) {
	if dep.Module != nil {
		if dep.Module.Version == "" {
			err = &NoVersionError{err: fmt.Errorf("no version found for dep")}
			return
		}

		project.Name = dep.Module.Path
		project.Version = dep.Module.Version
		return
	}

	return
}

func parseSpaceSeparatedDependency(scanner *bufio.Scanner, deps *types.ProjectList, criteria func(s []string) bool) {
	text := scanner.Text()
	rewrite := strings.Split(text, "=>")

	if len(rewrite) == 2 {
		v2 := strings.Split(strings.TrimSpace(rewrite[1]), " ")
		addProjectDep(criteria, v2, deps)
	}else{
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

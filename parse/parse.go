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

	"github.com/sonatype-nexus-community/nancy/packages"
	"github.com/sonatype-nexus-community/nancy/types"
)

var goListDependencyCriteria = func(s []string) bool {
	return len(s) > 1
}

// GoListAgnostic will take a io.Reader that is likely the os.StdIn, try to parse it as
// if `go list -json -m all` was ran, and then try to reparse it as if `go list -m all`
// was ran instead. It returns either an error, or a deps of types.ProjectList
func GoListAgnostic(stdIn io.Reader) (dependencies map[string]types.Dependency, err error) {
	dependencies = make(map[string]types.Dependency)

	// stdIn should never be massive, so taking this approach over reading from a stream
	// multiple times
	johnnyFiveNeedInput, err := ioutil.ReadAll(stdIn)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(strings.NewReader(string(johnnyFiveNeedInput)))

	for {
		var mod types.GoListModule

		err = decoder.Decode(&mod)
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}

		dep, err := modToProjectList(mod)
		if _, ok := err.(*NoVersionError); ok {
			continue
		}
		dependencies[packages.GimmeAPURL(dep.Name, dep.Version)] = dep
	}

	if err != nil {
		err = nil
		scanner := bufio.NewScanner(strings.NewReader(string(johnnyFiveNeedInput)))

		for scanner.Scan() {
			dep := parseSpaceSeparatedDependency(scanner, goListDependencyCriteria)
			if dep != nil {
				dependencies[packages.GimmeAPURL(dep.Name, dep.Version)] = *dep
			}
		}

		if err != nil {
			return
		}
	}

	return
}

func modToProjectList(mod types.GoListModule) (dep types.Dependency, err error) {
	if mod.Replace != nil {
		if mod.Replace.Version == "" {
			err = &NoVersionError{err: fmt.Errorf("no version found for mod")}
			return
		}

		dep = populateMod(mod, mod.Replace.Path, mod.Replace.Version)

		return
	}
	if mod.Version == "" {
		err = &NoVersionError{err: fmt.Errorf("no version found for mod")}
		return
	}

	dep = populateMod(mod, mod.Path, mod.Version)

	return
}

func populateMod(mod types.GoListModule, name string, version string) (dep types.Dependency) {
	if mod.Update != nil {
		dep.Update = &types.ProjectUpdate{
			Path:    mod.Update.Path,
			Version: mod.Update.Version,
			Time:    *mod.Update.Time,
		}
	}
	dep.Valid = true
	dep.Name = name
	dep.Version = version

	return
}

func parseSpaceSeparatedDependency(scanner *bufio.Scanner, criteria func(s []string) bool) (dep *types.Dependency) {
	text := scanner.Text()
	rewrite := strings.Split(text, "=>")

	if len(rewrite) == 2 {
		v2 := strings.Split(strings.TrimSpace(rewrite[1]), " ")
		return addProjectDep(criteria, v2)
	}

	s := strings.Split(text, " ")
	return addProjectDep(criteria, s)
}

func addProjectDep(criteria func(s []string) bool, s []string) (dep *types.Dependency) {
	if criteria(s) {
		if len(s) > 3 {
			return &types.Dependency{Valid: true, Name: s[0], Version: s[4]}
		} else {
			return &types.Dependency{Valid: true, Name: s[0], Version: s[1]}
		}
	}

	return nil
}

type NoVersionError struct {
	err error
}

func (n *NoVersionError) Error() string {
	return n.err.Error()
}

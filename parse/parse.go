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
	"fmt"
	"os"
	"strings"

	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
)

var goModDependencyCriteria = func(s []string) bool {
	return len(s) > 1 && !strings.HasSuffix(s[1], "/go.mod")
}
var goListDependencyCriteria = func(s []string) bool {
	return len(s) > 1
}

func GoList(stdIn *bufio.Scanner) (deps types.ProjectList, err error) {
	for stdIn.Scan() {
		parseSpaceSeparatedDependency(stdIn, &deps, goListDependencyCriteria)
	}

	LogLady.Info("Beginning to dedupe list")
	deps = removeDuplicates(&deps)

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
		parseSpaceSeparatedDependency(scanner, &deps, goModDependencyCriteria)
	}

	LogLady.Info("Beginning to dedupe list")
	deps = removeDuplicates(&deps)

	return deps, nil
}

func parseSpaceSeparatedDependency(scanner *bufio.Scanner, deps *types.ProjectList, criteria func(s []string) bool) {
	text := scanner.Text()
	s := strings.Split(text, " ")
	if criteria(s) {
		deps.Projects = append(deps.Projects, types.Projects{Name: s[0], Version: s[1]})
	}
}

func removeDuplicates(deps *types.ProjectList) (newDeps types.ProjectList) {
	encountered := map[string]bool{}

	for _, v := range deps.Projects {
		if encountered[fmt.Sprintf("%s-%s", v.Name, v.Version)] == true {
			LogLady.WithField("dep", v).Debug("Found duplicate dependency, eliminating it")
		} else {
			LogLady.WithField("dep", v).Debug("Unique dependency, adding it")
			encountered[fmt.Sprintf("%s-%s", v.Name, v.Version)] = true
			newDeps.Projects = append(newDeps.Projects, v)
		}
	}

	return
}

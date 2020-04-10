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
	"fmt"
	"github.com/Flaque/filet"
	"github.com/golang/dep"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestExtractPurlsFromManifestUsingDep(t *testing.T) {
	path, projectDir, err := doGoPathSimulatedSetup(t)
	require.NoError(t, err)
	defer filet.CleanUp(t)

	ctx := dep.Ctx{
		WorkingDir: projectDir,
		GOPATHs:    []string{path},
		Out:        log.New(os.Stdout, "", 0),
		Err:        log.New(os.Stderr, "", 0),
	}
	project, err2 := ctx.LoadProject()
	require.NoError(t, err2)

	purls, invalidPurls := ExtractPurlsUsingDep(project)
	if len(invalidPurls) != 7 {
		t.Errorf("Number of invalid purls not as expected. Expected : %d, Got %d", 7, len(purls))
	}
	if len(purls) != 7 {
		t.Errorf("Number of purls not as expected. Expected : %d, Got %d", 7, len(purls))
	}
	assertPurlFound("pkg:golang/github.com/Masterminds/semver@2.x", invalidPurls, t)
	assertPurlFound("pkg:golang/github.com/armon/go-radix@master", invalidPurls, t)
	assertPurlFound("pkg:golang/github.com/nightlyone/lockfile@master", invalidPurls, t)
	assertPurlFound("pkg:golang/github.com/sdboyer/constext@master", invalidPurls, t)
	assertPurlFound("pkg:golang/golang.org/x/net@master", invalidPurls, t)
	assertPurlFound("pkg:golang/golang.org/x/sync@master", invalidPurls, t)
	assertPurlFound("pkg:golang/golang.org/x/sys@master", invalidPurls, t)

	assertPurlFound("pkg:golang/github.com/go-yaml/yaml@2", purls, t)
	assertPurlFound("pkg:golang/github.com/Masterminds/vcs@1.11.1", purls, t)
	assertPurlFound("pkg:golang/github.com/boltdb/bolt@1.3.1", purls, t)
	assertPurlFound("pkg:golang/github.com/golang/protobuf@1.0.0", purls, t)
	assertPurlFound("pkg:golang/github.com/jmank88/nuts@0.3.0", purls, t)
	assertPurlFound("pkg:golang/github.com/pelletier/go-toml@1.2.0", purls, t)
	assertPurlFound("pkg:golang/github.com/pkg/errors@0.8.0", purls, t)
}

func assertPurlFound(expectedPurl string, result []string, t *testing.T) {
	if !inArray(expectedPurl, result) {
		t.Errorf("Expected purl %s not found. List of purls was %s", expectedPurl, result)
	}
}

func doGoPathSimulatedSetup(t *testing.T) (string, string, error) {
	dir, _ := os.Getwd()
	path := filet.TmpDir(t, dir)
	fakeGoPath := fmt.Sprint(path, "/src")
	e := os.Mkdir(fakeGoPath, os.ModePerm)
	if e != nil {
		t.Error(e)
	}
	projectDir := fmt.Sprint(fakeGoPath, "/projectname")
	e = os.Mkdir(projectDir, os.ModePerm)
	if e != nil {
		t.Error(e)
	}
	lockBytes, e := ioutil.ReadFile("testdata/Gopkg.lock")
	if e != nil {
		t.Error(e)
	}
	e = ioutil.WriteFile(fmt.Sprint(projectDir, "/Gopkg.lock"), lockBytes, 0644)
	if e != nil {
		t.Error(e)
	}

	tomlBytes, e := ioutil.ReadFile("testdata/Gopkg.toml")
	if e != nil {
		t.Error(e)
	}
	e = ioutil.WriteFile(fmt.Sprint(projectDir, "/Gopkg.toml"), tomlBytes, 0644)
	if e != nil {
		t.Error(e)
	}

	files, e := ioutil.ReadDir(projectDir)
	if e != nil {
		t.Error(e)
	}
	for _, file := range files {
		fmt.Println(file.Name())
	}
	return path, projectDir, e
}

func inArray(val string, array []string) (exists bool) {
	exists = false

	for _, v := range array {
		if val == v {
			exists = true
			return
		}
	}

	return
}

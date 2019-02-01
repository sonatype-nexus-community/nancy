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
	"testing"
)

var testGoSumName = "go.sum"

func TestModCheckExistenceOfManifestExists(t *testing.T) {
	mod := Mod{}
	mod.GoSumPath = testGoSumName
	exists := mod.CheckExistenceOfManifest()

	if !exists {
		t.Errorf("Expected existence of %s", testGoSumName)
	}
}

func TestModExtractPurlsFromManifest(t *testing.T) {
	var err error
	mod := Mod{}
	mod.GoSumPath = testGoSumName
	mod.ProjectList = getProjectList()
	if err != nil {
		t.Error(err)
	}

	result := mod.ExtractPurlsFromManifest()
	if len(result) != 5 {
		t.Error(result)
	}
}

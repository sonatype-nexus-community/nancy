// Copyright 2020 Sonatype Inc.
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

// Package cyclonedx has definitions and functions for processing golang purls into a minimal CycloneDX 1.1 Sbom
package cyclonedx

import "testing"
import "github.com/sonatype-nexus-community/nancy/types"
import "github.com/stretchr/testify/assert"

const expectedResult = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n <bom xmlns=\"http://cyclonedx.org/schema/bom/1.1\" xmlns:v=\"http://cyclonedx.org/schema/ext/vulnerability/1.0\" version=\"1\">\n      <components>\n           <component type=\"library\" bom-ref=\"pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2\">\n                <name>crypto</name>\n                <version>v0.0.0-20190308221718-c2843e01d9a2</version>\n                <purl>pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2</purl>\n                <v:vulnerabilities></v:vulnerabilities>\n           </component>\n           <component type=\"library\" bom-ref=\"pkg:golang/github.com/go-yaml/yaml@v2.2.2\">\n                <name>yaml</name>\n                <version>v2.2.2</version>\n                <purl>pkg:golang/github.com/go-yaml/yaml@v2.2.2</purl>\n                <v:vulnerabilities></v:vulnerabilities>\n           </component>\n      </components>\n </bom>"

func TestProcessPurlsIntoSBOM(t *testing.T) {
	results := []types.Coordinate{}
	results = append(results, types.Coordinate{
		Coordinates:     "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
		Reference:       "https://ossindex.sonatype.org/component/pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
		Vulnerabilities: []types.Vulnerability{},
	})
	results = append(results, types.Coordinate{
		Coordinates:     "pkg:golang/github.com/go-yaml/yaml@v2.2.2",
		Reference:       "https://ossindex.sonatype.org/component/pkg:golang/github.com/go-yaml/yaml@v2.2.2",
		Vulnerabilities: []types.Vulnerability{},
	})
	result := ProcessPurlsIntoSBOM(results)

	assert.Equal(t, result, expectedResult)
}

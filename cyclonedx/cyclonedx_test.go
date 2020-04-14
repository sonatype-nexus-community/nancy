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

// Package cyclonedx has definitions and functions for processing golang purls into a minimal CycloneDX 1.1 Sbom
package cyclonedx

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/package-url/packageurl-go"
	"github.com/shopspring/decimal"
	"github.com/sonatype-nexus-community/nancy/types"
	assert "gopkg.in/go-playground/assert.v1"
)

func TestProcessPurlsIntoSBOM(t *testing.T) {
	results := []types.Coordinate{}
	crypto := types.Coordinate{
		Coordinates:     "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
		Reference:       "https://ossindex.sonatype.org/component/pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
		Vulnerabilities: []types.Vulnerability{},
	}
	dec, _ := decimal.NewFromString("5.8")
	crypto.Vulnerabilities = append(crypto.Vulnerabilities,
		types.Vulnerability{
			Id:          "CVE-123",
			Title:       "CVE-123",
			Description: "Hello I am a CVE",
			CvssScore:   dec,
			CvssVector:  "WhatsYourVectorVictor",
			Cve:         "CVE-123",
			Reference:   "http://www.google.com",
		})
	results = append(results, crypto)

	results = append(results, types.Coordinate{
		Coordinates:     "pkg:golang/github.com/go-yaml/yaml@v2.2.2",
		Reference:       "https://ossindex.sonatype.org/component/pkg:golang/github.com/go-yaml/yaml@v2.2.2",
		Vulnerabilities: []types.Vulnerability{},
	})
	result := ProcessPurlsIntoSBOM(results)

	doc := etree.NewDocument()

	if err := doc.ReadFromString(result); err != nil {
		t.Error("Uh Oh")
	}

	root := doc.SelectElement("bom")
	assert.Equal(t, root.Tag, "bom")
	assert.Equal(t, root.Attr[0].Key, "xmlns")
	assert.Equal(t, root.Attr[0].Value, "http://cyclonedx.org/schema/bom/1.1")
	assert.Equal(t, root.Attr[1].Space, "xmlns")
	assert.Equal(t, root.Attr[1].Key, "v")
	assert.Equal(t, root.Attr[1].Value, "http://cyclonedx.org/schema/ext/vulnerability/1.0")
	assert.Equal(t, root.Attr[2].Key, "version")
	assert.Equal(t, root.Attr[2].Value, "1")
	components := root.SelectElement("components")
	for i, component := range components.SelectElements("component") {
		coordinate, _ := packageurl.FromString(results[i].Coordinates)
		assert.Equal(t, component.Tag, "component")
		assert.Equal(t, component.Attr[0].Key, "type")
		assert.Equal(t, component.Attr[0].Value, "library")
		assert.Equal(t, component.Attr[1].Key, "bom-ref")
		assert.Equal(t, component.Attr[1].Value, results[i].Coordinates)
		name := component.SelectElement("name")
		assert.Equal(t, name.Tag, "name")
		assert.Equal(t, name.Text(), coordinate.Name)
		version := component.SelectElement("version")
		assert.Equal(t, version.Tag, "version")
		assert.Equal(t, version.Text(), coordinate.Version)
		purl := component.SelectElement("purl")
		assert.Equal(t, purl.Tag, "purl")
		assert.Equal(t, purl.Text(), coordinate.ToString())
		if purl.Text() == "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2" {
			vulnerabilities := component.SelectElement("vulnerabilities")
			assert.Equal(t, vulnerabilities.Space, "v")
			assert.Equal(t, vulnerabilities.Tag, "vulnerabilities")
			for x, vulnerability := range vulnerabilities.SelectElements("vulnerability") {
				assert.Equal(t, vulnerability.Tag, "vulnerability")
				assert.Equal(t, vulnerability.Space, "v")
				assert.Equal(t, vulnerability.Attr[0].Key, "ref")
				assert.Equal(t, vulnerability.Attr[0].Value, coordinate.ToString())
				id := vulnerability.SelectElement("id")
				assert.Equal(t, id.Tag, "id")
				assert.Equal(t, id.Space, "v")
				assert.Equal(t, id.Text(), results[0].Vulnerabilities[x].Title)
				source := vulnerability.SelectElement("source")
				assert.Equal(t, source.Tag, "source")
				assert.Equal(t, source.Space, "v")
				assert.Equal(t, source.Attr[0].Key, "name")
				assert.Equal(t, source.Attr[0].Value, "ossindex")
				url := source.SelectElement("url")
				assert.Equal(t, url.Tag, "url")
				assert.Equal(t, url.Space, "v")
				assert.Equal(t, url.Text(), results[0].Vulnerabilities[x].Reference)
				ratings := vulnerability.SelectElement("ratings")
				assert.Equal(t, ratings.Tag, "ratings")
				assert.Equal(t, ratings.Space, "v")
				rating := ratings.SelectElement("rating")
				assert.Equal(t, rating.Tag, "rating")
				assert.Equal(t, rating.Space, "v")
				score := rating.SelectElement("score")
				assert.Equal(t, score.Tag, "score")
				assert.Equal(t, score.Space, "v")
				base := score.SelectElement("base")
				assert.Equal(t, base.Tag, "base")
				assert.Equal(t, base.Space, "v")
				assert.Equal(t, base.Text(), results[0].Vulnerabilities[x].CvssScore.String())
				vector := rating.SelectElement("vector")
				assert.Equal(t, vector.Tag, "vector")
				assert.Equal(t, vector.Space, "v")
				assert.Equal(t, vector.Text(), results[0].Vulnerabilities[x].CvssVector)
			}
		}
	}
}

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
	"encoding/xml"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
)

const cycloneDXBomXmlns1_1 = "http://cyclonedx.org/schema/bom/1.1"
const cycloneDXBomXmlns1_0V = "http://cyclonedx.org/schema/ext/vulnerability/1.0"

const version = "1"

// ProcessPurlsIntoSBOM will take a slice of packageurl.PackageURL and convert them
// into a minimal 1.1 CycloneDX sbom
func ProcessPurlsIntoSBOM(results []types.Coordinate) string {
	return processPurlsIntoSBOMSchema1_1(results)
}

// SBOMFromPackageURLs will take a slice of packageurl.PackageURL and convert them
// into a minimal 1.1 CycloneDX sbom
func SBOMFromPackageURLs(results []packageurl.PackageURL) string {
	return processPackageURLsIntoSBOMSchema1_1(results)
}

// SBOMFromSHA1 will take a slice of Sha1SBOM and convert them
// into a minimal 1.1 CycloneDX sbom
func SBOMFromSHA1(results []types.Sha1SBOM) string {
	return createMinimalSha1Sbom(results)
}

func createMinimalSha1Sbom(results []types.Sha1SBOM) string {
	sbom := createSbomDocument()
	for _, v := range results {
		hash := types.Hash{
			Alg: "SHA-1", Attribute: v.Sha1,
		}
		component := types.Component{
			Type:    "library",
			BomRef:  v.Sha1,
			Name:    v.Location,
			Version: "0",
		}
		component.Hashes.Hash = append(component.Hashes.Hash, hash)

		sbom.Components.Component = append(sbom.Components.Component, component)
	}

	return processAndReturnSbom(sbom)
}

func processPackageURLsIntoSBOMSchema1_1(results []packageurl.PackageURL) string {
	sbom := createSbomDocument()
	for _, v := range results {
		component := types.Component{
			Type:    "library",
			BomRef:  v.ToString(),
			Purl:    v.ToString(),
			Group:   v.Namespace,
			Name:    v.Name,
			Version: v.Version,
		}

		sbom.Components.Component = append(sbom.Components.Component, component)
	}

	return processAndReturnSbom(sbom)
}

func processPurlsIntoSBOMSchema1_1(results []types.Coordinate) string {
	sbom := createSbomDocument()
	for _, v := range results {
		purl, err := packageurl.FromString(v.Coordinates)
		customerrors.Check(err, "Error parsing purl from given coordinate")

		component := types.Component{
			Type:    "library",
			BomRef:  purl.String(),
			Purl:    purl.String(),
			Name:    purl.Name,
			Version: purl.Version,
			Group:   purl.Namespace,
		}

		if v.IsVulnerable() {
			vulns := types.Vulnerabilities{}
			for _, x := range v.Vulnerabilities {
				rating := types.Rating{Score: types.Score{Base: x.CvssScore}}
				rating.Vector = x.CvssVector
				ratings := types.Ratings{}
				ratings.Rating = rating
				source := types.Source{Name: "ossindex"}
				source.URL = x.Reference
				vuln := types.SbomVulnerability{ID: x.Cve, Source: source, Description: x.Description, Ref: v.Coordinates}
				vuln.Ratings = append(vuln.Ratings, ratings)
				vulns.Vulnerability = append(vulns.Vulnerability, vuln)
			}
			component.Vulnerabilities = vulns
		}

		sbom.Components.Component = append(sbom.Components.Component, component)
	}

	return processAndReturnSbom(sbom)
}

func createSbomDocument() *types.Sbom {
	sbom := types.Sbom{}
	sbom.Xmlns = cycloneDXBomXmlns1_1
	sbom.XMLNSV = cycloneDXBomXmlns1_0V
	sbom.Version = version

	return &sbom
}

func processAndReturnSbom(sbom *types.Sbom) string {
	output, err := xml.MarshalIndent(sbom, " ", "     ")
	if err != nil {
		LogLady.Error(err)
	}

	output = []byte(xml.Header + string(output))

	return string(output)
}

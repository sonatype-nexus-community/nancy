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

import (
	"encoding/xml"
	"fmt"

	"github.com/package-url/packageurl-go"
)

type sbom struct {
	XMLName    xml.Name   `xml:"bom"`
	Xmlns      string     `xml:"xmlns,attr"`
	Version    string     `xml:"version,attr"`
	Components components `xml:"components"`
}

type components struct {
	Component []component `xml:"component"`
}

type component struct {
	Type    string `xml:"type,attr"`
	BomRef  string `xml:"bom-ref,attr"`
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Purl    string `xml:"purl"`
}

const cycloneDXBomXmlns1_1 = "http://cyclonedx.org/schema/bom/1.1"

const version = "1"

// ProcessPurlsIntoSBOM will take a slice of packageurl.PackageURL and convert them
// into a minimal 1.1 CycloneDX sbom
func ProcessPurlsIntoSBOM(purls []packageurl.PackageURL) string {
	return processPurlsIntoSBOMSchema1_1(purls)
}

func processPurlsIntoSBOMSchema1_1(purls []packageurl.PackageURL) string {
	sbom := sbom{}
	sbom.Xmlns = cycloneDXBomXmlns1_1
	sbom.Version = version
	for _, v := range purls {
		component := component{Type: "library", BomRef: v.String(), Purl: v.String(), Name: v.Name, Version: v.Version}
		sbom.Components.Component = append(sbom.Components.Component, component)
	}

	output, err := xml.MarshalIndent(sbom, " ", "     ")
	if err != nil {
		fmt.Print(err)
	}

	output = []byte(xml.Header + string(output))

	return string(output)
}

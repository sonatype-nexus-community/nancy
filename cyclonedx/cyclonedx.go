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

// Definitions and functions for processing golang purls into a CycloneDX Sbom
package cyclonedx

import (
	"encoding/xml"
	"fmt"

	"github.com/package-url/packageurl-go"
)

type Sbom struct {
	XMLName    xml.Name   `xml:"bom"`
	Xmlns      string     `xml:"xmlns,attr"`
	Version    string     `xml:"version,attr"`
	Components Components `xml:"components"`
}

type Components struct {
	Component []Component `xml:"component"`
}

type Component struct {
	Type    string `xml:"type,attr"`
	BomRef  string `xml:"bom-ref,attr"`
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Purl    string `xml:"purl"`
}

func ProcessPurlsIntoSBOM(purls []packageurl.PackageURL) string {
	sbom := Sbom{}
	sbom.Xmlns = "http://cyclonedx.org/schema/bom/1.1"
	sbom.Version = "1"
	for _, v := range purls {
		component := Component{Type: "library", BomRef: v.String(), Purl: v.String(), Name: v.Name, Version: v.Version}
		sbom.Components.Component = append(sbom.Components.Component, component)
	}

	output, err := xml.MarshalIndent(sbom, " ", "     ")
	if err != nil {
		fmt.Print(err)
	}

	output = []byte(xml.Header + string(output))

	return string(output)
}

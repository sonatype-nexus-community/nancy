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
package audit

import (
	"github.com/shopspring/decimal"
	"github.com/sonatype-nexus-community/nancy/types"
	"testing"
)

func createCoordinates(num int, vulnerable bool) (coordinates []types.Coordinate) {
	for i := 0; i < num; i++ {
		coordinates = append(coordinates, createCoordinate(vulnerable))
	}

	return coordinates
}

func createCoordinate(vulnerable bool) types.Coordinate {
	if vulnerable {
		return types.Coordinate{
			Coordinates:     "github/thing:2.0.0",
			Reference:       "Reference",
			Vulnerabilities: createVulnerabilities(10),
		}
	}
	return types.Coordinate{
		Coordinates: "github/thing:1.0.0",
		Reference:   "Reference",
	}
}

func createVulnerabilities(num int) (vulnerabilities []types.Vulnerability) {
	for i := 0; i < num; i++ {
		vulnerabilities = append(vulnerabilities, createVulnerability())
	}

	return vulnerabilities
}

func createVulnerability() (vulnerability types.Vulnerability) {
	vulnerability.Cve = "7.8"
	vulnerability.CvssScore, _ = decimal.NewFromString("7.88")
	vulnerability.CvssVector = "What"
	vulnerability.Description = "Description"
	vulnerability.Id = "123"
	vulnerability.Reference = "Reference"
	vulnerability.Title = "Vulnerability"

	return vulnerability
}

func TestLogResultsWithVulnerabilitiesNoColor(t *testing.T) {
	projects := 20
	coordinates := createCoordinates(projects, true)
	i := LogResults(false, 20, coordinates)

	if i != projects {
		t.Errorf("Expected %d vulnerabilites but found %d", projects, i)
	}
}

func TestLogResultsWithoutVulnerabilitiesNoColor(t *testing.T) {
	projects := 20
	coordinates := createCoordinates(projects, false)
	i := LogResults(false, 20, coordinates)

	if i != 0 {
		t.Errorf("Expected %d vulnerabilites but found %d", 0, i)
	}
}

func TestLogResultsWithVulnerabilitiesColor(t *testing.T) {
	projects := 20
	coordinates := createCoordinates(projects, true)
	i := LogResults(true, 20, coordinates)

	if i != projects {
		t.Errorf("Expected %d vulnerabilites but found %d", projects, i)
	}
}

func TestLogResultsWithoutVulnerabilitiesColor(t *testing.T) {
	projects := 20
	coordinates := createCoordinates(projects, false)
	i := LogResults(true, 20, coordinates)

	if i != 0 {
		t.Errorf("Expected %d vulnerabilites but found %d", 0, i)
	}
}

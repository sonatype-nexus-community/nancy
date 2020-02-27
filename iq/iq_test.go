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

// Package iq has definitions and functions for processing golang purls with Nexus IQ Server
package iq

import (
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
)

const applicationsResponse = `{
	"applications": [
		{
			"id": "4bb67dcfc86344e3a483832f8c496419",
			"publicId": "testapp",
			"name": "TestApp",
			"organizationId": "bb41817bd3e2403a8a52fe8bcd8fe25a",
			"contactUserName": "NewAppContact",
			"applicationTags": [
				{
					"id": "9beee80c6fc148dfa51e8b0359ee4d4e",
					"tagId": "cfea8fa79df64283bd64e5b6b624ba48",
					"applicationId": "4bb67dcfc86344e3a483832f8c496419"
				}
			]
		}
	]
}`

const thirdPartyAPIResultJSON = `{
		"statusUrl": "api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/status/9cee2b6366fc4d328edc318eae46b2cb"
}`

const pollingResult = `{
	"policyAction": "None",
	"reportHtmlUrl": "http://sillyplace.com:8090/ui/links/application/test-app/report/95c4c14e",
	"isError": false
}`

func setupIqConfiguration() (config configuration.IqConfiguration) {
	config.Application = "testapp"
	config.Server = "http://sillyplace.com:8090"
	config.Stage = "develop"
	config.User = "admin"
	config.Token = "admin123"
	return
}

func TestAuditPackages(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jsonCoordinates, _ := json.Marshal([]types.Coordinate{
		{
			Coordinates:     "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
			Reference:       "https://ossindex.sonatype.org/component/pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
			Vulnerabilities: []types.Vulnerability{},
		},
		{
			Coordinates:     "pkg:golang/github.com/go-yaml/yaml@v2.2.2",
			Reference:       "https://ossindex.sonatype.org/component/pkg:golang/github.com/go-yaml/yaml@v2.2.2",
			Vulnerabilities: []types.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		httpmock.NewStringResponder(200, string(jsonCoordinates)))

	httpmock.RegisterResponder("GET", "http://sillyplace.com:8090/api/v2/applications?publicId=testapp",
		httpmock.NewStringResponder(200, applicationsResponse))

	httpmock.RegisterResponder("POST", "http://sillyplace.com:8090/api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/sources/nancy?stageId=develop",
		httpmock.NewStringResponder(202, thirdPartyAPIResultJSON))

	httpmock.RegisterResponder("GET", "http://sillyplace.com:8090/api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/status/9cee2b6366fc4d328edc318eae46b2cb",
		httpmock.NewStringResponder(200, pollingResult))

	var purls []string
	purls = append(purls, "pkg:golang/github.com/go-yaml/yaml@v2.2.2")
	purls = append(purls, "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2")

	result, _ := AuditPackages(purls, "testapp", setupIqConfiguration())

	statusExpected := types.StatusURLResult{PolicyAction: "None", ReportHTMLURL: "http://sillyplace.com:8090/ui/links/application/test-app/report/95c4c14e", IsError: false}

	assert.Equal(t, result, statusExpected)
}

func TestAuditPackagesIqDownOrUnreachable(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jsonCoordinates, _ := json.Marshal([]types.Coordinate{
		{
			Coordinates:     "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
			Reference:       "https://ossindex.sonatype.org/component/pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
			Vulnerabilities: []types.Vulnerability{},
		},
		{
			Coordinates:     "pkg:golang/github.com/go-yaml/yaml@v2.2.2",
			Reference:       "https://ossindex.sonatype.org/component/pkg:golang/github.com/go-yaml/yaml@v2.2.2",
			Vulnerabilities: []types.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		httpmock.NewStringResponder(200, string(jsonCoordinates)))

	httpmock.RegisterResponder("GET", "http://sillyplace.com:8090/api/v2/applications?publicId=testapp",
		httpmock.NewBytesResponder(404, []byte("")))

	var purls []string
	purls = append(purls, "pkg:golang/github.com/go-yaml/yaml@v2.2.2")
	purls = append(purls, "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2")

	_, err := AuditPackages(purls, "testapp", setupIqConfiguration())
	if err == nil {
		t.Error("There is an error")
	}
}

func TestAuditPackagesIqUpButBadThirdPartyAPIResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jsonCoordinates, _ := json.Marshal([]types.Coordinate{
		{
			Coordinates:     "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
			Reference:       "https://ossindex.sonatype.org/component/pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2",
			Vulnerabilities: []types.Vulnerability{},
		},
		{
			Coordinates:     "pkg:golang/github.com/go-yaml/yaml@v2.2.2",
			Reference:       "https://ossindex.sonatype.org/component/pkg:golang/github.com/go-yaml/yaml@v2.2.2",
			Vulnerabilities: []types.Vulnerability{},
		},
	})

	httpmock.RegisterResponder("POST", "https://ossindex.sonatype.org/api/v3/component-report",
		httpmock.NewStringResponder(200, string(jsonCoordinates)))

	httpmock.RegisterResponder("GET", "http://sillyplace.com:8090/api/v2/applications?publicId=testapp",
		httpmock.NewStringResponder(200, applicationsResponse))

	httpmock.RegisterResponder("POST", "http://sillyplace.com:8090/api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/sources/nancy?stageId=develop",
		httpmock.NewBytesResponder(500, []byte("")))

	var purls []string
	purls = append(purls, "pkg:golang/github.com/go-yaml/yaml@v2.2.2")
	purls = append(purls, "pkg:golang/golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2")

	_, err := AuditPackages(purls, "testapp", setupIqConfiguration())
	if err == nil {
		t.Error("There is an error")
	}
}

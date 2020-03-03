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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/cyclonedx"
	"github.com/sonatype-nexus-community/nancy/ossindex"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/nancy/useragent"
)

const internalApplicationIDURL = "/api/v2/applications?publicId="

const thirdPartyAPILeft = "/api/v2/scan/applications/"

const thirdPartyAPIRight = "/sources/nancy?stageId="

const (
	pollInterval = 1 * time.Second
)

var (
	localConfig configuration.IqConfiguration
	tries       = 0
)

// Internal types for use by this package, don't need to expose them
type applicationResponse struct {
	Applications []application `json:"applications"`
}

type application struct {
	ID string `json:"id"`
}

type thirdPartyAPIResult struct {
	StatusURL string `json:"statusUrl"`
}

var statusURLResp types.StatusURLResult

// AuditPackages accepts a slice of purls, public application ID, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
func AuditPackages(purls []string, applicationID string, config configuration.IqConfiguration) (types.StatusURLResult, error) {
	localConfig = config

	if localConfig.User == "admin" && localConfig.Token == "admin123" {
		warnUserOfBadLifeChoices()
	}

	internalID := getInternalApplicationID(applicationID)
	if internalID == "" {
		return statusURLResp, fmt.Errorf("Internal ID for %s could not be found, or Nexus IQ Server is down", applicationID)
	}

	resultsFromOssIndex, err := ossindex.AuditPackages(purls)
	customerrors.Check(err, "There was an issue auditing packages using OSS Index")

	sbom := cyclonedx.ProcessPurlsIntoSBOM(resultsFromOssIndex)
	statusURL := submitToThirdPartyAPI(sbom, internalID)
	if statusURL == "" {
		return statusURLResp, fmt.Errorf("There was an issue submitting your sbom to the Nexus IQ Third Party API, sbom: %s", sbom)
	}

	statusURLResp = types.StatusURLResult{}

	finished := make(chan bool)

	go func() {
		for {
			select {
			case <-finished:
				return
			default:
				pollIQServer(fmt.Sprintf("%s/%s", localConfig.Server, statusURL), finished, localConfig.MaxRetries)
				time.Sleep(pollInterval)
			}
		}
	}()

	<-finished
	return statusURLResp, nil
}

func getInternalApplicationID(applicationID string) (internalID string) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s%s", localConfig.Server, internalApplicationIDURL, applicationID),
		nil,
	)

	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)
	customerrors.Check(err, "There was an error communicating with Nexus IQ Server to get your internal application ID")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		customerrors.Check(err, "There was an error retrieving the bytes of the response for getting your internal application ID from Nexus IQ Server")

		var response applicationResponse
		json.Unmarshal(bodyBytes, &response)
		return response.Applications[0].ID
	}
	return ""
}

func submitToThirdPartyAPI(sbom string, internalID string) string {
	client := &http.Client{}

	url := fmt.Sprintf("%s%s", localConfig.Server, fmt.Sprintf("%s%s%s%s", thirdPartyAPILeft, internalID, thirdPartyAPIRight, localConfig.Stage))

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(sbom)),
	)

	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())
	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	customerrors.Check(err, "There was an issue communicating with the Nexus IQ Third Party API")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		customerrors.Check(err, "There was an issue submitting your sbom to the Nexus IQ Third Party API")

		var response thirdPartyAPIResult
		json.Unmarshal(bodyBytes, &response)
		return response.StatusURL
	}

	return ""
}

func pollIQServer(statusURL string, finished chan bool, maxRetries int) {
	if tries > maxRetries {
		finished <- true
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", statusURL, nil)

	req.SetBasicAuth(localConfig.User, localConfig.Token)

	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)

	if err != nil {
		finished <- true
		customerrors.Check(err, "There was an error polling Nexus IQ Server")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		customerrors.Check(err, "There was an error with processing the response from polling Nexus IQ Server")

		var response types.StatusURLResult
		json.Unmarshal(bodyBytes, &response)
		statusURLResp = response
		if response.IsError {
			finished <- true
		}
		finished <- true
	}
	tries++
	fmt.Print(".")
}

func warnUserOfBadLifeChoices() {
	fmt.Println()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!! WARNING : You are using the default username and password for Nexus IQ. !!!!")
	fmt.Println("!!!! You are strongly encouraged to change these, and use a token.           !!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println()
}

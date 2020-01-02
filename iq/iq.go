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
	"os"
	"time"

	"github.com/package-url/packageurl-go"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/cyclonedx"
)

const internalApplicationIDURL = "/api/v2/applications?publicId="

const thirdPartyAPILeft = "/api/v2/scan/applications/"

const thirdPartyAPIRight = "/sources/nancy?stageId="

const (
	pollInterval = 1 * time.Second
)

type applicationResponse struct {
	Applications []application `json:"applications"`
}

type application struct {
	ID string `json:"id"`
}

type thirdPartyAPIResult struct {
	StatusURL string `json:"statusUrl"`
}

// StatusURLResult is a struct to let the consumer know what the response from Nexus IQ Server was
type StatusURLResult struct {
	PolicyAction  string `json:"policyAction"`
	ReportHTMLURL string `json:"reportHtmlUrl"`
	IsError       bool   `json:"isError"`
	ErrorMessage  string `json:"errorMessage"`
}

var localConfig configuration.Configuration

var statusURLResp StatusURLResult

func getPurls(purls []string) (result []packageurl.PackageURL) {
	for _, v := range purls {
		purl, _ := packageurl.FromString(v)
		result = append(result, purl)
	}

	return
}

// AuditPackages accepts a slice of purls, public application ID, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
func AuditPackages(purls []string, applicationID string, config configuration.Configuration) StatusURLResult {
	localConfig = config

	if localConfig.User == "admin" && localConfig.Token == "admin123" {
		warnUserOfBadLifeChoices()
	}

	internalID := getInternalApplicationID(applicationID)
	newPurls := getPurls(purls)
	sbom := cyclonedx.ProcessPurlsIntoSBOM(newPurls)
	statusURL := submitToThirdPartyAPI(sbom, internalID)

	finished := make(chan bool)

	statusURLResp = StatusURLResult{}

	go func() {
		for {
			select {
			case <-finished:
				return
			default:
				pollIQServer(fmt.Sprintf("%s/%s", localConfig.Server, statusURL), finished)
				time.Sleep(pollInterval)
			}
		}
	}()

	<-finished
	return statusURLResp
}

func getInternalApplicationID(applicationID string) (internalID string) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s%s", localConfig.Server, internalApplicationIDURL, applicationID),
		nil,
	)

	req.SetBasicAuth(localConfig.User, localConfig.Token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("User-Agent", fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var response thirdPartyAPIResult
		json.Unmarshal(bodyBytes, &response)
		return response.StatusURL
	}

	return ""
}

func pollIQServer(statusURL string, finished chan bool) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", statusURL, nil)

	req.SetBasicAuth(localConfig.User, localConfig.Token)

	req.Header.Set("User-Agent", fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion))

	resp, err := client.Do(req)

	if err != nil {
		finished <- true
		fmt.Println(err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Print(err)
		}
		var response StatusURLResult
		json.Unmarshal(bodyBytes, &response)
		statusURLResp = response
		if response.IsError {
			finished <- true
		}
		finished <- true
	}
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

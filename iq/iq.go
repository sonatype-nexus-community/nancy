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

	"github.com/package-url/packageurl-go"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/cyclonedx"
	. "github.com/sonatype-nexus-community/nancy/logger"
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
	localConfig configuration.Config
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

// Audit accepts a slice of packageurl.PackageURL, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
func Audit(purls []packageurl.PackageURL, config configuration.Config) (types.StatusURLResult, error) {
	return doAudit(purls, config)
}

func doAudit(purls []packageurl.PackageURL, config configuration.Config) (types.StatusURLResult, error) {
	LogLady.WithFields(logrus.Fields{
		"purls":          purls,
		"application_id": config.IQConfig.Application,
	}).Info("Beginning audit with IQ")
	localConfig = config

	if localConfig.IQConfig.Username == "admin" && localConfig.IQConfig.Token == "admin123" {
		LogLady.Info("Warning user of questionable life choices related to username and password")
		warnUserOfBadLifeChoices()
	}

	internalID, err := getInternalApplicationID(config.IQConfig.Application)
	if internalID == "" && err != nil {
		LogLady.Error("Internal ID not obtained from Nexus IQ")
		return statusURLResp, err
	}

	resultsFromOssIndex, err := ossindex.Audit(purls, &config)
	customerrors.Check(err, "There was an issue auditing packages using OSS Index")

	sbom := cyclonedx.ProcessPurlsIntoSBOM(resultsFromOssIndex)
	LogLady.WithField("sbom", sbom).Debug("Obtained cyclonedx SBOM")

	LogLady.WithFields(logrus.Fields{
		"internal_id": internalID,
		"sbom":        sbom,
	}).Debug("Submitting to Third Party API")
	statusURL := submitToThirdPartyAPI(sbom, internalID)
	if statusURL == "" {
		LogLady.Error("StatusURL not obtained from Third Party API")
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
				pollIQServer(fmt.Sprintf("%s/%s", localConfig.IQConfig.Server, statusURL), finished, localConfig.IQConfig.MaxRetries)
				time.Sleep(pollInterval)
			}
		}
	}()

	<-finished
	return statusURLResp, nil
}

func getInternalApplicationID(applicationID string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s%s", localConfig.IQConfig.Server, internalApplicationIDURL, applicationID),
		nil,
	)

	req.SetBasicAuth(localConfig.IQConfig.Username, localConfig.IQConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)
	customerrors.Check(err, "There was an error communicating with Nexus IQ Server to get your internal application ID")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		customerrors.Check(err, "There was an error retrieving the bytes of the response for getting your internal application ID from Nexus IQ Server")

		var response applicationResponse
		json.Unmarshal(bodyBytes, &response)

		if response.Applications != nil && len(response.Applications) > 0 {
			LogLady.WithFields(logrus.Fields{
				"internal_id": response.Applications[0].ID,
			}).Debug("Retrieved internal ID from Nexus IQ Server")

			return response.Applications[0].ID, nil
		}

		LogLady.WithFields(logrus.Fields{
			"application_id": applicationID,
		}).Error("Unable to retrieve an internal ID for the specified public application ID")

		return "", fmt.Errorf("Unable to retrieve an internal ID for the specified public application ID: %s", applicationID)
	}
	LogLady.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
	}).Error("Error communicating with Nexus IQ Server application endpoint")
	return "", fmt.Errorf("Unable to communicate with Nexus IQ Server, status code returned is: %d", resp.StatusCode)
}

func submitToThirdPartyAPI(sbom string, internalID string) string {
	LogLady.Debug("Beginning to submit to Third Party API")
	client := &http.Client{}

	url := fmt.Sprintf("%s%s", localConfig.IQConfig.Server, fmt.Sprintf("%s%s%s%s", thirdPartyAPILeft, internalID, thirdPartyAPIRight, localConfig.IQConfig.Stage))
	LogLady.WithField("url", url).Debug("Crafted URL for submission to Third Party API")

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(sbom)),
	)

	req.SetBasicAuth(localConfig.IQConfig.Username, localConfig.IQConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())
	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	customerrors.Check(err, "There was an issue communicating with the Nexus IQ Third Party API")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		LogLady.WithField("body", string(bodyBytes)).Info("Request accepted")
		customerrors.Check(err, "There was an issue submitting your sbom to the Nexus IQ Third Party API")

		var response thirdPartyAPIResult
		json.Unmarshal(bodyBytes, &response)
		return response.StatusURL
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	LogLady.WithFields(logrus.Fields{
		"body":        string(bodyBytes),
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	}).Info("Request not accepted")
	customerrors.Check(err, "There was an issue submitting your sbom to the Nexus IQ Third Party API")

	return ""
}

func pollIQServer(statusURL string, finished chan bool, maxRetries int) {
	LogLady.WithFields(logrus.Fields{
		"attempt_number": tries,
		"max_retries":    maxRetries,
		"status_url":     statusURL,
	}).Trace("Polling Nexus IQ for response")
	if tries > maxRetries {
		LogLady.Error("Maximum tries exceeded, finished polling, consider bumping up Max Retries")
		finished <- true
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", statusURL, nil)

	req.SetBasicAuth(localConfig.IQConfig.Username, localConfig.IQConfig.Token)

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

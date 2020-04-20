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

// Package iq has definitions and functions for processing golang purls with Nexus IQ Server
package iq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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

type resultError struct {
	finished bool
	err      error
}

// AuditPackages accepts a slice of purls, public application ID, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
func AuditPackages(purls []string, applicationID string, config configuration.IqConfiguration) (types.StatusURLResult, error) {
	LogLady.WithFields(logrus.Fields{
		"purls":          purls,
		"application_id": applicationID,
	}).Info("Beginning audit with IQ")
	localConfig = config

	if localConfig.User == "admin" && localConfig.Token == "admin123" {
		LogLady.Info("Warning user of questionable life choices related to username and password")
		warnUserOfBadLifeChoices()
	}

	internalID, err := getInternalApplicationID(applicationID)
	if internalID == "" && err != nil {
		LogLady.Error("Internal ID not obtained from Nexus IQ")
		return statusURLResp, err
	}

	resultsFromOssIndex, err := ossindex.AuditPackages(purls) //nolint deprecation
	if err != nil {
		return statusURLResp, customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an issue auditing packages using OSS Index"}
	}

	sbom := cyclonedx.ProcessPurlsIntoSBOM(resultsFromOssIndex)
	LogLady.WithField("sbom", sbom).Debug("Obtained cyclonedx SBOM")

	LogLady.WithFields(logrus.Fields{
		"internal_id": internalID,
		"sbom":        sbom,
	}).Debug("Submitting to Third Party API")
	statusURL, err := submitToThirdPartyAPI(sbom, internalID)
	if err != nil {
		return statusURLResp, customerrors.ErrorExit{ExitCode: 3, Err: err}
	}
	if statusURL == "" {
		LogLady.Error("StatusURL not obtained from Third Party API")
		return statusURLResp, fmt.Errorf("There was an issue submitting your sbom to the Nexus IQ Third Party API, sbom: %s", sbom)
	}

	statusURLResp = types.StatusURLResult{}

	finishedChan := make(chan resultError)

	go func() resultError {
		for {
			select {
			case <-finishedChan:
				return resultError{finished: true}
			default:
				if err = pollIQServer(fmt.Sprintf("%s/%s", localConfig.Server, statusURL), finishedChan, localConfig.MaxRetries); err != nil {
					return resultError{finished: false, err: err}
				}
				time.Sleep(pollInterval)
			}
		}
	}()

	r := <-finishedChan
	return statusURLResp, r.err
}

func getInternalApplicationID(applicationID string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s%s", localConfig.Server, internalApplicationIDURL, applicationID),
		nil,
	)
	if err != nil {
		return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "Request to get internal application id failed"}
	}

	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an error communicating with Nexus IQ Server to get your internal application ID"}
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an error retrieving the bytes of the response for getting your internal application ID from Nexus IQ Server"}
		}

		var response applicationResponse
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "failed to unmarshal response"}
		}

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

func submitToThirdPartyAPI(sbom string, internalID string) (string, error) {
	LogLady.Debug("Beginning to submit to Third Party API")
	client := &http.Client{}

	url := fmt.Sprintf("%s%s", localConfig.Server, fmt.Sprintf("%s%s%s%s", thirdPartyAPILeft, internalID, thirdPartyAPIRight, localConfig.Stage))
	LogLady.WithField("url", url).Debug("Crafted URL for submission to Third Party API")

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(sbom)),
	)
	if err != nil {
		return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "Could not POST to Nexus iQ Third Party API"}
	}

	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())
	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an issue communicating with the Nexus IQ Third Party API"}
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		LogLady.WithField("body", string(bodyBytes)).Info("Request accepted")
		if err != nil {
			return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an issue submitting your sbom to the Nexus IQ Third Party API"}
		}

		var response thirdPartyAPIResult
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "Could not unmarshal response from iQ server"}
		}
		return response.StatusURL, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	LogLady.WithFields(logrus.Fields{
		"body":        string(bodyBytes),
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	}).Info("Request not accepted")
	if err != nil {
		return "", customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an issue submitting your sbom to the Nexus IQ Third Party API"}
	}

	return "", err
}

func pollIQServer(statusURL string, finished chan resultError, maxRetries int) error {
	LogLady.WithFields(logrus.Fields{
		"attempt_number": tries,
		"max_retries":    maxRetries,
		"status_url":     statusURL,
	}).Trace("Polling Nexus IQ for response")
	if tries > maxRetries {
		LogLady.Error("Maximum tries exceeded, finished polling, consider bumping up Max Retries")
		finished <- resultError{finished: true, err: nil}
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "Could not poll iQ server"}
	}

	req.SetBasicAuth(localConfig.User, localConfig.Token)

	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)

	if err != nil {
		finished <- resultError{finished: true, err: err}
		return customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an error polling Nexus IQ Server"}
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "There was an error with processing the response from polling Nexus IQ Server"}
		}

		var response types.StatusURLResult
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return customerrors.ErrorExit{ExitCode: 3, Err: err, Message: "Could not unmarshal response from iQ server"}
		}
		statusURLResp = response
		if response.IsError {
			finished <- resultError{finished: true, err: nil}
		}
		finished <- resultError{finished: true, err: nil}
	}
	tries++
	fmt.Print(".")
	return err
}

func warnUserOfBadLifeChoices() {
	fmt.Println()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!! WARNING : You are using the default username and password for Nexus IQ. !!!!")
	fmt.Println("!!!! You are strongly encouraged to change these, and use a token.           !!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println()
}

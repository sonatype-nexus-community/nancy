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

// Definitions and functions for processing the OSS Index Feed
package ossindex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/ossindex/internal/cache"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/nancy/useragent"
)

const defaultOssIndexUrl = "https://ossindex.sonatype.org/api/v3/component-report"

const MAX_COORDS = 128

var (
	ossIndexUrl string
)

func getOssIndexUrl() string {
	if ossIndexUrl == "" {
		ossIndexUrl = defaultOssIndexUrl
	}
	return ossIndexUrl
}

// RemoveCacheDirectory deletes the local database directory.
func RemoveCacheDirectory() error {
	return cache.RemoveCacheDirectory()
}

// AuditPackages will given a list of Package URLs, run an OSS Index audit.
//
// Deprecated: AuditPackages is old and being maintained for upstream compatibility at the moment.
// It will be removed when we go to a major version release. Use AuditPackagesWithOSSIndex instead.
func AuditPackages(purls []string) ([]types.Coordinate, error) {
	return doAuditPackages(purls, nil)
}

// AuditPackagesWithOSSIndex will given a list of Package URLs, run an OSS Index audit, and takes OSS Index configuration
func AuditPackagesWithOSSIndex(purls []string, config *configuration.Configuration) ([]types.Coordinate, error) {
	return doAuditPackages(purls, config)
}

func doAuditPackages(purls []string, config *configuration.Configuration) ([]types.Coordinate, error) {
	newPurls, results, err := cache.HydrateNewPurlsFromCache(purls)
	customerrors.Check(err, "Error initializing cache")

	chunks := chunk(newPurls, MAX_COORDS)

	for _, chunk := range chunks {
		if len(chunk) > 0 {
			var request types.AuditRequest
			request.Coordinates = chunk
			LogLady.WithField("request", request).Info("Prepping request to OSS Index")
			var jsonStr, _ = json.Marshal(request)

			coordinates, err := doRequestToOSSIndex(jsonStr, config)
			if err != nil {
				return nil, err
			}

			for _, v := range coordinates {
				results = append(results, v)
			}

			LogLady.WithField("coordinates", coordinates).Info("Coordinates unmarshalled from OSS Index")
			err = cache.InsertValuesIntoCache(coordinates)
			if err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

func doRequestToOSSIndex(jsonStr []byte, config *configuration.Configuration) (coordinates []types.Coordinate, err error) {
	req, err := setupRequest(jsonStr, config)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		LogLady.WithField("resp_status_code", resp.Status).Error("Error accessing OSS Index due to Rate Limiting")
		return nil, &types.OSSIndexRateLimitError{}
	}

	if resp.StatusCode != http.StatusOK {
		LogLady.WithField("resp_status_code", resp.Status).Error("Error accessing OSS Index")
		return nil, fmt.Errorf("[%s] error accessing OSS Index", resp.Status)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			LogLady.WithField("error", err).Error("Error closing response body")
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogLady.WithField("error", err).Error("Error accessing OSS Index")
		return
	}

	// Process results
	if err = json.Unmarshal([]byte(body), &coordinates); err != nil {
		LogLady.WithField("error", err).Error("Error unmarshalling response from OSS Index")
		return
	}
	return
}

func setupRequest(jsonStr []byte, config *configuration.Configuration) (req *http.Request, err error) {
	LogLady.WithField("json_string", string(jsonStr)).Debug("Setting up new POST request to OSS Index")
	req, err = http.NewRequest(
		"POST",
		getOssIndexUrl(),
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", useragent.GetUserAgent())
	req.Header.Set("Content-Type", "application/json")
	if config != nil && config.Username != "" && config.Token != "" {
		LogLady.Info("Set OSS Index Basic Auth")
		req.SetBasicAuth(config.Username, config.Token)
	}

	return req, nil
}

func chunk(purls []string, chunkSize int) [][]string {
	var divided [][]string

	for i := 0; i < len(purls); i += chunkSize {
		end := i + chunkSize

		if end > len(purls) {
			end = len(purls)
		}

		divided = append(divided, purls[i:end])
	}

	return divided
}

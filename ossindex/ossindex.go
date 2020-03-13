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

// Definitions and functions for processing the OSS Index Feed
package ossindex

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/nancy/useragent"
)

const dbValueDirName = "golang"

const defaultOssIndexUrl = "https://ossindex.sonatype.org/api/v3/component-report"

const MAX_COORDS = 128

var (
	ossIndexUrl string
)

func getDatabaseDirectory() (dbDir string) {
	LogLady.Trace("Attempting to get database directory")
	usr, err := user.Current()
	customerrors.Check(err, "Error getting user home")

	LogLady.WithField("home_dir", usr.HomeDir).Trace("Obtained user directory")
	var leftPath = path.Join(usr.HomeDir, types.OssIndexDirName)
	var fullPath string
	if flag.Lookup("test") == nil {
		fullPath = path.Join(leftPath, dbValueDirName)
	} else {
		fullPath = path.Join(leftPath, "test-nancy")
	}

	return fullPath
}

// RemoveCacheDirectory deletes the local database directory.
func RemoveCacheDirectory() error {
	return os.RemoveAll(getDatabaseDirectory())
}

func getOssIndexUrl() string {
	if ossIndexUrl == "" {
		ossIndexUrl = defaultOssIndexUrl
	}
	return ossIndexUrl
}

func openDb(dbDir string) (db *badger.DB, err error) {
	LogLady.Debug("Attempting to open Badger DB")
	opts := badger.DefaultOptions

	opts.Dir = getDatabaseDirectory()
	opts.ValueDir = getDatabaseDirectory()
	LogLady.WithField("badger_opts", opts).Debug("Set Badger Options")

	db, err = badger.Open(opts)
	return
}

// AuditPackages will given a list of Package URLs, run an OSS Index audit
func AuditPackages(purls []string) ([]types.Coordinate, error) {
	dbDir := getDatabaseDirectory()
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Initialize the cache
	db, err := openDb(dbDir)
	customerrors.Check(err, "Error initializing cache")
	defer db.Close()

	var newPurls []string
	var results []types.Coordinate

	err = db.View(func(txn *badger.Txn) error {
		for _, purl := range purls {
			item, err := txn.Get([]byte(strings.ToLower(purl)))
			if err == nil {
				err := item.Value(func(val []byte) error {
					var coordinate types.Coordinate
					err := json.Unmarshal(val, &coordinate)
					results = append(results, coordinate)
					return err
				})
				if err != nil {
					newPurls = append(newPurls, purl)
				}
			} else {
				newPurls = append(newPurls, purl)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var chunks = chunk(newPurls, MAX_COORDS)

	for _, chunk := range chunks {
		if len(chunk) > 0 {
			var request types.AuditRequest
			request.Coordinates = chunk
			var jsonStr, _ = json.Marshal(request)

			req, err := setupRequest(jsonStr)
			if err != nil {
				return nil, err
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, err
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
				return nil, err
			}

			// Process results
			var coordinates []types.Coordinate
			if err = json.Unmarshal([]byte(body), &coordinates); err != nil {
				return nil, err
			}

			// Cache the new results
			if err := db.Update(func(txn *badger.Txn) error {
				for i := 0; i < len(coordinates); i++ {
					var coord = coordinates[i].Coordinates
					results = append(results, coordinates[i])

					var coordJson, _ = json.Marshal(coordinates[i])

					err := txn.SetWithTTL([]byte(strings.ToLower(coord)), []byte(coordJson), time.Hour*12)
					if err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return nil, err
			}
		}
	}
	return results, nil
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

func setupRequest(jsonStr []byte) (req *http.Request, err error) {
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

	return req, nil
}

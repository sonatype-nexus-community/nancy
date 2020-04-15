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
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/nancy/useragent"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"time"
)

const dbValueDirName = "golang"

const defaultOssIndexUrl = "https://ossindex.sonatype.org/api/v3/component-report"

const MAX_COORDS = 128

var (
	ossIndexUrl string
	now = time.Now()
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

func openDb(dbDir string) (db *sql.DB, err error) {
	LogLady.Debug("Attempting to open Badger DB")
	db, err = sql.Open("sqlite3", dbDir+"/nancy.sqlite")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS nancy_cache (coordinates TEXT, vulnerabilities_json TEXT, insert_time INTEGER)")
	if err != nil {
		return nil, err
	}
	return
}

func purgeExpiredEntries(db *sql.DB) error{
	db.Exec("DELETE FROM nancy_cache WHERE insert_time < ?", )
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
	dbDir := getDatabaseDirectory()
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Initialize the cache
	db, err := openDb(dbDir)
	customerrors.Check(err, "Error initializing cache")
	err = purgeExpiredEntries(db, now.Add())
	customerrors.Check(err, "Error initializing cache")
	defer db.Close()

	var newPurls []string
	var results []types.Coordinate

	lookupStmt, err := db.Prepare("SELECT vulnerabilities_json from nancy_cache where coordinates = ?")
	if err != nil {
		return nil, err
	}
	for _, purl := range purls {
		var coordJson string
		err = lookupStmt.QueryRow(purl).Scan(&coordJson)
		if err != nil {
			//not in cache
			if err == sql.ErrNoRows {
				newPurls = append(newPurls, purl)
				continue
			}
			//something weird happened, error out
			return nil, err
		}

		var coordinate types.Coordinate
		err := json.Unmarshal([]byte(coordJson), &coordinate)
		if err != nil {
			newPurls = append(newPurls, purl)
			continue
		}
		results = append(results, coordinate)

	}
	var chunks = chunk(newPurls, MAX_COORDS)

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

			LogLady.WithField("coordinates", coordinates).Info("Coordinates unmarshalled from OSS Index")

			// Cache the new results
			cacheInsertStatement, err := db.Prepare("INSERT INTO nancy_cache (coordinates, vulnerabilities_json) values (?, ?) ")
			if err != nil {
				return nil, err
			}
			time.Now().Unix()
			for i := 0; i < len(coordinates); i++ {
				coord := coordinates[i].Coordinates

				results = append(results, coordinates[i])

				coordJSON, _ := json.Marshal(coordinates[i])
				LogLady.WithField("json", coordinates[i]).Info("Marshall coordinate into json for insertion into DB")
				if err != nil {
					fmt.Println(err)
				}
				_, err = cacheInsertStatement.Exec(coord, coordJSON)
				if err != nil {
					fmt.Println(err)
				}
				//		err := txn.SetWithTTL([]byte(strings.ToLower(coord)), []byte(coordJSON), time.Hour*12)
				//		if err != nil {
				//			LogLady.WithField("error", err).Error("Unable to add coordinate to cache DB")
				//			return err
				//		}
			}
			//
			//	return nil
			//}); err != nil {
			//	return nil, err
			//}
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

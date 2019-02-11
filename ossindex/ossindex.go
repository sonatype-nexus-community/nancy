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
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"
)

const dbValueDirName = "golang"

const defaultOssIndexUrl = "https://ossindex.sonatype.org/api/v3/component-report"

var (
	ossIndexUrl string
)

func getDatabaseDirectory() (dbDir string) {
	usr, err := user.Current()
	customerrors.Check(err, "Error getting user home")

	return usr.HomeDir + "/.ossindex"
}

func getOssIndexUrl() string {
	if ossIndexUrl == "" {
		ossIndexUrl = defaultOssIndexUrl
	}
	return ossIndexUrl
}

// AuditPackages will given a list of Package URLs, run an OSS Index audit
func AuditPackages(purls []string) ([]types.Coordinate, error) {
	dbDir := getDatabaseDirectory()
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Initialize the cache
	opts := badger.DefaultOptions
	opts.Dir = dbDir + "/" + dbValueDirName
	opts.ValueDir = dbDir + "/" + dbValueDirName
	db, err := badger.Open(opts)
	customerrors.Check(err, "Error initializing cache")

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %s\n", err)
		}
	}()

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

	if len(newPurls) > 0 {
		var request types.AuditRequest
		request.Coordinates = newPurls
		var jsonStr, _ = json.Marshal(request)

		var req *http.Request
		if req, err = http.NewRequest(
			"POST",
			getOssIndexUrl(),
			bytes.NewBuffer(jsonStr)); err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		switch {
		case strings.Contains(resp.Status, "200"):
			fmt.Println(resp)
		default:
			return nil, errors.New("[" + resp.Status + "] error accessing OSS Index")
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("error closing response body: %s\n", err)
			}
		}()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
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
	return results, nil
}

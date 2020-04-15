//
// Copyright 2020-present Sonatype Inc.
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

// Package cache has definitions and functions for processing the OSS Index Feed
package cache

import (
	"encoding/json"
	"flag"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
)

const dbValueDirName = "golang"

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

func openDb(dbDir string) (db *badger.DB, err error) {
	LogLady.Debug("Attempting to open Badger DB")
	opts := badger.DefaultOptions

	opts.Dir = getDatabaseDirectory()
	opts.ValueDir = getDatabaseDirectory()
	LogLady.WithField("badger_opts", opts).Debug("Set Badger Options")

	db, err = badger.Open(opts)
	return
}

func InsertValuesIntoCache(coordinates []types.Coordinate) (err error) {
	dbDir := getDatabaseDirectory()
	if err = os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return
	}
	// Initialize the cache
	db, err := openDb(dbDir)
	customerrors.Check(err, "Error initializing cache")
	defer db.Close()

	// Cache the new results
	if err = db.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(coordinates); i++ {
			coord := coordinates[i].Coordinates

			coordJSON, _ := json.Marshal(coordinates[i])
			LogLady.WithField("json", coordinates[i]).Info("Marshall coordinate into json for insertion into DB")

			err := txn.SetWithTTL([]byte(strings.ToLower(coord)), []byte(coordJSON), time.Hour*12)
			if err != nil {
				LogLady.WithField("error", err).Error("Unable to add coordinate to cache DB")
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}
	return
}

func HydrateNewPurlsFromCache(purls []string) (newPurls []string, results []types.Coordinate, err error) {
	dbDir := getDatabaseDirectory()
	if err = os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return
	}
	// Initialize the cache
	db, err := openDb(dbDir)
	customerrors.Check(err, "Error initializing cache")
	defer db.Close()

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
		return err
	})
	return
}

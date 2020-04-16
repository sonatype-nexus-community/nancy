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
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/recoilme/pudge"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
)

// Cache is a struct with methods meant to be used for getting values from a DB cache
// DBName can be used to override the actual database name, primarily for testing
// TTL can be used to set the amount of time a specific object can live in the cache
type Cache struct {
	DBName string
	TTL    time.Time
}

const dbDirName = "nancy"

// DBValue is a local struct used for adding a TTL to a Coordinates struct
type DBValue struct {
	Coordinates types.Coordinate
	TTL         int64
}

func (c *Cache) getDatabaseDirectory() (dbDir string) {
	usr, err := user.Current()
	customerrors.Check(err, "Error getting user home")

	return path.Join(usr.HomeDir, types.OssIndexDirName, dbDirName, c.DBName)
}

// RemoveCacheDirectory deletes the local database directory.
func (c *Cache) RemoveCacheDirectory() error {
	defer func() {
		if err := pudge.CloseAll(); err != nil {
			LogLady.WithField("error", err).Error("An error occurred with closing the Pudge DB")
		}
	}()

	err := pudge.DeleteFile(c.getDatabaseDirectory())
	if err == nil {
		return nil
	}
	if _, ok := err.(*os.PathError); ok {
		LogLady.WithField("error", err).Error("Unable to delete database, looks like it doesn't exist")
		return nil
	}
	err = pudge.BackupAll(c.getDatabaseDirectory())
	if err != nil {
		return err
	}

	return err
}

// Insert takes a slice of Coordinates, and inserts them into the cache database.
// By default, values are given a TTL of 12 hours. An error is returned if there is an issue setting
// a key into the cache.
func (c *Cache) Insert(coordinates []types.Coordinate) (err error) {
	defer func() {
		if err := pudge.CloseAll(); err != nil {
			LogLady.WithField("error", err).Error("An error occurred with closing the Pudge DB")
		}
	}()

	doSet := func(coordinate types.Coordinate) error {
		err = pudge.Set(c.getDatabaseDirectory(), strings.ToLower(coordinate.Coordinates), DBValue{Coordinates: coordinate, TTL: c.TTL.Unix()})
		if err != nil {
			LogLady.WithField("error", err).Error("Unable to add coordinate to cache DB")
			return err
		}
		return nil
	}

	for i := 0; i < len(coordinates); i++ {
		var exists DBValue
		err = pudge.Get(c.getDatabaseDirectory(), strings.ToLower(coordinates[i].Coordinates), &exists)
		if err != nil {
			if errors.Is(err, pudge.ErrKeyNotFound) {
				err = doSet(coordinates[i])
				if err != nil {
					return
				}
			}
			continue
		}
		if exists.TTL < time.Now().Unix() {
			err = pudge.Delete(c.getDatabaseDirectory(), strings.ToLower(coordinates[i].Coordinates))
			if err != nil {
				fmt.Println(err)
				LogLady.WithField("error", err).Error("Unable to delete coordinate from cache DB")
				continue
			}
			err = doSet(coordinates[i])
			if err != nil {
				continue
			}
		}
	}
	return
}

// GetCacheValues takes a slice of purls, and checks to see if they are in the cache.
// It will return a new slice of purls used to talk to OSS Index (if any are not in the cache),
// a partially hydrated results slice if there are results in the cache, and an error if the world
// ends, or we wrote bad code, whichever comes first.
func (c *Cache) GetCacheValues(purls []string) ([]string, []types.Coordinate, error) {
	defer func() {
		if err := pudge.CloseAll(); err != nil {
			LogLady.WithField("error", err).Error("An error occurred with closing the Pudge DB")
		}
	}()

	var newPurls []string
	var results []types.Coordinate

	for i := 0; i < len(purls); i++ {
		var item DBValue
		err := pudge.Get(c.getDatabaseDirectory(), strings.ToLower(purls[i]), &item)
		if err != nil {
			if errors.Is(err, pudge.ErrKeyNotFound) {
				newPurls = append(newPurls, purls[i])
				continue
			} else {
				return nil, nil, err
			}
		}

		if item.TTL < time.Now().Unix() {
			newPurls = append(newPurls, purls[i])
			err = pudge.Delete(c.getDatabaseDirectory(), strings.ToLower(item.Coordinates.Coordinates))
			if err != nil {
				LogLady.WithField("error", err).Error("Unable to delete value from pudge db")
			}
			continue
		} else {
			LogLady.WithField("coordinate", item.Coordinates).Info("Result found in cache, moving forward and hydrating results")
			results = append(results, item.Coordinates)
		}
	}

	return newPurls, results, nil
}

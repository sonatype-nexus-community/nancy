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
	"flag"
	"os"
	"os/user"
	"path"

	"github.com/dgraph-io/badger"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	. "github.com/sonatype-nexus-community/nancy/logger"
	"github.com/sonatype-nexus-community/nancy/types"
)

const dbValueDirName = "golang"

func GetDatabaseDirectory() (dbDir string) {
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
	return os.RemoveAll(GetDatabaseDirectory())
}

func OpenDb(dbDir string) (db *badger.DB, err error) {
	LogLady.Debug("Attempting to open Badger DB")
	opts := badger.DefaultOptions

	opts.Dir = GetDatabaseDirectory()
	opts.ValueDir = GetDatabaseDirectory()
	LogLady.WithField("badger_opts", opts).Debug("Set Badger Options")

	db, err = badger.Open(opts)
	return
}

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
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func exists(filePath string) (exists bool) {
	exists = true

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	return
}

func setupTestCaseMoveCacheDb(t *testing.T) func(t *testing.T) {
	// temporarily move existing cache db
	cacheValueDir := getDatabaseDirectory() + "/" + dbValueDirName
	var mustRestoreExistingValueDir = exists(cacheValueDir)
	backupValueDir := getDatabaseDirectory() + "/" + dbValueDirName + "-NancyOssIndexTestBackup"
	if mustRestoreExistingValueDir {
		// move existing valueDir to backup name
		assert.Nil(t, os.Rename(cacheValueDir, backupValueDir))
	}

	return func(t *testing.T) {
		// remove valueDir created during test
		assert.Nil(t, os.RemoveAll(cacheValueDir))

		if mustRestoreExistingValueDir {
			// restore existing valueDir from backup name
			assert.Nil(t, os.Rename(backupValueDir, cacheValueDir))
		}
	}
}

func TestAuditPackages_Empty(t *testing.T) {
	coordinates, err := AuditPackages([]string{})
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	assert.Nil(t, err)
}

func TestAuditPackages_Nil(t *testing.T) {
	coordinates, err := AuditPackages(nil)
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	assert.Nil(t, err)
}

func TestAuditPackages_ErrorNonExistentPurl(t *testing.T) {
	coordinates, err := AuditPackages([]string{"nonexistent-purl"})
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	assert.Equal(t, "[400 Bad Request] error accessing OSS Index", err.Error())
}

func TestAuditPackages_NewPackage(t *testing.T) {
	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	purl := "pkg:github/BurntSushi/toml@0.3.1"
	coordinates, err := AuditPackages([]string{purl})

	lowerCasePurl := strings.ToLower(purl)
	expectedCoordinate := types.Coordinate{
		Coordinates:     lowerCasePurl,
		Reference:       "https://ossindex.sonatype.org/component/" + lowerCasePurl,
		Vulnerabilities: []types.Vulnerability{},
	}
	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

func TestAuditPackages_SinglePackage_Cached(t *testing.T) {
	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	purl := "pkg:github/BurntSushi/toml@0.3.1"
	// call twice to ensure second call always finds package in local cache
	coordinates, err := AuditPackages([]string{purl})
	assert.Nil(t, err)
	coordinates, err = AuditPackages([]string{purl})

	lowerCasePurl := strings.ToLower(purl)
	expectedCoordinate := types.Coordinate{
		Coordinates:     lowerCasePurl,
		Reference:       "https://ossindex.sonatype.org/component/" + lowerCasePurl,
		Vulnerabilities: []types.Vulnerability{},
	}
	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

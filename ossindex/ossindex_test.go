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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sonatype-nexus-community/nancy/configuration"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
)

const purl = "pkg:github/BurntSushi/toml@0.3.1"

var lowerCasePurl = strings.ToLower(purl)
var expectedCoordinate = types.Coordinate{
	Coordinates:     lowerCasePurl,
	Reference:       "https://ossindex.sonatype.org/component/" + lowerCasePurl,
	Vulnerabilities: []types.Vulnerability{},
}

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

func TestOssIndexUrlDefault(t *testing.T) {
	ossIndexUrl = ""
	assert.Equal(t, defaultOssIndexUrl, getOssIndexUrl())
}

func TestAuditPackages_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("No call should occur with empty package. called: %v", r)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages([]string{})
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	assert.Nil(t, err)
}

func TestAuditPackages_Nil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("No call should occur with nil package. called: %v", r)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages(nil)
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	assert.Nil(t, err)
}

func TestAuditPackages_ErrorHttpRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("No call should occur with nil package. called: %v", r)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL + "\\"

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages([]string{"nonexistent-purl"})
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	parseError := err.(*url.Error)
	assert.Equal(t, "parse", parseError.Op)
}

func TestAuditPackages_ErrorNonExistentPurl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/", r.URL.EscapedPath())

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages([]string{"nonexistent-purl"})
	assert.Equal(t, []types.Coordinate(nil), coordinates)
	assert.Equal(t, "[400 Bad Request] error accessing OSS Index", err.Error())
}

func TestAuditPackages_ErrorBadResponseBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/", r.URL.EscapedPath())

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("badStuff"))
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages([]string{purl})

	assert.Equal(t, []types.Coordinate(nil), coordinates)
	jsonError := err.(*json.SyntaxError)
	assert.Equal(t, int64(1), jsonError.Offset)
	assert.Equal(t, "invalid character 'b' looking for beginning of value", jsonError.Error())
}

func TestAuditPackages_NewPackage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyClientCallAndWriteValidPackageResponse(t, r, w)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages([]string{purl})

	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

func verifyClientCallAndWriteValidPackageResponse(t *testing.T, r *http.Request, w http.ResponseWriter) {
	assert.Equal(t, http.MethodPost, r.Method)
	assert.Equal(t, "/", r.URL.EscapedPath())
	w.WriteHeader(http.StatusOK)
	coordinates := []types.Coordinate{
		{
			Coordinates:     "pkg:github/burntsushi/toml@0.3.1",
			Reference:       "https://ossindex.sonatype.org/component/pkg:github/burntsushi/toml@0.3.1",
			Vulnerabilities: []types.Vulnerability{},
		},
	}
	jsonCoordinates, _ := json.Marshal(coordinates)
	_, _ = w.Write(jsonCoordinates)
}

func TestAuditPackages_SinglePackage_Cached(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("No call should occur with previously cached package. called: %v", r)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	// create the cached package
	db, err := openDb(getDatabaseDirectory())
	assert.Nil(t, err)
	assert.Nil(t, db.Update(func(txn *badger.Txn) error {
		var coordJson, _ = json.Marshal(expectedCoordinate)
		err := txn.SetWithTTL([]byte(strings.ToLower(lowerCasePurl)), []byte(coordJson), time.Hour*12)
		if err != nil {
			return err
		}
		return nil
	}))
	assert.Nil(t, db.Close())

	coordinates, err := AuditPackages([]string{purl})
	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

func TestAuditPackages_SinglePackage_Cached_WithExpiredTTL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyClientCallAndWriteValidPackageResponse(t, r, w)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	// create the cached package with short TTL for the cached item to ensure item TTL expires before we read it
	db, err := openDb(getDatabaseDirectory())
	assert.Nil(t, err)
	assert.Nil(t, db.Update(func(txn *badger.Txn) error {
		var coordJson, _ = json.Marshal(expectedCoordinate)
		err := txn.SetWithTTL([]byte(strings.ToLower(lowerCasePurl)), []byte(coordJson), time.Second*1)
		if err != nil {
			return err
		}
		return nil
	}))
	assert.Nil(t, db.Close())
	time.Sleep(2 * time.Second)

	coordinates, err := AuditPackages([]string{purl})
	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

func TestSetupRequest(t *testing.T) {
	coordJson, _ := setupJson(t)
	config := configuration.Configuration{Username: "testuser", Token: "test"}
	req, err := setupRequest(coordJson, &config)

	assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, req.Method, "POST")
	user, token, ok := req.BasicAuth()
	assert.Equal(t, user, "testuser")
	assert.Equal(t, token, "test")
	assert.Equal(t, ok, true)
	assert.Nil(t, err)
}

// TODO: Use this for more than just TestSetupRequest
func setupJson(t *testing.T) (coordJson []byte, err error) {
	coordJson, err = json.Marshal(expectedCoordinate)
	if err != nil {
		t.Errorf("Couldn't setup json")
	}

	return
}

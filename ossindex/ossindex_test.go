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
	"fmt"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
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
	assert.Equal(t, "invalid character \"\\\\\" in host name", parseError.Err.Error())
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
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	coordinates, err := AuditPackages([]string{purl})

	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

// File copies a single file from src to dst
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer func() {
		_ = srcfd.Close()
	}()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer func() {
		_ = dstfd.Close()
	}()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func TestAuditPackages_SinglePackage_Cached(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("No call should occur with nil package. called: %v", r)
	}))
	defer ts.Close()
	ossIndexUrl = ts.URL

	teardownTestCase := setupTestCaseMoveCacheDb(t)
	defer teardownTestCase(t)

	// put test db cache dir in expected location
	cacheValueDir := getDatabaseDirectory() + "/" + dbValueDirName
	assert.Nil(t, copyDir("testdata/golang", cacheValueDir))

	coordinates, err := AuditPackages([]string{purl})
	assert.Equal(t, []types.Coordinate{expectedCoordinate}, coordinates)
	assert.Nil(t, err)
}

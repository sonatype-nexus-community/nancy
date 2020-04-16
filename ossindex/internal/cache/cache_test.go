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
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
)

var coordinates []types.Coordinate

var purls []string

func TestInsert(t *testing.T) {
	cache := setupTestsAndCache(t)

	err := cache.Insert(coordinates)
	assert.Nil(t, err)

	var result DBValue
	err = cache.getKeyAndHydrate(coordinates[0].Coordinates, &result)

	assert.Equal(t, coordinates[0], result.Coordinates)
	assert.Nil(t, err)
}

func TestGetWithRegularTTL(t *testing.T) {
	cache := setupTestsAndCache(t)

	err := cache.Insert(coordinates)
	assert.Nil(t, err)

	newPurls, results, err := cache.GetCacheValues(purls)

	assert.Empty(t, newPurls)
	assert.Equal(t, results, coordinates)
	assert.Nil(t, err)
}

func TestGetWithExpiredTTL(t *testing.T) {
	cache := setupTestsAndCache(t)
	cache.TTL = time.Now().AddDate(0, 0, -1)

	err := cache.Insert(coordinates)
	assert.Nil(t, err)

	newPurls, results, err := cache.GetCacheValues(purls)

	assert.Equal(t, purls, newPurls)
	assert.Empty(t, results)
	assert.Nil(t, err)
}

func setupTestsAndCache(t *testing.T) *Cache {
	dec, _ := decimal.NewFromString("9.8")
	coordinate := types.Coordinate{
		Coordinates: "test",
		Reference:   "http://www.innernet.com",
		Vulnerabilities: []types.Vulnerability{
			{
				Id:          "id",
				Title:       "test",
				Description: "description",
				CvssScore:   dec,
				CvssVector:  "vectorvictor",
				Cve:         "CVE-123-123",
				Reference:   "http://www.internet.com",
				Excluded:    false,
			},
		},
	}

	purls = append(purls, "test")

	coordinates = append(coordinates, coordinate)
	cache := Cache{DBName: "nancy-test", TTL: time.Now().Local().Add(time.Hour * 12)}
	err := cache.RemoveCacheDirectory()
	if err != nil {
		t.Error(err)
	}
	return &cache
}

//
// Copyright 2018-present Sonatype Inc.
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

package guide

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sonatype-nexus-community/nancy/internal/ossindex"
)

const cacheTTLHours = 12

type cacheEntry struct {
	Coordinate ossindex.Coordinate `json:"coordinate"`
	ExpiresAt  time.Time           `json:"expires_at"`
}

type jsonCache struct {
	mu      sync.RWMutex
	data    map[string]cacheEntry
	path    string
	dirty   bool
}

func newJSONCache(basePath string) *jsonCache {
	c := &jsonCache{
		data: make(map[string]cacheEntry),
	}
	if basePath != "" {
		c.path = filepath.Join(basePath, "nancy-guide-cache.json")
	} else {
		home, _ := os.UserHomeDir()
		c.path = filepath.Join(home, ossindex.OssIndexDirName, "nancy-guide-cache.json")
	}
	c.load()
	return c
}

func (c *jsonCache) get(purl string) (ossindex.Coordinate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.data[purl]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return ossindex.Coordinate{}, false
	}
	return entry.Coordinate, true
}

func (c *jsonCache) set(purl string, coord ossindex.Coordinate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[purl] = cacheEntry{
		Coordinate: coord,
		ExpiresAt:  time.Now().Add(cacheTTLHours * time.Hour),
	}
	c.dirty = true
	_ = c.save()
}

func (c *jsonCache) removeAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
	if c.path == "" {
		return nil
	}
	err := os.Remove(c.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (c *jsonCache) load() {
	if c.path == "" {
		return
	}
	data, err := os.ReadFile(c.path) // #nosec
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &c.data)
}

func (c *jsonCache) save() error {
	if c.path == "" {
		return nil
	}
	_ = os.MkdirAll(filepath.Dir(c.path), 0700)
	data, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}

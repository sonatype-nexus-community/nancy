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

// Package guide provides a client for the Sonatype Guide API.
package guide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/internal/ossindex"
)

const (
	// DefaultGuideURL is the default Sonatype Guide API URL for component reports.
	DefaultGuideURL = "https://ossindex.sonatype.org/api/v3/component-report"

	// maxCoords is the maximum number of coordinates to send per request.
	maxCoords = 128
)

// Options configures the Guide API client.
type Options struct {
	// Username is the OSS Index/Guide username (email). Leave empty to use bearer token auth.
	Username string
	// Token is either the OSS Index API token (when Username is set) or a Guide Bearer token.
	Token string
	// ServerURL overrides the default Guide API URL.
	ServerURL string
	// DBCachePath is the directory for the response cache.
	DBCachePath string
}

// Server implements ossindex.IServer using the Sonatype Guide API.
type Server struct {
	opts    Options
	logger  *logrus.Logger
	client  *http.Client
	cache   *jsonCache
}

// New creates a new Guide API Server.
func New(logger *logrus.Logger, opts Options) *Server {
	if opts.ServerURL == "" {
		opts.ServerURL = DefaultGuideURL
	}
	return &Server{
		opts:   opts,
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  newJSONCache(opts.DBCachePath),
	}
}

// NoCacheNoProblems removes the local cache.
func (s *Server) NoCacheNoProblems() error {
	return s.cache.removeAll()
}

// AuditPackages checks a list of PURLs against the Sonatype Guide API.
func (s *Server) AuditPackages(purls []string) ([]ossindex.Coordinate, error) {
	var allCoords []ossindex.Coordinate

	for i := 0; i < len(purls); i += maxCoords {
		end := i + maxCoords
		if end > len(purls) {
			end = len(purls)
		}
		batch := purls[i:end]

		// Check cache first
		var uncached []string
		cachedCoords := make(map[string]ossindex.Coordinate)
		for _, p := range batch {
			if c, ok := s.cache.get(p); ok {
				cachedCoords[p] = c
			} else {
				uncached = append(uncached, p)
			}
		}

		var batchResult []ossindex.Coordinate
		if len(uncached) > 0 {
			fetched, err := s.fetchBatch(uncached)
			if err != nil {
				return nil, err
			}
			for _, c := range fetched {
				s.cache.set(c.Coordinates, c)
				batchResult = append(batchResult, c)
			}
		}

		// Preserve order: cached first in order of purls
		for _, p := range batch {
			if c, ok := cachedCoords[p]; ok {
				allCoords = append(allCoords, c)
			}
		}
		allCoords = append(allCoords, batchResult...)
	}

	return allCoords, nil
}

type auditRequest struct {
	Coordinates []string `json:"coordinates"`
}

type auditResponseItem struct {
	Coordinates     string          `json:"coordinates"`
	Reference       string          `json:"reference"`
	Vulnerabilities []apiVuln       `json:"vulnerabilities"`
}

type apiVuln struct {
	ID          string  `json:"id"`
	Title       string  `json:"displayName"`
	Description string  `json:"description"`
	CvssScore   float64 `json:"cvssScore"`
	CvssVector  string  `json:"cvssVector"`
	Cve         string  `json:"cve"`
	Reference   string  `json:"reference"`
}

func (s *Server) fetchBatch(purls []string) ([]ossindex.Coordinate, error) {
	body, err := json.Marshal(auditRequest{Coordinates: purls})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.opts.ServerURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("nancy-client/%s (%s/%s)",
		buildversion.BuildVersion, runtime.GOOS, runtime.GOARCH))

	if s.opts.Username != "" {
		req.SetBasicAuth(s.opts.Username, s.opts.Token)
	} else if s.opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.opts.Token)
	}

	s.logger.WithField("url", s.opts.ServerURL).Debug("Sending batch to Guide API")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("guide API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading guide API response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests:
		return nil, &ossindex.RateLimitError{}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &ossindex.ServerError{Message: fmt.Sprintf("authentication failed (%d): check your credentials", resp.StatusCode)}
	default:
		return nil, &ossindex.ServerError{Message: fmt.Sprintf("Guide API returned %d: %s", resp.StatusCode, string(respBody))}
	}

	var items []auditResponseItem
	if err = json.Unmarshal(respBody, &items); err != nil {
		return nil, fmt.Errorf("parsing guide API response: %w", err)
	}

	coords := make([]ossindex.Coordinate, 0, len(items))
	for _, item := range items {
		c := ossindex.Coordinate{
			Coordinates: item.Coordinates,
			Reference:   item.Reference,
		}
		for _, v := range item.Vulnerabilities {
			c.Vulnerabilities = append(c.Vulnerabilities, ossindex.Vulnerability{
				ID:          v.ID,
				Title:       v.Title,
				Description: v.Description,
				CvssScore:   decimal.NewFromFloat(v.CvssScore),
				CvssVector:  v.CvssVector,
				Cve:         v.Cve,
				Reference:   v.Reference,
			})
		}
		coords = append(coords, c)
	}
	return coords, nil
}

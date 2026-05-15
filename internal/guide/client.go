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
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	sonatypeguide "github.com/sonatype-nexus-community/sonatype-guide-api-client-go"
	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/internal/ossindex"
)

const (
	// DefaultGuideURL is the default Sonatype Guide API base URL.
	DefaultGuideURL = "https://api.guide.sonatype.com"

	// maxCoords is the maximum number of coordinates to send per request.
	maxCoords = 128
)

// Options configures the Guide API client.
type Options struct {
	// Username is the OSS Index/Guide username (email). Leave empty to use bearer token auth.
	Username string
	// Token is either the OSS Index API token (when Username is set) or a Guide Bearer token.
	Token string
	// ServerURL overrides the default Guide API base URL.
	ServerURL string
	// DBCachePath is the directory for the response cache.
	DBCachePath string
}

// Server implements ossindex.IServer using the Sonatype Guide API.
type Server struct {
	opts    Options
	logger  *logrus.Logger
	api     *sonatypeguide.APIClient
	authCtx context.Context
	cache   *jsonCache
}

// New creates a new Guide API Server.
func New(logger *logrus.Logger, opts Options) *Server {
	if opts.ServerURL == "" {
		opts.ServerURL = DefaultGuideURL
	}

	cfg := sonatypeguide.NewConfiguration()
	cfg.Servers = sonatypeguide.ServerConfigurations{{URL: opts.ServerURL}}
	cfg.UserAgent = fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion)

	authCtx := context.Background()
	if opts.Username != "" {
		authCtx = context.WithValue(authCtx, sonatypeguide.ContextBasicAuth, sonatypeguide.BasicAuth{
			UserName: opts.Username,
			Password: opts.Token,
		})
	} else if opts.Token != "" {
		authCtx = context.WithValue(authCtx, sonatypeguide.ContextAccessToken, opts.Token)
	}

	return &Server{
		opts:    opts,
		logger:  logger,
		api:     sonatypeguide.NewAPIClient(cfg),
		authCtx: authCtx,
		cache:   newJSONCache(opts.DBCachePath),
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

		var uncached []string
		cachedCoords := make(map[string]ossindex.Coordinate)
		for _, p := range batch {
			if c, ok := s.cache.get(p); ok {
				cachedCoords[p] = c
			} else {
				uncached = append(uncached, p)
			}
		}

		if len(uncached) > 0 {
			fetched, err := s.fetchBatch(uncached)
			if err != nil {
				return nil, err
			}
			for _, c := range fetched {
				s.cache.set(c.Coordinates, c)
			}
			// Merge fetched into cachedCoords for ordering pass below
			for _, c := range fetched {
				cachedCoords[c.Coordinates] = c
			}
		}

		for _, p := range batch {
			if c, ok := cachedCoords[p]; ok {
				allCoords = append(allCoords, c)
			}
		}
	}

	return allCoords, nil
}

func (s *Server) fetchBatch(purls []string) ([]ossindex.Coordinate, error) {
	s.logger.WithField("count", len(purls)).Debug("Sending batch to Guide API")

	body := sonatypeguide.NewPurlRequestPost(purls)
	reports, _, err := s.api.OSSIndexCompatibilityAPI.GetComponentReports(s.authCtx).
		PurlRequestPost(*body).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("guide API request failed: %w", err)
	}

	coords := make([]ossindex.Coordinate, 0, len(reports))
	for _, r := range reports {
		c := ossindex.Coordinate{
			Coordinates: r.GetCoordinates(),
			Reference:   r.GetReference(),
		}
		for _, v := range r.GetVulnerabilities() {
			c.Vulnerabilities = append(c.Vulnerabilities, ossindex.Vulnerability{
				ID:          v.GetId(),
				Title:       v.GetTitle(),
				Description: v.GetDescription(),
				CvssScore:   decimal.NewFromFloat(v.GetCvssScore()),
				CvssVector:  v.GetCvssVector(),
				Cve:         v.GetCve(),
				Reference:   v.GetReference(),
			})
		}
		coords = append(coords, c)
	}
	return coords, nil
}

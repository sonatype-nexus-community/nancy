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
	"github.com/sonatype-nexus-community/nancy/v2/buildversion"
	"github.com/sonatype-nexus-community/nancy/v2/internal/ossindex"
)

const (
	// DefaultGuideURL is the default Sonatype Guide API base URL.
	DefaultGuideURL = "https://api.guide.sonatype.com"

	// maxCoords is the maximum number of coordinates to send per request.
	maxCoords = 128
)

// Options configures the Guide API client.
type Options struct {
	// GuideToken is the Sonatype Guide Bearer token (preferred auth method).
	GuideToken string
	// Username is the OSS Index username (email). Deprecated: use GuideToken.
	Username string
	// Token is the OSS Index API token when Username is set. Deprecated: use GuideToken.
	// For backward compat, if Username is empty and Token is non-empty, Token is used as a bearer token.
	Token string
	// ServerURL overrides the default Guide API base URL.
	ServerURL string
	// DBCachePath is the directory for the response cache.
	DBCachePath string
}

// Server implements ossindex.IServer using the Sonatype Guide API.
type Server struct {
	opts   Options
	logger *logrus.Logger
	api    *sonatypeguide.APIClient
	cache  *jsonCache
}

// New creates a new Guide API Server.
func New(logger *logrus.Logger, opts Options) *Server {
	if opts.ServerURL == "" {
		opts.ServerURL = DefaultGuideURL
	}

	cfg := sonatypeguide.NewConfiguration()
	cfg.Servers = sonatypeguide.ServerConfigurations{{URL: opts.ServerURL}}
	cfg.UserAgent = fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion)

	return &Server{
		opts:   opts,
		logger: logger,
		api:    sonatypeguide.NewAPIClient(cfg),
		cache:  newJSONCache(opts.DBCachePath),
	}
}

// authContext builds a context carrying the configured credentials.
func (s *Server) authContext() context.Context {
	ctx := context.Background()
	switch {
	case s.opts.GuideToken != "":
		return context.WithValue(ctx, sonatypeguide.ContextAccessToken, s.opts.GuideToken)
	case s.opts.Username != "":
		return context.WithValue(ctx, sonatypeguide.ContextBasicAuth, sonatypeguide.BasicAuth{
			UserName: s.opts.Username,
			Password: s.opts.Token,
		})
	case s.opts.Token != "":
		// Legacy: bearer token passed via --token with empty --username (pre-v2 pattern).
		return context.WithValue(ctx, sonatypeguide.ContextAccessToken, s.opts.Token)
	default:
		return ctx
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
		coords, err := s.processBatch(purls[i:end])
		if err != nil {
			return nil, err
		}
		allCoords = append(allCoords, coords...)
	}
	return allCoords, nil
}

func (s *Server) processBatch(batch []string) ([]ossindex.Coordinate, error) {
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
			cachedCoords[c.Coordinates] = c
		}
	}

	coords := make([]ossindex.Coordinate, 0, len(batch))
	for _, p := range batch {
		if c, ok := cachedCoords[p]; ok {
			coords = append(coords, c)
		}
	}
	return coords, nil
}

func (s *Server) fetchBatch(purls []string) ([]ossindex.Coordinate, error) {
	s.logger.WithField("count", len(purls)).Debug("Sending batch to Guide API")

	body := sonatypeguide.NewPurlRequestPost(purls)
	reports, _, err := s.api.OSSIndexCompatibilityAPI.GetComponentReports(s.authContext()).
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

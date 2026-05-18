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

// Package iq provides a client for Sonatype Lifecycle (Nexus IQ Server).
package iq

import (
	"context"
	"fmt"
	"strings"
	"time"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/buildversion"
)

// Valid policy action values returned by Lifecycle.
const (
	PolicyActionNone    = "None"
	PolicyActionWarning = "Warning"
	PolicyActionFailure = "Failure"
)

const defaultMaxRetries = 300

// StatusURLResult holds the result of a Lifecycle policy evaluation.
type StatusURLResult struct {
	PolicyAction          string
	ReportHTMLURL         string
	AbsoluteReportHTMLURL string
	IsError               bool
	ErrorMessage          string
}

// IServer is the interface for Lifecycle policy evaluation.
type IServer interface {
	AuditPackages(purls []string) (StatusURLResult, error)
}

// Options configures the Lifecycle client.
type Options struct {
	User        string
	Token       string
	Stage       string
	Application string
	Server      string
	MaxRetries  int
}

// Server implements IServer using nexus-iq-api-client-go.
type Server struct {
	Options Options
	logger  *logrus.Logger
	client  *sonatypeiq.APIClient
}

// New creates a new Lifecycle Server.
func New(logger *logrus.Logger, opts Options) (*Server, error) {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = defaultMaxRetries
	}

	cfg := sonatypeiq.NewConfiguration()
	cfg.Servers = sonatypeiq.ServerConfigurations{
		{URL: opts.Server, Description: "Sonatype Lifecycle"},
	}
	cfg.UserAgent = fmt.Sprintf("nancy-client/%s", buildversion.BuildVersion)

	client := sonatypeiq.NewAPIClient(cfg)

	return &Server{
		Options: opts,
		logger:  logger,
		client:  client,
	}, nil
}

// AuditPackages submits a list of PURLs to Lifecycle for policy evaluation.
func (s *Server) AuditPackages(purls []string) (StatusURLResult, error) {
	ctx := context.WithValue(context.Background(),
		sonatypeiq.ContextBasicAuth,
		sonatypeiq.BasicAuth{UserName: s.Options.User, Password: s.Options.Token},
	)

	internalAppID, err := s.resolveAppID(ctx)
	if err != nil {
		return StatusURLResult{IsError: true, ErrorMessage: err.Error()}, err
	}

	scanResp, _, err := s.client.ThirdPartyAnalysisAPI.
		ScanComponents(ctx, internalAppID, "nancy").
		StageId(s.Options.Stage).
		Body(buildCycloneDXSBOM(purls)).
		Execute()
	if err != nil {
		return StatusURLResult{IsError: true, ErrorMessage: err.Error()}, err
	}

	return s.pollForResult(ctx, internalAppID, extractScanRequestId(scanResp.GetStatusUrl()))
}

func (s *Server) resolveAppID(ctx context.Context) (string, error) {
	appResp, _, err := s.client.ApplicationsAPI.GetApplications(ctx).
		PublicId([]string{s.Options.Application}).
		Execute()
	if err != nil {
		return "", err
	}
	if appResp == nil || len(appResp.GetApplications()) == 0 {
		return "", fmt.Errorf("application with public ID %q not found", s.Options.Application)
	}
	return appResp.GetApplications()[0].GetId(), nil
}

func (s *Server) pollForResult(ctx context.Context, internalAppID, scanRequestID string) (StatusURLResult, error) {
	for i := 0; i < s.Options.MaxRetries; i++ {
		time.Sleep(1 * time.Second)

		result, _, err := s.client.ThirdPartyAnalysisAPI.
			GetScanStatus(ctx, internalAppID, scanRequestID).
			Execute()
		if err != nil {
			return StatusURLResult{IsError: true, ErrorMessage: err.Error()}, err
		}
		if result.IsError != nil && *result.IsError {
			msg := ""
			if result.ErrorMessage != nil {
				msg = *result.ErrorMessage
			}
			return StatusURLResult{IsError: true, ErrorMessage: msg}, nil
		}
		if result.PolicyAction != nil {
			reportURL := ""
			if result.ReportHtmlUrl != nil {
				reportURL = s.Options.Server + "/" + *result.ReportHtmlUrl
			}
			return StatusURLResult{
				PolicyAction:          *result.PolicyAction,
				AbsoluteReportHTMLURL: reportURL,
			}, nil
		}
	}
	return StatusURLResult{IsError: true, ErrorMessage: "timed out waiting for Lifecycle evaluation"}, nil
}

func buildCycloneDXSBOM(purls []string) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" ?><bom xmlns="http://cyclonedx.org/schema/bom/1.1" version="1"><components>`)
	for _, p := range purls {
		sb.WriteString(`<component type="library"><purl>`)
		sb.WriteString(p)
		sb.WriteString(`</purl></component>`)
	}
	sb.WriteString(`</components></bom>`)
	return sb.String()
}

func extractScanRequestId(statusURL string) string {
	// statusUrl: api/v2/scan/applications/{appId}/status/{scanRequestId}
	parts := strings.Split(statusURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return statusURL
}

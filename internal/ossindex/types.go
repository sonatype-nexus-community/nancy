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

// Package ossindex provides types and interfaces for vulnerability scanning.
package ossindex

import (
	"fmt"
	"path/filepath"

	"github.com/shopspring/decimal"
)

const (
	OssIndexDirName        = ".ossindex"
	OssIndexConfigFileName = ".oss-index-config"
	IQServerDirName        = ".iqserver"
	IQServerConfigFileName = ".iq-server-config"
)

// GetOssIndexDirectory returns the path to the OSS Index config directory.
func GetOssIndexDirectory(homeDir string) string {
	return filepath.Join(homeDir, OssIndexDirName)
}

// GetOssIndexConfigFile returns the path to the OSS Index config file.
func GetOssIndexConfigFile(homeDir string) string {
	return filepath.Join(GetOssIndexDirectory(homeDir), OssIndexConfigFileName)
}

// GetIQServerDirectory returns the path to the IQ Server config directory.
func GetIQServerDirectory(homeDir string) string {
	return filepath.Join(homeDir, IQServerDirName)
}

// GetIQServerConfigFile returns the path to the IQ Server config file.
func GetIQServerConfigFile(homeDir string) string {
	return filepath.Join(GetIQServerDirectory(homeDir), IQServerConfigFileName)
}

// IServer is the interface for vulnerability scanning backends.
type IServer interface {
	NoCacheNoProblems() error
	AuditPackages(p []string) ([]Coordinate, error)
}

// Coordinate represents a package coordinate with its associated vulnerabilities.
type Coordinate struct {
	Coordinates     string
	Reference       string
	Vulnerabilities []Vulnerability
	InvalidSemVer   bool
}

// Vulnerability represents a security vulnerability associated with a package.
type Vulnerability struct {
	ID          string
	Title       string
	Description string
	CvssScore   decimal.Decimal
	CvssVector  string
	Cve         string
	Reference   string
	Excluded    bool
}

// IsVulnerable returns true if the coordinate has at least one non-excluded vulnerability.
func (c Coordinate) IsVulnerable() bool {
	for _, v := range c.Vulnerabilities {
		if !v.Excluded {
			return true
		}
	}
	return false
}

// ExcludeVulnerabilities marks vulnerabilities as excluded if they appear in the exclusion list.
func (c *Coordinate) ExcludeVulnerabilities(exclusions []string) {
	for i := range c.Vulnerabilities {
		c.Vulnerabilities[i].maybeExclude(exclusions)
	}
}

func (v *Vulnerability) maybeExclude(exclusions []string) {
	for _, ex := range exclusions {
		if v.Cve == ex || v.ID == ex {
			v.Excluded = true
		}
	}
}

// RateLimitError is returned when the server rate-limits the request.
type RateLimitError struct{}

func (o *RateLimitError) Error() string {
	return `You have been rate limited by OSS Index.
If you do not have an account, please visit https://ossindex.sonatype.org/user/register to register.
After registering and verifying your account, retrieve your username (Email Address) and API Token
at https://ossindex.sonatype.org/user/settings. Then run 'nancy config' to save credentials.`
}

// ServerError represents an error from the vulnerability scanning server.
type ServerError struct {
	Err     error
	Message string
}

func (o *ServerError) Error() string {
	if o.Err != nil {
		return fmt.Sprintf("An error occurred: %s, err: %s", o.Message, o.Err.Error())
	}
	return fmt.Sprintf("An error occurred: %s", o.Message)
}

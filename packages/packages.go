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
package packages

import "regexp"

var gopkg1Pattern = regexp.MustCompile("^gopkg.in/([^.]+).*")
var gopkg2Pattern = regexp.MustCompile("^gopkg.in/([^/]+)/([^.]+).*")
var githubPattern = regexp.MustCompile("^github.com/([^/]+)/([^/]+).*")

// Packages is meant to be implemented for any package format such as dep, go mod, etc..
type Packages interface {
	ExtractPurlsFromManifest() []string
	CheckExistenceOfManifest() bool
}

// convertGopkgNameToPurl will convert the Gopkg name into a Package URL
//
// FIXME: Research the various Gopkg name formats and convert them correctly
func convertGopkgNameToPurl(name string) (rename string) {
	switch {
	case githubPattern.MatchString(name):
		rename = githubPattern.ReplaceAllString(name, "golang/$1/$2")
	case gopkg2Pattern.MatchString(name):
		rename = gopkg2Pattern.ReplaceAllString(name, "golang/$1/$2")
	case gopkg1Pattern.MatchString(name):
		rename = gopkg1Pattern.ReplaceAllString(name, "golang/go-$1/$1")
	default:
		rename = "golang/" + name
	}
	return
}

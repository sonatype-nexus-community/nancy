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

import (
	"strings"
)

// Packages is meant to be implemented for any package format such as dep, go mod, etc..
type Packages interface {
	ExtractPurlsFromManifest() []string
	CheckExistenceOfManifest() bool
}

// convertGopkgNameToPurl will convert the Gopkg name into a Package URL
//
// FIXME: Research the various Gopkg name formats and convert them correctly
func convertGopkgNameToPurl(name string) string {
	switch {
	case strings.Contains(name, "github.com"):
		name = strings.Replace(name, "github.com", "github", 1)
	case strings.Contains(name, "gopkg.in"):
		name = strings.Replace(name, "gopkg.in", "github", 1)
	case strings.Contains(name, "golang.org"):
		name = strings.Replace(name, "golang.org", "golang", 1)
	}
	return name
}

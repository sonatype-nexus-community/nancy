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

package buildversion

var (
	// these are overwritten/populated via build CLI
	BuildVersion = "0.0.0-dev"
	BuildTime    = ""
	BuildCommit  = ""
)

// PackageManager defines the package manager which was used to install the CLI.
// You can override this value using -X flag to the compiler ldflags. This is
// overridden when we build for Homebrew.
var packageManager = "source"

func PackageManager() string {
	return packageManager
}

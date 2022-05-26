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

import (
	"fmt"
	"runtime/debug"
)

const DefaultVersion = "0.0.0-dev"

var (
	// these are overwritten/populated via build CLI
	BuildVersion = DefaultVersion
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

func init() {
	// Use build info from debug package if available, and if no build info is
	// provided via build CLI.
	info, available := debug.ReadBuildInfo()
	// info.Main.Version will be "" when debugging, and "(devel)" when building with no arguments
	if available && info.Main.Version != "" && info.Main.Version != "(devel)" && BuildTime == "" && BuildCommit == "" && BuildVersion == DefaultVersion {
		BuildVersion = info.Main.Version
		BuildCommit = fmt.Sprintf("(unknown, mod sum: %q)", info.Main.Sum)
		BuildTime = "(unknown)"
	}

}

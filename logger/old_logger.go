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

// Package logger has functions to obtain a logger, and helpers for setting up where the logger writes
package logger

import (
	"fmt"
	"os"
	"path"

	"github.com/sonatype-nexus-community/nancy/types"
)

// GetLogFileLocation will return the location on disk of the log file
//
// Deprecated: Please use LogFileLocation() instead
func GetLogFileLocation() (result string) {
	result, _ = os.UserHomeDir()
	err := os.MkdirAll(path.Join(result, types.OssIndexDirName), os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	result = path.Join(result, types.OssIndexDirName, DefaultLogFile)
	return
}

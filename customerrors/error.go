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
package customerrors

import (
	"fmt"
	"os"

	"github.com/sonatype-nexus-community/nancy/buildversion"
	. "github.com/sonatype-nexus-community/nancy/logger"
)

type SwError struct {
	Message string
	Err     error
}

func (sw SwError) Error() string {
	return fmt.Sprintf("%s - error: %s", sw.Message, sw.Err.Error())
}

func (sw SwError) Exit() {
	os.Exit(3)
}

func Check(err error, message string) {
	if err != nil {
		location, _ := LogFileLocation()
		myErr := SwError{Message: message, Err: err}
		LogLady.WithField("error", err).Error(message)
		fmt.Println(myErr.Error())
		fmt.Printf("For more information, check the log file at %s\n", location)
		fmt.Println("nancy version:", buildversion.BuildVersion)
		myErr.Exit()
	}
}

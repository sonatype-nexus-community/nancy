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

package customerrors

import (
	"fmt"

	"github.com/sonatype-nexus-community/nancy/buildversion"
	"github.com/sonatype-nexus-community/nancy/logger"
)

type ErrorShowLogPath struct {
	Err error
}

func (es ErrorShowLogPath) Error() string {
	var errString string
	if es.Err != nil {
		errString = es.Err.Error()
	} else {
		errString = ""
	}

	return errString + "\n" + getLogFileMessage()
}

type ErrorExit struct {
	Message  string
	Err      error
	ExitCode int
}

func (ee ErrorExit) Error() string {
	var errString string
	if ee.Err != nil {
		errString = ee.Err.Error()
	} else {
		errString = ""
	}

	if ee.Message != "" {
		return fmt.Sprintf("exit code: %d - %s - error: %s", ee.ExitCode, ee.Message, errString)
	} else {
		return fmt.Sprintf("exit code: %d - error: %s", ee.ExitCode, errString)
	}
}

func NewErrorExitPrintHelp(errCause error, message string) ErrorExit {
	myErr := ErrorExit{message, errCause, 3}
	// LogLady.WithField("error", errCause).Error(message)
	fmt.Println(myErr.Error())

	fmt.Print(getLogFileMessage())
	return myErr
}

func getLogFileMessage() string {
	var logFile string
	var logFileErr error
	if logFile, logFileErr = logger.LogFileLocation(); logFileErr != nil {
		logFile = "unknown"
	}

	return fmt.Sprintf("For more information, check the log file at %s\n"+
		"nancy version: %s\n", logFile, buildversion.BuildVersion)
}

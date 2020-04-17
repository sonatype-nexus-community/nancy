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
	. "github.com/sonatype-nexus-community/nancy/logger"
	"os"
	"runtime"
)

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
	return fmt.Sprintf("exit code: %d - %s - error: %s", ee.ExitCode, ee.Message, errString)
}

func NewErrorExitPrintHelp(errCause error, message string) ErrorExit {
	myErr := ErrorExit{message, errCause, 3}
	LogLady.WithField("error", errCause).Error(message)
	fmt.Println(myErr.Error())

	var logFile string
	var logFileErr error
	if logFile, logFileErr = LogFileLocation(); logFileErr != nil {
		logFile = "unknown"
	}

	fmt.Printf("For more information, check the log file at %s\n", logFile)
	fmt.Println("nancy version:", buildversion.BuildVersion)
	return myErr
}

func getCallerFunction(skip int) string {
	if skip > 10 {
		LogLady.Errorf("getCallerFunction called with invalid skip value: %d", skip)
	}
	programCounters := make([]uintptr, 10)
	runtime.Callers(0, programCounters)

	// for debugging
	/*	callerNames := [10]string{}
		for idx, pc := range programCounters {
			if pc != 0 {
				callerNames[idx] = runtime.FuncForPC(programCounters[idx]).Name()
			}
		}
	*/
	return runtime.FuncForPC(programCounters[skip]).Name()
}

func Exit(code int) error {
	activeExiter.Exit(code)
	return GetBypassError(code, getCallerFunction(3))
}

func GetBypassError(code int, callerFunction string) error {
	return fmt.Errorf("exit was bypassed, code: %d, called by: %s", code, callerFunction)
}

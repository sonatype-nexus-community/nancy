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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError(t *testing.T) {
	assert.Equal(t, "exit code: 2 - MyMessage - error: MyError",
		ErrorExit{Message: "MyMessage", Err: fmt.Errorf("MyError"), ExitCode: 2}.Error())
	assert.Equal(t, "exit code: 0 - MyMessage - error: MyError",
		ErrorExit{Message: "MyMessage", Err: fmt.Errorf("MyError")}.Error())
	assert.Equal(t, "exit code: 0 - MyMessage - error: ",
		ErrorExit{Message: "MyMessage"}.Error())
	assert.Equal(t, "exit code: 0 - error: ",
		ErrorExit{}.Error())
}

func TestNewErrorExitPrintHelp(t *testing.T) {
	assert.Equal(t, "exit code: 3 - MyMessage - error: MyError",
		NewErrorExitPrintHelp(fmt.Errorf("MyError"), "MyMessage").Error())
	assert.Equal(t, "exit code: 3 - MyMessage - error: ",
		NewErrorExitPrintHelp(nil, "MyMessage").Error())
	assert.Equal(t, "exit code: 3 - error: ",
		NewErrorExitPrintHelp(nil, "").Error())
}

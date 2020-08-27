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

package audit

import (
	"testing"

	. "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestJsonOutpu(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := JsonFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)
	assert.Equal(t, `{"another":"me","stuff":1}`, string(logMessage))
}

func TestJsonPrettyPrintOutpu(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := JsonFormatter{PrettyPrint: true}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)
	assert.Equal(t, `{
  "another": "me",
  "stuff": 1
}`, string(logMessage))
}

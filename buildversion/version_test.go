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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultVersion(t *testing.T) {
	assert.Equal(t, "0.0.0-dev", BuildVersion)
}

func TestDefaultBuildTime(t *testing.T) {
	assert.Equal(t, "", BuildTime)
}

func TestDefaultBuildCommit(t *testing.T) {
	assert.Equal(t, "", BuildCommit)
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		expected   string
		input      string
		shouldFail bool
	}{
		{
			input:    "1.1.1",
			expected: "1.1.1",
		},
		{
			input:    "v1.1.1",
			expected: "1.1.1",
		},
		{
			input:      "x1.1.1",
			shouldFail: true,
		},
		{
			input:      "vv1.1.1",
			shouldFail: true,
		},
		{
			input:      "1.1",
			shouldFail: true,
		},
		{
			input:      "foobar",
			shouldFail: true,
		},
	}
	for _, test := range tests {
		actual, err := NormalizeVersion(test.input)
		if test.shouldFail {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expected, actual)
	}
}

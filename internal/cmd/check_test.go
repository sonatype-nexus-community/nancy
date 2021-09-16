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

package cmd

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdates(t *testing.T) {
	logLady, _ = test.NewNullLogger()

	// NOTE: will not actually run check unless last_update_check is old or file is removed.
	// Still, just having this test helped me find a slug bug. Hey, that rhymes.
	// Can add real harness setup/teardown later if desired.
	assert.Nil(t, checkForUpdates("", false))
}

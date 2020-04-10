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

package packages

import (
	"testing"
)

const (
	GolangResult        = "golang/golang.org/x/net"
	GitHubResult        = "golang/github.com/sonatype-nexus-community/nancy"
	GitHubSubPathResult = "golang/github.com/subutai-io/base/agent/lib/net/p2p"
	GoPkgIn1Result      = "golang/github.com/go-name/name"
	GoPkgIn2Result      = "golang/github.com/owner/name"
)

func TestConvertGopkgNameToPurl(t *testing.T) {
	result := convertGopkgNameToPurl("github.com/sonatype-nexus-community/nancy")

	if result != GitHubResult {
		t.Errorf("Conversion did not work, got back %s, but expected %s", result, GitHubResult)
	}

	result = convertGopkgNameToPurl("golang.org/x/net")

	if result != GolangResult {
		t.Errorf("Conversion did not work, got back %s, but expected %s", result, GolangResult)
	}

	result = convertGopkgNameToPurl("gopkg.in/name")

	if result != GoPkgIn1Result {
		t.Errorf("Conversion did not work, got back %s, but expected %s", result, GoPkgIn1Result)
	}

	result = convertGopkgNameToPurl("gopkg.in/owner/name")

	if result != GoPkgIn2Result {
		t.Errorf("Conversion did not work, got back %s, but expected %s", result, GoPkgIn2Result)
	}

	result = convertGopkgNameToPurl("github.com/subutai-io/base/agent/lib/net/p2p")
	if result != GitHubSubPathResult {
		t.Errorf("Conversion did not work, got back %s, but expected %s", result, GoPkgIn2Result)
	}

}

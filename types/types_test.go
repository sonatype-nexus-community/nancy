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
package types

import "testing"

func TestCveRemoveWhiteSpace(t *testing.T) {
	cve := CveListFlag{}
	test := "CVE-123, CVE-456, CVE-0020-20"
	result := []string{"CVE-123", "CVE-456", "CVE-0020-20"}
	cve.Set(test)

	if len(cve.Cves) != 3 {
		t.Errorf("Split unsuccessful")
	}

	for i := range cve.Cves {
		if cve.Cves[i] != result[i] {
			t.Errorf("Slices do not match")
		}
	}
}

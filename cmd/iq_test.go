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
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestIqApplicationFlagMissing(t *testing.T) {
	output, err := executeCommand(rootCmd, "iq")
	checkStringContains(t, output, "Error: required flag(s) \"application\" not set")
	assert.NotNil(t, err)
	checkStringContains(t, err.Error(), "required flag(s) \"application\" not set")
}

func TestIqHelp(t *testing.T) {
	output, err := executeCommand(rootCmd, "iq", "--help")
	checkStringContains(t, output, "go list -m -json all | nancy iq --application your_public_application_id --server ")
	assert.Nil(t, err)
}

func TestInitIQConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := setupConfig(t)
	defer resetConfig(t, tempDir)

	cfgDir := path.Join(tempDir, types.OssIndexDirName)
	assert.Nil(t, os.Mkdir(cfgDir, 0700))

	cfgFile = path.Join(tempDir, types.OssIndexDirName, types.OssIndexConfigFileName)

	const credentials = "username: iqUsername\n" +
		"token: iqToken\n" +
		"server: iqServer"
	assert.Nil(t, ioutil.WriteFile(cfgFile, []byte(credentials), 0644))

	initIQConfig()

	assert.Equal(t, "iqUsername", viper.GetString("username"))
	assert.Equal(t, "iqToken", viper.GetString("token"))
	assert.Equal(t, "iqServer", viper.GetString("server"))
}

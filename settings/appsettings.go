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

package settings

import (
	ossIndexTypes "github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

// UpdateCheck is used to represent settings for checking for updates of the CLI.
type UpdateCheck struct {
	LastUpdateCheck time.Time `yaml:"last_update_check"`
	FileUsed        string    `yaml:"-"`
}

// WriteToDisk will write the last update check to disk by serializing the YAML
func (upd *UpdateCheck) WriteToDisk() error {
	enc, err := yaml.Marshal(&upd)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(upd.FileUsed, enc, 0600)
	return err
}

// Load will read the update check settings from the user's disk and then deserialize it into the current instance.
func (upd *UpdateCheck) Load() error {
	appSettingsPath := filepath.Join(AppSettingsPath(), updateCheckFilename())

	if err := ensureSettingsFileExists(appSettingsPath); err != nil {
		return err
	}

	upd.FileUsed = appSettingsPath

	content, err := ioutil.ReadFile(appSettingsPath) // #nosec
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, &upd)
	return err
}

// Config is used to represent the current state of a CLI instance.
type Config struct {
	GitHubAPI       string `yaml:"-"`
	SkipUpdateCheck bool   `yaml:"-"`
}

const (
	NancyConfigFileName = ".nancy-config"
)

// AppSettingsPath returns the path of the CLI settings directory
func AppSettingsPath() string {
	// TODO: Make this configurable
	home, _ := os.UserHomeDir()
	return path.Join(home, ossIndexTypes.OssIndexDirName, NancyConfigFileName)
}

// updateCheckFilename returns the name of the cli update checks file
func updateCheckFilename() string {
	return "update_check.yml"
}

// ensureSettingsFileExists does just that.
func ensureSettingsFileExists(path string) error {
	// TODO - handle invalid YAML config files.

	_, err := os.Stat(path)

	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		// Filesystem error
		return err
	}

	dir := filepath.Dir(path)

	// Create folder
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	_, err = os.Create(path)
	if err != nil {
		return err
	}

	err = os.Chmod(path, 0600)

	return err
}

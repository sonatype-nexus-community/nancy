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

// settingsPath returns the path of the CLI settings directory
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

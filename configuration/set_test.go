package configuration

import (
	"bytes"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestGetConfigFromCommandLineOssIndex(t *testing.T) {
	HomeDir = "/tmp"
	var buffer bytes.Buffer
	buffer.Write([]byte("ossindex\ntestuser\ntoken\n"))

	err := GetConfigFromCommandLine(&buffer)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}

	var ossIndexConfig OSSIndexConfig

	b, err := ioutil.ReadFile(ConfigLocation)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}
	err = yaml.Unmarshal(b, &ossIndexConfig)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}

	if ossIndexConfig.Username != "testuser" && ossIndexConfig.Token != "token" {
		t.Errorf("Config not set properly, expected 'testuser' && 'token' but got %s and %s", ossIndexConfig.Username, ossIndexConfig.Token)
	}
}

func TestGetConfigFromCommandLineIqServer(t *testing.T) {
	HomeDir = "/tmp"
	var buffer bytes.Buffer
	buffer.Write([]byte("iq\nhttp://localhost:8070\nadmin\nadmin123\nn"))

	err := GetConfigFromCommandLine(&buffer)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}

	var iqConfig IQConfig

	b, err := ioutil.ReadFile(ConfigLocation)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}
	err = yaml.Unmarshal(b, &iqConfig)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}

	if iqConfig.Username != "admin" && iqConfig.Token != "admin123" && iqConfig.Server != "http://localhost:8070" {
		t.Errorf("Config not set properly, expected 'admin', 'admin123' and 'http://localhost:8070' but got %s, %s and %s", iqConfig.Username, iqConfig.Token, iqConfig.Server)
	}
}

func TestGetConfigFromCommandLineIqServerWithNoAnswer(t *testing.T) {
	HomeDir = "/tmp"
	var buffer bytes.Buffer
	buffer.Write([]byte("iq\nhttp://localhost:8070\nadmin\nadmin123\ny\nhttp://localhost:8080\nadmin1\nadmin1234\n"))

	err := GetConfigFromCommandLine(&buffer)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}

	var iqConfig IQConfig

	b, err := ioutil.ReadFile(ConfigLocation)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}
	err = yaml.Unmarshal(b, &iqConfig)
	if err != nil {
		t.Errorf("Test failed: %s", err.Error())
	}

	if iqConfig.Username != "admin1" && iqConfig.Token != "admin1234" && iqConfig.Server != "http://localhost:8080" {
		t.Errorf("Config not set properly, expected 'admin1', 'admin1234' and 'http://localhost:8080' but got %s, %s and %s", iqConfig.Username, iqConfig.Token, iqConfig.Server)
	}
}

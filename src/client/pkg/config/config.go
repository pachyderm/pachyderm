package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
)

var configDirPath = filepath.Join(os.Getenv("HOME"), ".pachyderm")
var configPath = filepath.Join(configDirPath, "config.json")

//Read loads pachyderm user config
//If an existing configuration cannot be found, it sets up the defaults
func Read() (*Config, error) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		// File doesn't exist, so create the UID
		return createDefaults()
	}
	var c *Config
	err = json.Unmarshal(raw, &c)
	return c, err
}

func createDefaults() (*Config, error) {
	c := &Config{
		UserID: uuid.NewWithoutDashes(),
	}
	rawConfig, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(configDirPath, 0755)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(configPath, rawConfig, 0644)
	if err != nil {
		return nil, err
	}
	return c, nil
}

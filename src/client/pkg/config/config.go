package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/satori/go.uuid"
)

const configEnvVar = "PACH_CONFIG"

var defaultConfigDir = filepath.Join(os.Getenv("HOME"), ".pachyderm")
var defaultConfigPath = filepath.Join(defaultConfigDir, "config.json")

func configPath() string {
	if env, ok := os.LookupEnv(configEnvVar); ok {
		return env
	}
	return defaultConfigPath
}

// Read loads the Pachyderm config on this machine.
// If an existing configuration cannot be found, it sets up the defaults. Read
// returns a nil Config if and only if it returns a non-nil error.
func Read() (*Config, error) {
	var c *Config

	// Read json file
	p := configPath()
	if raw, err := ioutil.ReadFile(p); err == nil {
		err = json.Unmarshal(raw, &c)
		if err != nil {
			return nil, err
		}
	} else {
		// File doesn't exist, so create a new config
		fmt.Println("No config detected. Generating new config...")
		c = &Config{}
	}
	if c.UserID == "" {
		fmt.Printf("No UserID present in config. Generating new UserID and "+
			"updating config at %s\n", p)
		uuid, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		c.UserID = uuid.String()
		if err := c.Write(); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Write writes the configuration in 'c' to this machine's Pachyderm config
// file.
func (c *Config) Write() error {
	rawConfig, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// If we're not using a custom config path, create the default config path
	p := configPath()
	if _, ok := os.LookupEnv(configEnvVar); ok {
		// using overridden config path -- just make sure the parent dir exists
		d := filepath.Dir(p)
		if _, err := os.Stat(d); err != nil {
			return fmt.Errorf("cannot use config at %s: could not stat parent directory (%v)", p, err)
		}
	} else {
		// using the default config path, create the config directory
		err = os.MkdirAll(defaultConfigDir, 0755)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(p, rawConfig, 0644)
}

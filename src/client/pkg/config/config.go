package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client/pkg/erronce"
	uuid "github.com/satori/go.uuid"
)

const configEnvVar = "PACH_CONFIG"

var (
	defaultPachConfigDir  string = filepath.Join(os.Getenv("HOME"), ".pachyderm")
	defaultPachConfigPath string = filepath.Join(defaultPachConfigDir, "config.json")
	pachConfigOnce        erronce.ErrOnce
	pachConfigValue       *Config
)

func pachConfigPath() string {
	if env, ok := os.LookupEnv(configEnvVar); ok {
		return env
	}
	return defaultPachConfigPath
}

// ReadPachConfig loads the Pachyderm config on this machine.
// If an existing configuration cannot be found, it sets up the defaults. Read
// returns a nil Config if and only if it returns a non-nil error.
func ReadPachConfig() (*Config, error) {
	err := pachConfigOnce.Do(func() error {
		// Read json file
		p := pachConfigPath()
		if raw, err := ioutil.ReadFile(p); err == nil {
			err = json.Unmarshal(raw, &pachConfigValue)
			if err != nil {
				return err
			}
		} else if os.IsNotExist(err) {
			// File doesn't exist, so create a new config
			fmt.Fprintf(os.Stderr, "No config detected at %q. Generating new config...\n", p)
			pachConfigValue = &Config{}
		} else {
			return fmt.Errorf("fatal: could not read config at %q: %v", p, err)
		}

		updated := false

		if pachConfigValue.UserID == "" {
			updated = true
			fmt.Fprintln(os.Stderr, "No UserID present in config - generating new one.")
			uuid := uuid.NewV4()
			pachConfigValue.UserID = uuid.String()
		}

		if pachConfigValue.V3 == nil {
			updated = true
			fmt.Fprintln(os.Stderr, "No config V3 present in config - generating a new one.")

			pachConfigValue.V3 = &ConfigV3{
				Metrics: true,
			}

			if pachConfigValue.V2 != nil {
				pachConfigValue.V3.Metrics = pachConfigValue.V2.Metrics
				pachConfigValue.V2 = nil
			}
		}

		if updated {
			fmt.Fprintf(os.Stderr, "Rewriting config at %q.\n", p)

			if err := pachConfigValue.Write(); err != nil {
				return fmt.Errorf("could not rewrite config at %q: %v", p, err)
			}
		}

		return nil
	})

	return pachConfigValue, err
}

// Write writes the configuration in 'c' to this machine's Pachyderm config
// file.
func (c *Config) Write() error {
	if c.V1 != nil {
		panic("config V1 included (this is a bug)")
	}
	if c.V2 != nil {
		panic("config V2 included (this is a bug)")
	}

	rawConfig, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// If we're not using a custom config path, create the default config path
	p := pachConfigPath()
	if _, ok := os.LookupEnv(configEnvVar); ok {
		// using overridden config path -- just make sure the parent dir exists
		d := filepath.Dir(p)
		if _, err := os.Stat(d); err != nil {
			return fmt.Errorf("cannot use config at %s: could not stat parent directory (%v)", p, err)
		}
	} else {
		// using the default config path, create the config directory
		err = os.MkdirAll(defaultPachConfigDir, 0755)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(p, rawConfig, 0644)
}

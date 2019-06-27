package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/satori/go.uuid"
)

const configEnvVar = "PACH_CONFIG"
const contextEnvVar = "PACH_CONTEXT"

var defaultConfigDir = filepath.Join(os.Getenv("HOME"), ".pachyderm")
var defaultConfigPath = filepath.Join(defaultConfigDir, "config.json")

var readerOnce sync.Once
var value *Config
var readErr error

func configPath() string {
	if env, ok := os.LookupEnv(configEnvVar); ok {
		return env
	}
	return defaultConfigPath
}

// ActiveContext gets the active context in the config
func (c *Config) ActiveContext() (*Context, error) {
	if env, ok := os.LookupEnv(contextEnvVar); ok {
		context := c.V2.Contexts[env]
		if context == nil {
			return nil, fmt.Errorf("`%s` refers to a context that does not exist", contextEnvVar)
		}
		return context, nil
	}
	context := c.V2.Contexts[c.V2.ActiveContext]
	if context == nil {
		return nil, fmt.Errorf("the active context references one that does exist; set the active context first like so: pachctl config set active-context [value]")
	}
	return context, nil
}

// Read loads the Pachyderm config on this machine.
// If an existing configuration cannot be found, it sets up the defaults. Read
// returns a nil Config if and only if it returns a non-nil error.
func Read() (*Config, error) {
	readerOnce.Do(func() {
		// Read json file
		p := configPath()
		if raw, err := ioutil.ReadFile(p); err == nil {
			err = json.Unmarshal(raw, &value)
			if err != nil {
				readErr = err
				return
			}
		} else if os.IsNotExist(err) {
			// File doesn't exist, so create a new config
			fmt.Fprintf(os.Stderr, "No config detected at %q. Generating new config...\n", p)
			value = &Config{}
		} else {
			readErr = fmt.Errorf("fatal: could not read config at %q: %v", p, err)
			return
		}

		updated := false

		if value.UserID == "" {
			updated = true
			fmt.Fprintln(os.Stderr, "No UserID present in config - generating new one.")
			uuid, err := uuid.NewV4()
			if err != nil {
				readErr = fmt.Errorf("could not generate new user ID: %v", err)
				return
			}
			value.UserID = uuid.String()
		}

		if value.V2 == nil {
			updated = true
			fmt.Fprintln(os.Stderr, "No config V2 present in config - generating a new one.")
			if err := value.initV2(); err != nil {
				readErr = err
				return
			}
		}

		if updated {
			fmt.Fprintf(os.Stderr, "Rewriting config at %q.\n", p)

			if err := value.Write(); err != nil {
				readErr = fmt.Errorf("could not rewrite config at %q: %v", p, err)
				return
			}
		}

	})

	return value, readErr
}

func (c *Config) initV2() error {
	c.V2 = &ConfigV2{
		ActiveContext: "default",
		Contexts:      map[string]*Context{},
		Metrics:       true,
	}

	if c.V1 != nil {
		c.V2.Contexts["default"] = &Context{
			Source:            ContextSource_CONFIG_V1,
			PachdAddress:      c.V1.PachdAddress,
			ServerCAs:         c.V1.ServerCAs,
			SessionToken:      c.V1.SessionToken,
			ActiveTransaction: c.V1.ActiveTransaction,
		}

		c.V1 = nil
	} else {
		c.V2.Contexts["default"] = &Context{
			Source: ContextSource_NONE,
		}
	}
	return nil
}

// Write writes the configuration in 'c' to this machine's Pachyderm config
// file.
func (c *Config) Write() error {
	if c.V1 != nil {
		panic("config V1 included (this is a bug)")
	}

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

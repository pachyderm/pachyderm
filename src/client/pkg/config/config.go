package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const configEnvVar = "PACH_CONFIG"
const contextEnvVar = "PACH_CONTEXT"

var defaultConfigDir = filepath.Join(os.Getenv("HOME"), ".pachyderm")
var defaultConfigPath = filepath.Join(defaultConfigDir, "config.json")

var configMu sync.Mutex
var value *Config

func configPath() string {
	if env, ok := os.LookupEnv(configEnvVar); ok {
		return env
	}
	return defaultConfigPath
}

// ActiveContext gets the active context in the config
func (c *Config) ActiveContext() (string, *Context, error) {
	if c.V2 == nil {
		return "", nil, fmt.Errorf("cannot get active context from non-v2 config")
	}
	if envContext, ok := os.LookupEnv(contextEnvVar); ok {
		context := c.V2.Contexts[envContext]
		if context == nil {
			return "", nil, fmt.Errorf("`%s` refers to a context (%q) that does not exist", contextEnvVar, envContext)
		}
		return envContext, context, nil
	}
	context := c.V2.Contexts[c.V2.ActiveContext]
	if context == nil {
		return "", nil, fmt.Errorf("pachctl config error: pachctl's active "+
			"context is %q, but no context named %q has been configured.\n\nYou can fix "+
			"your config by setting the active context like so: pachctl config set "+
			"active-context <context>",
			c.V2.ActiveContext, c.V2.ActiveContext)
	}
	return c.V2.ActiveContext, context, nil
}

// Read loads the Pachyderm config on this machine.
// If an existing configuration cannot be found, it sets up the defaults. Read
// returns a nil Config if and only if it returns a non-nil error.
func Read() (*Config, error) {
	configMu.Lock()
	defer configMu.Unlock()

	if value == nil {
		// Read json file
		p := configPath()
		if raw, err := ioutil.ReadFile(p); err == nil {
			err = json.Unmarshal(raw, &value)
			if err != nil {
				return nil, fmt.Errorf("could not parse config json at %q: %v", p, err)
			}
		} else if os.IsNotExist(err) {
			// File doesn't exist, so create a new config
			log.Debugf("No config detected at %q. Generating new config...", p)
			value = &Config{}
		} else {
			return nil, fmt.Errorf("could not read config at %q: %v", p, err)
		}

		updated := false

		if value.UserID == "" {
			updated = true
			log.Debugln("No UserID present in config - generating new one.")
			uuid := uuid.NewV4()
			value.UserID = uuid.String()
		}

		if value.V2 == nil {
			updated = true
			log.Debugln("No config V2 present in config - generating a new one.")
			if err := value.initV2(); err != nil {
				return nil, err
			}
		}

		for contextName, context := range value.V2.Contexts {
			pachdAddress, err := grpcutil.ParsePachdAddress(context.PachdAddress)
			if err != nil {
				if err != grpcutil.ErrNoPachdAddress {
					return nil, fmt.Errorf("could not parse pachd address for context '%s': %v", contextName, err)
				}
			} else {
				if qualifiedPachdAddress := pachdAddress.Qualified(); qualifiedPachdAddress != context.PachdAddress {
					log.Debugf("Non-qualified pachd address set for context '%s' - fixing", contextName)
					context.PachdAddress = qualifiedPachdAddress
					updated = true
				}
			}
		}

		if updated {
			log.Debugf("Rewriting config at %q.", p)

			if err := value.write(); err != nil {
				return nil, fmt.Errorf("could not rewrite config at %q: %v", p, err)
			}
		}
	}

	return proto.Clone(value).(*Config), nil
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
	configMu.Lock()
	defer configMu.Unlock()
	return c.write()
}

// write() is a helper for Write() that assumes the caller has already locked
// configMu. Because Go mutexes are not reentrant, this is safe to call from
// inside Read()
func (c *Config) write() error {
	if c.V1 != nil {
		panic("config V1 included (this is a bug)")
	}

	rawConfig, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	p := configPath()

	// Because we're writing the config back to disk, we'll also need to make sure
	// that the directory we're writing the config into exists. The approach we
	// use for doing this depends on whether PACH_CONFIG is set.
	if _, ok := os.LookupEnv(configEnvVar); ok {
		// using overridden config path: check that the parent dir exists, but don't
		// create any new directories
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

	// Write to a temporary file first, then rename the temporary file to `p`.
	// This ensures the write is atomic on POSIX.
	tmpfile, err := ioutil.TempFile("", "pachyderm-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if _, err = tmpfile.Write(rawConfig); err != nil {
		return err
	}
	if err = tmpfile.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpfile.Name(), p); err != nil {
		return err
	}

	// essentially short-cuts reading the new config back from disk
	value = proto.Clone(c).(*Config)
	return nil
}

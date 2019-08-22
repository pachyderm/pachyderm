package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client/pkg/erronce"
	uuid "github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	configEnvVar  = "PACH_CONFIG"
	contextEnvVar = "PACH_CONTEXT"
)

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
				if err := pachConfigValue.migrateV3(); err != nil {
					return err
				}
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

func (c *Config) migrateV3() error {
	c.V3.Metrics = c.V2.Metrics

	_, pachActiveContext, err := c.ActiveContext()
	if err != nil {
		return err
	}

	kubeConfig := ReadKubeConfig()
	kubeConfigAccess := kubeConfig.ConfigAccess()
	kubeStartingConfig, err := kubeConfigAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("could not fetch kubernetes' starting config: %v", err)
	}

	if len(kubeStartingConfig.CurrentContext) == 0 {
		return errors.New("kubernetes' current context has not been set")
	}

	kubeContext, ok := kubeStartingConfig.Contexts[kubeStartingConfig.CurrentContext]
	if !ok {
		return errors.New("kubernetes' current config refers to one that does not exist")
	}

	kubeContext.Extensions["pachyderm:v1"] = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"pachd_address":      pachActiveContext.PachdAddress,
			"server_cas":         pachActiveContext.ServerCAs,
			"session_token":      pachActiveContext.SessionToken,
			"active_transaction": pachActiveContext.ActiveTransaction,
		},
	}

	if err := clientcmd.ModifyConfig(kubeConfigAccess, *kubeStartingConfig, true); err != nil {
		return fmt.Errorf("could not modify kubernetes config: %v", err)
	}

	c.V2 = nil
	return nil
}

// ActiveContext gets the active context in the config
func (c *Config) ActiveContext() (string, *Context, error) {
	if env, ok := os.LookupEnv(contextEnvVar); ok {
		context := c.V2.Contexts[env]
		if context == nil {
			return "", nil, fmt.Errorf("`%s` refers to a context that does not exist", contextEnvVar)
		}
		return env, context, nil
	}
	context := c.V2.Contexts[c.V2.ActiveContext]
	if context == nil {
		return "", nil, fmt.Errorf("the active context references one that does exist; set the active context first like so: pachctl config set active-context [value]")
	}
	return c.V2.ActiveContext, context, nil
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

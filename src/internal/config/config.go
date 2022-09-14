package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	configEnvVar            = "PACH_CONFIG"
	contextEnvVar           = "PACH_CONTEXT"
	enterpriseContextEnvVar = "PACH_ENTERPRISE_CONTEXT"
)

var defaultConfigDir = filepath.Join(os.Getenv("HOME"), ".pachyderm")
var defaultConfigPath = filepath.Join(defaultConfigDir, "config.json")
var pachctlConfigPath = filepath.Join("/pachctl", "config.json")

var configMu sync.Mutex
var cachedConfig *Config

func configPath() string {
	if env, ok := os.LookupEnv(configEnvVar); ok {
		return env
	}
	if _, err := os.Stat(pachctlConfigPath); err == nil {
		return pachctlConfigPath
	}
	return defaultConfigPath
}

// ActiveContext gets the active context in the config
func (c *Config) ActiveContext(errorOnNoActive bool) (string, *Context, error) {
	if c.V2 == nil {
		return "", nil, errors.Errorf("cannot get active context from non-v2 config")
	}
	if envContext, ok := os.LookupEnv(contextEnvVar); ok {
		context := c.V2.Contexts[envContext]
		if context == nil {
			return "", nil, errors.Errorf("pachctl config error: `%s` refers to a context (%q) that does not exist", contextEnvVar, envContext)
		}
		return envContext, context, nil
	}
	context := c.V2.Contexts[c.V2.ActiveContext]
	if context == nil {
		if c.V2.ActiveContext == "" {
			if errorOnNoActive {
				return "", nil, errors.Errorf("pachctl config error: no active " +
					"context configured.\n\nYou can fix your config by setting " +
					"the active context like so: pachctl config set " +
					"active-context <context>")
			}
		} else {
			return "", nil, errors.Errorf("pachctl config error: pachctl's active "+
				"context is %q, but no context named %q has been configured.\n\nYou can fix "+
				"your config by setting the active context like so: pachctl config set "+
				"active-context <context>",
				c.V2.ActiveContext, c.V2.ActiveContext)
		}

	}
	return c.V2.ActiveContext, context, nil
}

// ActiveEnterpriseContext gets the active enterprise server context in the config.
// If no enterprise context is configured, this returns the active context.
func (c *Config) ActiveEnterpriseContext(errorOnNoActive bool) (string, *Context, error) {
	if c.V2 == nil {
		return "", nil, errors.Errorf("cannot get active context from non-v2 config")
	}
	if envContext, ok := os.LookupEnv(enterpriseContextEnvVar); ok {
		context := c.V2.Contexts[envContext]
		if context == nil {
			return "", nil, errors.Errorf("pachctl config error: `%s` refers to a context (%q) that does not exist", contextEnvVar, envContext)
		}
		return envContext, context, nil
	}

	if c.V2.ActiveEnterpriseContext == "" {
		return c.ActiveContext(errorOnNoActive)
	}

	context := c.V2.Contexts[c.V2.ActiveEnterpriseContext]
	if context == nil {
		return "", nil, errors.Errorf("pachctl config error: pachctl's active "+
			"enterprise context is %q, but no context named %q has been configured.\n\nYou can fix "+
			"your config by setting the active context like so: pachctl config set "+
			"active-enterprise-context <context>",
			c.V2.ActiveContext, c.V2.ActiveContext)
	}
	return c.V2.ActiveEnterpriseContext, context, nil
}

// fetchCachedConfig is a helper (for Read()) that fetches the pachctl config
// from one of several possible places on disk (see configPath()) and stores it
// in cachedConfig.
func fetchCachedConfig(p string) error {
	cachedConfig = &Config{}
	if raw, err := os.ReadFile(p); err == nil {
		err = serde.Decode(raw, cachedConfig)
		if err != nil {
			return errors.Wrapf(err, "could not parse config json at %q", p)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// File doesn't exist, so create a new config
		log.Debugf("No config detected at %q. Generating new config...", p)
	} else {
		return errors.Wrapf(err, "could not read config at %q", p)
	}
	return nil
}

// validateCachedConfig is a helper (for Read()) that validates 'cachedConfig'
// after parsing. This error should be returned to the caller (typically, of
// pachctl), indicating that their on-disk config is invalid.
func validateCachedConfig() (bool, error) {
	var updated bool
	if cachedConfig.UserID == "" {
		updated = true
		log.Debugln("No UserID present in config - generating new one.")
		cachedConfig.UserID = uuid.NewWithoutDashes()
	}

	if cachedConfig.V2 == nil {
		updated = true
		log.Debugln("No config V2 present in config - generating a new one.")
		if err := cachedConfig.InitV2(); err != nil {
			return updated, err
		}
	}

	for contextName, context := range cachedConfig.V2.Contexts {
		pachdAddress, err := grpcutil.ParsePachdAddress(context.PachdAddress)
		if err != nil {
			if !errors.Is(err, grpcutil.ErrNoPachdAddress) {
				return updated, errors.Wrapf(err, "could not parse pachd address for context '%s'", contextName)
			}
		} else {
			if qualifiedPachdAddress := pachdAddress.Qualified(); qualifiedPachdAddress != context.PachdAddress {
				updated = true
				log.Debugf("Non-qualified pachd address set for context '%s' - fixing", contextName)
				context.PachdAddress = qualifiedPachdAddress
			}
		}
	}
	return updated, nil
}

// Read loads the Pachyderm config on this machine.
// If an existing configuration cannot be found, it sets up the defaults. Read
// returns a nil Config if and only if it returns a non-nil error.
func Read(ignoreCache, readOnly bool) (*Config, error) {
	configMu.Lock()
	defer configMu.Unlock()

	if cachedConfig == nil || ignoreCache {
		// Read json file
		p := configPath()
		if err := fetchCachedConfig(p); err != nil {
			return nil, err
		}
		if updated, err := validateCachedConfig(); err != nil {
			return nil, err
		} else if updated && !readOnly {
			log.Debugf("Rewriting config at %q.", p)

			if err := cachedConfig.write(p); err != nil {
				return nil, errors.Wrapf(err, "could not rewrite config at %q", p)
			}
		}
	}

	cloned := proto.Clone(cachedConfig).(*Config)
	// In the case of an empty map, `proto.Clone` clones `Contexts` as nil. This
	// fixes the issue.
	if cloned.V2.Contexts == nil {
		cloned.V2.Contexts = map[string]*Context{}
	}

	return cloned, nil
}

// InitV2 initializes the V2 object of the config
func (c *Config) InitV2() error {
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

func (c *Config) Write() error {
	configMu.Lock()
	defer configMu.Unlock()
	return c.write(configPath())
}

// Write writes the configuration in 'c' to this machine's Pachyderm config
// file.
// Note: Write() overwrites both the on-disk config and the cachedConfig;
// configMu must be locked by the caller to ensure that Write() calls are
// serialized and that these two representations stay in sync.
func (c *Config) write(path string) error {
	if c.V1 != nil {
		panic("config V1 included (this is a bug)")
	}

	rawConfig, err := serde.EncodeJSON(c, serde.WithIndent(2))
	if err != nil {
		return err
	}

	// Because we're writing the config back to disk, we'll also need to make sure
	// that the directory we're writing the config into exists. The approach we
	// use for doing this depends on whether PACH_CONFIG is being used (if it is,
	// error rather than create new parent dir, in case PACH_CONFIG was
	// simply mistyped).
	if _, ok := os.LookupEnv(configEnvVar); ok {
		// using overridden config path: check that the parent dir exists, but don't
		// create any new directories
		d := filepath.Dir(path)
		if _, err := os.Stat(d); err != nil {
			return errors.Wrapf(err, "cannot use config at %s: could not stat parent directory", path)
		}
	} else {
		// using the default config path, create the config directory
		err = os.MkdirAll(defaultConfigDir, 0755)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}

	// Write to a temporary file first, then rename the temporary file to `p`.
	// This ensures the write is atomic on POSIX.
	tmpfile, err := os.CreateTemp("", "pachyderm-config-*.json")
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err = tmpfile.Write(rawConfig); err != nil {
		return errors.EnsureStack(err)
	}
	if err = tmpfile.Close(); err != nil {
		return errors.EnsureStack(err)
	}
	if err = os.Rename(tmpfile.Name(), path); err != nil {
		// A rename could fail if the temporary directory is mounted on a
		// different device than the config path. If the rename failed, try to
		// just copy the bytes instead. Note that a destructive disk error could
		// leave cachedConfig out of date.
		// TODO(msteffen) attempt to backup the config if it exists & restore on
		// failure.
		if err = os.WriteFile(path, rawConfig, 0644); err != nil {
			return errors.Wrapf(err, "failed to copy updated config file from %s to %s", tmpfile.Name(), path)
		}
	}

	// essentially short-cuts reading the new config back from disk
	// Note: this will only fail in the case that the caller passes
	// in a config path to the API with WritePachTokenToConfigPath()
	if path == configPath() {
		cachedConfig = proto.Clone(c).(*Config)
	}
	return nil
}

func WritePachTokenToConfigPath(token string, path string, enterpriseContext bool) error {
	config := &Config{}
	var raw []byte
	var err error
	if raw, err = os.ReadFile(path); err != nil {
		return errors.Wrapf(err, "could not read config at %q", path)
	}
	if err = serde.Decode(raw, config); err != nil {
		return errors.Wrapf(err, "could not parse config json at %q", path)
	}
	return writePachTokenToConfig(token, config, path, enterpriseContext)
}

// WritePachTokenToConfig sets the auth token for the current pachctl config.
// Used during tests to ensure we don't lose access to a cluster if a test fails.
func WritePachTokenToConfig(token string, enterpriseContext bool) error {
	cfg, err := Read(false, false)
	if err != nil {
		return errors.Wrapf(err, "error reading Pachyderm config (for cluster address)")
	}
	return writePachTokenToConfig(token, cfg, configPath(), enterpriseContext)
}

func writePachTokenToConfig(token string, cfg *Config, path string, enterpriseContext bool) error {
	if enterpriseContext {
		_, context, err := cfg.ActiveEnterpriseContext(true)
		if err != nil {
			return errors.Wrapf(err, "error getting the active enterprise context")
		}
		context.SessionToken = token
	} else {
		_, context, err := cfg.ActiveContext(true)
		if err != nil {
			return errors.Wrapf(err, "error getting the active context")
		}
		context.SessionToken = token
	}
	configMu.Lock()
	defer configMu.Unlock()
	if err := cfg.write(path); err != nil {
		return errors.Wrapf(err, "error writing pachyderm config")
	}
	return nil
}

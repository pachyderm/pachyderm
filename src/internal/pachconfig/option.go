package pachconfig

type ConfigOption = func(*Configuration)

func ApplyOptions(config *Configuration, opts ...ConfigOption) {
	for _, opt := range opts {
		opt(config)
	}
}

// ConfigFromOptions is for use in tests where some tests may want to generate a
// config with certain default values overridden via options.
func ConfigFromOptions(opts ...ConfigOption) *Configuration {
	result := &Configuration{
		GlobalConfiguration: &GlobalConfiguration{},
		PachdSpecificConfiguration: &PachdSpecificConfiguration{
			StorageConfiguration: StorageConfiguration{
				StorageGCPeriod:      0,
				StorageChunkGCPeriod: 0,
			},
		},
		WorkerSpecificConfiguration:     &WorkerSpecificConfiguration{},
		EnterpriseSpecificConfiguration: &EnterpriseSpecificConfiguration{},
	}
	ApplyOptions(result, opts...)
	return result
}

func WithEtcdHostPort(host string, port string) ConfigOption {
	return func(config *Configuration) {
		config.EtcdHost = host
		config.EtcdPort = port
	}
}

func WithPachdPeerPort(port uint16) ConfigOption {
	return func(config *Configuration) {
		config.PeerPort = port
	}
}

func WithOidcPort(port uint16) ConfigOption {
	return func(config *Configuration) {
		config.OidcPort = port
	}
}

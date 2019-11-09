package chunk

import "github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

// StorageOption configures a storage.
type StorageOption func(s *Storage)

// ServiceEnvToOptions converts a service environment configuration (specifically
// the storage configuration) to a set of storage options.
func ServiceEnvToOptions(env *serviceenv.ServiceEnv) []StorageOption {
	return nil
}

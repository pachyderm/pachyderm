package chunk

import "github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

// StorageOption configures a storage.
type StorageOption func(s *Storage)

func ServiceEnvToOptions(env *serviceenv.ServiceEnv) []StorageOption {
	return nil
}

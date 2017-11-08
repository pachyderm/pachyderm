package testutil

import "github.com/pachyderm/pachyderm/src/client/pkg/uuid"

// UniqueString adds a UUID suffix to 'prefix'. This helps avoid name conflicts
// between tests that share the same Pachyderm cluster
func UniqueString(prefix string) string {
	return prefix + uuid.NewWithoutDashes()[0:12]
}

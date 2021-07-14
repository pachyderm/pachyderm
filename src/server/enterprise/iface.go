package enterprise

import (
	enterprise_client "github.com/pachyderm/pachyderm/v2/src/enterprise"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	enterprise_client.APIServer
}

package role

import (
	"github.com/pachyderm/pachyderm/src/pfs/route"
)

type roler struct {
	addresser route.Addresser
	server    Server
}

func newRoler(addresser route.Addresser, server Server) *roler {
	return &roler{addresser, server}
}

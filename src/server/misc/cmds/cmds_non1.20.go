//go:build !go1.20

package cmds

import (
	"net"
)

func setupControl(d *net.Dialer) {}

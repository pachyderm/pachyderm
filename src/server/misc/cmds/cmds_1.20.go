//go:build go1.20

package cmds

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

func setupControl(d *net.Dialer) {
	d.ControlContext = func(ctx context.Context, network, address string, c syscall.RawConn) error {
		var fd uintptr
		c.Control(func(f uintptr) { fd = f }) //nolint:errcheck
		log.Debug(ctx, "created socket for connection", zap.String("network", network), zap.String("address", address), zap.Any("rawConn", c), zap.String("rawConn.type", fmt.Sprintf("%T", c)), zap.Uintptr("fd", fd))
		return nil
	}
}

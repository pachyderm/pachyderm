package version

import (
	"context"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/version"

	"google.golang.org/grpc"
)

// packages is a list of V1 pachyderm APIs - when a user requests these,
// throw a special error prompting them to use a V2 client.
var packages = map[string]struct{}{
	"admin.API":       {},
	"auth.API":        {},
	"debug.API":       {},
	"enterprise.API":  {},
	"identity.API":    {},
	"pfs.API":         {},
	"pps.API":         {},
	"transaction.API": {},
	"versionpb.API":   {},
}

func errOnOldPackage(fullMethod string) error {
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 3 {
		return errors.Errorf("unexpected method name %q", fullMethod)
	}
	if _, ok := packages[parts[1]]; ok {
		return errors.Errorf("this server is running pachyderm %v.x, but your client appears to be an older version - please upgrade", version.MajorVersion)
	}
	return nil
}

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := errOnOldPackage(info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := errOnOldPackage(info.FullMethod); err != nil {
		return err
	}

	return handler(srv, stream)
}

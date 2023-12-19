package cmds

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	ci "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/pachyderm/pachyderm/v2/src/worker"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

type gRPCParams struct {
	Address string
	TLS     bool
	Headers []string
}

func GrpcFromAddress(address string) gRPCParams {
	return gRPCParams{Address: address}
}

func (p gRPCParams) Run(ctx context.Context, pachctlCfg *pachctl.Config, w io.Writer, args []string) error {
	handlers := buildRPCs(
		admin.File_admin_admin_proto,
		auth.File_auth_auth_proto,
		debug.File_debug_debug_proto,
		enterprise.File_enterprise_enterprise_proto,
		identity.File_identity_identity_proto,
		license.File_license_license_proto,
		logs.File_logs_logs_proto,
		pfs.File_pfs_pfs_proto,
		pps.File_pps_pps_proto,
		proxy.File_proxy_proxy_proto,
		transaction.File_transaction_transaction_proto,
		versionpb.File_version_versionpb_version_proto,
		worker.File_worker_worker_proto,
	)

	// If no args, print all available RPCs.  The names can be difficult to
	// guess.
	if len(args) == 0 {
		// Print RPCs.
		methods := maps.Keys(handlers)
		sort.Strings(methods)
		for _, m := range methods {
			fmt.Fprint(w, m)
		}
		return nil
	}

	// If there's an arg, find the method.
	f, ok := handlers[args[0]]
	if !ok {
		return errors.Errorf("no method %v", args[0])
	}

	var msgs []string
	// If there's no more args, read messages from stdin.
	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "Reading messages from stdin; will send after EOF...\n")
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				msgs = append(msgs, s.Text())
			}
		}
		if err := s.Err(); err != nil {
			return errors.Wrap(err, "scan lines")
		}
	} else {
		msgs = args[1:]
	}

	// Get a client connection.
	ctx, c := signal.NotifyContext(ctx, signals.TerminationSignals...)
	defer c()

	authCtx := ctx
	var cc grpc.ClientConnInterface
	if p.Address == "" {
		// If --address isn't set, use the pach context to connect.
		cl, err := pachctlCfg.NewOnUserMachine(ctx, false)
		if err != nil {
			return errors.Wrap(err, "connect")
		}
		authCtx = cl.Ctx()
		cc = cl.ClientConn()
		defer cl.Close()
	} else {
		// If --address is set, just dial the server.
		creds := insecure.NewCredentials()
		if p.TLS {
			creds = credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})
		}
		var err error
		cc, err = grpc.DialContext(ctx, p.Address,
			grpc.WithTransportCredentials(creds),
			grpc.WithUnaryInterceptor(ci.LogUnary),
			grpc.WithStreamInterceptor(ci.LogStream),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                20 * time.Second,
				Timeout:             20 * time.Second,
				PermitWithoutStream: true,
			}),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(grpcutil.MaxMsgSize),
				grpc.MaxCallSendMsgSize(grpcutil.MaxMsgSize),
			),
		)
		if err != nil {
			return errors.Wrap(err, "dial --address")
		}
	}

	// Add headers
	for _, h := range p.Headers {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) != 2 {
			return errors.Errorf("malformed --header %q; use Key=Value", h)
		}
		authCtx = metadata.AppendToOutgoingContext(authCtx, parts[0], parts[1])
	}

	// Add a note during "long" RPCs.
	go func() {
		select {
		case <-time.After(5 * time.Second):
			fmt.Fprintln(os.Stderr, "RPC running; press Control-c to cancel...")
		case <-ctx.Done():
			return
		}
	}()
	if err := f(authCtx, cc, w, msgs...); err != nil {
		return errors.Wrap(err, "do RPC")
	}
	return nil
}

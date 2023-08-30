package cmds

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/archiveserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	ci "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/preflight"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/pachyderm/pachyderm/v2/src/worker"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func Cmds(ctx context.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var d net.Dialer
	setupControl(&d)

	r := new(net.Resolver)
	d.Resolver = r
	r.Dial = func(ctx context.Context, network, address string) (_ net.Conn, retErr error) {
		defer log.Span(ctx, "DialDNS", zap.String("network", network), zap.String("address", address))(log.Errorp(&retErr))
		return d.DialContext(ctx, network, address) //nolint:wrapcheck
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = d.DialContext
	t.TLSClientConfig.VerifyConnection = func(cs tls.ConnectionState) error {
		for _, cert := range cs.PeerCertificates {
			fmt.Printf("tls: cert: %v\n", cert.Subject.String())
		}
		fmt.Printf("tls: server name: %v\n", cs.ServerName)
		fmt.Printf("tls: negotiated protocol: %v\n", cs.NegotiatedProtocol)
		fmt.Println()
		return nil
	}
	t.TLSClientConfig.InsecureSkipVerify = true
	hc := &http.Client{
		Transport: promutil.InstrumentRoundTripper("pachctl", t),
	}

	dnsLookup := &cobra.Command{
		Use:   "{{alias}} <hostname>",
		Short: "Do a DNS lookup on a hostname.",
		Long:  "Do a DNS lookup on a hostname.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			var errs error
			if result, err := r.LookupHost(ctx, args[0]); err != nil {
				errors.JoinInto(&errs, errors.Wrap(err, "lookup host"))
			} else {
				fmt.Printf("A: %v\n", result)
			}
			if result, err := r.LookupCNAME(ctx, args[0]); err != nil {
				errors.JoinInto(&errs, errors.Wrap(err, "lookup cname"))
			} else {
				fmt.Printf("CNAME: %v\n", result)
			}
			if result, err := r.LookupAddr(ctx, args[0]); err != nil {
				errors.JoinInto(&errs, errors.Wrap(err, "lookup reverse address"))
			} else {
				fmt.Printf("IP: %v\n", result)
			}
			if errs != nil {
				fmt.Fprintf(os.Stderr, "some lookups not successful, this is normally fine:\n%v", errs)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dnsLookup, "misc dns-lookup"))

	httpHead := &cobra.Command{
		Use:   "{{alias}} <url>",
		Short: "Make an HTTP HEAD request.",
		Long:  "Make an HTTP HEAD request.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			ctx := pctx.Background("")
			req, err := http.NewRequestWithContext(ctx, "HEAD", args[0], nil)
			if err != nil {
				return errors.Wrap(err, "NewRequest")
			}
			if content, err := httputil.DumpRequestOut(req, false); err == nil {
				fmt.Printf("%s", content)
			}
			res, err := hc.Do(req)
			if res != nil {
				defer res.Body.Close()
				if content, err := httputil.DumpResponse(res, true); err == nil {
					fmt.Printf("%s", content)
				}
			}
			if err != nil {
				return errors.Wrap(err, "Do")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(httpHead, "misc http-head"))

	dial := &cobra.Command{
		Use:   "{{alias}} <network(tcp|udp)> <address>",
		Short: "Dials a network server and then disconnects.",
		Long:  "Dials a network server and then disconnects.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			ctx := pctx.Background("")
			conn, err := d.DialContext(ctx, args[0], args[1])
			if err != nil {
				return errors.Wrap(err, "Dial")
			}
			fmt.Printf("OK: %s -> %s\n", conn.LocalAddr(), conn.RemoteAddr())
			return errors.Wrap(conn.Close(), "Close")
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dial, "misc dial"))

	// NOTE(jonathan): This can move out of misc when we add OAuth support to the download
	// endpoint, and tell pachd what its externally-accessible URL is (so the link works when
	// you click it).
	generateURL := &cobra.Command{
		Use:   "{{alias}} project/repo@branch_or_commit:/file_or_directory ...",
		Short: "Generates the encoded part of an archive download URL.",
		Long:  "Generates the encoded part of an archive download URL.",
		Run: cmdutil.Run(func(args []string) error {
			path, err := archiveserver.EncodeV1(args)
			if err != nil {
				return errors.Wrap(err, "encode")
			}

			u := &url.URL{}
			getPrefix := func() error {
				tctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				c, err := pachctlCfg.NewOnUserMachine(tctx, false)
				if err != nil {
					return err
				}
				defer c.Close()

				info, _ := c.ClusterInfo()
				fmt.Println(info.GetWebResources().GetArchiveDownloadBaseUrl() + path + ".zip")
				return nil
			}
			if err := getPrefix(); err != nil {
				fmt.Println(path)
				return err
			}
			fmt.Println(u.String())
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(generateURL, "misc generate-download-url"))

	decodeURL := &cobra.Command{
		Use:   "{{alias}} <url>",
		Short: "Decodes the encoded part of an archive download URL.",
		Long:  "Decodes the encoded part of an archive download URL.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			u, err := url.Parse(args[0])
			if err != nil {
				return errors.Wrap(err, "url.Parse")
			}
			if !strings.HasPrefix(u.Path, "/archive/") {
				u.Path = "/archive/" + u.Path
			}
			if !strings.HasSuffix(u.Path, ".zip") {
				u.Path = u.Path + ".zip"
			}
			req, err := archiveserver.ArchiveFromURL(u)
			if err != nil {
				return errors.Wrap(err, "ArchiveFromURL")
			}
			if err := req.ForEachPath(func(path string) error {
				fmt.Println(path)
				return nil
			}); err != nil {
				return errors.Wrap(err, "ForEachPath")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(decodeURL, "misc decode-download-url"))

	testMigrations := &cobra.Command{
		Use:   "{{alias}} <postgres dsn>",
		Short: "Runs the database migrations against the supplied database, then rolls them back.",
		Long:  "Runs the database migrations against the supplied database, then rolls them back.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			ctx, c := signal.NotifyContext(ctx, signals.TerminationSignals...)
			defer c()

			dsn := args[0]
			db, err := sqlx.Open("pgx", dsn)
			if err != nil {
				return errors.Wrap(err, "open database")
			}
			if err := dbutil.WaitUntilReady(ctx, db); err != nil {
				return errors.Wrap(err, "wait for database ready")
			}
			if err := preflight.TestMigrations(ctx, db); err != nil {
				return errors.Wrap(err, "apply migrations")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(testMigrations, "misc test-migrations"))

	var grpcAddress string
	var grpcTLS bool
	var grpcHeaders []string
	grpc := &cobra.Command{
		Use:   "{{alias}} service.Method {msg}... ",
		Short: "Call a gRPC method on the server.",
		Long:  "Call a gRPC method on the server.",
		Run: cmdutil.Run(func(args []string) error {
			handlers := buildRPCs(
				admin.File_admin_admin_proto,
				auth.File_auth_auth_proto,
				debug.File_debug_debug_proto,
				enterprise.File_enterprise_enterprise_proto,
				identity.File_identity_identity_proto,
				license.File_license_license_proto,
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
				fmt.Fprintf(os.Stderr, "Specify the RPC you'd like to call:\n%s",
					strings.Join(methods, "\n"))
				return nil
			}

			// If there's an arg, find the method.
			f, ok := handlers[args[0]]
			if !ok {
				return errors.Errorf("no method %v", args[0])
			}

			ctx, c := signal.NotifyContext(ctx, signals.TerminationSignals...)
			defer c()

			// Get a client connection.
			authCtx := ctx
			var cc grpc.ClientConnInterface
			if grpcAddress == "" {
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
				if grpcTLS {
					creds = credentials.NewTLS(&tls.Config{
						InsecureSkipVerify: true,
					})
				}
				var err error
				cc, err = grpc.Dial(grpcAddress, grpc.WithTransportCredentials(creds), grpc.WithUnaryInterceptor(ci.LogUnary), grpc.WithStreamInterceptor(ci.LogStream))
				if err != nil {
					return errors.Wrap(err, "dial --address")
				}
			}

			// Add headers
			md := make(metadata.MD)
			for _, h := range grpcHeaders {
				parts := strings.SplitN(h, "=", 2)
				if len(parts) != 2 {
					return errors.Errorf("malformed --header %q; use Key=Value", h)
				}
				md.Append(parts[0], parts[1])
			}
			authCtx = metadata.NewOutgoingContext(authCtx, md)

			// Add a note during "long" RPCs.
			go func() {
				select {
				case <-time.After(5 * time.Second):
					fmt.Fprintln(os.Stderr, "RPC running; press Control-c to cancel...")
				case <-ctx.Done():
					return
				}
			}()
			if err := f(authCtx, cc, os.Stdout, args[1:]...); err != nil {
				return errors.Wrap(err, "do RPC")
			}
			return nil
		}),
	}
	grpc.PersistentFlags().StringVar(&grpcAddress, "address", "", "If set, don't use the pach client to connect, but manually dial the provided GRPC address instead; url must be in a form like dns:/// or passthrough:///, not http:// or grpc://")
	grpc.PersistentFlags().BoolVar(&grpcTLS, "tls", false, "If set along with --address, use TLS to connect to the server.  The certificate is NOT checked for validity.")
	grpc.PersistentFlags().StringSliceVarP(&grpcHeaders, "header", "H", nil, "foo")
	commands = append(commands, cmdutil.CreateAlias(grpc, "misc grpc"))

	misc := &cobra.Command{
		Short:  "Miscellaneous utilities unrelated to Pachyderm itself.",
		Long:   "Miscellaneous utilities unrelated to Pachyderm itself.  These utilities can be removed or changed in minor releases; do not rely on them.",
		Hidden: true,
	}
	commands = append(commands, cmdutil.CreateAlias(misc, "misc"))
	return commands
}

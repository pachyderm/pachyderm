// Package cmds implements miscellaneous pachctl commands.
package cmds

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/archiveserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/preflight"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var d net.Dialer
	d.ControlContext = func(ctx context.Context, network, address string, c syscall.RawConn) error {
		var fd uintptr
		c.Control(func(f uintptr) { fd = f }) //nolint:errcheck
		log.Debug(ctx, "created socket for connection", zap.String("network", network), zap.String("address", address), zap.Any("rawConn", c), zap.String("rawConn.type", fmt.Sprintf("%T", c)), zap.Uintptr("fd", fd))
		return nil
	}

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
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
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
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			req, err := http.NewRequestWithContext(ctx, "HEAD", args[0], nil)
			if err != nil {
				return errors.Wrap(err, "NewRequest")
			}
			if content, err := httputil.DumpRequestOut(req, false); err == nil {
				fmt.Printf("%s", content)
			}
			res, err := hc.Do(req)
			if res != nil {
				defer errors.Close(&retErr, res.Body, "close body")
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
		Run: cmdutil.RunFixedArgs(2, func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
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
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			path, err := archiveserver.EncodeV1(args)
			if err != nil {
				return errors.Wrap(err, "encode")
			}

			u := &url.URL{}
			getPrefix := func() (retErr error) {
				tctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				c, err := pachctlCfg.NewOnUserMachine(tctx, false)
				if err != nil {
					return err
				}
				defer errors.Close(&retErr, c, "close client")

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
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) error {
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
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
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
		Long:  "Call a gRPC method on the server.  With no args; prints all available methods.  With 1 arg; reads messages to send as JSON lines from stdin.  With >1 arg, sends each JSON-encoded argument as a message.",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return gRPCParams{
				Address: grpcAddress,
				TLS:     grpcTLS,
				Headers: grpcHeaders,
			}.Run(ctx, pachctlCfg, os.Stdout, args)
		}),
	}
	grpc.PersistentFlags().StringVar(&grpcAddress, "address", "", "If set, don't use the pach client to connect, but manually dial the provided GRPC address instead; url must be in a form like dns:/// or passthrough:///, not http:// or grpc://.")
	grpc.PersistentFlags().BoolVar(&grpcTLS, "tls", false, "If set along with --address, use TLS to connect to the server.  The certificate is NOT checked for validity.")
	grpc.PersistentFlags().StringSliceVarP(&grpcHeaders, "header", "H", nil, "Key=Value metadata to add to the request; repeatable.")
	commands = append(commands, cmdutil.CreateAlias(grpc, "misc grpc"))

	var decodeProtoFormat string
	var decodeCompact bool
	decodeProto := &cobra.Command{
		Use:   "{{alias}} <message type> <message bytes>",
		Short: "Decodes a protocol buffer message",
		Long:  "Decodes the provided bytes as the named proto message type and prints the result as JSON.  Without the last arg, reads from stdin.  If the message type is @, then try all message types and print any messages that result in non-empty JSON.",
		Run: cmdutil.RunBoundedArgs(0, 2, func(cmd *cobra.Command, args []string) error {
			all := allProtoMessages()

			// If no args, print all message types.
			if len(args) == 0 {
				keys := maps.Keys(all)
				sort.Strings(keys)
				fmt.Println(strings.Join(keys, "\n"))
				return nil
			}

			// If at least one arg, figure out if we can decode it.
			// If the message type is "@", then we'll try every possible message
			// below.  md will be nil. "@" was chosen for not needing to be
			// escaped from shells.
			var suppressErrors bool
			var mds []protoreflect.MessageDescriptor
			if args[0] != "@" {
				md, ok := all[args[0]]
				if !ok {
					return errors.Errorf("no known message %q", args[0])
				}
				mds = append(mds, md)
			} else {
				suppressErrors = true
				try := maps.Keys(all)
				sort.Strings(try)
				for _, name := range try {
					mds = append(mds, all[name])
				}
			}

			// If a second arg, don't read stdin.
			var input []byte
			if len(args) > 1 {
				input = []byte(args[1])
			} else {
				fmt.Fprintln(os.Stderr, "Reading from stdin...")
				var err error
				input, err = io.ReadAll(os.Stdin)
				if err != nil {
					return errors.Wrap(err, "read stdin")
				}
			}

			// Decode the transport encoding.
			raw, err := decodeBinfmt(input, decodeProtoFormat)
			if err != nil {
				return errors.Wrapf(err, "decode transport format")
			}

			// Decode the actual protobuf data.
			for _, md := range mds {
				m, err := decodeBinaryProto(md, raw)
				if err != nil {
					if suppressErrors {
						continue
					}
					n := min(len(raw), 10)
					dots := ""
					if len(raw) > n {
						dots = "..."
					}
					return errors.Wrapf(err, "unmarshal binary %x%s", raw[:n], dots)
				}
				// Print the message out as JSON.
				mo := protojson.MarshalOptions{
					Indent:    "  ",
					Multiline: true,
				}
				if decodeCompact {
					mo = protojson.MarshalOptions{}
				}
				js, err := mo.Marshal(m)
				if err != nil {
					// If we can't do JSON for some reason, we can at least do
					// something.  And exit non-zero to not mess up scripts.
					fmt.Fprintf(os.Stderr, "%s\n", m)
					if suppressErrors {
						continue
					}
					return errors.Wrap(err, "marshal to json")
				}
				if !suppressErrors || !bytes.Equal(js, []byte("{}")) {
					if suppressErrors {
						fmt.Printf("%s: ", md.FullName())
					}
					fmt.Printf("%s\n", js)
				}
			}
			return nil
		}),
	}
	decodeProto.PersistentFlags().StringVarP(&decodeProtoFormat, "format", "f", "hex", "The format of binary data to read; 'go', 'hex', 'base64', or 'raw'")
	decodeProto.PersistentFlags().BoolVarP(&decodeCompact, "compact", "c", false, "If true, print the output on a single line.")
	commands = append(commands, cmdutil.CreateAlias(decodeProto, "misc decode-proto"))

	misc := &cobra.Command{
		Short:  "Miscellaneous utilities unrelated to Pachyderm itself.",
		Long:   "Miscellaneous utilities unrelated to Pachyderm itself.  These utilities can be removed or changed in minor releases; do not rely on them.",
		Hidden: true,
	}
	commands = append(commands, cmdutil.CreateAlias(misc, "misc"))
	return commands
}

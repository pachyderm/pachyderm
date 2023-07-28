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
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/archiveserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/preflight"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Cmds(ctx context.Context) []*cobra.Command {
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
			u, err := archiveserver.EncodeV1(args)
			if err != nil {
				return errors.Wrap(err, "encode")
			}
			fmt.Println(u)
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
			ctx, c := signal.NotifyContext(pctx.Background(""), signals.TerminationSignals...)
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

	misc := &cobra.Command{
		Short:  "Miscellaneous utilities unrelated to Pachyderm itself.",
		Long:   "Miscellaneous utilities unrelated to Pachyderm itself.  These utilities can be removed or changed in minor releases; do not rely on them.",
		Hidden: true,
	}
	commands = append(commands, cmdutil.CreateAlias(misc, "misc"))
	return commands
}

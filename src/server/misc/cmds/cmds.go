package cmds

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var d net.Dialer
	d.ControlContext = func(ctx context.Context, network, address string, c syscall.RawConn) error {
		log.Debug(ctx, "ControlContext", zap.String("network", network), zap.String("address", address))
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
	hc := &http.Client{
		Transport: promutil.InstrumentRoundTripper("pachctl", t),
	}

	dnsLookup := &cobra.Command{
		Use:   "{{alias}} <hostname>",
		Short: "Do a DNS lookup on a hostname.",
		Long:  "Do a DNS lookup on a hostname.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			ctx := pctx.Background("")
			var errs error
			if result, err := r.LookupHost(ctx, args[0]); err != nil {
				multierr.AppendInto(&errs, errors.Wrap(err, "lookup host"))
			} else {
				fmt.Printf("A: %v\n", result)
			}
			if result, err := r.LookupCNAME(ctx, args[0]); err != nil {
				multierr.AppendInto(&errs, errors.Wrap(err, "lookup cname"))
			} else {
				fmt.Printf("CNAME: %v\n", result)
			}
			if result, err := r.LookupAddr(ctx, args[0]); err != nil {
				multierr.AppendInto(&errs, errors.Wrap(err, "lookup reverse address"))
			} else {
				fmt.Printf("IP: %v\n", result)
			}
			if errs != nil {
				fmt.Printf("\nDo not be alarmed by the below errors; some lookups are expected to fail.\n")
			}
			return errs
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

	misc := &cobra.Command{
		Short:  "Miscellaneous utilities unrelated to Pachyderm itself.",
		Long:   "Miscellaneous utilities unrelated to Pachyderm itself.",
		Hidden: true,
	}
	commands = append(commands, cmdutil.CreateAlias(misc, "misc"))
	return commands
}

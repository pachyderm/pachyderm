package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/server/debug/shell"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds(mainCtx context.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var duration time.Duration
	var pachd bool
	var database bool
	var pipeline string
	var worker string
	var setGRPCLevel bool
	var levelChangeDuration time.Duration
	var recursivelySetLogLevel bool
	profile := &cobra.Command{
		Use:   "{{alias}} <profile> <file>",
		Short: "Collect a set of pprof profiles.",
		Long:  "Collect a set of pprof profiles.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			var d *types.Duration
			if duration != 0 {
				d = types.DurationProto(duration)
			}
			p := &debug.Profile{
				Name:     args[0],
				Duration: d,
			}
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(args[1], func(f *os.File) error {
				return client.Profile(p, filter, f)
			})
		}),
	}
	profile.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration to run a CPU profile for.")
	profile.Flags().BoolVar(&pachd, "pachd", false, "Only collect the profile from pachd.")
	profile.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the profile from the worker pods for the given pipeline.")
	profile.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the profile from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(profile, "debug profile"))

	binary := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a set of binaries.",
		Long:  "Collect a set of binaries.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(args[0], func(f *os.File) error {
				return client.Binary(filter, f)
			})
		}),
	}
	binary.Flags().BoolVar(&pachd, "pachd", false, "Only collect the binary from pachd.")
	binary.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the binary from the worker pods for the given pipeline.")
	binary.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the binary from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(binary, "debug binary"))

	dumpV2Template := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a standard set of debugging information.",
		Long:  "Collect a standard set of debugging information.",
		Run: cmdutil.Run(func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			r, err := client.GetDumpV2Template(client.Ctx(), &debug.GetDumpV2TemplateRequest{})
			if err != nil {
				return err
			}
			bytes, err := serde.EncodeYAML(r.Request)
			if err != nil {
				return err
			}
			fmt.Println(string(bytes))
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dumpV2Template, "debug template"))

	var template string
	dumpV2 := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a standard set of debugging information.",
		Long:  "Collect a standard set of debugging information.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()
			var req *debug.DumpV2Request
			if template == "" {
				r, err := client.DebugClient.GetDumpV2Template(client.Ctx(), &debug.GetDumpV2TemplateRequest{})
				if err != nil {
					return errors.Wrap(err, "get dump template")
				}
				req = r.Request
			} else {
				f, err := os.Open(template)
				if err != nil {
					return errors.Wrap(err, "open template file")
				}
				req = &debug.DumpV2Request{}
				if err := jsonpb.Unmarshal(f, req); err != nil {
					return errors.Wrap(err, "unmarhsal template to DumpV2Request")
				}
			}
			ctx, cf := context.WithCancel(client.Ctx())
			defer cf()
			c, err := client.DebugClient.DumpV2(ctx, req)
			if err != nil {
				return err
			}
			return withFile(args[0], func(f *os.File) error {
				var d *debug.DumpChunk
				for d, err = c.Recv(); !errors.Is(err, io.EOF); d, err = c.Recv() {
					if err != nil {
						return err
					}
					if content := d.GetContent(); content != nil {
						if _, err = f.Write(content.Content); err != nil {
							return errors.Wrap(err, "write dump contents")
						}
					}
					// erroring here should not stop the dump from working
					if prgs := d.GetProgress(); prgs != nil {
						progress.WriteProgress(prgs.Task, prgs.Progress, prgs.Total)
					}
				}
				return nil
			})
		}),
	}
	dumpV2.Flags().StringVarP(&template, "template", "t", "", "A template to customize the output of the debug dump operation.")
	commands = append(commands, cmdutil.CreateAlias(dumpV2, "debug dump"))

	var serverPort int
	analyze := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Start a local pachd server to analyze a debug dump.",
		Long:  "Start a local pachd server to analyze a debug dump.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			dump := shell.NewDumpServer(args[0], uint16(serverPort))
			fmt.Println("listening on", dump.Address())
			select {}
		}),
	}
	analyze.Flags().IntVarP(&serverPort, "port", "p", 0,
		"launch a debug server on the given port. If unset, choose a free port automatically")
	commands = append(commands, cmdutil.CreateAlias(analyze, "debug analyze"))

	log := &cobra.Command{
		Use:   "{{alias}} <level>",
		Short: "Change the log level across Pachyderm.",
		Long:  "Change the log level across Pachyderm.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return err
			}
			defer client.Close()

			want := strings.ToUpper(args[0])
			lvl, ok := debug.SetLogLevelRequest_LogLevel_value[want]
			if !ok {
				return errors.Errorf("no log level %v", want)
			}
			req := &debug.SetLogLevelRequest{
				Duration: types.DurationProto(levelChangeDuration),
				Recurse:  recursivelySetLogLevel,
			}
			if setGRPCLevel {
				req.Level = &debug.SetLogLevelRequest_Grpc{
					Grpc: debug.SetLogLevelRequest_LogLevel(lvl),
				}
			} else {
				req.Level = &debug.SetLogLevelRequest_Pachyderm{
					Pachyderm: debug.SetLogLevelRequest_LogLevel(lvl),
				}
			}
			res, err := client.DebugClient.SetLogLevel(client.Ctx(), req)
			if pods := res.GetAffectedPods(); len(pods) > 0 {
				fmt.Printf("adjusted log level on pods: %v\n", pods)
			}
			if pods := res.GetErroredPods(); len(pods) > 0 {
				fmt.Printf("failed to adjust log level on some pods: %v\n", pods)
			}
			if err != nil {
				return errors.Wrap(err, "SetLogLevel")
			}
			return nil
		}),
	}
	log.Flags().DurationVarP(&levelChangeDuration, "duration", "d", 5*time.Minute, "how long to log at the non-default level")
	log.Flags().BoolVarP(&setGRPCLevel, "grpc", "g", false, "adjust the grpc log level instead of the pachyderm log level")
	log.Flags().BoolVarP(&recursivelySetLogLevel, "recursive", "r", true, "set the log level on all pachyderm pods; if false, only the pachd that handles this RPC")
	commands = append(commands, cmdutil.CreateAlias(log, "debug log-level"))

	var tapConfigID string
	var envoyAddresses []string
	tap := &cobra.Command{
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			out := args[0]
			var nCap int

			// Open the output.
			w, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return errors.Wrap(err, "open pcap destination")
			}
			defer errors.Close(&retErr, w, "close output file")

			// Start the tap.
			client, err := pachctlCfg.NewOnUserMachine(mainCtx, false)
			if err != nil {
				return errors.Wrap(err, "new pachyderm client")
			}
			ts, err := client.DebugClient.Tap(client.Ctx())
			if err != nil {
				return errors.Wrap(err, "start tap stream")
			}
			if err := ts.Send(&debug.TapRequest{
				Request: &debug.TapRequest_Start_{
					Start: &debug.TapRequest_Start{
						ConfigId:       tapConfigID,
						EnvoyAddresses: envoyAddresses,
					},
				},
			}); err != nil {
				return errors.Wrap(err, "send tap request")
			}

			// The tap has started; collect data.
			defer func() {
				fmt.Fprintf(os.Stderr, "captured %d bytes\n", nCap)
			}()

			msgCh := make(chan *debug.TapResponse)
			errCh := make(chan error)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)
			defer signal.Stop(sigCh)
			defer close(sigCh)

			fmt.Fprintf(os.Stderr, "Writing captured packets to %v; ^C to end tap\n", out)
			go func() {
				// Direct messages from the server to the appropriate channel.
				for {
					msg, err := ts.Recv()
					if err != nil {
						if errors.Is(err, io.EOF) {
							// Everything finished OK.
							close(errCh)
							close(msgCh)
							return
						}
						// The upstream RPC returned an error.
						errCh <- errors.Wrap(err, "recv")
						close(errCh)
						close(msgCh)
						return
					}
					msgCh <- msg
				}
			}()
		main:
			for {
				select {
				case err, ok := <-errCh:
					if ok && err != nil {
						return errors.Wrap(err, "tap")
					}
				case <-sigCh:
					signal.Stop(sigCh) // Now ^C kills the program.
					fmt.Fprintf(os.Stderr, "\nEnding tap and writing final packets... ^C to kill\n")
					if err := ts.Send(&debug.TapRequest{
						Request: &debug.TapRequest_End_{
							End: &debug.TapRequest_End{},
						},
					}); err != nil {
						return errors.Wrap(err, "send end stream request")
					}
					if err := ts.CloseSend(); err != nil {
						return errors.Wrap(err, "close send")
					}
				case msg, ok := <-msgCh:
					if !ok { // chan closed
						fmt.Fprintf(os.Stderr, "Tap stream complete\n")
						break main
					}
					switch x := msg.Response.(type) {
					case *debug.TapResponse_Capture:
						n, err := w.Write(x.Capture.GetValue())
						if err != nil {
							return errors.Wrap(err, "write captured packets")
						}
						nCap += n
					case *debug.TapResponse_Log_:
						fmt.Fprintf(os.Stderr, "server log: %s\n", x.Log.Msg)
					}

				}
			}
			return nil
		}),
	}
	// Generate knownTaps with:
	// cat etc/helm/pachyderm/envoy.json  | grep "config_id" | sed 's/^/{/g; s/$/}/g;'
	// | jq .config_id -r | sort | sed 's/^/"/g;s/$/",/g'
	knownTaps := []string{
		"tap-cluster-console",
		"tap-cluster-pachd-archive",
		"tap-cluster-pachd-grpc",
		"tap-cluster-pachd-identity",
		"tap-cluster-pachd-metrics",
		"tap-cluster-pachd-oidc",
		"tap-cluster-pachd-s3",
		"tap-http-direct-console",
		"tap-http-direct-pachd-archive",
		"tap-http-direct-pachd-grpc",
		"tap-http-direct-pachd-identity",
		"tap-http-direct-pachd-metrics",
		"tap-http-direct-pachd-oidc",
		"tap-http-direct-pachd-s3",
		"tap-http-https-warning",
		"tap-http-proxy-http",
		"tap-listener-direct-console",
		"tap-listener-direct-pachd-archive",
		"tap-listener-direct-pachd-grpc",
		"tap-listener-direct-pachd-identity",
		"tap-listener-direct-pachd-metrics",
		"tap-listener-direct-pachd-oidc",
		"tap-listener-direct-pachd-s3",
		"tap-listener-https-warning",
		"tap-listener-proxy-http",
	}
	tap.Flags().StringVarP(&tapConfigID, "tap", "t", "tap-listener-proxy-http", "The tap config to tap; "+strings.Join(knownTaps, ", "))
	tap.Flags().StringArrayVarP(&envoyAddresses, "proxy", "p", nil, "The addresses of proxies to connect to; if not specified, all available.  Each address should include a port, typically 9901.")
	commands = append(commands, cmdutil.CreateAliases(tap, "debug tap"))

	debug := &cobra.Command{
		Short: "Debug commands for analyzing a running cluster.",
		Long:  "Debug commands for analyzing a running cluster.",
	}
	commands = append(commands, cmdutil.CreateAlias(debug, "debug"))
	return commands
}

// FIXME(CORE-1078): handle projects
func createFilter(pachd, database bool, pipeline, worker string) (*debug.Filter, error) {
	var f *debug.Filter
	if pachd {
		f = &debug.Filter{Filter: &debug.Filter_Pachd{Pachd: true}}
	}
	if database {
		f = &debug.Filter{Filter: &debug.Filter_Database{Database: true}}
	}
	if pipeline != "" {
		if f != nil {
			return nil, errors.Errorf("only one debug filter allowed")
		}
		f = &debug.Filter{Filter: &debug.Filter_Pipeline{Pipeline: &pps.Pipeline{Name: pipeline}}}
	}
	if worker != "" {
		if f != nil {
			return nil, errors.Errorf("only one debug filter allowed")
		}
		f = &debug.Filter{Filter: &debug.Filter_Worker{Worker: &debug.Worker{Pod: worker}}}
	}
	return f, nil
}

func withFile(file string, cb func(*os.File) error) (retErr error) {
	f, err := os.Create(file)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(f)
}

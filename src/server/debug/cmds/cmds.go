package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/server/debug/server/debugstar"
	"github.com/pachyderm/pachyderm/v2/src/server/debug/shell"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
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
		Long:  "This command collects a set of pprof profiles. Options include heap (memory), CPU, block, mutex, and goroutine profiles.",
		Example: "\t- {{alias}} cpu cpu.tgz \n" +
			"\t- {{alias}} heap heap.tgz \n" +
			"\t- {{alias}} goroutine goroutine.tgz \n" +
			"\t- {{alias}} goroutine --pachd goroutine.tgz \n" +
			"\t- {{alias}} cpu --pachd -d 30s cpu.tgz \n" +
			"\t- {{alias}} cpu --pipeline foo -d 30s foo-pipeline.tgz \n" +
			"\t- {{alias}} cpu --worker foo-v1-r6pdq -d 30s worker.tgz \n",
		Run: cmdutil.RunFixedArgs(2, func(cmd *cobra.Command, args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer client.Close()
			var d *durationpb.Duration
			if duration != 0 {
				d = durationpb.New(duration)
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
				return client.Profile(client.Ctx(), p, filter, f)
			})
		}),
	}
	profile.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Specify a duration for compiling a CPU profile.")
	profile.Flags().BoolVar(&pachd, "pachd", false, "Collect only pachd's profile.")
	profile.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Collect only a specific pipeline's profile from the worker pods.")
	profile.Flags().StringVarP(&worker, "worker", "w", "", "Collect only the profile of a given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(profile, "debug profile"))

	binary := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a set of binaries.",
		Long:  "This command collects a set of binaries.",
		Example: "\t- {{alias}} binaries.tgz \n" +
			"\t- {{alias}} --pachd pachd-binary.tgz \n" +
			"\t- {{alias}} --worker foo-v1-r6pdq foo-pod-binary.tgz \n" +
			"\t- {{alias}} --pipeline foo foo-binary.tgz \n",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
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
	binary.Flags().BoolVar(&pachd, "pachd", false, "Collect only pachd's binary.")
	binary.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Collect only the binary from a given pipeline.")
	binary.Flags().StringVarP(&worker, "worker", "w", "", "Collect only the binary from a given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(binary, "debug binary"))

	dumpV2Template := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Print a yaml debugging template.",
		Long: "This command outputs a customizable yaml template useful for debugging. This is often used by Customer Engineering to support your troubleshooting needs. \n" +
			"Use the modified template with the `debug dump` command (e.g., `pachctl debug dump --template debug-template.yaml out.tgz`) \n",
		Example: "\t- {{alias}} \n" +
			"\t- {{alias}} > debug-template.yaml\n",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer client.Close()
			r, err := client.GetDumpV2Template(client.Ctx(), &debug.GetDumpV2TemplateRequest{})
			if err != nil {
				return err
			}
			e := serde.NewYAMLEncoder(os.Stdout)
			if err := e.EncodeProto(r.Request); err != nil {
				return err
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dumpV2Template, "debug template"))

	var template string
	dumpV2 := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a standard set of debugging information.",
		Long: "This command collects a standard set of debugging information related to the version, database, source repos, helm, profiles, binaries, loki-logs, pipelines, describes, and logs. \n \n" +
			"You can customize this output by passing in a customized template (made from `pachctl debug template` via the `--template` flag.",
		Example: "\t- {{alias}} dump.tgz \n" +
			"\t- {{alias}} -t template.yaml out.tgz\n",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
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
				bytes, err := os.ReadFile(template)
				if err != nil {
					return errors.Wrap(err, "open template file")
				}
				req = &debug.DumpV2Request{}
				if err := serde.Decode(bytes, req); err != nil {
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
				bytesWritten := int64(0)
				for d, err = c.Recv(); !errors.Is(err, io.EOF); d, err = c.Recv() {
					if err != nil {
						return err
					}
					if content := d.GetContent(); content != nil {
						if _, err = f.Write(content.Content); err != nil {
							return errors.Wrap(err, "write dump contents")
						}
						bytesWritten += int64(len(content.Content))
						progress.WriteProgressCountBytes("Downloaded", bytesWritten, false, 100)
					}
					// erroring here should not stop the dump from working
					if prgs := d.GetProgress(); prgs != nil {
						progress.WriteProgress(prgs.Task, prgs.Progress, prgs.Total)
					}
				}
				progress.WriteProgressCountBytes("Downloaded", bytesWritten, true, 100)
				return nil
			})
		}),
	}
	dumpV2.Flags().StringVarP(&template, "template", "t", "", "Download a template to customize the output of the debug dump operation.")
	commands = append(commands, cmdutil.CreateAlias(dumpV2, "debug dump"))

	localDump := &cobra.Command{
		Use:     "{{alias}} <template>",
		Short:   "Collect debugging information without connecting to a Pachyderm server.",
		Long:    "This command collects debugging information based on a special template provided by Pachyderm Support.\n\nIt works in cases where Pachyderm isn't already installed.",
		Example: "\t- {{alias}} template.yaml \n",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cmdutil.FileMustExist(0)),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			bytes, err := os.ReadFile(args[0])
			if err != nil {
				return errors.Wrap(err, "read template file")
			}
			template := new(debug.DumpV2Request)
			if err := serde.Decode(bytes, template); err != nil {
				return errors.Wrap(err, "unmarshal template")
			}
			fs := new(debugstar.LocalDumpFS)
			env := debugstar.Env{FS: fs}
			defer errors.Close(&retErr, fs, "finalize archive")
			var errs error
			for _, s := range template.GetStarlarkScripts() {
				var name, program string
				switch s.GetSource().(type) {
				case *debug.Starlark_Builtin:
					name = s.GetBuiltin()
					var ok bool
					program, ok = debugstar.BuiltinScripts[name]
					if !ok {
						errors.JoinInto(&errs, errors.Errorf("built-in script %q not available", name))
						continue
					}
				case *debug.Starlark_Literal:
					name = s.GetLiteral().GetName()
					program = s.GetLiteral().GetProgramText()
				}
				if err := env.RunStarlark(cmd.Context(), name, program); err != nil {
					errors.JoinInto(&errs, errors.Wrapf(err, "run script %q", name))
				}
			}
			if err := errs; err != nil {
				if err := fs.Write("local-errors.txt", func(w io.Writer) error {
					fmt.Fprintf(w, "%+v\n", err)
					return nil
				}); err != nil {
					fmt.Fprintf(os.Stderr, "Some errors occurred and could not be included in the dump. The dump may still be usable.\n%v\n%+v\n", err, errs)
				}
			}
			fmt.Fprintf(os.Stderr, "Debug file available at:\n")
			fmt.Fprintf(os.Stdout, "%v\n", fs.Name())
			return nil
		},
	}
	commands = append(commands, cmdutil.CreateAlias(localDump, "debug local"))

	var serverPort int
	analyze := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Start a local pachd server to analyze a debug dump.",
		Long:  "This command starts a local pachd server to analyze a debug dump.",
		Example: "\t- {{alias}} dump.tgz \n" +
			"\t- {{alias}} dump.tgz --port 1650 \n",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) error {
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
		Long:  "This command changes the log level across Pachyderm.",
		Example: "\t- {{alias}} debug \n" +
			"\t- {{alias}} info --duration 5m \n" +
			"\t- {{alias}} info --grpc --duration 5m \n" +
			"\t- {{alias}} info --recursive false --duration 5m \n",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) error {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
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
				Duration: durationpb.New(levelChangeDuration),
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
	log.Flags().DurationVarP(&levelChangeDuration, "duration", "d", 5*time.Minute, "Specify a duration for how long to log at the non-default level.")
	log.Flags().BoolVarP(&setGRPCLevel, "grpc", "g", false, "Set the grpc log level instead of the Pachyderm log level.")
	log.Flags().BoolVarP(&recursivelySetLogLevel, "recursive", "r", true, "Set the log level on all Pachyderm pods; if false, only the pachd that handles this RPC")
	commands = append(commands, cmdutil.CreateAlias(log, "debug log-level"))

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

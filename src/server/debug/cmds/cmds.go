package cmds

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pachyderm/pachyderm/v2/src/server/debug/shell"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var duration time.Duration
	var pachd bool
	var database bool
	var pipeline string
	var worker string
	profile := &cobra.Command{
		Use:   "{{alias}} <profile> <file>",
		Short: "Collect a set of pprof profiles.",
		Long:  "Collect a set of pprof profiles.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-profile")
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
			return withFile(os.Stderr, args[1], func(w io.Writer) error {
				return client.Profile(p, filter, w)
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
			client, err := client.NewOnUserMachine("debug-binary")
			if err != nil {
				return err
			}
			defer client.Close()
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(os.Stderr, args[0], func(w io.Writer) error {
				return client.Binary(filter, w)
			})
		}),
	}
	binary.Flags().BoolVar(&pachd, "pachd", false, "Only collect the binary from pachd.")
	binary.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the binary from the worker pods for the given pipeline.")
	binary.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the binary from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(binary, "debug binary"))

	var limit int64
	dump := &cobra.Command{
		Use:   "{{alias}} <file>",
		Short: "Collect a standard set of debugging information.",
		Long:  "Collect a standard set of debugging information.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			filter, err := createFilter(pachd, database, pipeline, worker)
			if err != nil {
				return err
			}
			return withFile(os.Stderr, args[0], func(w io.Writer) error {
				return client.Dump(filter, limit, w)
			})
		}),
	}
	dump.Flags().BoolVar(&pachd, "pachd", false, "Only collect the dump from pachd.")
	dump.Flags().BoolVar(&database, "database", false, "Only collect the dump from pachd's database.")
	dump.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the dump from the worker pods for the given pipeline.")
	dump.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the dump from the given worker pod.")
	dump.Flags().Int64VarP(&limit, "limit", "l", 0, "Limit sets the limit for the number of commits / jobs that are returned for each repo / pipeline in the dump.")
	commands = append(commands, cmdutil.CreateAlias(dump, "debug dump"))

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

	debug := &cobra.Command{
		Short: "Debug commands for analyzing a running cluster.",
		Long:  "Debug commands for analyzing a running cluster.",
	}
	commands = append(commands, cmdutil.CreateAlias(debug, "debug"))

	return commands
}

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

// progressWriter is an io.Writer that logs a progress update after each write.  The output is in
// the form "1.2KiB 42%" with a spinner that changes with each write even if the number doesn't.
// The updates occur in-place and the progress update is replaced with "done." when done.
type progressWriter struct {
	log                          io.Writer
	wrote                        string
	spinnerState, current, total int
}

// Write implements io.Writer.
func (w *progressWriter) Write(buf []byte) (n int, err error) {
	w.spinnerState++
	w.current += len(buf)

	backspaces, msg, clear := new(strings.Builder), new(strings.Builder), new(strings.Builder)
	for i := 0; i < len(w.wrote); i++ {
		// Backspace over the last write.
		backspaces.WriteByte('\b')
	}

	// Print current number of bytes downloaded.
	msg.WriteString(humanize.IBytes(uint64(w.current)))

	if w.current == w.total {
		// Print "done."
		msg.WriteString(" done.")
	} else {
		// Print the % indicator.
		fmt.Fprintf(msg, "%3.f%% ", 100*float64(w.current)/float64(w.total))

		// Print the spinner.
		switch w.spinnerState % 4 {
		case 0:
			msg.WriteByte('|')
		case 1:
			msg.WriteByte('/')
		case 2:
			msg.WriteByte('-')
		case 3:
			msg.WriteByte('\\')
		}
	}

	// Clear any bytes remaining at the end of the line.
	for _, b := range []byte{' ', '\b'} {
		for i := msg.Len(); i < len(w.wrote); i++ {
			clear.WriteByte(b)
		}
	}

	fmt.Fprintf(w.log, "%s%s%s", backspaces.String(), msg.String(), clear.String())
	w.wrote = msg.String()

	return len(buf), nil
}

func withFile(log io.Writer, file string, cb func(io.Writer) error) (retErr error) {
	// Create the named file on disk.
	f, err := os.Create(file)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()

	// This syncronization allows tests to ensure we do not keep the reader running forever,
	// which would result in lost data in real life.
	doneCh := make(chan struct{})
	defer func() { <-doneCh }()

	// Create a pipe which will let the progress producer see the bytes written to the
	// callback's writer (a MultiWriter between f and w).
	r, w := io.Pipe()
	defer w.Close()

	// In the background, print progress information as tar.gz bytes arrive.
	go func() {
		defer close(doneCh)
		// Consume all bytes even if we error out.
		defer io.Copy(io.Discard, r) //nolint:errcheck

		gr, err := gzip.NewReader(r)
		if err != nil {
			fmt.Fprintf(log, "debug: problem creating gzip reader: %v\n", err)
			return
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Fprintf(log, "debug: problem reading tar header: %v\n", err)
				return
			}

			fmt.Fprintf(log, "debug: receiving sub-file %s (%s): ", hdr.Name, humanize.IBytes(uint64(hdr.Size)))
			pw := &progressWriter{
				total: int(hdr.Size),
				log:   log,
			}
			if _, err := io.Copy(pw, tr); err != nil {
				fmt.Fprintf(log, "\ndebug: problem reading sub-file %s: %v", hdr.Name, err)
			}
			fmt.Fprintln(log, "")
		}
	}()

	return cb(io.MultiWriter(f, w))
}

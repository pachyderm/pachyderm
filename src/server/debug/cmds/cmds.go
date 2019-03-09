package cmds

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"

	units "github.com/docker/go-units"
	"github.com/mholt/archiver"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds(noMetrics *bool, noPortForwarding *bool) []*cobra.Command {
	debug := &cobra.Command{
		Use:   "debug-dump",
		Short: "Return a dump of running goroutines.",
		Long:  "Return a dump of running goroutines.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Dump(os.Stdout)
		}),
	}

	var duration time.Duration
	profile := &cobra.Command{
		Use:   "debug-profile profile",
		Short: "Return a profile from the server.",
		Long:  "Return a profile from the server.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Profile(args[0], duration, os.Stdout)
		}),
	}
	profile.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration of the CPU profile.")

	binary := &cobra.Command{
		Use:   "debug-binary",
		Short: "Return the binary the server is running.",
		Long:  "Return the binary the server is running.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Binary(os.Stdout)
		}),
	}

	var profileFile string
	var binaryFile string
	pprof := &cobra.Command{
		Use:   "debug-pprof profile",
		Short: "Analyze a profile of pachd in pprof.",
		Long:  "Analyze a profile of pachd in pprof.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			var eg errgroup.Group
			// Download the profile
			eg.Go(func() (retErr error) {
				if args[0] == "cpu" {
					fmt.Printf("Downloading cpu profile, this will take %s...", units.HumanDuration(duration))
				}
				f, err := os.Create(profileFile)
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				return client.Profile(args[0], duration, f)
			})
			// Download the binary
			eg.Go(func() (retErr error) {
				f, err := os.Create(binaryFile)
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				return client.Binary(f)
			})
			if err := eg.Wait(); err != nil {
				return err
			}
			cmd := exec.Command("go", "tool", "pprof", binaryFile, profileFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}),
	}
	pprof.Flags().StringVar(&profileFile, "profile-file", "profile", "File to write the profile to.")
	pprof.Flags().StringVar(&binaryFile, "binary-file", "binary", "File to write the binary to.")
	pprof.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration of the CPU profile.")

	var outputFileName string
	sos := &cobra.Command{
		Use:   "debug-sos",
		Short: "Collect the full set of debug information, and write it to a file.",
		Long:  "Collect the full set of debug information, and write it to a file.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(!*noMetrics, !*noPortForwarding, "debug-sos")
			if err != nil {
				return err
			}
			defer client.Close()
			// Setup debug files to collect.
			debugFiles := []struct {
				name    string
				collect func(io.Writer) error
			}{
				{
					name: "goroutine-dump",
					collect: func(w io.Writer) error {
						return client.Dump(w)
					},
				},
				{
					name: "profile-cpu",
					collect: func(w io.Writer) error {
						return client.Profile("cpu", duration, w)
					},
				},
				{
					name: "binary",
					collect: func(w io.Writer) error {
						return client.Binary(w)
					},
				},
			}
			// Collect debug files.
			debugDir := filepath.Join(os.TempDir(), "pachyderm-debug")
			if err := os.MkdirAll(debugDir, 0700); err != nil {
				return err
			}
			defer os.RemoveAll(debugDir)
			var filesCollected []string
			for _, debugFile := range debugFiles {
				file, err := os.Create(filepath.Join(debugDir, debugFile.name))
				if err != nil {
					return err
				}
				fmt.Println("Collecting", debugFile.name)
				if err := debugFile.collect(file); err != nil {
					fmt.Printf("Failed to collect %v: %v\n", debugFile.name, err)
					continue
				}
				filesCollected = append(filesCollected, file.Name())
			}
			// Tar and gzip debug files.
			outputFileName = outputFileName + ".tar.gz"
			os.Remove(outputFileName)
			return archiver.NewTarGz().Archive(filesCollected, outputFileName)
		}),
	}
	sos.Flags().StringVar(&outputFileName, "output-file-name", "pachyderm-debug", "Name of the output file.")
	sos.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration of the CPU profile.")

	return []*cobra.Command{
		debug,
		profile,
		binary,
		pprof,
		sos,
	}
}

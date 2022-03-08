package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	pachshell "github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/shell"
	pfscmds "github.com/pachyderm/pachyderm/v2/src/server/pfs/cmds"
	ppscmds "github.com/pachyderm/pachyderm/v2/src/server/pps/cmds"

	"github.com/c-bata/go-prompt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{}
	var subcommands []*cobra.Command

	subcommands = append(subcommands, pfscmds.DebugCmds()...)
	subcommands = append(subcommands, ppscmds.DebugCmds()...)

	profile := &cobra.Command{
		Use:   "{{alias}} <profile-type>",
		Short: "Open the pprof CLI to investigate a pachd profile.",
		Long:  "Open pprof to investigate a pachd profile. Golang must be installed and on the path.",
		Example: `
# view pachd heap profile
$ {{alias}} heap

# view pachd goroutine profile
$ {{alias}} goro`,
		Run: func(_ *cobra.Command, _ []string) {}, // just to show up in suggestions
	}
	pachshell.RegisterCompletionFunc(profile, func(_ *client.APIClient, flag, text string, _ int64) ([]prompt.Suggest, pachshell.CacheFunc) {
		return []prompt.Suggest{{
			Text: "heap",
		}, {
			Text: "goroutine",
		}}, nil
	})
	subcommands = append(subcommands, cmdutil.CreateAlias(profile, "profile"))

	cmdutil.MergeCommands(rootCmd, subcommands)
	return rootCmd
}

func (d *debugDump) execute(in string) {
	if in == "" {
		return
	}
	if in == "exit" {
		os.Exit(0)
	}

	if strings.HasPrefix(in, "profile") {
		d.handleProfile(in)
		return
	}

	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("PACH_DEBUG_DIRECT_ADDRESS=%s pachctl "+in, d.mock.Addr.String()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func Run(maxCompletions int, d *debugDump) {
	c, err := client.NewFromURI(d.mock.Addr.String())
	if err != nil {
		log.Fatalf("Couldn't create a client for analysis: %v", err)
	}
	defer c.Close()
	pachshell.RunCustom(rootCmd(), "Debug Shell", c, int64(maxCompletions), d.execute)
}

func (d *debugDump) handleProfile(in string) {
	parts := strings.Fields(in)
	if len(parts) < 2 {
		fmt.Println("Must choose either heap or goroutine (goro) profile.")
		return
	}

	var chosenProfile string
	switch parts[1] {
	case "heap":
		chosenProfile = "heap"
	case "goro", "goroutine":
		chosenProfile = "goroutine"
	default:
		fmt.Printf("Unrecognzied profile type %s, choose either heap or goroutine\n", parts[1])
		return
	}

	tmpFile, err := os.CreateTemp("", "pach_debug_shell_profile_"+chosenProfile)
	if err != nil {
		fmt.Printf("Couldn't create a temp file for the profile: %v", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	// TODO: select pachd? run multiple?
	glob := fmt.Sprintf("pachd/*/pachd/%s", chosenProfile)
	err = d.globTar(glob, func(_ string, r io.Reader) error {
		_, err := io.Copy(tmpFile, r)
		if err != nil {
			return err
		}
		return errutil.ErrBreak // only want one
	})
	if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Printf("Failed to copy the profile data: %v", err)
	}

	profilePath := tmpFile.Name()

	fmt.Println("Running pprof\n\n")
	cmd := exec.Command("go", "tool", "pprof", filepath.Base(profilePath))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(profilePath)
	cmd.Run()
	fmt.Println()
}

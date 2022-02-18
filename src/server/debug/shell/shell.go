package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	completionAnnotation  string = "completion"
	ldThreshold           int    = 2
	defaultMaxCompletions int64  = 64
)

// CacheFunc is a function which returns whether or not cached results from a
// previous call to a CompletionFunc can be reused.
type CacheFunc func(flag, text string) bool

type shell struct {
	serverAddr string
}

func newShell(addr string) *shell {
	return &shell{
		serverAddr: addr,
	}
}

func (s *shell) executor(in string) {
	if in == "exit" {
		os.Exit(0)
	}

	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("PACH_DIRECT_PACHD_ADDRESS=%s pachctl "+in, s.serverAddr))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (s *shell) suggestor(in prompt.Document) []prompt.Suggest {
	return nil
}

func (s *shell) run() {
	fmt.Printf("Type 'exit' or press Ctrl-D to exit.\n")
	color.NoColor = true // color doesn't work in terminal
	prompt.New(
		s.executor,
		s.suggestor,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("Debug Shell"),
	).Run()
}

// Run runs a prompt, it does not return.
func Run(rootCmd *cobra.Command, maxCompletions int64, addr string) {
	newShell(addr).run()
}

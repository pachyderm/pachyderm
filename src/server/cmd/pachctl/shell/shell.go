package shell

import (
	"log"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type shell struct {
	rootCmd *cobra.Command
}

func newShell(rootCmd *cobra.Command) *shell {
	return &shell{rootCmd: rootCmd}
}

func (s *shell) executor(in string) {
	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader("pachctl " + in)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return
}

func (s *shell) suggestor(in prompt.Document) []prompt.Suggest {
	args := strings.Fields(in.Text)
	if len(strings.TrimSuffix(in.Text, " ")) < len(in.Text) {
		args = append(args, "")
	}
	cmd := s.rootCmd
	text := ""
	if len(args) > 0 {
		var err error
		cmd, _, err = s.rootCmd.Traverse(args[:len(args)-1])
		if err != nil {
			log.Fatal(err)
		}
		text = args[len(args)-1]
	}
	suggestions := cmd.SuggestionsFor(text)
	var result []prompt.Suggest
	for _, suggestion := range suggestions {
		cmd, _, err := cmd.Traverse([]string{suggestion})
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, prompt.Suggest{
			Text:        suggestion,
			Description: cmd.Short,
		})
	}
	return result
}

func (s *shell) run() {
	prompt.New(
		s.executor,
		s.suggestor,
		prompt.OptionPrefix(">>> "),
	).Run()
}

// Run runs a prompt, it does not return.
func Run(rootCmd *cobra.Command) {
	newShell(rootCmd).run()
}

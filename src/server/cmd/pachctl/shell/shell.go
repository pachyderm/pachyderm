package shell

import (
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

func executor(in string) {
	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader("pachctl " + in)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return
}

func suggestor(in prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

// Run runs a prompt, it does not return.
func Run() {
	prompt.New(
		executor,
		suggestor,
		prompt.OptionPrefix(">>> "),
	).Run()
}

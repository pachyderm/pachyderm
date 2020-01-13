package shell

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/spf13/cobra"
)

const (
	completionAnnotation  string = "completion"
	ldThreshold           int    = 2
	defaultMaxCompletions int64  = 64
)

type CompletionFunc func(flag, arg string, maxCompletions int64) []prompt.Suggest

var completions map[string]CompletionFunc = make(map[string]CompletionFunc)

func RegisterCompletionFunc(cmd *cobra.Command, completionFunc CompletionFunc) {
	id := uuid.NewWithoutDashes()

	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	} else if _, ok := cmd.Annotations[completionAnnotation]; ok {
		panic("duplicate completion func registration")
	}
	cmd.Annotations[completionAnnotation] = id
	completions[id] = completionFunc
}

type shell struct {
	rootCmd        *cobra.Command
	maxCompletions int64
}

func newShell(rootCmd *cobra.Command, maxCompletions int64) *shell {
	if maxCompletions == 0 {
		maxCompletions = defaultMaxCompletions
	}
	return &shell{
		rootCmd:        rootCmd,
		maxCompletions: maxCompletions,
	}
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
	flag := ""
	if len(args) > 1 {
		if args[len(args)-2][0] == '-' {
			flag = args[len(args)-2]
		}
	}
	suggestions := cmd.SuggestionsFor(text)
	if len(suggestions) > 0 {
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
	if id, ok := cmd.Annotations[completionAnnotation]; ok {
		completionFunc := completions[id]
		suggests := completionFunc(flag, text, s.maxCompletions)
		var result []prompt.Suggest
		for _, sug := range suggests {
			sText := sug.Text
			if len(text) < len(sText) {
				sText = sText[:len(text)]
			}
			if ld(sText, text, true) < ldThreshold {
				result = append(result, sug)
			}
			if int64(len(result)) > s.maxCompletions {
				break
			}
		}
		return result
	}
	return nil
}

func (s *shell) run() {
	color.NoColor = true // color doesn't work in terminal
	prompt.New(
		s.executor,
		s.suggestor,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("Pachyderm Shell"),
		prompt.OptionLivePrefix(func() (string, bool) {
			cfg, err := config.Read(true)
			if err != nil {
				return "", false
			}
			activeContext, _, err := cfg.ActiveContext()
			if err != nil {
				return "", false
			}
			return fmt.Sprintf("context:(%s) >>> ", activeContext), true
		}),
	).Run()
}

// Run runs a prompt, it does not return.
func Run(rootCmd *cobra.Command, maxCompletions int64) {
	newShell(rootCmd, maxCompletions).run()
}

// ld computes the Levenshtein Distance for two strings.
func ld(s, t string, ignoreCase bool) int {
	if ignoreCase {
		s = strings.ToLower(s)
		t = strings.ToLower(t)
	}
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}
	return d[len(s)][len(t)]
}

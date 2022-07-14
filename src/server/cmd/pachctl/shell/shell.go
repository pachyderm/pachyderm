package shell

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
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

// CacheAll is a CacheFunc that always returns true (always use cached results).
func CacheAll(_, _ string) bool { return true }

// CacheNone is a CacheFunc that always returns false (never cache anything).
func CacheNone(_, _ string) bool { return false }

// SameFlag is a CacheFunc that returns true if the flags are the same.
func SameFlag(flag string) CacheFunc {
	return func(_flag, _ string) bool {
		return flag == _flag
	}
}

// AndCacheFunc ands 0 or more cache funcs together.
func AndCacheFunc(fs ...CacheFunc) CacheFunc {
	return func(flag, text string) bool {
		for _, f := range fs {
			if !f(flag, text) {
				return false
			}
		}
		return true
	}
}

type CompletionResult struct {
	Suggestions []prompt.Suggest
	Prefix      string
	CacheFunc   CacheFunc
}

// CompletionFunc is a function which returns completions for a command.
type CompletionFunc func(c *client.APIClient, flag, text string, maxCompletions int64) *CompletionResult

var completions map[string]CompletionFunc = make(map[string]CompletionFunc)

// RegisterCompletionFunc registers a completion function for a command.
// NOTE: RegisterCompletionFunc must be called before cmd is passed to
// functions that make copies of it (such as cmdutil.CreateAlias. This is
// because RegisterCompletionFunc modifies cmd in a superficial way by adding
// an annotation (to the Annotations field) that associates it with the
// completion function. This means that
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

	// variables for caching completion calls
	completionID     string
	completionResult *CompletionResult

	getClient func() *client.APIClient
}

func newShell(rootCmd *cobra.Command, maxCompletions int64, getClient func() *client.APIClient) *shell {
	if maxCompletions == 0 {
		maxCompletions = defaultMaxCompletions
	}
	return &shell{
		rootCmd:        rootCmd,
		maxCompletions: maxCompletions,
		getClient:      getClient,
	}
}

func (s *shell) executor(in string) {
	if in == "exit" {
		os.Exit(0)
	}

	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader("pachctl " + in)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func (s *shell) suggestor(in prompt.Document) []prompt.Suggest {
	return SuggestFromCommand(s.rootCmd, in, s.maxCompletions, func(id, flag, text string, fn CompletionFunc) (suggests []prompt.Suggest, prefix string) {
		if s.completionID != id || s.completionResult == nil ||
			s.completionResult.CacheFunc == nil || !s.completionResult.CacheFunc(flag, text) {
			s.completionID = id
			s.completionResult = fn(s.getClient(), flag, text, s.maxCompletions)
		}
		if s.completionResult != nil {
			suggests = s.completionResult.Suggestions
			prefix = s.completionResult.Prefix
		}
		return
	})
}

func SuggestFromCommand(
	cmd *cobra.Command,
	in prompt.Document,
	maxCompletions int64,
	cache func(string, string, string, CompletionFunc) ([]prompt.Suggest, string)) []prompt.Suggest {
	args := strings.Fields(in.Text)
	if strings.HasSuffix(in.Text, " ") {
		args = append(args, "")
	}
	text := ""
	if len(args) > 0 {
		var err error
		cmd, _, err = cmd.Traverse(args[:len(args)-1])
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
		suggestions, prefix := cache(id, flag, text, completions[id])
		return TrimSuggestionList(text, prefix, suggestions, int(maxCompletions))
	}
	return nil
}

func TrimSuggestionList(text, prefix string, suggestions []prompt.Suggest, maxCompletions int) []prompt.Suggest {
	var result []prompt.Suggest
	for _, sug := range suggestions {
		sText := prefix + sug.Text
		if len(text) < len(sText) {
			sText = sText[:len(text)]
		}
		if ld(sText, text, true) < ldThreshold {
			result = append(result, sug)
		}
		if len(result) > maxCompletions {
			break
		}
	}
	return result
}

func (s *shell) clearCache() {
	s.completionID = ""
	s.completionResult = nil
}

func (s *shell) run() {
	fmt.Printf("Type 'exit' or press Ctrl-D to exit.\n")
	color.NoColor = true // color doesn't work in terminal
	prompt.New(
		s.executor,
		s.suggestor,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("Pachyderm Shell"),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.F5,
			Fn:  func(*prompt.Buffer) { s.clearCache() },
		}),
		prompt.OptionLivePrefix(func() (string, bool) {
			cfg, err := config.Read(true, false)
			if err != nil {
				return "", false
			}
			activeContext, _, err := cfg.ActiveContext(false)
			if err != nil {
				return "", false
			}
			return fmt.Sprintf("context:(%s) >>> ", activeContext), true
		}),
		prompt.OptionCompletionWordSeparator(" @"), // split on @ to replace only commit/job IDs
	).Run()
	if pachClient != nil {
		if err := pachClient.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

// Run runs a prompt, it does not return.
func Run(rootCmd *cobra.Command, maxCompletions int64) {
	newShell(rootCmd, maxCompletions, getPachClient).run()
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

func RunCustom(rootCmd *cobra.Command, title string, c *client.APIClient, maxCompletions int64, execute func(in string)) {
	s := newShell(rootCmd, maxCompletions, func() *client.APIClient { return c })

	fmt.Printf("Type 'exit' or press Ctrl-D to exit.\n")
	color.NoColor = true
	prompt.New(
		execute,
		s.suggestor,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle(title),
		prompt.OptionCompletionWordSeparator(" @"), // split on @ to replace only commit/job IDs
	).Run()
}

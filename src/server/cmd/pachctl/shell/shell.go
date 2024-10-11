// Package shell needs to be documented.
//
// TODO: document
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
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

// CompletionFunc is a function which returns completions for a command.
type CompletionFunc func(flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc)

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
	completionID string
	suggests     []prompt.Suggest
	cacheF       CacheFunc
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
	if in == "exit" {
		os.Exit(0)
	}

	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader("pachctl " + in)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() //nolint:errcheck
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
			Fatal(err)
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
				Fatal(err)
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
		if s.completionID != id || s.cacheF == nil || !s.cacheF(flag, text) {
			s.completionID = id
			s.suggests, s.cacheF = completionFunc(flag, text, s.maxCompletions)
		}
		var result []prompt.Suggest
		for _, sug := range s.suggests {
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

func (s *shell) clearCache() {
	s.completionID = ""
	s.suggests = nil
	s.cacheF = nil
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
	).Run()
	if err := closePachClient(); err != nil {
		Fatal(err)
	}
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

func Fatal(args ...any) {
	fmt.Fprint(os.Stderr, args...)
	os.Exit(1)
}

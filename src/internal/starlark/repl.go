package starlark

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/wader/readline"

	"github.com/adrg/xdg"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func RunShell(ctx context.Context, path string, opts Options) error {
	_, err := run(ctx, path, opts, func(fileOpts *syntax.FileOptions, thread *starlark.Thread, module string, globals starlark.StringDict) (starlark.StringDict, error) {
		if module != "" {
			fmt.Printf("Running %v...", module)
			result, err := starlark.ExecFileOptions(fileOpts, thread, module, nil, globals)
			fmt.Println("done.")
			if err != nil {
				return nil, errors.Wrapf(err, "run %v", module)
			}
			globals = result
		}
		for k, v := range opts.REPLPredefined {
			globals[k] = v
		}
		rlcfg := &readline.Config{
			Prompt:       ">>> ",
			HistoryFile:  filepath.Join(xdg.StateHome, "starpach.history"),
			HistoryLimit: 10000,
			AutoComplete: &completer{t: thread, globals: globals},
		}
		rl, err := readline.NewEx(rlcfg)
		if err != nil {
			printError(err)
			return nil, err
		}
		defer rl.Close()
		for {
			signal.Reset(os.Interrupt) // This is rude, but oh well.
			oneCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
			thread.SetLocal(goContextKey, oneCtx)
			if err := rep(fileOpts, rl, thread, globals); err != nil {
				if err == readline.ErrInterrupt {
					fmt.Println(err)
					stop()
					continue
				}
				stop()
				break
			}
			stop()
		}
		return nil, nil
	})
	return err
}

type completer struct {
	t       *starlark.Thread
	globals starlark.StringDict
}

func (c *completer) Do(line []rune, pos int) ([][]rune, int) {
	if len(line) != pos {
		// Completing mid-line is difficult, so we opt out here.
		return nil, 0
	}
	var trim int
	for i, r := range line {
		if !unicode.IsSpace(r) {
			break
		} else {
			trim = i + 1
			pos--
		}
	}
	line = line[trim:]

	var results [][]rune
	// If a Universe function completes the entire line, include it.
	for k := range starlark.Universe {
		if strings.HasPrefix(k, string(line)) {
			results = append(results, []rune(k[pos:]))
		}
	}

	// If a global completes the entire line, include it.
	for k := range c.globals {
		if strings.HasPrefix(k, string(line)) {
			results = append(results, []rune(k[pos:]))
		}
	}

	// If there is an opening paren, treat it like the start of a new expression.
	lastParen := -1
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == '(' || line[i] == ',' || line[i] == '[' || line[i] == '=' {
			lastParen = i
			break
		}
	}
	if lastParen > 0 {
		newLine := line[lastParen+1:]
		results, _ = c.Do(newLine, len(newLine))
		return results, pos
	}

	// If there is a single period, assume it's <some global>.<want an attr>, and complete the
	// attr.
	lastDot := -1
	for i := len(line) - 1; i >= 0; i-- {
		// Search backwards for a .
		if line[i] == '.' {
			switch {
			case lastDot == -1:
				// Found one.
				lastDot = i
			case lastDot != -2:
				// Found another one, which invalidates this method of completion.
				// It's an expression like foo.bar.something, and we'd have to eval
				// foo.bar to figure out how to complete "something".
				lastDot = -2
			}
		}
	}
	if lastDot > 0 {
		sym := string(line[:lastDot])
		part := string(line[lastDot+1:])
		if sym, ok := c.globals[sym]; ok {
			if ha, ok := sym.(starlark.HasAttrs); ok {
				for _, attr := range ha.AttrNames() {
					if strings.HasPrefix(attr, part) {
						results = append(results, []rune(attr[len(part):]))
					}
				}
			}
		}
	}
	return results, pos
}

// printError prints the error to stderr,
// or its backtrace if it is a Starlark evaluation error.
func printError(err error) {
	if evalErr, ok := err.(*starlark.EvalError); ok {
		fmt.Fprintln(os.Stderr, evalErr.Backtrace())
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}

// rep is copied from go.starlark.net/repl.
func rep(opts *syntax.FileOptions, rl *readline.Instance, thread *starlark.Thread, globals starlark.StringDict) error {
	eof := false

	// readline returns EOF, ErrInterrupted, or a line including "\n".
	rl.SetPrompt(">>> ")
	readline := func() ([]byte, error) {
		line, err := rl.Readline()
		rl.SetPrompt("... ")
		if err != nil {
			if err == io.EOF {
				eof = true
			}
			return nil, err
		}
		return []byte(line + "\n"), nil
	}

	// Treat load bindings as global (like they used to be) in the REPL.
	// Fixes github.com/google/starlark-go/issues/224.
	opts2 := *opts
	opts2.LoadBindsGlobally = true
	opts = &opts2

	// parse
	f, err := opts.ParseCompoundStmt("<stdin>", readline)
	if err != nil {
		if eof {
			return io.EOF
		}
		printError(err)
		return nil
	}

	if expr := soleExpr(f); expr != nil {
		// eval
		v, err := starlark.EvalExprOptions(f.Options, thread, expr, globals)
		if err != nil {
			printError(err)
			return nil
		}
		globals["_"] = v // Python does this.

		// print
		if v != starlark.None {
			fmt.Println(v)
		}
	} else if err := starlark.ExecREPLChunk(f, thread, globals); err != nil {
		printError(err)
		return nil
	}

	return nil
}

func soleExpr(f *syntax.File) syntax.Expr {
	if len(f.Stmts) == 1 {
		if stmt, ok := f.Stmts[0].(*syntax.ExprStmt); ok {
			return stmt.X
		}
	}
	return nil
}

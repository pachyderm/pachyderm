package testutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
	"unicode"
)

// TestCmd aliases 'exec.Cmd' so that we can add
// func (c *TestCmd) Run(t *testing.T)
type TestCmd exec.Cmd

// dedent is a helper function that trims repeated, leading spaces from several
// lines of text. This is useful for long, inline commands that are indented to
// match the surrounding code but should be interpreted without the space:
// e.g.:
// func() {
//   BashCmd(`
//     cat <<EOF
//     This is a test
//     EOF
//   `)
// }
// Should evaluate:
// $ cat <<EOF
// > This is a test
// > EOF
// not:
// $      cat <<EOF
// >      This is a test
// >      EOF  # doesn't terminate heredoc b/c of space
// such that the heredoc doesn't end properly
func dedent(cmd string) string {
	notSpace := func(r rune) bool {
		return !unicode.IsSpace(r)
	}

	// Ignore leading space-only lines
	var prefix string
	s := bufio.NewScanner(strings.NewReader(cmd))
	var dedentedCmd bytes.Buffer
	for s.Scan() {
		i := strings.IndexFunc(s.Text(), notSpace)
		if i == -1 {
			continue // blank line -- ignore completely
		} else if prefix == "" {
			prefix = s.Text()[:i] // first non-blank line, set prefix
			dedentedCmd.WriteString(s.Text()[i:])
			dedentedCmd.WriteRune('\n')
		} else if strings.HasPrefix(s.Text(), prefix) {
			dedentedCmd.WriteString(s.Text()[len(prefix):]) // remove prefix
			dedentedCmd.WriteRune('\n')
		} else {
			return cmd // no common prefix--break early
		}
	}
	fmt.Printf("(dedentedCmd)---\n%s---\n", dedentedCmd.String())
	return dedentedCmd.String()
}

// Cmd is a convenience function that replaces exec.Command. It's both shorter
// and it uses the current process's stderr as output for the command, which
// makes debugging failures much easier (i.e. you get an error message
// rather than "exit status 1")
func Cmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	// for convenience, simulate hitting "enter" after any prompt. This can easily
	// be replaced
	cmd.Stdin = strings.NewReader("\n")
	cmd.Env = os.Environ()
	return cmd
}

// BashCmd is a convenience function that:
// 1. Performs a Go template substitution on 'cmd' using the strings in 'subs'
// 2. Ruturns a command that runs the result string from 1 as a Bash script
func BashCmd(cmd string, subs ...string) *TestCmd {
	fmt.Printf(">>> %s\n", cmd)
	if len(subs)%2 == 1 {
		panic("some variable does not have a corresponding value")
	}

	// copy 'subs' into a map
	subsMap := make(map[string]string)
	for i := 0; i < len(subs); i += 2 {
		subsMap[subs[i]] = subs[i+1]
	}

	// Warn users that they must install 'match' if they want to run tests with
	// this library, and enable 'pipefail' so that if any 'match' in a chain
	// fails, the whole command fails.
	buf := &bytes.Buffer{}
	buf.WriteString(`
set -e -o pipefail
which match >/dev/null || {
	echo "You must have 'match' installed to run these tests. Please run:" >&2
	echo "  cd ${GOPATH}/src/github.com/pachyderm/pachyderm/ && go install ./src/server/cmd/match" >&2
	exit 1
}`)
	buf.WriteRune('\n')

	// do the substitution
	template.Must(template.New("").Parse(dedent(cmd))).Execute(buf, subsMap)
	res := exec.Command("bash", "-c", buf.String())
	res.Stderr = os.Stderr
	res.Stdin = strings.NewReader("\n")
	res.Env = os.Environ()
	return (*TestCmd)(res)
}

// Run wraps (*exec.Cmd).Run(), though in the particular case that the command
// is run on Linux and fails with a SIGPIPE error (which often happens because
// TestCmd is created by BashCmd() above, which also sets -o pipefail), this
// throws away the error. See "The Infamous SIGPIPE Signal"
// http://www.tldp.org/LDP/lpg/node20.html
func (c *TestCmd) Run() error {
	cmd := (*exec.Cmd)(c)
	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 141 {
				return nil
			}
		}
	}
	return err
}

package testutil

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
)

// TestCmd aliases 'exec.Cmd' so that we can add
// func (c *TestCmd) Run(t *testing.T)
type TestCmd exec.Cmd

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
	return cmd
}

// BashCmd is a convenience function that:
// 1. Performs a Go template substitution on 'cmd' using the strings in 'subs'
// 2. Ruturns a command that runs the result string from 1 as a Bash script
func BashCmd(cmd string, subs ...string) *TestCmd {
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
	which match || {
		echo "You must have 'match' installed to run these tests. Please run:"
		echo "  go install ${GOPATH}/src/github.com/pachyderm/pachyderm/src/server/cmd/match"
		exit 1
	}
	`)
	// do the substitution
	template.Must(template.New("").Parse(cmd)).Execute(buf, subsMap)
	res := exec.Command("bash", "-c", buf.String())
	res.Stderr = os.Stderr
	res.Stdin = strings.NewReader("\n")
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

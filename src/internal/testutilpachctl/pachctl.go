// Package testutilpachctl contains utilities for running pachctl commands directly.
package testutilpachctl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

const contextName = "test-context"

type Pachctl struct {
	configPath string
}

// NewPachctl returns a new Pachctl object configured with the given client
// information, with the configuration stored in configPath.
func NewPachctl(ctx context.Context, c *client.APIClient, configPath string) (*Pachctl, error) {
	var p = &Pachctl{configPath}

	// If the config file exists, then let the caller determine if it
	// expected the situation and wishes to use the Pachctl object.
	if _, err := os.Open(configPath); err == nil {
		return p, errors.Wrap(fs.ErrExist, configPath)
	}

	// test filesystem access by creating a temporary file
	if _, err := os.Create(configPath); err != nil {
		return nil, errors.Wrap(err, "could not create placeholder pachctl config")
	}
	// remove the empty file so that a config can be generated
	if err := os.Remove(configPath); err != nil {
		return nil, errors.Wrap(err, "could not delete placeholder pachctl config")
	}
	cmd, err := p.CommandTemplate(ctx, `
		pachctl config set context  --overwrite {{ .context }} <<EOF
		{
		  "source": 2,
		  "session_token": "{{.token}}",
		  "pachd_address": "grpc://{{.host}}:{{.port}}",
		  "cluster_deployment_id": "dev"
		}
		EOF
		pachctl config set active-context {{ .context }}
		`,
		map[string]string{
			"config":  configPath,
			"context": contextName,
			"token":   c.AuthToken(),
			"host":    c.GetAddress().Host,
			"port":    fmt.Sprint(c.GetAddress().Port),
		})
	if err != nil {
		return nil, errors.Wrap(err, "could not create pachctl-configuration command")
	}
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "could not configure pachctl")
	}
	return p, nil
}

// Close cleans up the pachctl config file.  It does not delete its parent
// directory.
func (p Pachctl) Close() error {
	return errors.Wrapf(os.Remove(p.configPath), "could not delete pachctl config %q", p.configPath)
}

func (p Pachctl) bashPrelude(w io.Writer) error {
	if err := writeBashPrelude(w); err != nil {
		return errors.Wrap(err, "write bash prelude")
	}
	if _, err := fmt.Fprintf(w, `export PACH_CONFIG="%s"`+"\n", p.configPath); err != nil {
		return errors.Wrap(err, "write PACH_CONFIG")
	}
	return nil
}

// writeTemplate dedents the given script, parses it as a Go template and writes
// the filled-in template to w.
func writeTemplate(w io.Writer, s string, data any) error {
	tmpl, err := template.New(uuid.UniqueString("template")).Parse(dedent(s))
	if err != nil {
		return errors.Wrap(err, "could not create new template")
	}
	if err := tmpl.Execute(w, data); err != nil {
		return errors.Wrap(err, "could not execute script template")
	}
	return nil
}

// bashTemplate interprets scriptTemplate as an indented Go template for a Bash
// script; it returns an io.Reader from which the filled-in script may be read.
func (p Pachctl) bashTemplate(scriptTemplate string, data any) (io.Reader, error) {
	// Warn users that they must install 'match' if they want to run tests with
	// this library, and enable 'pipefail' so that if any 'match' in a chain
	// fails, the whole command fails.
	buf := &bytes.Buffer{}
	if err := p.bashPrelude(buf); err != nil {
		return nil, errors.Wrap(err, "could not write prelude before template")
	}
	if err := writeTemplate(buf, scriptTemplate, data); err != nil {
		return nil, errors.Wrap(err, "could not write template")
	}
	return buf, nil
}

type Cmd struct {
	*exec.Cmd
	stdout, stderr *bytes.Buffer
}

// Stdout reads all of the command’s standard output.  It panics in case of an
// error.
func (cmd Cmd) Stdout() string {
	b, err := io.ReadAll(cmd.stdout)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// Stderr reads all of the command’s error output.  It panics in case of an
// error.
func (cmd Cmd) Stderr() string {
	b, err := io.ReadAll(cmd.stderr)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func newCmd(ctx context.Context, name string, args []string, stdin io.Reader) Cmd {
	var cmd Cmd
	cmd.Cmd = exec.CommandContext(ctx, name, args...)
	cmd.Cmd.Stdin = stdin
	cmd.stdout = new(bytes.Buffer)
	cmd.Cmd.Stdout = cmd.stdout
	cmd.stderr = new(bytes.Buffer)
	cmd.Cmd.Stderr = cmd.stderr
	cmd.Env = os.Environ()
	return cmd
}

func (cmd *Cmd) Run() error {
	if err := cmd.Cmd.Run(); err != nil {
		if cmd.Cmd.Stderr != cmd.stderr {
			return errors.Wrap(err, "command failed without buffered stderr")
		}
		return errors.Wrapf(err, "command failed\n====== BEGIN STDERR ======\n%s\n====== END STDERR ======\n", cmd.stderr.String())
	}
	return nil
}

// Command provides a Cmd to execute script with Bash.  If context is cancelled
// then the command will be terminated.
func (p Pachctl) Command(ctx context.Context, script string) (Cmd, error) {
	var buf = new(bytes.Buffer)
	if err := p.bashPrelude(buf); err != nil {
		return Cmd{}, errors.Wrap(err, "could not insert prelude")
	}
	fmt.Fprintf(buf, "export PACH_CONFIG=\"%s\"\n", p.configPath)
	buf.WriteString(script)
	return newCmd(ctx, "/bin/bash", nil, buf), nil
}

// CommandTemplate provides a Cmd to execute script with Bash.  Script is
// interpreted as a Go template and provided with data.  If context is cancelled
// then the command will be terminated.
func (p *Pachctl) CommandTemplate(ctx context.Context, scriptTemplate string, data any) (Cmd, error) {
	r, err := p.bashTemplate(scriptTemplate, data)
	if err != nil {
		return Cmd{}, errors.Wrap(err, "could not execute script template")
	}
	return newCmd(ctx, "/bin/bash", nil, r), nil
}

// RunCommand runs command in sh (rather than bash), returning the combined
// stdout & stderr.
func (p Pachctl) RunCommand(ctx context.Context, command string) (string, error) {
	env := os.Environ()

	// Adjust PATH to point at pachctl if it's in the runfiles for this invocation.
	if pachctl, ok := bazel.FindBinary("//src/server/cmd/pachctl", "pachctl"); ok {
		for i, entry := range env {
			if strings.HasPrefix(entry, "PATH=") {
				oldPath := entry[len("PATH="):]
				env[i] = fmt.Sprintf("PATH=%s:%s", filepath.Dir(pachctl), oldPath)
			}
		}
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Env = append(env, fmt.Sprintf(`PACH_CONFIG=%s`, shellescape.Quote(p.configPath)))
	b, err := cmd.CombinedOutput()
	return string(b), err
}

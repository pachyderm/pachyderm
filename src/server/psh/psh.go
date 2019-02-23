package psh

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sync/errgroup"
)

const (
	pfsPath = "/pfs"
)

var (
	prompts = []string{"🐘 ", "🐎 ", "🐴 ", "🐮 ", "🐂 "}
)

type Shell struct {
	*terminal.Terminal
	state *terminal.State
	c     *client.APIClient
}

func NewShell(c *client.APIClient) (*Shell, error) {
	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		return nil, fmt.Errorf("stdin/stdout not a terminal")
	}
	state, err := terminal.MakeRaw(0)
	if err != nil {
		return nil, err
	}
	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	return &Shell{
		Terminal: terminal.NewTerminal(screen, prompts[rand.Intn(len(prompts))]),
		state:    state,
		c:        c,
	}, nil
}

func (s *Shell) Run() error {
	defer terminal.Restore(0, s.state)
	for {
		line, err := s.ReadLine()
		if err == io.EOF {
			fmt.Fprintln(os.Stdout, "exit")
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.runLine(line); err != nil {
			return err
		}
	}
}

func (s *Shell) runLine(line string) error {
	parts := strings.Split(line, " ")
	if len(parts) == 0 {
		return nil
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	terminal.Restore(0, s.state)
	var eg errgroup.Group
	puller := sync.NewPuller()
	for _, arg := range parts[1:] {
		if strings.HasPrefix(arg, pfsPath) {
			pfsPath := strings.TrimPrefix(arg, pfsPath)
			pathParts := strings.Split(pfsPath, "/")
			fmt.Println(pathParts)
			eg.Go(func() error {
				return puller.Pull(s.c, pfsPath, pathParts[1], "master", path.Join(pathParts[2:]...), false, false, 20, nil, "")
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	var err error
	s.state, err = terminal.MakeRaw(0)
	if err != nil {
		return err
	}
	return nil
}

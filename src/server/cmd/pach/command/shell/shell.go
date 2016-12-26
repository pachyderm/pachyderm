package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/urfave/cli"
)

type shell struct {
	app      *cli.App
	prompt   string
	exitCmds []string
	sigs     []os.Signal
}

// Option initializes the shell
type Option func(*shell)

// WithPrompt specifies shell prompt
func WithPrompt(p string) Option {
	return func(s *shell) {
		s.prompt = p
	}
}

// WithSignalMask specifies signals to mask.
func WithSignalMask(sigs ...os.Signal) Option {
	return func(s *shell) {
		s.sigs = sigs
	}
}

// WithExitCmds specifies shell exit command.
func WithExitCmds(exitCmds ...string) Option {
	return func(s *shell) {
		s.exitCmds = exitCmds
	}
}

func newShell(app *cli.App, opts ...Option) *shell {
	s := &shell{
		app:      app,
		prompt:   ">",
		exitCmds: []string{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *shell) run() error {
	if len(s.sigs) == 0 {
		return s.loop()
	}
	cli.OsExiter = func(c int) {}
	sigs := make(chan os.Signal, 1)
	go func() {
		for range sigs {
			if len(s.exitCmds) != 0 {
				fmt.Printf("\n(type %v exit)\n\n", s.exitCmds)
			}
			fmt.Printf("%s", s.prompt)
		}
	}()
	defer close(sigs)
	signal.Notify(sigs, s.sigs...)
	err := s.loop()
	signal.Stop(sigs)
	return err
}

func (s *shell) loop() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s", s.prompt)
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if text == "" {
			time.Sleep(time.Second)
			continue
		}
		for _, pat := range s.exitCmds {
			if text == pat {
				return nil
			}
		}
		if err := s.app.Run(append(os.Args[1:], strings.Split(text, " ")...)); err != nil {
			// discard the error, as we want to exit simply because of a sub command fails.
		}
	}
}

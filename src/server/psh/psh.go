package psh

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	prompts = []string{"🐘 ", "🐎 ", "🐴 ", "🐮 ", "🐂 "}
)

func Shell() error {
	if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
		return fmt.Errorf("stdin/stdout not a terminal")
	}
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		return err
	}
	defer terminal.Restore(0, oldState)
	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	term := terminal.NewTerminal(screen, prompts[rand.Intn(len(prompts))])
	for {
		line, err := term.ReadLine()
		if err == io.EOF {
			fmt.Fprintln(os.Stdout, "exit")
			return nil
		}
		if err != nil {
			return err
		}
		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			return nil
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		terminal.Restore(0, oldState)
		if err := cmd.Run(); err != nil {
			return err
		}
		oldState, err = terminal.MakeRaw(0)
		if err != nil {
			return err
		}
	}
}

func Run(line string) error {
	parts := strings.Split(line, " ")
	if len(parts) == 0 {
		return nil
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

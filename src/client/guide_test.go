package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func parseCommandFromShellString(shell string) string {
	lines := strings.Split(shell, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "$") {
			normalizedCommand := strings.TrimSpace(strings.TrimLeft(line, "$"))
			if len(normalizedCommand) > 0 {
				return normalizedCommand
			}
		}
	}
	return ""
}

type Command struct {
	Cmd            string
	Args           []string
	ExpectedOutput string
	Fork           bool
}

func runHelper(c *Command) (string, error) {
	/*	var finalArgs []string
		finalArgs = append(finalArgs, "-cE")
		finalArgs = append(finalArgs, c.Cmd)
		finalArgs = append(finalArgs, c.Args...)*/

	shellCommand := exec.Command(c.Cmd, c.Args...)
	//	shellCommand := exec.Command("bash", finalArgs...)

	basePath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/pachyderm/pachyderm")
	shellCommand.Dir = basePath

	raw, err := shellCommand.CombinedOutput()
	fmt.Printf("Output: [%v]\n", string(raw))
	return string(raw), err

	//return "didnotrun", nil
}

func (c *Command) Run(input string) (string, error) {
	fmt.Printf("Running [%v %v] ... forked? %v \n", c.Cmd, c.Args, c.Fork)

	homePath := fmt.Sprintf("%v/", os.Getenv("HOME"))

	for i, arg := range c.Args {
		c.Args[i] = strings.Replace(arg, "CHAINED_INPUT", input, 2)
		c.Args[i] = strings.Replace(c.Args[i], "~/", homePath, 2)
	}

	if c.Fork {
		fmt.Printf("I should be forking: [%v]\n", c)

		done := make(chan bool)

		func(cmd *Command) {
			fmt.Printf("2 - I should be forking: [%v | %v]\n", c, cmd)
			go func() {
				fmt.Printf("3 - I should be forking: [%v | %v]\n", c, cmd)
				// Remove the &
				//				var finalArgs []string
				//				finalArgs = append(finalArgs, "-cE")
				/*
					var nestedCommand []string
					nestedCommand = append(nestedCommand, c.Cmd)
					nestedCommand = append(nestedCommand, c.Args...)
					rawNestedCommand := fmt.Sprintf("\"%v\"", strings.Join(nestedCommand, " "))
					finalArgs = append(finalArgs, rawNestedCommand)
				*/

				//				finalArgs = append(finalArgs, c.Cmd)
				//				finalArgs = append(finalArgs, c.Args...)
				//				c.Args = finalArgs
				//				c.Cmd = "bash"

				/*				var args []string
								for _, arg := range cmd.Args {
									if arg != "&" {
										args = append(args, arg)
									}
								}
								cmd.Args = args*/

				fmt.Printf("!!! forked job for command: %v\n", c)
				done <- true

				_, err := runHelper(c)

				if err != nil {
					fmt.Printf("Error Forking Mounting: %v\n", err.Error())
				}
			}()
		}(c)

		// lazy
		<-done

		return "", nil
	}

	out, err := runHelper(c)

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func parseCommand(raw string) *Command {
	cmd := &Command{}

	// Detect if there are any special directives
	if strings.Contains(raw, "[//]") {
		// Options are SKIP | CHECK_OUTPUT | custom command

		re, err := regexp.Compile(`\[\/\/\].*?\n`)
		if err != nil {
			fmt.Printf(err.Error())
		}
		result := re.FindAllStringSubmatch(raw, -1)
		line := result[0][0]

		re, err = regexp.Compile("\\((.*?)\\)")
		if err != nil {
			fmt.Printf(err.Error())
		}
		result = re.FindAllStringSubmatch(line, -1)
		if len(result) > 0 && len(result[0]) == 2 {
			directive := result[len(result)-1][1]
			switch directive {
			case "SKIP":
				return nil
			case "CHECK_OUTPUT":
				shellString := strings.Split(raw, "```")
				cmd.ExpectedOutput = parseExpectedOutput(shellString[0])
			case "FORK":
				cmd.Fork = true
			default:
				rawCommand := strings.TrimSpace(strings.TrimLeft(directive, "$"))
				cmd.Cmd, cmd.Args = splitCommand(rawCommand)
				return cmd
			}
		}

	}
	secondaryTokens := strings.Split(raw, "```")
	rawCommand := parseCommandFromShellString(secondaryTokens[0])

	if rawCommand == "" {
		return nil
	}

	cmd.Cmd, cmd.Args = splitCommand(rawCommand)

	return cmd
}

func splitCommand(rawCommand string) (string, []string) {
	tokens := strings.SplitAfterN(rawCommand, " ", 2)
	cmd := strings.TrimSpace(tokens[0])
	args := strings.Split(strings.TrimSpace(tokens[1]), " ")
	return cmd, args
}

func parseExpectedOutput(raw string) string {
	var output []string

	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "$") || len(line) == 0 {
			continue
		}
		output = append(output, line)
	}

	return strings.Join(output, "\n")
}

func runCommands(t *testing.T, commands []Command) {
	pipe := ""
	var last Command
	for _, c := range commands {
		done := false
		runOnce := true
		if last.Fork {
			// Last command was a fork
			runOnce = false
		}

		for !done {
			out, err := c.Run(pipe)
			if err != nil {
				t.Errorf(err.Error())
				t.Fail()
			}
			if out != "" {
				pipe = out
			}
			if runOnce {
				done = true
			} else {
				fmt.Printf("sleeping...\n")
				time.Sleep(1 * time.Second)
			}
		}

		last = c
	}
}

func TestGuide(t *testing.T) {
	data, _ := ioutil.ReadFile("../../examples/fruit_stand/GUIDE.md")
	tokens := strings.Split(string(data), "```shell")

	var commands []Command
	for _, token := range tokens {
		c := parseCommand(token)
		if c != nil {
			commands = append(commands, *c)
		}
	}

	runCommands(t, commands)
}

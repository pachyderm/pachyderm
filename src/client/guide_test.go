package client

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
	"testing"
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
	Chain          bool
}

func (c *Command) Run(input string) (string, error) {

	if c.Chain {
		c.Cmd = strings.Replace(c.Cmd, "CHAINED_INPUT", input, -1)
	}

	fmt.Printf("Running [%v %v] ... \n", c.Cmd, c.Args)
	shellCommand := exec.Command(c.Cmd, c.Args...)
	out, err := shellCommand.CombinedOutput()
	fmt.Printf("Output: [%v]\n", string(out))

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func parseCommand(raw string) *Command {
	cmd := &Command{}

	// Detect if there are any special directives
	if strings.Contains(raw, "[//]") {
		// Options are SKIP | CHAIN_OUTPUT | CHECK_OUTPUT | custom command

		re, err := regexp.Compile("\\((.*?)\\)")
		if err != nil {
			fmt.Printf(err.Error())
		}

		result := re.FindAllStringSubmatch(raw, -1)

		if len(result) > 0 && len(result[0]) == 2 {
			directive := result[0][1]
			switch directive {
			case "SKIP":
				return nil
			case "CHAIN_OUTPUT":
				cmd.Chain = true
			case "CHECK_OUTPUT":
				shellString := strings.Split(raw, "```")
				cmd.ExpectedOutput = parseExpectedOutput(shellString[0])
			case "FORK":
				fmt.Printf("I SHOULD FORK THIS\n")
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
	for _, c := range commands {
		fmt.Printf("Running chained? %v, [%v %v], expected output [%v]\n", c.Chain, c.Cmd, c.Args, c.ExpectedOutput)
		out, err := c.Run(pipe)
		if err != nil {
			t.Errorf(err.Error())
		}
		pipe = out
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

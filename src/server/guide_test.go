package server

import (
	"fmt"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"google.golang.org/grpc"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
	Mount          bool
}

func collapseArgs(args []string) string {
	//	return fmt.Sprintf("\"%v\"", strings.Join(args, " "))
	return strings.Join(args, " ")
}
func finalizeArgs(c *Command) []string {
	var finalArgs []string
	var argBuffer []string

	finalArgs = append(finalArgs, "-cE")
	//	finalArgs = append(finalArgs, c.Cmd)

	var rawArgs []string
	rawArgs = append(rawArgs, c.Cmd)
	rawArgs = append(rawArgs, c.Args...)

	finalArgs = append(finalArgs, strings.Join(rawArgs, " "))

	return finalArgs

	for _, rawArg := range c.Args {
		if rawArg != ">" {
			argBuffer = append(argBuffer, rawArg)
		} else {

			finalArgs = append(finalArgs, collapseArgs(argBuffer))
			argBuffer = make([]string, 0)
		}
	}
	finalArgs = append(finalArgs, collapseArgs(argBuffer))

	return finalArgs
}

func runHelper(c *Command) (string, error) {

	//	shellCommand := exec.Command(c.Cmd, c.Args...)
	args := finalizeArgs(c)
	fmt.Printf("final args: (%v)\n", args)
	shellCommand := exec.Command("bash", args...)

	basePath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/pachyderm/pachyderm")
	shellCommand.Dir = basePath

	raw, err := shellCommand.CombinedOutput()
	fmt.Printf("Output: [%v]\n", string(raw))
	return string(raw), err

	//	return "didnotrun", nil
}

func getAPIClient(address string) (pfsclient.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfsclient.NewAPIClient(clientConn), nil
}

func shard(fileNumber int, fileModulus int, blockNumber int, blockModulus int) *pfsclient.Shard {

	return &pfsclient.Shard{
		FileNumber:   uint64(fileNumber),
		FileModulus:  uint64(fileModulus),
		BlockNumber:  uint64(blockNumber),
		BlockModulus: uint64(blockModulus),
	}
}

func (c *Command) Run(input string) (string, error) {
	fmt.Printf("Running [%v %v] ... forked? %v \n", c.Cmd, c.Args, c.Mount)
	var wg sync.WaitGroup
	defer wg.Wait()

	homePath := fmt.Sprintf("%v/", os.Getenv("HOME"))

	for i, arg := range c.Args {
		c.Args[i] = strings.Replace(arg, "CHAINED_INPUT", input, 2)
		c.Args[i] = strings.Replace(c.Args[i], "~/", homePath, 2)
	}

	if c.Mount {
		fmt.Printf("I should be forking: [%v]\n", c)
		mountPoint := filepath.Join(os.Getenv("HOME"), "pfs")
		address := "0.0.0.0:30650"

		apiClient, err := getAPIClient(address)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		mounter := fuse.NewMounter(address, apiClient)

		ready := make(chan bool)

		go func() {
			fmt.Printf("XXX mounting\n")
			err := mounter.Mount(mountPoint, shard(0, 1, 0, 1), nil, ready)
			//		err = mounter.Mount(mountPoint, shard(0, 1, 0, 1), nil, nil)
			fmt.Printf("completed mount call\n")
			if err != nil {
				fmt.Printf("Error mounting: %v\n", err)
			}
		}()
		/*
			err = <-didError
			if err != nil {
				return "", err
			}*/
		fmt.Printf("Waiting for mount...\n")
		<-ready
		fmt.Printf("Mounted!\n")
		fmt.Printf("Sleeping for 20s\n")
		time.Sleep(20 * time.Second)
		fmt.Printf("Done Sleeping!\n")

		return "", err
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
		// Options are SKIP | CHECK_OUTPUT | MOUNT | custom command

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
			case "MOUNT":
				cmd.Mount = true
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
		out, err := c.Run(pipe)
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
		out = strings.TrimSpace(out)
		if c.ExpectedOutput != "" {
			fmt.Printf("Checking output ... expected[%v] == actual[%v] ?\n", c.ExpectedOutput, out)
			if c.ExpectedOutput != out {
				t.Errorf("Mismatched output. Expectd [%v], Actual [%v]\n", c.ExpectedOutput, out)
				t.FailNow()
			}
		}
		if out != "" {
			pipe = out
		}
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

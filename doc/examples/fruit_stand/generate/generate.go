package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func item() string {
	return [3]string{"apple", "banana", "orange"}[rand.Intn(3)]
}

func amount() string {
	return fmt.Sprintf("%d", rand.Intn(9)+1)
}

func line() string {
	return fmt.Sprintf("%s\t%s\n", item(), amount())
}

type reader struct {
	lines int
}

func (r *reader) Read(p []byte) (int, error) {
	size := 0
	for ; r.lines != 0; r.lines-- {
		line := line()
		if len(line) > len(p[size:]) {
			break
		}
		copy(p[size:], line)
		size += len(line)
	}
	if r.lines == 0 {
		return size, io.EOF
	}
	return size, nil
}

func main() {
	var commits int
	var lines int
	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Generate pfs traffic.",
		Long:  "Generate pfs traffic.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			rand.Seed(time.Now().UnixNano())
			client, err := client.NewFromAddress("0.0.0.0:30650")
			if err != nil {
				return err
			}
			for i := 0; i < commits; i++ {
				if _, err := client.StartCommit("data", "master"); err != nil {
					return err
				}
				if _, err := client.PutFile("data", "master", "sales", &reader{lines}); err != nil {
					return err
				}
				if err := client.FinishCommit("data", "master"); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	cmd.Flags().IntVarP(&commits, "commits", "c", 100, "commits to write")
	cmd.Flags().IntVarP(&lines, "lines", "l", 100, "lines to write for each commit")
	cmd.Execute()
}

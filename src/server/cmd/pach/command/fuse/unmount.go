package fuse

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli"

	"go.pedge.io/pkg/exec"
)

func newUnmountCommand() cli.Command {
	return cli.Command{
		Name:        "unmount",
		Aliases:     []string{"u"},
		Usage:       "Unmount pfs.",
		ArgsUsage:   "path/to/mount/point",
		Description: "Unmount pfs.",
		Action:      actUnmount,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all, a",
				Usage: "Unmount all pfs mounts.",
			},
			cli.IntFlag{Name: "file-shard, s", Usage: "file shard to read", Value: 0},
			cli.IntFlag{Name: "file-modulus, m", Usage: "file shard to read", Value: 1},
			cli.IntFlag{Name: "block-shard, b", Usage: "block shard to read", Value: 0},
			cli.IntFlag{Name: "block-modulus, n", Usage: "file shard to read", Value: 1},
		},
	}
}

func actUnmount(c *cli.Context) error {
	if c.NArg() == 1 {
		return syscall.Unmount(c.Args().First(), 0)
	}

	if c.Bool("all") {
		stdin := strings.NewReader(`mount | grep pfs:// | cut -f 3 -d " " `)
		var stdout bytes.Buffer
		if err := pkgexec.RunIO(pkgexec.IO{Stdin: stdin, Stdout: &stdout, Stderr: os.Stderr}, "sh"); err != nil {
			return err
		}
		scanner := bufio.NewScanner(&stdout)
		var mounts []string
		for scanner.Scan() {
			mounts = append(mounts, scanner.Text())
		}
		if len(mounts) == 0 {
			fmt.Println("No mounts found.")
			return nil
		}
		fmt.Printf("Unmount the following filesystems? yN\n")
		for _, mount := range mounts {
			fmt.Printf("%s\n", mount)
		}
		r := bufio.NewReader(os.Stdin)
		bytes, err := r.ReadBytes('\n')
		if err != nil {
			return err
		}
		if bytes[0] == 'y' || bytes[0] == 'Y' {
			for _, mount := range mounts {
				if err := syscall.Unmount(mount, 0); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

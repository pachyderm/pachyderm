// package pipeline implements a system for running data pipelines on top of
// the filesystem
package pipeline

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pachyderm/pfs/lib/container"
	"github.com/samalba/dockerclient"
)

func addImport(hostPath string, readOnly bool, container *dockerclient.ContainerConfig) {
	bind := fmt.Sprintf("%s:%s", hostPath, path.Join("/in", path.Base(hostPath)))
	if readOnly {
		bind = ":ro"
	}
	container.Binds = append(container.Binds, bind)
}

func RunPachFile(inPath, outPath string, r io.Reader) error {
	// Structures for launching containers
	// TODO figure out how to make the container read only
	containerConfig := &dockerclient.ContainerConfig{Image: "ubuntu", AttachStdout: true, AttachStderr: true}
	hostConfig := dockerclient.HostConfig{}

	outBind := fmt.Sprintf("%s:%s", outPath, "/out")
	hostConfig.Binds = append(hostConfig.Binds, outBind)

	lines := bufio.Scanner(r)

	for lines.Scan() {
		tokens := strings.Fields(lines.Text())
		if len(tokens) == 0 {
			continue
		}

		switch strings.ToUpper(tokens[0]) {
		case "image":
			containerConfig.Image = tokens[1]
		case "import":
			if len(tokens) < 1 {
				log.Print("Empty import line.")
				continue
			}
			// notice that this bind is read-only (ro)
			bind := fmt.Sprintf("%s:%s:ro", path.Join(inPath, tokens[1]), path.Join("/in", tokens[1]))
			hostConfig.Binds = append(hostConfig.Binds, bind)
		case "run":
			containerConfig.Cmd = tokens[1:]
			containerId, err := container.RawStartContainer(containerConfig, hostConfig)
			if err != nil {
				log.Print(err)
				return err
			}
			r, err := client.ContainerLogs(containerId, &dockerclient.LogOptions{Follow: true, Stdout: true, Stderr: true, Timestamps: true})
			if err != nil {
				log.Print(err)
				return err
			}
			f, err := os.Create(path.Join(outPath, ".log"))
			if err != nil {
				log.Print(err)
				return err
			}
			// Blocks until the command has exited.
			_, err := io.Copy(f, r)
			f.Close()
			if err != nil {
				log.Print(err)
				return err
			}
		}
	}
}

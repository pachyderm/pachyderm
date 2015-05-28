// package pipeline implements a system for running data pipelines on top of
// the filesystem
package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/container"
	"github.com/samalba/dockerclient"
)

type Pipeline struct {
	name              string
	containerConfig   dockerclient.ContainerConfig
	hostConfig        dockerclient.HostConfig
	dataRepo, outRepo string
	commit, branch    string
	counter           int
}

func NewPipeline(name, dataRepo, outRepo, commit, branch string) *Pipeline {
	return &Pipeline{
		name:     name,
		dataRepo: dataRepo,
		outRepo:  outRepo,
		commit:   commit,
		branch:   branch,
	}
}

// Import makes a dataset available for computations in the container.
func (p *Pipeline) Import(name string) error {
	hostPath := btrfs.HostPath(path.Join(p.dataRepo, p.commit, name))
	containerPath := path.Join("/in", name)

	bind := fmt.Sprintf("%s:%s:ro", hostPath, containerPath)
	p.hostConfig.Binds = append(p.hostConfig.Binds, bind)
	return nil
}

// Image sets the image that is being used for computations.
func (p *Pipeline) Image(image string) error {
	p.containerConfig.Image = image
	return nil
}

func (p *Pipeline) Run(cmd []string) error {
	defer func() { p.counter++ }()
	p.containerConfig.Cmd = cmd

	hostPath := btrfs.HostPath(path.Join(p.outRepo, p.branch))

	// Map in the out directory as a bind
	bind := fmt.Sprintf("%s:/out", hostPath)
	p.hostConfig.Binds = append(p.hostConfig.Binds, bind)
	// Make sure this bind is only visible for the duration of run
	defer func() { p.hostConfig.Binds = p.hostConfig.Binds[:len(p.hostConfig.Binds)-1] }()

	containerId, err := container.RawStartContainer(fmt.Sprintf("%d-%s", p.counter, p.name),
		&p.containerConfig, &p.hostConfig)
	if err != nil {
		log.Print(err)
		return err
	}
	r, err := container.ContainerLogs(containerId, &dockerclient.LogOptions{Follow: true, Stdout: true, Stderr: true, Timestamps: true})
	if err != nil {
		log.Print(err)
		return err
	}
	f, err := btrfs.CreateAll(path.Join(p.outRepo, p.branch, ".log"))
	if err != nil {
		log.Print(err)
		return err
	}
	defer f.Close()
	// Blocks until the command has exited.
	_, err = io.Copy(f, r)
	if err != nil {
		log.Print(err)
		return err
	}

	err = btrfs.Commit(p.outRepo, strconv.Itoa(p.counter), p.branch)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (p *Pipeline) RunPachFile(r io.Reader) error {
	lines := bufio.NewScanner(r)

	counter := 0
	for lines.Scan() {
		tokens := strings.Fields(lines.Text())
		if len(tokens) < 2 {
			continue
		}

		var err error
		switch strings.ToLower(tokens[0]) {
		case "import":
			err = p.Import(tokens[1])
		case "image":
			err = p.Image(tokens[1])
		case "run":
			err = p.Run(tokens[1:])
		}
		if err != nil {
			log.Print(err)
			return err
		}
		counter++
	}
	return nil
}

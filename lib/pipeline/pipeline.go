// package pipeline implements a system for running data pipelines on top of
// the filesystem
package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

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

// runCommit returns the commit that the current run will create
func (p *Pipeline) runCommit() string {
	return path.Join(p.outRepo, strconv.Itoa(p.counter))
}

// alreadyDone returns true if the current run was committed after the commit in
// dataRepo
func (p *Pipeline) alreadyDone() bool {
	// check if the commit already exists
	exists, err := btrfs.FileExists(p.runCommit())
	if err != nil {
		return false
	}
	if !exists {
		return false
	}
	before, err := btrfs.Before(path.Join(p.dataRepo, p.commit), p.runCommit())
	if err != nil {
		return false
	}
	return before
}

// setupOutRepo creates the output directory for a call to
func (p *Pipeline) setupOutRepo() error {
	// Create the branch
	if p.counter == 0 {
		err := btrfs.Branch(p.outRepo, "", p.branch)
		if err != nil {
			log.Print(err)
			return err
		}
	} else {
		err := btrfs.Branch(p.outRepo, strconv.Itoa(p.counter-1), p.branch)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	// Remove the commit if it exists:
	// TODO map this commit in to the container so it can be used as reference.
	exists, err := btrfs.FileExists(p.runCommit())
	if err != nil {
		log.Print(err)
		return err
	}
	if exists {
		err := btrfs.SubvolumeDelete(p.runCommit())
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

func (p *Pipeline) Run(cmd []string) error {
	defer func() { p.counter++ }()
	if p.alreadyDone() {
		return nil
	}

	// Set the command
	p.containerConfig.Cmd = cmd

	// Make sure the outRepo is ready for our output
	err := p.setupOutRepo()
	if err != nil {
		log.Print(err)
		return err
	}

	// Map the out directory in as a bind
	hostPath := btrfs.HostPath(path.Join(p.outRepo, p.branch))
	bind := fmt.Sprintf("%s:/out", hostPath)
	p.hostConfig.Binds = append(p.hostConfig.Binds, bind)
	// Make sure this bind is only visible for the duration of run
	defer func() { p.hostConfig.Binds = p.hostConfig.Binds[:len(p.hostConfig.Binds)-1] }()

	// Start the container
	containerId, err := container.RawStartContainer(fmt.Sprintf("%d-%s", p.counter, p.name),
		&p.containerConfig, &p.hostConfig)
	if err != nil {
		log.Print(err)
		return err
	}
	// Block for logs from the container
	r, err := container.ContainerLogs(containerId, &dockerclient.LogOptions{Follow: true, Stdout: true, Stderr: true, Timestamps: true})
	if err != nil {
		log.Print(err)
		return err
	}
	// Create a place to put the logs
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

// RunPipelines runs all of the pipelines it finds in pipelineDir. Returns the
// first error it encounters.
func RunPipelines(pipelineDir, dataRepo, outRepo, commit, branch string) error {
	pipelines, err := btrfs.ReadDir(path.Join(dataRepo, commit, pipelineDir))
	if err != nil {
		return err
	}
	// A chanel for the errors, notice that it's capacity is the same as the
	// number of pipelines. The below code should make sure that each pipeline only
	// sends 1 error otherwise deadlock may occur.
	errors := make(chan error, len(pipelines))

	var wg sync.WaitGroup
	wg.Add(len(pipelines))
	for _, pInfo := range pipelines {
		go func(pInfo os.FileInfo) {
			p := NewPipeline(pInfo.Name(), dataRepo, outRepo, commit, branch)
			f, err := btrfs.Open(path.Join(dataRepo, commit, pipelineDir, pInfo.Name()))
			if err != nil {
				log.Print(err)
				errors <- err
				return
			}
			err = p.RunPachFile(f)
			if err != nil {
				log.Print(err)
				errors <- err
				return
			}
		}(pInfo)
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

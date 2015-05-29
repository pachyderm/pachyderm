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
	"strings"
	"sync"

	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/container"
	"github.com/samalba/dockerclient"
)

type Pipeline struct {
	containerConfig   dockerclient.ContainerConfig
	hostConfig        dockerclient.HostConfig
	dataRepo, outRepo string
	commit, branch    string
	counter           int
}

func NewPipeline(dataRepo, outRepo, commit, branch string) *Pipeline {
	return &Pipeline{
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
	return fmt.Sprintf("%s-%d", p.commit, p.counter)
}

// Run runs a command in the container, it assumes that `branch` has already
// been created.
func (p *Pipeline) Run(cmd []string) error {
	// this function always increments counter
	defer func() { p.counter++ }()
	// Check if the commit already exists
	exists, err := btrfs.FileExists(path.Join(p.outRepo, p.runCommit()))
	if err != nil {
		log.Print(err)
		return err
	}
	// if the commit exists there's no work to be done
	if exists {
		return nil
	}
	// Set the command
	p.containerConfig.Cmd = cmd
	// Map the out directory in as a bind
	hostPath := btrfs.HostPath(path.Join(p.outRepo, p.branch))
	bind := fmt.Sprintf("%s:/out", hostPath)
	p.hostConfig.Binds = append(p.hostConfig.Binds, bind)
	// Make sure this bind is only visible for the duration of run
	defer func() { p.hostConfig.Binds = p.hostConfig.Binds[:len(p.hostConfig.Binds)-1] }()
	// Start the container
	containerId, err := container.RawStartContainer(path.Join(p.outRepo, p.commit),
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

	// Commit the results
	err = btrfs.Commit(p.outRepo, p.runCommit(), p.branch)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (p *Pipeline) Finish() error {
	exists, err := btrfs.FileExists(path.Join(p.outRepo, p.commit))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return btrfs.Commit(p.outRepo, p.commit, p.branch)
}

func (p *Pipeline) RunPachFile(r io.Reader) error {
	lines := bufio.NewScanner(r)

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
	}
	err := p.Finish()
	if err != nil {
		return err
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
			defer wg.Done()
			p := NewPipeline(dataRepo, outRepo, commit, branch)
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

// package pipeline implements a system for running data pipelines on top of the filesystem
package pipeline

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/pachyderm/pachyderm/src/s3utils"
	"github.com/pachyderm/pachyderm/src/util"
)

var (
	ErrFailed          = errors.New("pfs: pipeline failed")
	ErrCancelled       = errors.New("pfs: cancelled")
	ErrArgCount        = errors.New("pfs: illegal argument count")
	ErrUnkownKeyword   = errors.New("pfs: unknown keyword")
	ErrUnknownProtocol = errors.New("pfs: unknown protocol")
)

type pipeline struct {
	name        string
	config      docker.CreateContainerOptions
	inRepo      string
	outRepo     string
	commit      string
	branch      string
	counter     int
	container   string
	cancelled   bool
	shard       string
	pipelineDir string
	cache       etcache.Cache
}

func newPipeline(name, dataRepo, outRepo, commit, branch, shard, pipelineDir string, cache etcache.Cache) *pipeline {
	return &pipeline{
		name,
		docker.CreateContainerOptions{Config: &DefaultConfig,
			HostConfig: &docker.HostConfig{}},
		dataRepo,
		outRepo,
		commit,
		branch,
		0,
		"",
		false,
		shard,
		pipelineDir,
		cache,
	}
}

// input makes a dataset available for computations in the container.
func (p *pipeline) input(name string) error {
	var trimmed string
	switch {
	case strings.HasPrefix(name, "s3://"):
		trimmed = strings.TrimPrefix(name, "s3://")
		if err := WaitPipeline(p.pipelineDir, trimmed, p.commit); err != nil {
			return err
		}
		hostPath := btrfs.HostPath(path.Join(p.pipelineDir, trimmed, p.commit))
		containerPath := path.Join("/in", trimmed)
		bind := fmt.Sprintf("%s:%s:ro", hostPath, containerPath)
		p.config.HostConfig.Binds = append(p.config.HostConfig.Binds, bind)
	case strings.HasPrefix(name, "pps://"):
		trimmed = strings.TrimPrefix(name, "pps://")
		err := WaitPipeline(p.pipelineDir, trimmed, p.commit)
		if err != nil {
			return err
		}
		hostPath := btrfs.HostPath(path.Join(p.pipelineDir, trimmed, p.commit))
		containerPath := path.Join("/in", trimmed)
		bind := fmt.Sprintf("%s:%s:ro", hostPath, containerPath)
		p.config.HostConfig.Binds = append(p.config.HostConfig.Binds, bind)
	case strings.HasPrefix(name, "pfs://"):
		fallthrough
	default:
		hostPath := btrfs.HostPath(path.Join(p.inRepo, p.commit, name))
		containerPath := path.Join("/in", name)
		bind := fmt.Sprintf("%s:%s:ro", hostPath, containerPath)
		p.config.HostConfig.Binds = append(p.config.HostConfig.Binds, bind)
	}
	return nil
}

// inject injects data from an external source into the output directory
func (p *pipeline) inject(name string) error {
	switch {
	case strings.HasPrefix(name, "s3://"):
		bucket, err := s3utils.NewBucket(name)
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		s3utils.ForEachFile(name, "", func(file string, modtime time.Time) error {
			// Grab the path, it's handy later
			_path, err := s3utils.GetPath(name)
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}
			// Check if the file belongs on shit shard
			match, err := route.Match(file, p.shard)
			if err != nil {
				return err
			}
			if !match {
				return nil
			}
			// Check if the file has changed
			changed, err := btrfs.Changed(path.Join(p.outRepo, p.branch,
				strings.TrimPrefix(file, _path)), modtime)
			log.Print("Changed: ", changed)
			if err != nil {
				return err
			}
			if !changed {
				return nil
			}
			// TODO match the on disk timestamps to s3's timestamps and make
			// sure we only pull data that has changed
			wg.Add(1)
			go func() {
				defer wg.Done()
				src, err := bucket.GetReader(file)
				if err != nil {
					return
				}
				dst, err := btrfs.CreateAll(path.Join(p.outRepo, p.branch, strings.TrimPrefix(file, _path)))
				if err != nil {
					return
				}
				defer dst.Close()
				_, err = io.Copy(dst, src)
				if err != nil {
					return
				}
				err = btrfs.Chtimes(path.Join(p.outRepo, p.branch, strings.TrimPrefix(file, _path)), modtime, modtime)
				if err != nil {
					return
				}
			}()
			return nil
		})
		wg.Wait()
	default:
		log.Print("Unknown protocol: ", name)
		return ErrUnknownProtocol
	}
	return nil
}

// Image sets the image that is being used for computations.
func (p *pipeline) image(image string) error {
	p.config.Config.Image = image
	// TODO(pedge): ensure images are on machine
	err := pullImage(image)
	if err != nil {
		log.Print("assuming image is local and continuing")
	}
	return nil
}

// Start gets an outRepo ready to be used. This is where clean up of dirty
// state from a crash happens.
func (p *pipeline) start() error {
	//TODO cleanup crashed state here.
	return nil
}

// runCommit returns the commit that the current run will create
func (p *pipeline) runCommit() string {
	return fmt.Sprintf("%s-%d", p.commit, p.counter)
}

// Run runs a command in the container, it assumes that `branch` has already
// been created.
// Notice that any failure in this function leads to the branch having
// uncommitted dirty changes. This state needs to be cleaned up before the
// pipeline is rerun. The reason we don't do it here is that even if we try our
// best the process crashing at the wrong time could still leave it in an
// inconsistent state.
func (p *pipeline) run(cmd []string) error {
	// this function always increments counter
	defer func() { p.counter++ }()
	// Check if the commit already exists
	exists, err := btrfs.FileExists(path.Join(p.outRepo, p.runCommit()))
	if err != nil {
		return err
	}
	// if the commit exists there's no work to be done
	if exists {
		return nil
	}
	// Set the command
	p.config.Config.Cmd = []string{"sh"}
	// Map the out directory in as a bind
	hostPath := btrfs.HostPath(path.Join(p.outRepo, p.branch))
	bind := fmt.Sprintf("%s:/out", hostPath)
	p.config.HostConfig.Binds = append(p.config.HostConfig.Binds, bind)
	log.Print(p.config.HostConfig.Binds)
	// Make sure this bind is only visible for the duration of run
	defer func() { p.config.HostConfig.Binds = p.config.HostConfig.Binds[:len(p.config.HostConfig.Binds)-1] }()
	// Start the container
	p.container, err = startContainer(p.config)
	if err != nil {
		return err
	}
	if err := pipeToStdin(p.container, strings.NewReader(strings.Join(cmd, " ")+"\n")); err != nil {
		return err
	}
	// Create a place to put the logs
	f, err := btrfs.CreateAll(path.Join(p.outRepo, p.branch, ".log"))
	if err != nil {
		return err
	}
	defer f.Close()
	// Copy the logs from the container in to the file.
	if err = containerLogs(p.container, f); err != nil {
		return err
	}
	// Wait for the command to finish:
	exit, err := waitContainer(p.container)
	if err != nil {
		return err
	}
	if exit != 0 {
		// The command errored
		return fmt.Errorf("Command:\n\t%s\nhad exit code: %d.\n",
			strings.Join(cmd, " "), exit)
	}
	return btrfs.Commit(p.outRepo, p.runCommit(), p.branch)
}

// Shuffle rehashes an output directory.
// If 2 shards each have a copy of the file `foo` with the content: `bar`.
// Then after shuffling 1 of those nodes will have a file `foo` with content
// `barbar` and the other will have no file foo.
func (p *pipeline) shuffle(dir string) error {
	// this function always increments counter
	defer func() { p.counter++ }()
	// First we clear the directory, notice that the previous commit from
	// which we're pulling has already been made so this doesn't destroy the
	// data that others are trying to pull.
	// TODO(jd) #performance this is a seriously unperformant part of the code
	// since it messes up our ability to do incremental results. We should do
	// something smarter here.
	if err := btrfs.RemoveAll(path.Join(p.outRepo, p.branch, dir)); err != nil {
		return err
	}
	if err := btrfs.MkdirAll(path.Join(p.outRepo, p.branch, dir)); err != nil {
		return err
	}
	// We want to pull files from the previous commit
	commit := fmt.Sprintf("%s-%d", p.commit, p.counter-1)
	// Notice we're just passing "host" here. Multicast will fill in the host
	// field so we don't actually need to specify it.
	req, err := http.NewRequest("GET", "http://host/"+path.Join("pipeline", p.name, "file", dir, "*")+"?commit="+commit+"&shard="+p.shard, nil)
	if err != nil {
		return err
	}
	// Dispatch the request
	resps, err := route.Multicast(p.cache, req, "/pfs/master")
	if err != nil {
		return err
	}

	// Set up some concurrency structures.
	errors := make(chan error, len(resps))
	var wg sync.WaitGroup
	wg.Add(len(resps))
	lock := util.NewPathLock()
	// for _, resp := range resps {
	// We used to iterate like the above but it exhibited racy behavior. I
	// don't fully understand why this was. Something to look in to.
	for _, resp := range resps {
		go func(resp *http.Response) {
			defer wg.Done()
			reader := multipart.NewReader(resp.Body, resp.Header.Get("Boundary"))

			for part, err := reader.NextPart(); err != io.EOF; part, err = reader.NextPart() {
				lock.Lock(part.FileName())
				_, err := btrfs.Append(path.Join(p.outRepo, p.branch, part.FileName()), part)
				lock.Unlock(part.FileName())
				if err != nil {
					errors <- err
					return
				}
			}
		}(resp)
	}
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}
	return btrfs.Commit(p.outRepo, p.runCommit(), p.branch)
}

// finish makes the final commit for the pipeline
func (p *pipeline) finish() error {
	exists, err := btrfs.FileExists(path.Join(p.outRepo, p.commit))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return btrfs.Commit(p.outRepo, p.commit, p.branch)
}

func (p *pipeline) finished() (bool, error) {
	return btrfs.FileExists(path.Join(p.outRepo, p.commit))
}

func (p *pipeline) fail() error {
	exists, err := btrfs.FileExists(path.Join(p.outRepo, p.commit+"-fail"))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return btrfs.DanglingCommit(p.outRepo, p.commit+"-fail", p.branch)
}

// Cancel stops a pipeline by force before it's finished
func (p *pipeline) cancel() error {
	p.cancelled = true
	return stopContainer(p.container)
}

// runPachFile parses r as a PachFile and executes. runPachFile is GUARANTEED
// to call either `finish` or `fail`
func (p *pipeline) runPachFile(r io.Reader) (retErr error) {
	defer func() {
		// This function GUARANTEES that if `p.finish()` didn't happen.
		// `p.fail()` does.
		finished, err := p.finished()
		if err != nil {
			if retErr == nil {
				retErr = err
			}
			return
		}
		if finished {
			return
		}
		if err := p.fail(); err != nil {
			if retErr == nil {
				retErr = err
			}
			return
		}
	}()
	lines := bufio.NewScanner(r)

	if err := p.start(); err != nil {
		return err
	}
	var tokens []string
	for lines.Scan() {
		if p.cancelled {
			return ErrCancelled
		}
		if len(tokens) > 0 && tokens[len(tokens)-1] == "\\" {
			// We have tokens from last loop, remove the \ token which designates the line wrap
			tokens = tokens[:len(tokens)-1]
		} else {
			// No line wrap, clear the tokens they were already execuated
			tokens = []string{}
		}
		tokens = append(tokens, strings.Fields(lines.Text())...)
		// These conditions are, empty line, comment line and wrapped line.
		// All 3 cause us to continue, the first 2 because we're skipping them.
		// The last because we need more input.
		if len(tokens) == 0 || tokens[0][0] == '#' || tokens[len(tokens)-1] == "\\" {
			// Comment or empty line, skip
			continue
		}

		var err error
		switch strings.ToLower(tokens[0]) {
		case "input":
			if len(tokens) != 2 {
				return ErrArgCount
			}
			err = p.input(tokens[1])
		case "image":
			if len(tokens) != 2 {
				return ErrArgCount
			}
			err = p.image(tokens[1])
		case "run":
			if len(tokens) < 2 {
				return ErrArgCount
			}
			err = p.run(tokens[1:])
		case "shuffle":
			if len(tokens) != 2 {
				return ErrArgCount
			}
			err = p.shuffle(tokens[1])
		default:
			return ErrUnkownKeyword
		}
		if err != nil {
			log.Print(err)
			return err
		}
	}
	if err := p.finish(); err != nil {
		return err
	}
	return nil
}

type Runner struct {
	pipelineDir string
	inRepo      string
	commit      string
	branch      string
	outPrefix   string // the prefix for out repos
	shard       string
	pipelines   []*pipeline
	wait        sync.WaitGroup
	lock        sync.Mutex // used to prevent races between `Run` and `Cancel`
	cancelled   bool
	cache       etcache.Cache
}

func NewRunner(pipelineDir string, inRepo string, outPrefix string, commit string, branch string, shard string, cache etcache.Cache) *Runner {
	return &Runner{
		pipelineDir: pipelineDir,
		inRepo:      inRepo,
		commit:      commit,
		branch:      branch,
		outPrefix:   outPrefix,
		shard:       shard,
		cache:       cache,
	}
}

func (r *Runner) makeOutRepo(pipeline string) error {
	if err := btrfs.Ensure(path.Join(r.outPrefix, pipeline)); err != nil {
		return err
	}

	exists, err := btrfs.FileExists(path.Join(r.outPrefix, pipeline, r.branch))
	if err != nil {
		return err
	}
	if !exists {
		// The branch doesn't exist, we need to create it We'll make our branch
		// have the same parent as the commit we're running off of if that
		// parent exists in the pipelines outRepo. This lets us carry over past
		// computation results when a new branch is created rather than having
		// to start from scratch.
		parent := btrfs.GetMeta(path.Join(r.inRepo, r.commit), "parent")
		if parent != "" {
			exists, err := btrfs.FileExists(path.Join(r.outPrefix, pipeline, parent))
			if err != nil {
				return err
			}
			if !exists {
				parent = ""
			}
		}
		if err := btrfs.Branch(path.Join(r.outPrefix, pipeline), parent, r.branch); err != nil {
			return err
		}
	}
	// The branch exists, so we're ready to return
	return nil
}

// Run runs all of the pipelines it finds in pipelineDir. Returns the
// first error it encounters.
func (r *Runner) Run() error {
	if err := btrfs.MkdirAll(r.outPrefix); err != nil {
		return err
	}
	if err := r.startInputPipelines(); err != nil {
		return err
	}
	pipelines, err := btrfs.ReadDir(path.Join(r.inRepo, r.commit, r.pipelineDir))
	if err != nil {
		// Notice we don't return this error but instead no-op. It's fine to not
		// have a pipeline dir.
		return nil
	}
	// A chanel for the errors, notice that it's capacity is the same as the
	// number of pipelines. The below code should make sure that each pipeline only
	// sends 1 error otherwise deadlock may occur.
	errors := make(chan error, len(pipelines))
	// Make sure we don't race with cancel this is held while we add pipelines.
	r.lock.Lock()
	if r.cancelled {
		// we were cancelled before we even started
		r.lock.Unlock()
		return ErrCancelled
	}
	// unlocker lets us defer unlocking and explicitly unlock
	var unlocker sync.Once
	defer unlocker.Do(r.lock.Unlock)
	r.wait.Add(len(pipelines))
	for _, pInfo := range pipelines {
		if err := r.makeOutRepo(pInfo.Name()); err != nil {
			return err
		}
		p := newPipeline(pInfo.Name(), r.inRepo, path.Join(r.outPrefix, pInfo.Name()), r.commit, r.branch, r.shard, r.outPrefix, r.cache)
		r.pipelines = append(r.pipelines, p)
		go func(pInfo os.FileInfo, p *pipeline) {
			defer r.wait.Done()
			f, err := btrfs.Open(path.Join(r.inRepo, r.commit, r.pipelineDir, pInfo.Name()))
			if err != nil {
				errors <- err
				return
			}
			defer f.Close()
			err = p.runPachFile(f)
			if err != nil {
				errors <- err
				return
			}
		}(pInfo, p)
	}
	// We're done adding pipelines so unlock
	unlocker.Do(r.lock.Unlock)
	// Wait for the pipelines to finish
	r.wait.Wait()
	close(errors)
	if r.cancelled {
		// Pipelines finished because we were cancelled
		return ErrCancelled
	}
	for err := range errors {
		return err
	}
	return nil
}

// RunPipelines lets you easily run the Pipelines in one line if you don't care about cancelling them.
func RunPipelines(pipelineDir string, inRepo string, outRepo string, commit string, branch string, shard string, cache etcache.Cache) error {
	return NewRunner(pipelineDir, inRepo, outRepo, commit, branch, shard, cache).Run()
}

func (r *Runner) Cancel() error {
	log.Print("Cancel: ", r)
	// A chanel for the errors, notice that it's capacity is the same as the
	// number of pipelines. The below code should make sure that each pipeline only
	// sends 1 error otherwise deadlock may occur.
	errors := make(chan error, len(r.pipelines))

	// Make sure we don't race with Run
	r.lock.Lock()
	// Indicate that we're cancelling the pipelines
	r.cancelled = true
	// A waitgroup for the goros that cancel the containers
	var wg sync.WaitGroup
	// We'll have one goro per pipelines
	wg.Add(len(r.pipelines))
	for _, p := range r.pipelines {
		go func(p *pipeline) {
			defer wg.Done()
			err := p.cancel()
			if err != nil {
				errors <- err
			}
		}(p)
	}
	// Wait for the cancellations to finish.
	wg.Wait()
	r.lock.Unlock()
	close(errors)
	for err := range errors {
		return err
	}
	// At the end we wait for the pipelines to actually finish, this means that
	// once Cancel is done you can safely fire off a new batch of pipelines.
	r.wait.Wait()
	return nil
}

// Inputs returns all of the inputs for the pipelines.
func (r *Runner) Inputs() ([]string, error) {
	pipelines, err := btrfs.ReadDir(path.Join(r.inRepo, r.commit, r.pipelineDir))
	if err != nil {
		// Notice we don't return this error but instead no-op. It's fine to not
		// have a pipeline dir.
		return nil, nil
	}
	var res []string
	for _, pInfo := range pipelines {
		f, err := btrfs.Open(path.Join(r.inRepo, r.commit, r.pipelineDir, pInfo.Name()))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		lines := bufio.NewScanner(f)
		// TODO we're copy-pasting code from runPachFile. Let's abstract that.
		var tokens []string
		for lines.Scan() {
			if len(tokens) > 0 && tokens[len(tokens)-1] == "\\" {
				// We have tokens from last loop, remove the \ token which designates the line wrap
				tokens = tokens[:len(tokens)-1]
			} else {
				// No line wrap, clear the tokens they were already considered
				tokens = []string{}
			}
			tokens = append(tokens, strings.Fields(lines.Text())...)
			if len(tokens) > 0 && tokens[0] == "input" {
				if len(tokens) < 2 {
					return nil, ErrArgCount
				}
				res = append(res, tokens[1])
			}
		}
	}
	return res, nil
}

// startInputPipelines starts pipelines which pull in external data (such as
// from s3)
func (r *Runner) startInputPipelines() error {
	inputs, err := r.Inputs()
	if err != nil {
		return err
	}
	r.lock.Lock()
	if r.cancelled {
		// we were cancelled before we even started
		r.lock.Unlock()
		return ErrCancelled
	}
	defer r.lock.Unlock()
	for _, input := range inputs {
		if !strings.HasPrefix(input, "s3://") {
			continue
		}
		trimmed := strings.TrimPrefix(input, "s3://")
		if err := r.makeOutRepo(trimmed); err != nil {
			return err
		}
		p := newPipeline(trimmed, r.inRepo, path.Join(r.outPrefix, trimmed), r.commit, r.branch, r.shard, r.outPrefix, r.cache)
		r.pipelines = append(r.pipelines, p)
		r.wait.Add(1)
		go func(input string, p *pipeline) {
			defer r.wait.Done()
			err := p.inject(input)
			if err != nil {
				p.fail()
				return
			}
			p.finish()
		}(input, p)
	}
	return nil
}

// WaitPipeline waits for a pipeline to complete. If the pipeline fails
// ErrFailed is returned.
func WaitPipeline(pipelineDir, pipeline, commit string) error {
	success := path.Join(pipelineDir, pipeline, commit)
	failure := path.Join(pipelineDir, pipeline, commit+"-fail")
	file, err := btrfs.WaitAnyFile(success, failure)
	if err != nil {
		return err
	}
	if file == failure {
		return ErrFailed
	}
	return nil
}

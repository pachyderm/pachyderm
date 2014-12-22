package mapreduce

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/pachyderm-io/pfs/lib/btrfs"
	"github.com/pachyderm-io/pfs/lib/route"
	"github.com/samalba/dockerclient"
)

var retries int = 5

// StartContainer pulls image and starts a container from it with command. It
// returns the container id or an error.
func spinupContainer(image string, command []string) (string, error) {
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		return "", err
	}
	if err := docker.PullImage(image, nil); err != nil {
		return "", err
	}

	containerConfig := &dockerclient.ContainerConfig{Image: image, Cmd: command}

	containerId, err := docker.CreateContainer(containerConfig, "")
	if err != nil {
		return "", nil
	}

	if err := docker.StartContainer(containerId, &dockerclient.HostConfig{}); err != nil {
		return "", err
	}

	return containerId, nil
}

func ipAddr(containerId string) (string, error) {
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		return "", err
	}
	containerInfo, err := docker.InspectContainer(containerId)
	if err != nil {
		return "", err
	}

	return containerInfo.NetworkSettings.IpAddress, nil
}

func retry(f func() error, retries int, pause time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		err = f()
		if err == nil {
			break
		} else {
			time.Sleep(pause)
		}
	}
	return err
}

// contains checks if set contains val. It assums that set has already been
// sorted.
func contains(set []string, val string) bool {
	index := sort.SearchStrings(set, val)
	return index < len(set) && set[index] == val
}

type Job struct {
	Type      string   `json:"type"`
	Input     string   `json:"input"`
	Container string   `json:"container"`
	Command   []string `json:"command"`
}

type materializeInfo struct {
	In, Out, Branch, Commit string
}

func PrepJob(job Job, jobPath string, m materializeInfo) error {
	if err := btrfs.MkdirAll(path.Join(m.Out, m.Branch, jobPath)); err != nil {
		return err
	}
	return nil
}

func Map(job Job, jobPath string, m materializeInfo, host string) error {
	err := PrepJob(job, path.Base(jobPath), m)
	if err != nil {
		return err
	}

	if job.Type != "map" {
		return fmt.Errorf("runMap called on a job of type \"%s\". Should be \"map\".", job.Type)
	}

	inFiles, err := btrfs.ReadDir(path.Join(m.In, m.Commit, job.Input))
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	defer wg.Wait()
	for _, inF := range inFiles {
		wg.Add(1)
		go func(inF os.FileInfo) {
			defer wg.Done()
			inFile, err := btrfs.Open(path.Join(m.In, m.Commit, job.Input, inF.Name()))
			if err != nil {
				log.Print(err)
				return
			}
			defer inFile.Close()

			var resp *http.Response
			err = retry(func() error {
				log.Print("Posting: ", inF.Name())
				resp, err = http.Post("http://"+path.Join(host, inF.Name()), "application/text", inFile)
				return err
			}, 5, 200*time.Millisecond)
			if err != nil {
				log.Print(err)
				return
			}
			defer resp.Body.Close()

			outFile, err := btrfs.Create(path.Join(m.Out, m.Branch, jobPath, inF.Name()))
			if err != nil {
				log.Print(err)
				return
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, resp.Body); err != nil {
				log.Print(err)
				return
			}
		}(inF)
	}
	return nil
}

func Reduce(job Job, jobPath string, m materializeInfo, host string) error {
	log.Print("Reduce: ", job, " ", jobPath, " ")
	if job.Type != "reduce" {
		return fmt.Errorf("runMap called on a job of type \"%s\". Should be \"reduce\".", job.Type)
	}

	// Notice we're just passing "host" here. Multicast will fill in the host
	// field so we don't actually need to specify it.
	req, err := http.NewRequest("GET", "http://host/"+job.Input, nil)
	reader, err := route.Multicast(req, "/pfs/master")
	defer reader.Close()

	var resp *http.Response
	err = retry(func() error {
		resp, err = http.Post("http://"+path.Join(host, job.Input), "application/text", reader)
		return err
	}, 5, 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outFile, err := btrfs.Create(path.Join(m.Out, m.Branch, jobPath))
	if err != nil {
		return err
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return err
	}
	return nil
}

// Materialize parses the jobs found in `in_repo`/`commit`/`jobDir` runs them
// with `in_repo/commit` as input, outputs the results to `out_repo`/`branch`
// and commits them as `out_repo`/`commit`
func Materialize(in_repo, branch, commit, out_repo, jobDir string) error {
	log.Printf("Materialize: %s %s %s %s %s.", in_repo, branch, commit, out_repo, jobDir)
	// We make sure that this function always commits so that we know the comp
	// repo stays in sync with the data repo.
	defer func() {
		if err := btrfs.Commit(out_repo, commit, branch); err != nil {
			log.Print("btrfs.Commit error in Materialize: ", err)
		}
	}()
	// First check if the jobs dir actually exists.
	exists, err := btrfs.FileExists(path.Join(in_repo, commit, jobDir))
	if err != nil {
		return err
	}
	if !exists {
		// Perfectly valid to have no jobs dir, it just means we have no work
		// to do.
		return nil
	}

	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		return err
	}
	newFiles, err := btrfs.NewFiles(in_repo, commit)
	if err != nil {
		return err
	}
	sort.Strings(newFiles)

	jobsPath := path.Join(in_repo, commit, jobDir)
	jobs, err := btrfs.ReadDir(jobsPath)
	if err != nil {
		return err
	}
	for _, jobInfo := range jobs {
		jobFile, err := btrfs.Open(path.Join(jobsPath, jobInfo.Name()))
		if err != nil {
			log.Print(err)
			continue
		}
		defer jobFile.Close()
		decoder := json.NewDecoder(jobFile)
		job := Job{}
		if err = decoder.Decode(&job); err != nil {
			log.Print(err)
			continue
		}
		log.Print("Job: ", job)
		m := materializeInfo{in_repo, out_repo, branch, commit}

		containerId, err := spinupContainer(job.Container, job.Command)
		if err != nil {
			log.Print(err)
			continue
		}
		defer docker.StopContainer(containerId, 5)

		containerAddr, err := ipAddr(containerId)
		if err != nil {
			log.Print(err)
			continue
		}

		if job.Type == "map" {
			err := Map(job, jobInfo.Name(), m, containerAddr)
			if err != nil {
				log.Print(err)
				continue
			}
		} else if job.Type == "reduce" {
			err := Reduce(job, jobInfo.Name(), m, containerAddr)
			if err != nil {
				log.Print(err)
				continue
			}
		} else {
			log.Printf("Job %s has unrecognized type: %s.", jobInfo.Name(), job.Type)
			continue
		}
	}
	return nil
}

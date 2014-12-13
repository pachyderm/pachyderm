package mapreduce

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/pachyderm-io/pfs/lib/btrfs"
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

type Job struct {
	Input     string   `json:"input"`
	Container string   `json:"container"`
	Command   []string `json:"command"`
}

// Materialize parses the jobs found in `in_repo`/`commit`/`jobDir` runs them
// with `in_repo/commit` as input, outputs the results to `out_repo`/`branch`
// and commits them as `out_repo`/`commit`
func Materialize(in_repo, branch, commit, out_repo, jobDir string) error {
	log.Printf("Materialize: %s %s %s %s %s.", in_repo, branch, commit, out_repo, jobDir)
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		return err
	}
	exists, err := btrfs.FileExists(path.Join(in_repo, commit, jobDir))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	jobsPath := path.Join(in_repo, commit, jobDir)
	jobs, err := btrfs.ReadDir(jobsPath)
	if err != nil {
		return err
	}
	for _, jobInfo := range jobs {
		jobFile, err := btrfs.Open(path.Join(jobsPath, jobInfo.Name()))
		if err != nil {
			return err
		}
		defer jobFile.Close()
		decoder := json.NewDecoder(jobFile)
		j := &Job{}
		if err = decoder.Decode(j); err != nil {
			return err
		}

		containerId, err := spinupContainer(j.Container, j.Command)
		if err != nil {
			return err
		}
		defer docker.StopContainer(containerId, 5)

		containerAddr, err := ipAddr(containerId)
		if err != nil {
			return err
		}

		inFiles, err := btrfs.ReadDir(path.Join(in_repo, commit, j.Input))
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		defer wg.Wait()
		for _, inF := range inFiles {
			wg.Add(1)
			go func() {
				defer wg.Done()
				inFile, err := btrfs.Open(path.Join(in_repo, commit, j.Input, inF.Name()))
				if err != nil {
					log.Print(err)
					return
				}
				defer inFile.Close()

				var resp *http.Response
				err = retry(func() error {
					resp, err = http.Post("http://"+path.Join(containerAddr, inF.Name()), "application/text", inFile)
					return err
				}, 5, 200*time.Millisecond)
				if err != nil {
					log.Print(err)
					return
				}
				defer resp.Body.Close()

				exists, err := btrfs.FileExists(path.Join(out_repo, branch))
				if err != nil {
					log.Print(err)
					return
				}
				if !exists {
					log.Printf("Invalid state. %s should already exists.", path.Join(out_repo, branch))
					return
				} else {
					if err := btrfs.MkdirAll(path.Join(out_repo, branch, jobInfo.Name())); err != nil {
						log.Print(err)
						return
					}
				}

				outFile, err := btrfs.Create(path.Join(out_repo, branch, jobInfo.Name(), inF.Name()))
				if err != nil {
					log.Print(err)
					return
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, resp.Body); err != nil {
					log.Print(err)
					return
				}
			}()
		}
	}

	if err := btrfs.Commit(out_repo, commit, branch); err != nil {
		return err
	}

	return nil
}

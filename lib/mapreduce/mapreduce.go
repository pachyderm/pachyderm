package mapreduce

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"testing/iotest"

	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/route"
	"github.com/pachyderm/pfs/lib/shell"
	"github.com/samalba/dockerclient"
)

var retries int = 5

/* This is a terrible hack, but dockerclient doesn't expose a `Build` method
* and it doesn't make it easy to extend it so I had to copy paste this piece of
* code. Please for the love of god remove this as soon as possible. */
func doRequest(client *dockerclient.DockerClient, method string, path string, r io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, client.URL.String()+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for header, value := range headers {
			req.Header.Add(header, value)
		}
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "connection refused") && client.TLSConfig == nil {
			return nil, fmt.Errorf("%v. Are you trying to connect to a TLS-enabled daemon without TLS?", err)
		}
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, dockerclient.ErrNotFound
	}
	if resp.StatusCode >= 400 {
		log.Print(data)
		return nil, dockerclient.Error{StatusCode: resp.StatusCode, Status: resp.Status}
	}
	return data, nil
}

func startContainer(image string, command []string) (string, error) {
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	containerConfig := &dockerclient.ContainerConfig{Image: image, Cmd: command}

	containerId, err := docker.CreateContainer(containerConfig, "")
	if err != nil {
		log.Print(err)
		return "", nil
	}

	if err := docker.StartContainer(containerId, &dockerclient.HostConfig{}); err != nil {
		log.Print(err)
		return "", err
	}

	return containerId, nil
}

// spinupContainer pulls image and starts a container from it with command. It
// returns the container id or an error.
func spinupContainer(image string, command []string) (string, error) {
	log.Print("spinupContainer", " ", image, " ", command)
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Print(err)
		return "", err
	}
	if err := docker.PullImage(image, nil); err != nil {
		//return "", err this is erroring due to failing to parse response json
	}

	return startContainer(image, command)
}

func spinupContainerFromGit(url string, command []string) (string, error) {
	log.Print("spinupContainerFromGit(", url, ", ", command)
	name := path.Join("/tmp", btrfs.RandSeq(5))
	tag := btrfs.RandSeq(5)
	c := exec.Command("git", "clone", url, name, "--depth", "1")
	if err := shell.RunStderr(c); err != nil {
		return "", err
	}
	defer os.RemoveAll(name)

	c = exec.Command("git", "archive", path.Join(name, "image"))
	err := shell.CallCont(c, func(r io.ReadCloser) error {
		docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
		uri := fmt.Sprintf("/build?t=%s", tag)

		res, err := doRequest(docker, "POST", uri, r, nil)
		log.Print("Response to build:\n", res)
		return err
	})
	if err != nil {
		return "", err
	}

	return startContainer(tag, command)
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
	Type    string   `json:"type"`
	Input   string   `json:"input"`
	Image   string   `json:"image"`
	Repo    string   `json:"repo"`
	Command []string `json:"command"`
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

	files := make(chan string)

	var wg sync.WaitGroup
	defer wg.Wait()

	defer close(files)
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for name := range files {
				inFile, err := btrfs.Open(path.Join(m.In, m.Commit, job.Input, name))
				if err != nil {
					log.Print(err)
					return
				}
				defer inFile.Close()

				var resp *http.Response
				err = retry(func() error {
					log.Print("Posting: ", name)
					resp, err = http.Post("http://"+path.Join(host, name), "application/text", inFile)
					return err
				}, retries, 200*time.Millisecond)
				if err != nil {
					log.Print(err)
					return
				}
				defer resp.Body.Close()

				outFile, err := btrfs.Create(path.Join(m.Out, m.Branch, jobPath, name))
				if err != nil {
					log.Print(err)
					return
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, resp.Body); err != nil {
					log.Print(err)
					return
				}
			}
		}(&wg)
	}
	err = btrfs.LazyWalk(path.Join(m.In, m.Commit, job.Input),
		func(name string) error {
			files <- name
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}

func Reduce(job Job, jobPath string, m materializeInfo, host string, shard, modulos uint64) error {
	if (route.HashResource(path.Join("/job", jobPath)) % modulos) != shard {
		// This resource isn't supposed to be located on this machine.
		return nil
	}
	log.Print("Reduce: ", job, " ", jobPath, " ")
	if job.Type != "reduce" {
		return fmt.Errorf("Reduce called on a job of type \"%s\". Should be \"reduce\".", job.Type)
	}

	// Notice we're just passing "host" here. Multicast will fill in the host
	// field so we don't actually need to specify it.
	var _reader io.ReadCloser
	err := retry(func() error {
		req, err := http.NewRequest("GET", "http://host/"+path.Join(job.Input, "file", "*")+"?commit="+m.Commit, nil)
		if err != nil {
			return err
		}
		_reader, err = route.Multicast(req, "/pfs/master")
		return err
	}, retries, time.Minute)
	if err != nil {
		return err
	}
	defer _reader.Close()
	reader := iotest.NewReadLogger("Reduce", _reader)

	var resp *http.Response
	err = retry(func() error {
		resp, err = http.Post("http://"+path.Join(host, job.Input), "application/text", reader)
		return err
	}, retries, 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Received error %s", resp.Status)
	}

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

type jobCond struct {
	sync.Cond
	Done bool
}

var jobs map[string]*jobCond = make(map[string]*jobCond)
var jobsAccess sync.Mutex

// jobCond returns the name of the condition variable for job.
func condKey(in_repo, commit, job string) string {
	return path.Join(in_repo, commit, job)
}

func ensureCond(name string) {
	jobsAccess.Lock()
	defer jobsAccess.Unlock()
	if _, ok := jobs[name]; !ok {
		jobs[name] = &jobCond{sync.Cond{L: &sync.Mutex{}}, false}
	}
}

func broadcast(in_repo, commit, job string) {
	name := condKey(in_repo, commit, job)
	ensureCond(name)
	jobs[name].L.Lock()
	jobs[name].Done = true
	jobs[name].Broadcast()
	jobs[name].L.Unlock()
}

func WaitJob(in_repo, commit, job string) {
	name := condKey(in_repo, commit, job)
	ensureCond(name)
	jobs[name].L.Lock()
	for !jobs[name].Done {
		jobs[name].Wait()
	}
	jobs[name].L.Unlock()
}

// Materialize parses the jobs found in `in_repo`/`commit`/`jobDir` runs them
// with `in_repo/commit` as input, outputs the results to `out_repo`/`branch`
// and commits them as `out_repo`/`commit`
func Materialize(in_repo, branch, commit, out_repo, jobDir string, shard, modulos uint64) error {
	log.Printf("Materialize: %s %s %s %s %s.", in_repo, branch, commit, out_repo, jobDir)
	exists, err := btrfs.FileExists(path.Join(out_repo, branch))
	if err != nil {
		log.Print(err)
		return err
	}
	if !exists {
		if err := btrfs.Branch(out_repo, "t0", branch); err != nil {
			log.Print(err)
			return err
		}
	}
	// We make sure that this function always commits so that we know the comp
	// repo stays in sync with the data repo.
	defer func() {
		if err := btrfs.Commit(out_repo, commit, branch); err != nil {
			log.Print("DEFERED: btrfs.Commit error in Materialize: ", err)
		}
	}()
	// First check if the jobs dir actually exists.
	exists, err = btrfs.FileExists(path.Join(in_repo, commit, jobDir))
	if err != nil {
		log.Print(err)
		return err
	}
	if !exists {
		// Perfectly valid to have no jobs dir, it just means we have no work
		// to do.
		log.Printf("Jobs dir doesn't exists:\n", path.Join(in_repo, commit, jobDir))
		return nil
	}

	log.Print("Make docker client.")
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Print(err)
		return err
	}
	//log.Print("Find new files.")
	//newFiles, err := btrfs.NewFiles(in_repo, commit)
	//if err != nil {
	//	log.Print(err)
	//	return err
	//}
	//sort.Strings(newFiles)

	jobsPath := path.Join(in_repo, commit, jobDir)
	jobs, err := btrfs.ReadDir(jobsPath)
	if err != nil {
		return err
	}
	log.Print("Jobs: ", jobs)
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, jobInfo := range jobs {
		wg.Add(1)
		go func(jobInfo os.FileInfo) {
			log.Print("Running job:\n", jobInfo.Name())
			defer wg.Done()
			defer broadcast(in_repo, commit, jobInfo.Name())
			jobFile, err := btrfs.Open(path.Join(jobsPath, jobInfo.Name()))
			if err != nil {
				log.Print(err)
				return
			}
			defer jobFile.Close()
			decoder := json.NewDecoder(jobFile)
			job := Job{}
			if err = decoder.Decode(&job); err != nil {
				log.Print(err)
				return
			}
			log.Print("Job: ", job)
			m := materializeInfo{in_repo, out_repo, branch, commit}

			var containerId string
			if job.Image != "" {
				containerId, err = spinupContainer(job.Image, job.Command)
			} else {
				containerId, err = spinupContainerFromGit(job.Repo, job.Command)
			}
			if err != nil {
				log.Print(err)
				return
			}
			defer docker.StopContainer(containerId, 5)

			containerAddr, err := ipAddr(containerId)
			if err != nil {
				log.Print(err)
				return
			}

			if job.Type == "map" {
				err := Map(job, jobInfo.Name(), m, containerAddr)
				if err != nil {
					log.Print(err)
					return
				}
			} else if job.Type == "reduce" {
				err := Reduce(job, jobInfo.Name(), m, containerAddr, shard, modulos)
				if err != nil {
					log.Print(err)
					return
				}
			} else {
				log.Printf("Job %s has unrecognized type: %s.", jobInfo.Name(), job.Type)
				return
			}
		}(jobInfo)
	}
	return nil
}

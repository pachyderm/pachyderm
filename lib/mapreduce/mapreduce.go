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
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/route"
	"github.com/samalba/dockerclient"
)

var retries int = 5

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
	Type     string   `json:"type"`
	Input    string   `json:"input"`
	Image    string   `json:"image"`
	Command  []string `json:"command"`
	Limit    int      `json:"limit"`
	Parallel int      `json:"parallel"`
	TimeOut  int      `json:"timeout"`
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

const (
	ProtoPfs = iota
	ProtoS3  = iota
)

// getProtocol extracts the protocol for an input
func getProtocol(input string) int {
	if strings.TrimPrefix(input, "s3://") != input {
		return ProtoS3
	} else {
		return ProtoPfs
	}
}

// An s3 input looks like: s3://bucket/dir
// Where dir can be a path

// getBucket extracts the bucket from an s3 input
func getBucket(input string) (string, error) {
	if getProtocol(input) != ProtoS3 {
		return "", fmt.Errorf("Input string: %s must begin with 's3://'.", input)
	}
	return strings.Split(strings.TrimPrefix(input, "s3://"), "/")[0], nil
}

// getPath extracts the path from an s3 input
func getPath(input string) (string, error) {
	if getProtocol(input) != ProtoS3 {
		return "", fmt.Errorf("Input string: %s must begin with 's3://'.", input)
	}
	return path.Join(strings.Split(strings.TrimPrefix(input, "s3://"), "/")[1:]...), nil
}

func Map(job Job, jobPath string, m materializeInfo, host string, shard, modulos uint64) {
	err := PrepJob(job, path.Base(jobPath), m)
	if err != nil {
		log.Print(err)
		return
	}

	if job.Type != "map" {
		log.Printf("runMap called on a job of type \"%s\". Should be \"map\".", job.Type)
		return
	}

	files := make(chan string)

	var bucket *s3.Bucket
	if getProtocol(job.Input) == ProtoS3 {
		auth, err := aws.EnvAuth()
		if err != nil {
			log.Print(err)
			return
		}
		client := s3.New(auth, aws.USWest)

		bucketName, err := getBucket(job.Input)
		if err != nil {
			log.Print(err)
			return
		}
		bucket = client.Bucket(bucketName)
	}

	var wg sync.WaitGroup
	defer wg.Wait()
	client := &http.Client{}

	nGoros := 300
	if job.Parallel > 0 {
		nGoros = job.Parallel
	}

	defer close(files)
	for i := 0; i < nGoros; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for name := range files {
				var inFile io.ReadCloser
				var err error
				switch {
				case getProtocol(job.Input) == ProtoPfs:
					inFile, err = btrfs.Open(path.Join(m.In, m.Commit, job.Input, name))
				case getProtocol(job.Input) == ProtoS3:
					inFile, err = bucket.GetReader(name)
				default:
					log.Print("It shouldn't be possible to get here.")
					continue
				}
				if err != nil {
					log.Print(err)
					return
				}
				defer inFile.Close()

				var resp *http.Response
				err = retry(func() error {
					log.Print(name, ": ", "Posting: ", "http://"+path.Join(host, name))
					resp, err = client.Post("http://"+path.Join(host, name), "application/text", inFile)
					log.Print(name, ": ", "Post done.")
					return err
				}, retries, 200*time.Millisecond)
				if err != nil {
					log.Print(err)
					return
				}
				defer resp.Body.Close()

				log.Print(name, ": ", "Creating file ", path.Join(m.Out, m.Branch, jobPath, name))
				outFile, err := btrfs.CreateAll(path.Join(m.Out, m.Branch, jobPath, name))
				if err != nil {
					log.Print(err)
					return
				}
				defer outFile.Close()
				log.Print(name, ": ", "Opened outfile.")

				wait := time.Minute * 10
				if job.TimeOut != 0 {
					wait = time.Duration(job.TimeOut) * time.Second
				}
				timer := time.AfterFunc(wait,
					func() {
						log.Print(name, ": ", "Timeout. Killing mapper.")
						err := resp.Body.Close()
						if err != nil {
							log.Print(err)
						}
					})
				defer log.Print("Result of timer.Stop(): ", timer.Stop())
				log.Print(name, ": ", "Copying output...")
				if _, err := io.Copy(outFile, resp.Body); err != nil {
					log.Print(err)
					return
				}
				log.Print(name, ": ", "Done copying.")
			}
		}(&wg)
	}
	fileCount := 0
	switch {
	case getProtocol(job.Input) == ProtoPfs:
		err = btrfs.LazyWalk(path.Join(m.In, m.Commit, job.Input),
			func(name string) error {
				files <- name
				fileCount++
				if job.Limit != 0 && fileCount >= job.Limit {
					return fmt.Errorf("STOP")
				}
				return nil
			})
		if err != nil && err.Error() != "STOP" {
			log.Print(err)
			return
		}
	case getProtocol(job.Input) == ProtoS3:
		inPath, err := getPath(job.Input)
		if err != nil {
			log.Print(err)
			return
		}
		nextMarker := ""
		for {
			log.Print("s3: before List nextMarker = ", nextMarker)
			lr, err := bucket.List(inPath, "", nextMarker, 0)
			if err != nil {
				log.Print(err)
				return
			}
			for _, key := range lr.Contents {
				if route.HashResource(key.Key)%modulos == shard {
					// This file belongs on this shard
					files <- key.Key
					fileCount++
					if job.Limit != 0 && fileCount >= job.Limit {
						break
					}
				}
				nextMarker = key.Key
			}
			if !lr.IsTruncated {
				// We've exhausted the output
				break
			}
			if job.Limit != 0 && fileCount >= job.Limit {
				// We've hit the user imposed limit
				break
			}
		}
	default:
		log.Printf("Unrecognized protocol.")
		return
	}
}

func Reduce(job Job, jobPath string, m materializeInfo, host string, shard, modulos uint64) {
	if (route.HashResource(path.Join("/job", jobPath)) % modulos) != shard {
		// This resource isn't supposed to be located on this machine so we
		// don't need to materialize it.
		return
	}
	log.Print("Reduce: ", job, " ", jobPath, " ")
	if job.Type != "reduce" {
		log.Printf("Reduce called on a job of type \"%s\". Should be \"reduce\".", job.Type)
		return
	}

	// Notice we're just passing "host" here. Multicast will fill in the host
	// field so we don't actually need to specify it.
	var reader io.ReadCloser
	err := retry(func() error {
		req, err := http.NewRequest("GET", "http://host/"+path.Join(job.Input, "file", "*")+"?commit="+m.Commit, nil)
		if err != nil {
			return err
		}
		reader, err = route.Multicast(req, "/pfs/master")
		return err
	}, retries, time.Minute)
	if err != nil {
		log.Print(err)
		return
	}
	defer reader.Close()

	var resp *http.Response
	err = retry(func() error {
		resp, err = http.Post("http://"+path.Join(host, job.Input), "application/text", reader)
		return err
	}, retries, 200*time.Millisecond)
	if err != nil {
		log.Print(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("Received error %s", resp.Status)
		return
	}

	outFile, err := btrfs.CreateAll(path.Join(m.Out, m.Branch, jobPath))
	if err != nil {
		log.Print(err)
		return
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		log.Print(err)
		return
	}
	log.Print(nil)
	return
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
		log.Print("Jobs dir doesn't exists:\n", path.Join(in_repo, commit, jobDir))
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
			containerId, err = spinupContainer(job.Image, job.Command)
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
				Map(job, jobInfo.Name(), m, containerAddr, shard, modulos)
			} else if job.Type == "reduce" {
				Reduce(job, jobInfo.Name(), m, containerAddr, shard, modulos)
			} else {
				log.Printf("Job %s has unrecognized type: %s.", jobInfo.Name(), job.Type)
			}
		}(jobInfo)
	}
	return nil
}

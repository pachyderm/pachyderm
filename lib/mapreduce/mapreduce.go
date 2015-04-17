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

const (
	retries         = 5
	defaultParallel = 100
	usesPerMapper   = 2000
)

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
		log.Print("Failed to pull ", image, " with error: ", err)
		// We keep going here because it might be a local image.
	}

	return startContainer(image, command)
}

func stopContainer(containerId string) error {
	log.Print("stopContainer", " ", containerId)
	docker, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		log.Print(err)
		return err
	}
	return docker.StopContainer(containerId, 5)
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
			log.Print("Retrying due to error: ", err)
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
	Cmd      []string `json:"command"`
	Limit    int      `json:"limit"`
	Parallel int      `json:"parallel"`
	TimeOut  int      `json:"timeout"`
}

type materializeInfo struct {
	In, Out, Branch, Commit string
}

func PrepJob(job Job, jobName string, m materializeInfo) error {
	if err := btrfs.MkdirAll(path.Join(m.Out, m.Branch, jobName)); err != nil {
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

func newBucket(bucketName string) (*s3.Bucket, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	client := s3.New(auth, aws.USWest)

	return client.Bucket(bucketName), nil
}

func mapFile(filename, jobName, host string, job Job, m materializeInfo) error {
	// This retry is because we don't have a great way to tell when
	// the containers http server is actually listening. It also
	// can help in cases where a user has written a job that fails
	// intermittently.
	err := retry(func() error {
		var err error
		var inFile io.ReadCloser
		switch {
		case getProtocol(job.Input) == ProtoPfs:
			inFile, err = btrfs.Open(path.Join(m.In, m.Commit, job.Input, filename))
		case getProtocol(job.Input) == ProtoS3:
			bucketName, err := getBucket(job.Input)
			if err != nil {
				log.Print(err)
				return err
			}
			bucket, err := newBucket(bucketName)
			if err != nil {
				log.Print(err)
				return err
			}
			inFile, err = bucket.GetReader(filename)
			if inFile == nil {
				return fmt.Errorf("Nil file returned.")
			}
		default:
			return fmt.Errorf("Invalid protocol.")
		}
		if err != nil {
			log.Print(err)
			return err
		}
		log.Print("In file: ", inFile)
		defer inFile.Close()

		log.Print(filename, ": ", "Posting: ", "http://"+path.Join(host, filename))
		client := &http.Client{Timeout: time.Duration(job.TimeOut) * time.Second}
		resp, err := client.Post("http://"+path.Join(host, filename), "application/text", inFile)
		if err != nil {
			log.Print(err)
			return err
		}
		defer resp.Body.Close()
		log.Print(filename, ": ", "Post done.")

		log.Print(filename, ": ", "Creating file ", path.Join(m.Out, m.Branch, jobName, filename))
		outFile, err := btrfs.CreateAll(path.Join(m.Out, m.Branch, jobName, filename))
		if err != nil {
			log.Print(err)
			return err
		}
		defer outFile.Close()
		log.Print(filename, ": ", "Created outfile.")
		log.Print(filename, ": ", "Copying output...")

		if _, err := io.Copy(outFile, resp.Body); err != nil {
			log.Print(err)
			return err
		}
		log.Print(filename, ": ", "Done copying.")
		return nil
	}, retries, 500*time.Millisecond)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func pumpFiles(files chan string, job Job, m materializeInfo, shard, modulos uint64) {
	defer close(files)

	fileCount := 0
	switch {
	case getProtocol(job.Input) == ProtoPfs:
		err := btrfs.LazyWalk(path.Join(m.In, m.Commit, job.Input),
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
		bucketName, err := getBucket(job.Input)
		if err != nil {
			log.Print(err)
			return
		}
		bucket, err := newBucket(bucketName)
		if err != nil {
			log.Print(err)
			return
		}
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

func Map(job Job, jobName string, m materializeInfo, shard, modulos uint64) {
	err := PrepJob(job, path.Base(jobName), m)
	if err != nil {
		log.Print(err)
		return
	}

	if job.Type != "map" {
		log.Printf("runMap called on a job of type \"%s\". Should be \"map\".", job.Type)
		return
	}

	files := make(chan string)

	// pumpFiles will close the channel when it's done
	go pumpFiles(files, job, m, shard, modulos)

	parallel := defaultParallel
	if job.Parallel > 0 {
		parallel = job.Parallel
	}

	for {
		// Spinup a Mapper()
		containerId, err := spinupContainer(job.Image, job.Cmd)
		if err != nil {
			log.Print(err)
			return
		}
		// Make sure that the Mapper gets cleaned up
		host, err := ipAddr(containerId)
		if err != nil {
			log.Print(err)
			return
		}

		semaphore := make(chan int, parallel)

		uses := 0
		for name := range files {
			log.Println("Wait for sem...")
			semaphore <- 1
			log.Println("Acquired sem.")
			go func(name string) {
				if err := mapFile(name, jobName, host, job, m); err != nil {
					log.Print(err)
				}
				log.Print("Releasing semaphore.")
				<-semaphore
			}(name)
			uses++
			if uses == usesPerMapper {
				log.Print("Used up the mapper, time to make a new one.")
				break
			}
		}
		// drain the semaphore
		for i := 0; i < parallel; i++ {
			semaphore <- 1
		}
		close(semaphore)
		if err := stopContainer(containerId); err != nil {
			log.Print(err)
			return
		}

		// This means we exited because we ran out of files to process
		if uses < usesPerMapper {
			break
		}
	}
}

func Reduce(job Job, jobName string, m materializeInfo, shard, modulos uint64) {
	if (route.HashResource(path.Join("/job", jobName)) % modulos) != shard {
		// This resource isn't supposed to be located on this machine so we
		// don't need to materialize it.
		return
	}
	log.Print("Reduce: ", job, " ", jobName, " ")
	if job.Type != "reduce" {
		log.Printf("Reduce called on a job of type \"%s\". Should be \"reduce\".", job.Type)
		return
	}

	// Spinup a Reducer()
	containerId, err := spinupContainer(job.Image, job.Cmd)
	if err != nil {
		log.Print(err)
		return
	}
	// Make sure that the Reducer gets cleaned up
	defer func() {
		if err := stopContainer(containerId); err != nil {
			log.Print(err)
		}
	}()
	host, err := ipAddr(containerId)
	if err != nil {
		log.Print(err)
		return
	}

	var resp *http.Response
	err = retry(func() error {
		var reader io.ReadCloser
		if modulos == 1 {
			// We're in single node mode so we can do something much simpler
			resp, err := http.Get("http://localhost/" + path.Join(job.Input, "file", "*") + "?commit=" + m.Commit)
			if err != nil {
				log.Print(err)
				return err
			}
			reader = resp.Body
		} else {
			err := retry(func() error {
				// Notice we're just passing "host" here. Multicast will fill in the host
				// field so we don't actually need to specify it.
				req, err := http.NewRequest("GET", "http://host/"+path.Join(job.Input, "file", "*")+"?commit="+m.Commit, nil)
				if err != nil {
					return err
				}
				reader, err = route.Multicast(req, "/pfs/master")
				return err
			}, retries, time.Minute)
			if err != nil {
				log.Print(err)
				return err
			}
		}
		defer reader.Close()

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

	outFile, err := btrfs.CreateAll(path.Join(m.Out, m.Branch, jobName))
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

// Materialize parses the jobs found in `in_repo`/`commit`/`jobDir` runs them
// with `in_repo/commit` as input, outputs the results to `outRepo`/`branch`
// and commits them as `outRepo`/`commit`
func Materialize(in_repo, branch, commit, outRepo, jobDir string, shard, modulos uint64) error {
	log.Printf("Materialize: %s %s %s %s %s.", in_repo, branch, commit, outRepo, jobDir)
	exists, err := btrfs.FileExists(path.Join(outRepo, branch))
	if err != nil {
		log.Print(err)
		return err
	}
	if !exists {
		if err := btrfs.Branch(outRepo, "t0", branch); err != nil {
			log.Print(err)
			return err
		}
	}
	/* We use the .progress dir to indicate if a job has been completed in this
	* materialization. */
	if err := btrfs.MkdirAll(path.Join(outRepo, branch, ".progress", commit)); err != nil {
		log.Print(err)
		return err
	}
	// We make sure that this function always commits so that we know the comp
	// repo stays in sync with the data repo.
	defer func() {
		if err := btrfs.Commit(outRepo, commit, branch); err != nil {
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
			defer func() {
				f, err := btrfs.Create(path.Join(outRepo, branch, ".progress", commit, jobInfo.Name()))
				if err != nil {
					log.Print(err)
					return
				}
				f.Close()
			}()
			// First create the job's directory and lock it.
			log.Print("Running job:\n", jobInfo.Name())
			defer wg.Done()
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
			m := materializeInfo{in_repo, outRepo, branch, commit}

			if job.Type == "map" {
				Map(job, jobInfo.Name(), m, shard, modulos)
			} else if job.Type == "reduce" {
				Reduce(job, jobInfo.Name(), m, shard, modulos)
			} else {
				log.Printf("Job %s has unrecognized type: %s.", jobInfo.Name(), job.Type)
			}
		}(jobInfo)
	}
	return nil
}

func WaitJob(outRepo, branch, commit, job string) error {
	log.Printf("WaitJob: %s %s %s %s", outRepo, branch, commit, job)
	return btrfs.WaitForFile(path.Join(outRepo, branch, ".progress", commit, job))
}

package s3

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/pachyderm/s2"
	"github.com/gogo/protobuf/types"
)

type Bucket struct {
	Repo   string
	Commit string
	Name   string
}

type Driver interface {
	// TODO(ys): make these methods private?
	ListBuckets(pc *client.APIClient, r *http.Request, buckets *[]s2.Bucket) error
	GetBucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error)
	CanModifyBuckets() bool
	CanGetHistoricObject() bool
}

type MasterDriver struct {}

func NewMasterDriver() *MasterDriver {
	return &MasterDriver{}
}

func (d *MasterDriver) ListBuckets(pc *client.APIClient, r *http.Request, buckets *[]s2.Bucket) error {
	repos, err := pc.ListRepo()
	if err != nil {
		return err
	}

	for _, repo := range repos {
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return err
		}
		for _, branch := range repo.Branches {
			*buckets = append(*buckets, s2.Bucket{
				Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
				CreationDate: t,
			})
		}
	}

	return nil
}

func (d *MasterDriver) GetBucket(pc *client.APIClient, r *http.Request, name string) (Bucket, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return Bucket{}, s2.InvalidBucketNameError(r)
	}
	repo := parts[1]
	branch := parts[0]
	return Bucket{
		Repo:   repo,
		Commit: branch,
		Name:   name,
	}, nil
}

func (d *MasterDriver) CanModifyBuckets() bool {
	return true
}

func (d *MasterDriver) CanGetHistoricObject() bool {
	return true
}

type WorkerDriver struct {
	inputBuckets []Bucket
	outputBucket *Bucket
	namesMap map[string]*Bucket
}

func NewWorkerDriver(inputBuckets []Bucket, outputBucket *Bucket) *WorkerDriver {
	namesMap := map[string]*Bucket{}
	
	for _, ib := range inputBuckets {
		namesMap[ib.Name] = &ib
	}

	if outputBucket != nil {
		namesMap[outputBucket.Name] = outputBucket
	}

	return &WorkerDriver{
		inputBuckets: inputBuckets,
		outputBucket: outputBucket,
		namesMap: namesMap,
	}
}

func (d *WorkerDriver) ListBuckets(pc *client.APIClient, r *http.Request, buckets *[]s2.Bucket) error {
	repos, err := pc.ListRepo()
	if err != nil {
		return err
	}
	timestamps := map[string]time.Time{}
	for _, repo := range repos {
		timestamp, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return err
		}
		timestamps[repo.Repo.Name] = timestamp
	}

	for _, bucket := range d.namesMap {
		timestamp, ok := timestamps[bucket.Repo]
		if !ok {
			return fmt.Errorf("worker s3gateway configuration includes repo %q, which does not exist", bucket.Repo)
		}
		*buckets = append(*buckets, s2.Bucket{
			Name:         inputRepo.Name,
			CreationDate: timestamp,
		})
	}

	return nil
}

func (d *WorkerDriver) GetBucket(pc *client.APIClient, r *http.Request, name string) (Bucket, error) {
	bucket := d.namesMap[name]
	if bucket == nil {
		return Bucket{}, s2.NoSuchBucketError(r)
	}
	return bucket, nil
}

func (d *WorkerDriver) CanModifyBuckets() bool {
	return false
}

func (d *WorkerDriver) CanGetHistoricObject() bool {
	return false
}

package s3

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/s2"
)

type Bucket struct {
	Repo   string
	Commit string
	Name   string
}

type bucketCapabilities struct {
	readable         bool
	writable         bool
	historicVersions bool
}

type Driver interface {
	listBuckets(pc *client.APIClient, r *http.Request, buckets *[]s2.Bucket) error
	bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error)
	bucketCapabilities(pc *client.APIClient, r *http.Request, bucket *Bucket) (bucketCapabilities, error)
	canModifyBuckets() bool
}

type MasterDriver struct{}

func NewMasterDriver() *MasterDriver {
	return &MasterDriver{}
}

func (d *MasterDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]s2.Bucket) error {
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

func (d *MasterDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return nil, s2.InvalidBucketNameError(r)
	}
	repo := parts[1]
	branch := parts[0]
	return &Bucket{
		Repo:   repo,
		Commit: branch,
		Name:   name,
	}, nil
}

func (d *MasterDriver) bucketCapabilities(pc *client.APIClient, r *http.Request, bucket *Bucket) (bucketCapabilities, error) {
	branchInfo, err := pc.InspectBranch(bucket.Repo, bucket.Commit)
	if err != nil {
		return bucketCapabilities{}, maybeNotFoundError(r, err)
	}

	return bucketCapabilities{
		readable:         branchInfo.Head != nil,
		writable:         true,
		historicVersions: true,
	}, nil
}

func (d *MasterDriver) canModifyBuckets() bool {
	return true
}

type WorkerDriver struct {
	inputBuckets []*Bucket
	outputBucket *Bucket
	namesMap     map[string]*Bucket
}

func NewWorkerDriver(inputBuckets []*Bucket, outputBucket *Bucket) *WorkerDriver {
	namesMap := map[string]*Bucket{}

	for _, ib := range inputBuckets {
		namesMap[ib.Name] = ib
	}

	if outputBucket != nil {
		namesMap[outputBucket.Name] = outputBucket
	}

	return &WorkerDriver{
		inputBuckets: inputBuckets,
		outputBucket: outputBucket,
		namesMap:     namesMap,
	}
}

func (d *WorkerDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]s2.Bucket) error {
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
			Name:         bucket.Name,
			CreationDate: timestamp,
		})
	}

	return nil
}

func (d *WorkerDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	bucket := d.namesMap[name]
	if bucket == nil {
		return &Bucket{
			Name: name,
		}, nil
	}
	return bucket, nil
}

func (d *WorkerDriver) bucketCapabilities(pc *client.APIClient, r *http.Request, bucket *Bucket) (bucketCapabilities, error) {
	if bucket.Repo == "" || bucket.Commit == "" {
		return bucketCapabilities{}, s2.NoSuchBucketError(r)
	} else if bucket == d.outputBucket {
		return bucketCapabilities{
			readable:         false,
			writable:         true,
			historicVersions: false,
		}, nil
	}
	return bucketCapabilities{
		readable:         true,
		writable:         false,
		historicVersions: false,
	}, nil
}

func (d *WorkerDriver) canModifyBuckets() bool {
	return false
}

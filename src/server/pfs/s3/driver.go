package s3

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/s2"
)

// Bucket represents an S3 bucket
type Bucket struct {
	// Repo is the PFS repo that this bucket points to
	Repo string
	// Commit is the PFS commit that this repo points to
	Commit string
	// Name is the name of the bucket
	Name string
}

type bucketCapabilities struct {
	readable         bool
	writable         bool
	historicVersions bool
}

// Driver implementations drive the underlying bucket-related functionality
// for an s3gateway instance
type Driver interface {
	listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error
	bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error)
	bucketCapabilities(pc *client.APIClient, r *http.Request, bucket *Bucket) (bucketCapabilities, error)
	canModifyBuckets() bool
}

// MasterDriver is the driver for the s3gateway instance running on pachd
// master
type MasterDriver struct{}

// NewMasterDriver constructs a new master driver
func NewMasterDriver() *MasterDriver {
	return &MasterDriver{}
}

func (d *MasterDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error {
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
			*buckets = append(*buckets, &s2.Bucket{
				Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
				CreationDate: t,
			})
		}
	}

	return nil
}

func (d *MasterDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	branch := "master"
	var repo string
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		branch = parts[0]
		repo = parts[1]
	} else {
		repo = parts[0]
	}

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

// WorkerDriver is the driver for the s3gateway instance running on pachd
// workers
type WorkerDriver struct {
	inputBuckets []*Bucket
	outputBucket *Bucket
	namesMap     map[string]*Bucket
}

// NewWorkerDriver creates a new worker driver. `inputBuckets` is a list of
// whitelisted buckets to be served from input repos. `outputBucket` is the
// whitelisted bucket to be served from an output repo. If `nil`, no output
// bucket will be available.
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

func (d *WorkerDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error {
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
			return errors.Errorf("worker s3gateway configuration includes repo %q, which does not exist", bucket.Repo)
		}
		*buckets = append(*buckets, &s2.Bucket{
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

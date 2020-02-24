package s3

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/pachyderm/s2"
	"github.com/gogo/protobuf/types"
)

type RepoReference struct {
	repo string
	commit string
}

type Driver interface {
	// TODO(ys): make these methods private?
	ListBuckets(pc *client.APIClient, buckets *[]s2.Bucket) error
	// TODO(ys): consider moving validation logic out
	DereferenceBucket(pc *client.APIClient, r *http.Request, bucket string) (RepoReference, error)
	CanModifyBuckets() bool
	CanGetHistoricObject() bool
}

type MasterDriver struct {}

func NewMasterDriver() *MasterDriver {
	return &MasterDriver{}
}

func (d *MasterDriver) ListBuckets(pc *client.APIClient, buckets *[]s2.Bucket) error {
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

func (d *MasterDriver) DereferenceBucket(pc *client.APIClient, r *http.Request, name string) (RepoReference, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return RepoReference{}, s2.InvalidBucketNameError(r)
	}
	repo := parts[1]
	branch := parts[0]
	return RepoReference{
		repo:   repo,
		commit: branch,
	}, nil
}

func (d *MasterDriver) CanModifyBuckets() bool {
	return true
}

func (d *MasterDriver) CanGetHistoricObject() bool {
	return true
}

type WorkerBucket struct {
	Repo   string
	Commit string
	Name   string
}

type WorkerDriver struct {
	inputBuckets []WorkerBucket
	outputBucket *WorkerBucket
	namesMap map[string]*WorkerBucket
}

func NewWorkerDriver(inputBuckets []WorkerBucket, outputBucket *WorkerBucket) *WorkerDriver {
	reposMap := map[string]*WorkerBucket{}
	namesMap := map[string]*WorkerBucket{}
	
	for _, ib := range inputBuckets {
		reposMap[ib.Repo] = &ib
		namesMap[ib.Name] = &ib
	}

	if outputBucket != nil {
		reposMap[outputBucket.Repo] = outputBucket
		namesMap[outputBucket.Name] = outputBucket
	}

	return &WorkerDriver{
		inputBuckets: inputBuckets,
		outputBucket: outputBucket,
		reposMap: reposMap,
		namesMap: namesMap,
	}
}

func (d *WorkerDriver) ListBuckets(pc *client.APIClient, buckets *[]s2.Bucket) error {
	repos, err := pc.ListRepo()
	if err != nil {
		return err
	}

	for _, repo := range repos {
		inputRepo := d.reposMap[repo.Repo.Name]
		if inputRepo == nil {
			continue
		}

		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return err
		}

		*buckets = append(*buckets, s2.Bucket{
			Name:         inputRepo.Name,
			CreationDate: t,
		})
	}

	return nil
}

func (d *WorkerDriver) DereferenceBucket(pc *client.APIClient, r *http.Request, name string) (RepoReference, error) {
	bucket := d.namesMap[name]
	if bucket == nil {
		return RepoReference{}, s2.NoSuchBucketError(r)
	}
	return RepoReference {
		repo: bucket.Repo,
		commit: bucket.Commit,
	}, nil
}

func (d *WorkerDriver) CanModifyBuckets() bool {
	return false
}

func (d *WorkerDriver) CanGetHistoricObject() bool {
	return false
}

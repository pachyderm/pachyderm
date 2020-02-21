package s3

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/pachyderm/s2"
	"github.com/gogo/protobuf/types"
)

type Driver interface {
	ListBuckets(pc *client.APIClient, buckets *[]s2.Bucket) error
	DereferenceBucket(pc *client.APIClient, r *http.Request, bucket string, validate bool) (string, string, error)
	CanModifyBuckets() bool
	CanGetHistoricObject() bool
}

type PFSDriver struct {}

func (d *PFSDriver) ListBuckets(pc *client.APIClient, buckets *[]s2.Bucket) error {
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

func (d *PFSDriver) DereferenceBucket(pc *client.APIClient, r *http.Request, name string, validate bool) (string, string, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return "", "", s2.InvalidBucketNameError(r)
	}
	repo := parts[1]
	branch := parts[0]

	if validate {
		branchInfo, err := pc.InspectBranch(repo, branch)
		if err != nil {
			return "", "", maybeNotFoundError(r, err)
		}
		if branchInfo.Head == nil {
			return "", "", s2.NoSuchKeyError(r)
		}
	}

	return repo, branch, nil
}

func (d *PFSDriver) CanModifyBuckets() bool {
	return true
}

func (d *PFSDriver) CanGetHistoricObject() bool {
	return true
}

type PPSBucket struct {
	Repo     string
	CommitID string
	Name     string
}

type PPSDriver struct {
	inputBuckets []PPSBucket
	outputBucket *PPSBucket
	reposMap map[string]*PPSBucket
	namesMap map[string]*PPSBucket
}

func NewPPSDriver(inputBuckets []PPSBucket, outputBucket *PPSBucket) *PPSDriver {
	reposMap := map[string]*PPSBucket{}
	namesMap := map[string]*PPSBucket{}
	
	for _, ib := range inputBuckets {
		reposMap[ib.Repo] = &ib
		namesMap[ib.Name] = &ib
	}

	if outputBucket != nil {
		reposMap[outputBucket.Repo] = outputBucket
		namesMap[outputBucket.Name] = outputBucket
	}

	return &PPSDriver{
		inputBuckets: inputBuckets,
		outputBucket: outputBucket,
		reposMap: reposMap,
		namesMap: namesMap,
	}
}

func (d *PPSDriver) ListBuckets(pc *client.APIClient, buckets *[]s2.Bucket) error {
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

func (d *PPSDriver) DereferenceBucket(pc *client.APIClient, r *http.Request, name string, validate bool) (string, string, error) {
	bucket := d.namesMap[name]
	if bucket == nil {
		return "", "", s2.NoSuchBucketError(r)
	}
	return bucket.Repo, bucket.CommitID, nil
}

func (d *PPSDriver) CanModifyBuckets() bool {
	return false
}

func (d *PPSDriver) CanGetHistoricObject() bool {
	return false
}

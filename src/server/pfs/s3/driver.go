//nolint:wrapcheck
// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors
package s3

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/s2"
)

// Bucket represents an S3 bucket
type Bucket struct {
	// Commit is the PFS commit that this bucket points to
	Commit *pfs.Commit
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

func branchBucketName(b *pfs.Branch) string {
	if b.Repo.Type == pfs.UserRepoType {
		return fmt.Sprintf("%s.%s", b.Name, b.Repo.Name)
	}
	return fmt.Sprintf("%s.%s.%s", b.Name, b.Repo.Type, b.Repo.Name)
}

func projectBranchBucketName(b *pfs.Branch) string {
	if b.Repo.Type == pfs.UserRepoType {
		return fmt.Sprintf("%s.%s.%s", b.Name, b.Repo.Name, b.Repo.Project.Name)
	}
	return fmt.Sprintf("%s.%s.%s.%s", b.Name, b.Repo.Type, b.Repo.Name, b.Repo.Project.Name)
}

func (d *MasterDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error {
	var isProjectAware = mux.Vars(r)["isProjectAware"] == "yes"
	repos, err := pc.ListRepoByType("") // get repos of all types
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if !isProjectAware && repo.Repo.Project.Name != pfs.DefaultProjectName {
			continue // skip repos not in the default project
		}
		if repo.Repo.Type == pfs.SpecRepoType {
			continue // hide spec repos, but allow meta/stats repos
		}
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return err
		}
		for _, branch := range repo.Branches {
			var name string
			if isProjectAware {
				name = projectBranchBucketName(branch)
			} else {
				name = branchBucketName(branch)
			}
			*buckets = append(*buckets, &s2.Bucket{
				Name:         name,
				CreationDate: t,
			})
		}
	}

	return nil
}

func bucketNameToCommit(bucketName string) *pfs.Commit {
	var id string
	branch := "master"
	var repo *pfs.Repo

	// the name is [commitID.][branch. | branch.type.]repoName
	// in particular, to access a non-user system repo, the branch name must be given
	parts := strings.SplitN(bucketName, ".", 4)
	if uuid.IsUUIDWithoutDashes(parts[0]) {
		id = parts[0]
		parts = parts[1:]
	}
	if len(parts) > 1 {
		branch = parts[0]
		parts = parts[1:]
	}
	if len(parts) == 1 {
		repo = client.NewProjectRepo(pfs.DefaultProjectName, parts[0])
	} else {
		repo = client.NewSystemProjectRepo(pfs.DefaultProjectName, parts[1], parts[0])
	}

	return repo.NewCommit(branch, id)
}

func bucketNameToProjectCommit(bucketName string) (*pfs.Commit, error) {
	var id string
	branch := "master"
	var repo *pfs.Repo

	// the name is [commitID.][branch. | branch.type.]repoName.projectName
	// in particular, to access a non-user system repo, the branch name must be given
	parts := strings.SplitN(bucketName, ".", 5)
	if uuid.IsUUIDWithoutDashes(parts[0]) {
		id = parts[0]
		parts = parts[1:]
	}
	if len(parts) > 2 {
		branch = parts[0]
		parts = parts[1:]
	}
	switch len(parts) {
	case 0:
		return nil, errors.Errorf("bad bucket name %s", bucketName)
	case 1:
		return nil, errors.Errorf("bad bucket name %s", bucketName)
	case 2:
		repo = client.NewProjectRepo(parts[1], parts[0])
	case 3:
		repo = client.NewSystemProjectRepo(parts[2], parts[1], parts[0])
	default:
		return nil, errors.Errorf("bad bucket name %s", bucketName)
	}

	return repo.NewCommit(branch, id), nil
}

func (d *MasterDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	if mux.Vars(r)["isProjectAware"] == "yes" {
		commit, err := bucketNameToProjectCommit(name)
		if err != nil {
			return nil, errors.Wrap(err, "could not map bucket name to commit")
		}
		return &Bucket{
			Commit: commit,
			Name:   name,
		}, nil
	}
	return &Bucket{
		Commit: bucketNameToCommit(name),
		Name:   name,
	}, nil
}

func (d *MasterDriver) bucketCapabilities(pc *client.APIClient, r *http.Request, bucket *Bucket) (bucketCapabilities, error) {
	_, err := pc.PfsAPIClient.InspectBranch(pc.Ctx(), &pfs.InspectBranchRequest{Branch: bucket.Commit.Branch})
	if err != nil {
		return bucketCapabilities{}, maybeNotFoundError(r, grpcutil.ScrubGRPC(err))
	}

	return bucketCapabilities{
		readable:         true,
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
	repos, err := pc.ListRepoByType("") // get repos of all types
	if err != nil {
		return err
	}
	timestamps := map[string]time.Time{}
	for _, repo := range repos {
		timestamp, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return err
		}
		timestamps[pfsdb.RepoKey(repo.Repo)] = timestamp
	}

	for _, bucket := range d.namesMap {
		timestamp, ok := timestamps[pfsdb.RepoKey(bucket.Commit.Branch.Repo)]
		if !ok {
			return errors.Errorf("worker s3gateway configuration includes repo %q, which does not exist", bucket.Commit.Branch.Repo)
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
	if bucket.Commit == nil {
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

// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors
package s3

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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

func (d *MasterDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error {
	repos, err := pc.ListRepo() // get all user-type repos
	if err != nil {
		return err
	}

	for _, repo := range repos {
		t := repo.Created.AsTime()
		for _, b := range repo.Branches {
			*buckets = append(*buckets, &s2.Bucket{
				Name:         fmt.Sprintf("%s.%s.%s", b.GetName(), b.GetRepo().GetName(), b.GetRepo().GetProject().GetName()),
				CreationDate: t,
			})
		}
	}

	return nil
}

func bucketNameToCommit(bucketName string) (*pfs.Commit, error) {
	var (
		id     string
		branch string
		repo   *pfs.Repo
	)

	// the name is [[commitID.]branch.]repoName[.project]
	// If unspecified, the branch is 'master', and the project is 'default'.
	// However, two-component buckets are always interpreted as '<branch>.<repo>'.
	// So, to specify a non-default project, the branch must be set explicitly.
	//
	// For example, 'master.myrepo.default' may be shortened to just 'myrepo' by
	// removing the default project and branch, and 'mybranch.myrepo.default' may
	// be shortened to just 'mybranch.myrepo' by removing the default project, but
	// 'master.myrepo.myproject' cannot be shortened because the project is
	// non-default.
	//
	// Note that buckets can never have 5 or more components--additional splitting
	// is unnecessary, as 5 will always error.
	parts := strings.SplitN(bucketName, ".", 5)
	if uuid.IsUUIDWithoutDashes(parts[0]) {
		id = parts[0]
		parts = parts[1:]
	}

	switch len(parts) {
	case 0:
		return nil, errors.Errorf("invalid bucket name %q; must include a repo", bucketName)
	case 1:
		// e.g. s3://myrepo
		repo, branch = client.NewRepo(pfs.DefaultProjectName, parts[0]), "master"
	case 2:
		// e.g. s3://mybranch.myrepo
		repo, branch = client.NewRepo(pfs.DefaultProjectName, parts[1]), parts[0]
	case 3:
		// e.g. s3://mybranch.myrepo.myproject
		repo, branch = client.NewRepo(parts[2], parts[1]), parts[0]
	default:
		return nil, errors.Errorf("invalid bucket name: %q", bucketName)
	}

	return repo.NewCommit(branch, id), nil
}

func (d *MasterDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	commit, err := bucketNameToCommit(name)
	if err != nil {
		return nil, err
	}
	return &Bucket{
		Commit: commit,
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
		timestamp := repo.Created.AsTime()
		timestamps[pfsdb.RepoKey(repo.Repo)] = timestamp
	}

	for _, bucket := range d.namesMap {
		timestamp, ok := timestamps[pfsdb.RepoKey(bucket.Commit.Repo)]
		if !ok {
			return errors.Errorf("worker s3gateway configuration includes repo %q, which does not exist", bucket.Commit.Repo)
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

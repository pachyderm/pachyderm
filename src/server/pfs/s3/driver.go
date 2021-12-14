package s3

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/s2"

	"github.com/gogo/protobuf/types"
)

// Bucket represents an S3 bucket
type Bucket struct {
	// Commit is the PFS commit that this bucket points to
	Commit *pfs.Commit
	// Name is the name of the bucket
	Name string
	// Path is the (implicity) root of the filesystem this bucket accesses,
	// understood as the root if empty
	Path string
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
	listObjects(pc *client.APIClient, bucket *Bucket, pattern, prefix, marker string, recursive bool, maxKeys int) (*s2.ListObjectsResult, error)
	getObject(pc *client.APIClient, bucket *Bucket, path, version string) (*s2.GetObjectResult, error)
	copyObject(pc *client.APIClient, srcBucket, destBucket *Bucket, srcFile, destFile string) (string, error)
	putObject(pc *client.APIClient, bucket *Bucket, path string, reader io.Reader) (*s2.PutObjectResult, error)
	deleteObject(pc *client.APIClient, bucket *Bucket, file string) error
}

// MasterDriver is the driver for the s3gateway instance running on pachd
// master
type MasterDriver struct {
	pachFS
}

// NewMasterDriver constructs a new master driver
func NewMasterDriver() *MasterDriver {
	return &MasterDriver{}
}

func (d *MasterDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error {
	repos, err := pc.ListRepoByType("") // get repos of all types
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if repo.Repo.Type == pfs.SpecRepoType {
			continue // hide spec repos, but allow meta/stats repos
		}
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return err
		}
		for _, branch := range repo.Branches {
			var name string
			if branch.Repo.Type == pfs.UserRepoType {
				name = fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name)
			} else {
				name = fmt.Sprintf("%s.%s.%s", branch.Name, branch.Repo.Type, branch.Repo.Name)
			}
			*buckets = append(*buckets, &s2.Bucket{
				Name:         name,
				CreationDate: t,
			})
		}
	}

	return nil
}

func (d *MasterDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	var id string
	branch := "master"
	var repo *pfs.Repo

	// the name is [commitID.][branch. | branch.type.]repoName
	// in particular, to access a non-user system repo, the branch name must be given
	parts := strings.SplitN(name, ".", 4)
	if uuid.IsUUIDWithoutDashes(parts[0]) {
		id = parts[0]
		parts = parts[1:]
	}
	if len(parts) > 1 {
		branch = parts[0]
		parts = parts[1:]
	}
	if len(parts) == 1 {
		repo = client.NewRepo(parts[0])
	} else {
		repo = client.NewSystemRepo(parts[1], parts[0])
	}

	return &Bucket{
		Commit: repo.NewCommit(branch, id),
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
	pachFS
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

var errNoLocalBucket = errors.New("No directory corresponding to the given bucket name")

// LocalDriver is the driver for the s3gateway instance running on pachd
// workers
type LocalDriver struct {
	localFS
	root string
}

// NewLocalDriver creates a new worker driver. `inputBuckets` is a list of
// whitelisted buckets to be served from input repos. `outputBucket` is the
// whitelisted bucket to be served from an output repo. If `nil`, no output
// bucket will be available.
func NewLocalDriver(root string) *LocalDriver {
	return &LocalDriver{
		root: root,
	}
}

func (d *LocalDriver) listBuckets(pc *client.APIClient, r *http.Request, buckets *[]*s2.Bucket) error {
	bucketPaths, err := filepath.Glob(filepath.Join(d.root, "*"))
	if err != nil {
		return err
	}

	for _, path := range bucketPaths {
		*buckets = append(*buckets, &s2.Bucket{
			Name: filepath.Base(path),
		})
	}
	return nil
}

func (d *LocalDriver) bucket(pc *client.APIClient, r *http.Request, name string) (*Bucket, error) {
	return &Bucket{
		Name: name,
		Path: filepath.Join(d.root, name),
	}, nil
}

func (d *LocalDriver) bucketCapabilities(pc *client.APIClient, r *http.Request, bucket *Bucket) (bucketCapabilities, error) {
	_, err := os.Stat(bucket.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = errNoLocalBucket
		}
		return bucketCapabilities{}, err
	}
	return bucketCapabilities{
		readable:         true,
		writable:         true,
		historicVersions: false,
	}, nil
}

func (d *LocalDriver) canModifyBuckets() bool {
	return false
}

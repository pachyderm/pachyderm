package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// NewRepo creates a pfs.Repo.
func NewRepo(repoName string) *pfs.Repo {
	return &pfs.Repo{Name: repoName}
}

// NewCommit creates a pfs.Commit.
func NewCommit(repoName string, commitID string) *pfs.Commit {
	return &pfs.Commit{
		Repo: NewRepo(repoName),
		ID:   commitID,
	}
}

// NewFile creates a pfs.File.
func NewFile(repoName string, commitID string, path string) *pfs.File {
	return &pfs.File{
		Commit: NewCommit(repoName, commitID),
		Path:   path,
	}
}

// NewBlock creates a pfs.Block.
func NewBlock(hash string) *pfs.Block {
	return &pfs.Block{
		Hash: hash,
	}
}

// CommitTypes alias pfs.CommitType_*
const (
	CommitTypeNone  = pfs.CommitType_COMMIT_TYPE_NONE
	CommitTypeRead  = pfs.CommitType_COMMIT_TYPE_READ
	CommitTypeWrite = pfs.CommitType_COMMIT_TYPE_WRITE
)

// CommitStatus alias pfs.CommitStatus_*
const (
	CommitStatusNormal    = pfs.CommitStatus_NORMAL
	CommitStatusArchived  = pfs.CommitStatus_ARCHIVED
	CommitStatusCancelled = pfs.CommitStatus_CANCELLED
	CommitStatusAll       = pfs.CommitStatus_ALL
)

// CreateRepo creates a new Repo object in pfs with the given name. Repos are
// the top level data object in pfs and should be used to store data of a
// similar type. For example rather than having a single Repo for an entire
// project you might have seperate Repos for logs, metrics, database dumps etc.
func (c APIClient) CreateRepo(repoName string) error {
	_, err := c.PfsAPIClient.CreateRepo(
		c.ctx(),
		&pfs.CreateRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return sanitizeErr(err)
}

// InspectRepo returns info about a specific Repo.
func (c APIClient) InspectRepo(repoName string) (*pfs.RepoInfo, error) {
	repoInfo, err := c.PfsAPIClient.InspectRepo(
		c.ctx(),
		&pfs.InspectRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return repoInfo, nil
}

// ListRepo returns info about all Repos.
// provenance specifies a set of provenance repos, only repos which have ALL of
// the specified repos as provenance will be returned unless provenance is nil
// in which case it is ignored.
func (c APIClient) ListRepo(provenance []string) ([]*pfs.RepoInfo, error) {
	request := &pfs.ListRepoRequest{}
	for _, repoName := range provenance {
		request.Provenance = append(request.Provenance, NewRepo(repoName))
	}
	repoInfos, err := c.PfsAPIClient.ListRepo(
		c.ctx(),
		request,
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return repoInfos.RepoInfo, nil
}

// DeleteRepo deletes a repo and reclaims the storage space it was using. Note
// that as of 1.0 we do not reclaim the blocks that the Repo was referencing,
// this is because they may also be referenced by other Repos and deleting them
// would make those Repos inaccessible. This will be resolved in later
// versions.
// If "force" is set to true, the repo will be removed regardless of errors.
// This argument should be used with care.
func (c APIClient) DeleteRepo(repoName string, force bool) error {
	_, err := c.PfsAPIClient.DeleteRepo(
		c.ctx(),
		&pfs.DeleteRepoRequest{
			Repo:  NewRepo(repoName),
			Force: force,
		},
	)
	return err
}

// StartCommit begins the process of committing data to a Repo. Once started
// you can write to the Commit with PutFile and when all the data has been
// written you must finish the Commit with FinishCommit. NOTE, data is not
// persisted until FinishCommit is called.
// parentCommit specifies the parent Commit, upon creation the new Commit will
// appear identical to the parent Commit, data can safely be added to the new
// commit without affecting the contents of the parent Commit. You may pass ""
// as parentCommit in which case the new Commit will have no parent and will
// initially appear empty.
// branch is a more convenient way to build linear chains of commits. When a
// commit is started with a non empty branch the value of branch becomes an
// alias for the created Commit. This enables a more intuitive access pattern.
// When the commit is started on a branch the previous head of the branch is
// used as the parent of the commit.
func (c APIClient) StartCommit(repoName string, parentCommit string) (*pfs.Commit, error) {
	commit, err := c.PfsAPIClient.StartCommit(
		c.ctx(),
		&pfs.StartCommitRequest{
			Parent: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				ID: parentCommit,
			},
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return commit, nil
}

// FinishCommit ends the process of committing data to a Repo and persists the
// Commit. Once a Commit is finished the data becomes immutable and future
// attempts to write to it with PutFile will error.
func (c APIClient) FinishCommit(repoName string, commitID string) error {
	_, err := c.PfsAPIClient.FinishCommit(
		c.ctx(),
		&pfs.FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return sanitizeErr(err)
}

// InspectCommit returns info about a specific Commit.
func (c APIClient) InspectCommit(repoName string, commitID string) (*pfs.CommitInfo, error) {
	commitInfo, err := c.PfsAPIClient.InspectCommit(
		c.ctx(),
		&pfs.InspectCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return commitInfo, nil
}

// ListCommit lists commits.
//
// If only `repo` is given, all commits in the repo are returned.
//
// If `from` is given, only the descendents of `from`, including `from`
// itself, are considered.
//
// If `to` is given, only the ancestors of `to`, including `to` itself,
// are considered.
//
// `number` determines how many commits are returned.  If `number` is 0,
// all commits that match the aforementioned criteria are returned.
func (c APIClient) ListCommit(repoName string, from string, to string, number uint64) ([]*pfs.CommitInfo, error) {
	req := &pfs.ListCommitRequest{
		Repo:   NewRepo(repoName),
		Number: number,
	}
	if from != "" {
		req.From = NewCommit(repoName, from)
	}
	if to != "" {
		req.To = NewCommit(repoName, to)
	}
	commitInfos, err := c.PfsAPIClient.ListCommit(
		c.ctx(),
		req,
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return commitInfos.CommitInfo, nil
}

// ListCommitByRepo lists all commits in a repo.
func (c APIClient) ListCommitByRepo(repoName string) ([]*pfs.CommitInfo, error) {
	return c.ListCommit(repoName, "", "", 0)
}

// ListBranch lists the active branches on a Repo.
func (c APIClient) ListBranch(repoName string, status pfs.CommitStatus) ([]string, error) {
	branches, err := c.PfsAPIClient.ListBranch(
		c.ctx(),
		&pfs.ListBranchRequest{
			Repo:   NewRepo(repoName),
			Status: status,
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return branches.Branches, nil
}

// DeleteCommit deletes a commit.
// Note it is currently not implemented.
func (c APIClient) DeleteCommit(repoName string, commitID string) error {
	_, err := c.PfsAPIClient.DeleteCommit(
		c.ctx(),
		&pfs.DeleteCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return sanitizeErr(err)
}

// FlushCommit blocks until all of the commits which have a set of commits as
// provenance have finished. For commits to be considered they must have all of
// the specified commits as provenance. This in effect waits for all of the
// jobs that are triggered by a set of commits to complete.
// It returns an error if any of the commits it's waiting on are cancelled due
// to one of the jobs encountering an error during runtime.
// If toRepos is not nil then only the commits up to and including those repos
// will be considered, otherwise all repos are considered.
// Note that it's never necessary to call FlushCommit to run jobs, they'll run
// no matter what, FlushCommit just allows you to wait for them to complete and
// see their output once they do.
func (c APIClient) FlushCommit(commits []*pfs.Commit, toRepos []*pfs.Repo) ([]*pfs.CommitInfo, error) {
	commitInfos, err := c.PfsAPIClient.FlushCommit(
		c.ctx(),
		&pfs.FlushCommitRequest{
			Commit: commits,
			ToRepo: toRepos,
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return commitInfos.CommitInfo, nil
}

// TODO: this API is temporary being used until the tag store is implemented
func (c APIClient) Put(reader io.Reader) (*pfs.BlockRef, error) {
	blocks, err := c.PutBlock(pfs.Delimiter_NONE, reader)
	if err != nil {
		return nil, err
	}
	return blocks.BlockRef[0], nil
}

// PutBlock takes a reader and splits the data in it into blocks.
// Blocks are guaranteed to be new line delimited.
// Blocks are content addressed and are thus identified by hashes of the content.
// NOTE: this is lower level function that's used internally and might not be
// useful to users.
func (c APIClient) PutBlock(delimiter pfs.Delimiter, reader io.Reader) (blockRefs *pfs.BlockRefs, retErr error) {
	writer, err := c.newPutBlockWriteCloser(delimiter)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	defer func() {
		err := writer.Close()
		if err != nil && retErr == nil {
			retErr = err
		}
		if retErr == nil {
			blockRefs = writer.blockRefs
		}
	}()
	_, retErr = io.Copy(writer, reader)
	return blockRefs, retErr
}

// GetBlock returns the content of a block using it's hash.
// offset specifies a number of bytes that should be skipped in the beginning of the block.
// size limits the total amount of data returned, note you will get fewer bytes
// than size if you pass a value larger than the size of the block.
// If size is set to 0 then all of the data will be returned.
// NOTE: this is lower level function that's used internally and might not be
// useful to users.
func (c APIClient) GetBlock(hash string, offset uint64, size uint64) (io.Reader, error) {
	apiGetBlockClient, err := c.BlockAPIClient.GetBlock(
		c.ctx(),
		&pfs.GetBlockRequest{
			Block:       NewBlock(hash),
			OffsetBytes: offset,
			SizeBytes:   size,
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return grpcutil.NewStreamingBytesReader(apiGetBlockClient), nil
}

// DeleteBlock deletes a block from the block store.
// NOTE: this is lower level function that's used internally and might not be
// useful to users.
func (c APIClient) DeleteBlock(block *pfs.Block) error {
	_, err := c.BlockAPIClient.DeleteBlock(
		c.ctx(),
		&pfs.DeleteBlockRequest{
			Block: block,
		},
	)
	return sanitizeErr(err)
}

// InspectBlock returns info about a specific Block.
func (c APIClient) InspectBlock(hash string) (*pfs.BlockInfo, error) {
	blockInfo, err := c.BlockAPIClient.InspectBlock(
		c.ctx(),
		&pfs.InspectBlockRequest{
			Block: NewBlock(hash),
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return blockInfo, nil
}

// ListBlock returns info about all Blocks.
func (c APIClient) ListBlock() ([]*pfs.BlockInfo, error) {
	blockInfos, err := c.BlockAPIClient.ListBlock(
		c.ctx(),
		&pfs.ListBlockRequest{},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return blockInfos.BlockInfo, nil
}

// PutFileWriter writes a file to PFS.
// NOTE: PutFileWriter returns an io.WriteCloser you must call Close on it when
// you are done writing.
func (c APIClient) PutFileWriter(repoName string, commitID string, path string, delimiter pfs.Delimiter) (io.WriteCloser, error) {
	return c.newPutFileWriteCloser(repoName, commitID, path, delimiter)
}

// PutFile writes a file to PFS from a reader.
func (c APIClient) PutFile(repoName string, commitID string, path string, reader io.Reader) (_ int, retErr error) {
	return c.PutFileWithDelimiter(repoName, commitID, path, pfs.Delimiter_LINE, reader)
}

//PutFileWithDelimiter writes a file to PFS from a reader
// delimiter is used to tell PFS how to break the input into blocks
func (c APIClient) PutFileWithDelimiter(repoName string, commitID string, path string, delimiter pfs.Delimiter, reader io.Reader) (_ int, retErr error) {
	writer, err := c.PutFileWriter(repoName, commitID, path, delimiter)
	if err != nil {
		return 0, sanitizeErr(err)
	}
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	written, err := io.Copy(writer, reader)
	return int(written), err
}

// PutFileURL puts a file using the content found at a URL.
// The URL is sent to the server which performs the request.
// recursive allow for recursive scraping of some types URLs for example on s3:// urls.
func (c APIClient) PutFileURL(repoName string, commitID string, path string, url string, recursive bool) (retErr error) {
	putFileClient, err := c.PfsAPIClient.PutFile(c.ctx())
	if err != nil {
		return sanitizeErr(err)
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = sanitizeErr(err)
		}
	}()
	if err := putFileClient.Send(&pfs.PutFileRequest{
		File:      NewFile(repoName, commitID, path),
		FileType:  pfs.FileType_FILE_TYPE_REGULAR,
		Url:       url,
		Recursive: recursive,
	}); err != nil {
		return sanitizeErr(err)
	}
	return nil
}

// GetFile returns the contents of a file at a specific Commit.
// offset specifies a number of bytes that should be skipped in the beginning of the file.
// size limits the total amount of data returned, note you will get fewer bytes
// than size if you pass a value larger than the size of the file.
// If size is set to 0 then all of the data will be returned.
func (c APIClient) GetFile(repoName string, commitID string, path string, offset int64, size int64, writer io.Writer) error {
	return c.getFile(repoName, commitID, path, offset, size, writer)
}

func (c APIClient) getFile(repoName string, commitID string, path string, offset int64, size int64, writer io.Writer) error {
	apiGetFileClient, err := c.PfsAPIClient.GetFile(
		c.ctx(),
		&pfs.GetFileRequest{
			File:        NewFile(repoName, commitID, path),
			OffsetBytes: offset,
			SizeBytes:   size,
		},
	)
	if err != nil {
		return sanitizeErr(err)
	}
	if err := grpcutil.WriteFromStreamingBytesClient(apiGetFileClient, writer); err != nil {
		return sanitizeErr(err)
	}
	return nil
}

// InspectFile returns info about a specific file.
func (c APIClient) InspectFile(repoName string, commitID string, path string) (*pfs.FileInfo, error) {
	return c.inspectFile(repoName, commitID, path)
}

func (c APIClient) inspectFile(repoName string, commitID string, path string) (*pfs.FileInfo, error) {
	fileInfo, err := c.PfsAPIClient.InspectFile(
		c.ctx(),
		&pfs.InspectFileRequest{
			File: NewFile(repoName, commitID, path),
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return fileInfo, nil
}

// ListFile returns info about all files in a Commit.
func (c APIClient) ListFile(repoName string, commitID string, path string) ([]*pfs.FileInfo, error) {
	fileInfos, err := c.PfsAPIClient.ListFile(
		c.ctx(),
		&pfs.ListFileRequest{
			File: NewFile(repoName, commitID, path),
		},
	)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	return fileInfos.FileInfo, nil
}

// WalkFn is the type of the function called for each file in Walk.
// Returning a non-nil error from WalkFn will result in Walk aborting and
// returning said error.
type WalkFn func(*pfs.FileInfo) error

// Walk walks the pfs filesystem rooted at path. walkFn will be called for each
// file found under path, this includes both regular files and directories.
func (c APIClient) Walk(repoName string, commitID string, path string, walkFn WalkFn) error {
	fileInfos, err := c.ListFile(repoName, commitID, path)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		if err := walkFn(fileInfo); err != nil {
			return err
		}
		if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
			if err := c.Walk(repoName, commitID, fileInfo.File.Path, walkFn); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteFile deletes a file from a Commit.
// DeleteFile leaves a tombstone in the Commit, assuming the file isn't written
// to later attempting to get the file from the finished commit will result in
// not found error.
// The file will of course remain intact in the Commit's parent.
func (c APIClient) DeleteFile(repoName string, commitID string, path string) error {
	_, err := c.PfsAPIClient.DeleteFile(
		c.ctx(),
		&pfs.DeleteFileRequest{
			File: NewFile(repoName, commitID, path),
		},
	)
	return err
}

// MakeDirectory creates a directory in PFS.
// Note directories are created implicitly by PutFile, so you technically never
// need this function unless you want to create an empty directory.
func (c APIClient) MakeDirectory(repoName string, commitID string, path string) (retErr error) {
	putFileClient, err := c.PfsAPIClient.PutFile(c.ctx())
	if err != nil {
		return sanitizeErr(err)
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = sanitizeErr(err)
		}
	}()
	return sanitizeErr(putFileClient.Send(
		&pfs.PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: pfs.FileType_FILE_TYPE_DIR,
		},
	))
}

type putFileWriteCloser struct {
	request       *pfs.PutFileRequest
	putFileClient pfs.API_PutFileClient
	sent          bool
}

func (c APIClient) newPutFileWriteCloser(repoName string, commitID string, path string, delimiter pfs.Delimiter) (*putFileWriteCloser, error) {
	putFileClient, err := c.PfsAPIClient.PutFile(c.ctx())
	if err != nil {
		return nil, err
	}
	return &putFileWriteCloser{
		request: &pfs.PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: pfs.FileType_FILE_TYPE_REGULAR,
		},
		putFileClient: putFileClient,
	}, nil
}

func (w *putFileWriteCloser) Write(p []byte) (int, error) {
	w.request.Value = p
	if err := w.putFileClient.Send(w.request); err != nil {
		return 0, sanitizeErr(err)
	}
	w.sent = true
	w.request.Value = nil
	// File is only needed on the first request
	w.request.File = nil
	return len(p), nil
}

func (w *putFileWriteCloser) Close() error {
	// we always send at least one request, otherwise it's impossible to create
	// an empty file
	if !w.sent {
		if err := w.putFileClient.Send(w.request); err != nil {
			return err
		}
	}
	_, err := w.putFileClient.CloseAndRecv()
	return sanitizeErr(err)
}

type putBlockWriteCloser struct {
	request        *pfs.PutBlockRequest
	putBlockClient pfs.BlockAPI_PutBlockClient
	blockRefs      *pfs.BlockRefs
}

func (c APIClient) newPutBlockWriteCloser(delimiter pfs.Delimiter) (*putBlockWriteCloser, error) {
	putBlockClient, err := c.BlockAPIClient.PutBlock(c.ctx())
	if err != nil {
		return nil, err
	}
	return &putBlockWriteCloser{
		request: &pfs.PutBlockRequest{
			Delimiter: delimiter,
		},
		putBlockClient: putBlockClient,
		blockRefs:      &pfs.BlockRefs{},
	}, nil
}

func (w *putBlockWriteCloser) Write(p []byte) (int, error) {
	w.request.Value = p
	if err := w.putBlockClient.Send(w.request); err != nil {
		return 0, sanitizeErr(err)
	}
	return len(p), nil
}

func (w *putBlockWriteCloser) Close() error {
	var err error
	w.blockRefs, err = w.putBlockClient.CloseAndRecv()
	return sanitizeErr(err)
}

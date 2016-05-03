package client

import (
	"io"
	"math"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

func NewRepo(repoName string) *pfs.Repo {
	return &pfs.Repo{Name: repoName}
}

func NewCommit(repoName string, commitID string) *pfs.Commit {
	return &pfs.Commit{
		Repo: NewRepo(repoName),
		ID:   commitID,
	}
}

func NewFile(repoName string, commitID string, path string) *pfs.File {
	return &pfs.File{
		Commit: NewCommit(repoName, commitID),
		Path:   path,
	}
}

func NewBlock(hash string) *pfs.Block {
	return &pfs.Block{
		Hash: hash,
	}
}

func NewDiff(repoName string, commitID string, shard uint64) *pfs.Diff {
	return &pfs.Diff{
		Commit: NewCommit(repoName, commitID),
		Shard:  shard,
	}
}

const (
	NONE  = pfs.CommitType_COMMIT_TYPE_NONE
	READ  = pfs.CommitType_COMMIT_TYPE_READ
	WRITE = pfs.CommitType_COMMIT_TYPE_WRITE
)

// CreateRepo creates a new Repo object in pfs with the given name. Repos are
// the top level data object in pfs and should be used to store data of a
// similar type. For example rather than having a single Repo for an entire
// project you might have seperate Repos for logs, metrics, database dumps etc.
func (c APIClient) CreateRepo(repoName string) error {
	_, err := c.PfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

// InspectRepo returns info about a specific Repo.
func (c APIClient) InspectRepo(repoName string) (*pfs.RepoInfo, error) {
	repoInfo, err := c.PfsAPIClient.InspectRepo(
		context.Background(),
		&pfs.InspectRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

// ListRepo returns info about all Repos.
func (c APIClient) ListRepo() ([]*pfs.RepoInfo, error) {
	repoInfos, err := c.PfsAPIClient.ListRepo(
		context.Background(),
		&pfs.ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

// DeleteRepo deletes a repo and reclaims the storage space it was using. Note
// that as of 1.0 we do not reclaim the blocks that the Repo was referencing,
// this is because they may also be referenced by other Repos and deleting them
// would make those Repos inaccessible. This will be resolved in later
// versions.
func (c APIClient) DeleteRepo(repoName string) error {
	_, err := c.PfsAPIClient.DeleteRepo(
		context.Background(),
		&pfs.DeleteRepoRequest{
			Repo: NewRepo(repoName),
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
func (c APIClient) StartCommit(repoName string, parentCommit string, branch string) (*pfs.Commit, error) {
	commit, err := c.PfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Repo:     NewRepo(repoName),
			ParentID: parentCommit,
			Branch:   branch,
		},
	)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

// FinishCommit ends the process of committing data to a Repo and persists the
// Commit. Once a Commit is finished the data becomes immutable and future
// attempts to write to it with PutFile will error.
func (c APIClient) FinishCommit(repoName string, commitID string) error {
	_, err := c.PfsAPIClient.FinishCommit(
		context.Background(),
		&pfs.FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

// CancelCommit ends the process of committing data to a repo. It differs from
// FinishCommit in that the Commit will not be used as a source for downstream
// pipelines. CancelCommit is used primarily by PPS for the output commits of
// errant jobs.
func (c APIClient) CancelCommit(repoName string, commitID string) error {
	_, err := c.PfsAPIClient.FinishCommit(
		context.Background(),
		&pfs.FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
			Cancel: true,
		},
	)
	return err
}

// InspectCommit returns info about a specific Commit.
func (c APIClient) InspectCommit(repoName string, commitID string) (*pfs.CommitInfo, error) {
	commitInfo, err := c.PfsAPIClient.InspectCommit(
		context.Background(),
		&pfs.InspectCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfo, nil
}

// ListCommit returns info about multiple commits.
// repoNames defines a set of Repos to consider commits from, if repoNames is left
// nil or empty then the result will be empty.
// fromCommitIDs lets you get info about Commits that occurred after this
// set of commits.
// commitType specifies the type of commit you want returned, normally READ is the most useful option
// block, when set to true, will cause ListCommit to block until at least 1 new CommitInfo is available.
// Using fromCommitIDs and block you can get subscription semantics from ListCommit.
// all, when set to true, will cause ListCommit to return cancelled commits as well.
func (c APIClient) ListCommit(repoNames []string, fromCommitIDs []string,
	commitType pfs.CommitType, block bool, all bool) ([]*pfs.CommitInfo, error) {
	var repos []*pfs.Repo
	for _, repoName := range repoNames {
		repos = append(repos, &pfs.Repo{Name: repoName})
	}
	var fromCommits []*pfs.Commit
	for i, fromCommitID := range fromCommitIDs {
		fromCommits = append(fromCommits, &pfs.Commit{
			Repo: repos[i],
			ID:   fromCommitID,
		})
	}
	commitInfos, err := c.PfsAPIClient.ListCommit(
		context.Background(),
		&pfs.ListCommitRequest{
			Repo:       repos,
			FromCommit: fromCommits,
			Block:      block,
			All:        all,
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

// ListBranch lists the active branches on a Repo.
func (c APIClient) ListBranch(repoName string) ([]*pfs.CommitInfo, error) {
	commitInfos, err := c.PfsAPIClient.ListBranch(
		context.Background(),
		&pfs.ListBranchRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

// DeleteCommit deletes a commit.
// Note it is currently not implemented.
func (c APIClient) DeleteCommit(repoName string, commitID string) error {
	_, err := c.PfsAPIClient.DeleteCommit(
		context.Background(),
		&pfs.DeleteCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

// PutBlock takes a reader and splits the data in it into blocks.
// Blocks are guaranteed to be new line delimited.
// Blocks are content addressed and are thus identified by hashes of the content.
// NOTE: this is lower level function that's used internally and might not be
// useful to users.
func (c APIClient) PutBlock(reader io.Reader) (*pfs.BlockRefs, error) {
	putBlockClient, err := c.BlockAPIClient.PutBlock(context.Background())
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(protostream.NewStreamingBytesWriter(putBlockClient), reader); err != nil {
		return nil, err
	}
	return putBlockClient.CloseAndRecv()
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
		context.Background(),
		&pfs.GetBlockRequest{
			Block:       NewBlock(hash),
			OffsetBytes: offset,
			SizeBytes:   size,
		},
	)
	if err != nil {
		return nil, err
	}
	return protostream.NewStreamingBytesReader(apiGetBlockClient), nil
}

// DeleteBlock deletes a block from the block store.
// NOTE: this is lower level function that's used internally and might not be
// useful to users.
func (c APIClient) DeleteBlock(block *pfs.Block) error {
	_, err := c.BlockAPIClient.DeleteBlock(
		context.Background(),
		&pfs.DeleteBlockRequest{
			Block: block,
		},
	)

	return err
}

// InspectBlock returns info about a specific Block.
func (c APIClient) InspectBlock(hash string) (*pfs.BlockInfo, error) {
	blockInfo, err := c.BlockAPIClient.InspectBlock(
		context.Background(),
		&pfs.InspectBlockRequest{
			Block: NewBlock(hash),
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

// ListBlock returns info about all Blocks.
func (c APIClient) ListBlock() ([]*pfs.BlockInfo, error) {
	blockInfos, err := c.BlockAPIClient.ListBlock(
		context.Background(),
		&pfs.ListBlockRequest{},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

// PutFileWriter writes a file to PFS.
// handle is used to perform multiple writes that are guaranteed to wind up
// contiguous in the final file. It may be safely left empty and likely won't
// be needed in most use cases.
// NOTE: PutFileWriter returns an io.WriteCloser you must call Close on it when
// you are done writing.
func (c APIClient) PutFileWriter(repoName string, commitID string, path string, handle string) (io.WriteCloser, error) {
	return c.newPutFileWriteCloser(repoName, commitID, path, handle)
}

// PutFile writes a file to PFS from a reader.
func (c APIClient) PutFile(repoName string, commitID string, path string, reader io.Reader) (_ int, retErr error) {
	writer, err := c.PutFileWriter(repoName, commitID, path, "")
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	written, err := io.Copy(writer, reader)
	return int(written), err
}

// GetFile returns the contents of a file at a specific Commit.
// offset specifies a number of bytes that should be skipped in the beginning of the file.
// size limits the total amount of data returned, note you will get fewer bytes
// than size if you pass a value larger than the size of the file.
// If size is set to 0 then all of the data will be returned.
// fromCommitID lets you get only the data which was added after this Commit.
// shard allows you to downsample the data, returning only a subset of the
// blocks in the file. shard may be left nil in which case the entire file will be returned
func (c APIClient) GetFile(repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *pfs.Shard, writer io.Writer) error {
	return c.getFile(repoName, commitID, path, offset, size, fromCommitID, shard, false, writer)
}

func (c APIClient) GetFileUnsafe(repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *pfs.Shard, writer io.Writer) error {
	return c.getFile(repoName, commitID, path, offset, size, fromCommitID, shard, true, writer)
}

func (c APIClient) getFile(repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *pfs.Shard, unsafe bool, writer io.Writer) error {
	if size == 0 {
		size = math.MaxInt64
	}
	apiGetFileClient, err := c.PfsAPIClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			File:        NewFile(repoName, commitID, path),
			Shard:       shard,
			OffsetBytes: offset,
			SizeBytes:   size,
			FromCommit:  newFromCommit(repoName, fromCommitID),
			Unsafe:      unsafe,
		},
	)
	if err != nil {
		return err
	}
	if err := protostream.WriteFromStreamingBytesClient(apiGetFileClient, writer); err != nil {
		return err
	}
	return nil
}

// InspectFile returns info about a specific file.  fromCommitID lets you get
// only info which was added after this Commit.  shard allows you to downsample
// the data, returning info about only a subset of the blocks in the file.
// shard may be left nil in which case info about the entire file will be
// returned
func (c APIClient) InspectFile(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard) (*pfs.FileInfo, error) {
	return c.inspectFile(repoName, commitID, path, fromCommitID, shard, false)
}

func (c APIClient) InspectFileUnsafe(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard) (*pfs.FileInfo, error) {
	return c.inspectFile(repoName, commitID, path, fromCommitID, shard, true)
}

func (c APIClient) inspectFile(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard, unsafe bool) (*pfs.FileInfo, error) {
	fileInfo, err := c.PfsAPIClient.InspectFile(
		context.Background(),
		&pfs.InspectFileRequest{
			File:       NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: newFromCommit(repoName, fromCommitID),
			Unsafe:     unsafe,
		},
	)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

// ListFile returns info about all files in a Commit.
// fromCommitID lets you get only info which was added after this Commit.
// shard allows you to downsample the data, returning info about only a subset
// of the blocks in the files or only a subset of files. shard may be left nil
// in which case info about all the files and all the blocks in those files
// will be returned.
// recurse causes ListFile to accurately report the size of data stored in directories, it makes the call more expensive
func (c APIClient) ListFile(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard, recurse bool) ([]*pfs.FileInfo, error) {
	return c.listFile(repoName, commitID, path, fromCommitID, shard, recurse, false)
}

func (c APIClient) ListFileUnsafe(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard, recurse bool) ([]*pfs.FileInfo, error) {
	return c.listFile(repoName, commitID, path, fromCommitID, shard, recurse, true)
}

func (c APIClient) listFile(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard, recurse bool, unsafe bool) ([]*pfs.FileInfo, error) {
	fileInfos, err := c.PfsAPIClient.ListFile(
		context.Background(),
		&pfs.ListFileRequest{
			File:       NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: newFromCommit(repoName, fromCommitID),
			Recurse:    recurse,
			Unsafe:     unsafe,
		},
	)
	if err != nil {
		return nil, err
	}
	return fileInfos.FileInfo, nil
}

// DeleteFile deletes a file from a Commit.
// DeleteFile leaves a tombstone in the Commit, assuming the file isn't written
// to later attempting to get the file from the finished commit will result in
// not found error.
// The file will of course remain intact in the Commit's parent.
func (c APIClient) DeleteFile(repoName string, commitID string, path string) error {
	_, err := c.PfsAPIClient.DeleteFile(
		context.Background(),
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
	putFileClient, err := c.PfsAPIClient.PutFile(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return putFileClient.Send(
		&pfs.PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: pfs.FileType_FILE_TYPE_DIR,
		},
	)
}

type putFileWriteCloser struct {
	request       *pfs.PutFileRequest
	putFileClient pfs.API_PutFileClient
}

func (c APIClient) newPutFileWriteCloser(repoName string, commitID string, path string, handle string) (*putFileWriteCloser, error) {
	putFileClient, err := c.PfsAPIClient.PutFile(context.Background())
	if err != nil {
		return nil, err
	}
	return &putFileWriteCloser{
		request: &pfs.PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: pfs.FileType_FILE_TYPE_REGULAR,
			Handle:   handle,
		},
		putFileClient: putFileClient,
	}, nil
}

func (w *putFileWriteCloser) Write(p []byte) (int, error) {
	w.request.Value = p
	if err := w.putFileClient.Send(w.request); err != nil {
		return 0, err
	}
	// File is only needed on the first request
	w.request.File = nil
	return len(p), nil
}

func (w *putFileWriteCloser) Close() error {
	_, err := w.putFileClient.CloseAndRecv()
	return err
}

func newFromCommit(repoName string, fromCommitID string) *pfs.Commit {
	if fromCommitID != "" {
		return NewCommit(repoName, fromCommitID)
	}
	return nil
}

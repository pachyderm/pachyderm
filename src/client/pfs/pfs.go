package pfs

import (
	"io"
	"math"

	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

const chunkSize = 1024 * 1024

func NewRepo(repoName string) *Repo {
	return &Repo{Name: repoName}
}

func NewCommit(repoName string, commitID string) *Commit {
	return &Commit{
		Repo: NewRepo(repoName),
		ID:   commitID,
	}
}

func NewFile(repoName string, commitID string, path string) *File {
	return &File{
		Commit: NewCommit(repoName, commitID),
		Path:   path,
	}
}

func NewBlock(hash string) *Block {
	return &Block{
		Hash: hash,
	}
}

func NewDiff(repoName string, commitID string, shard uint64) *Diff {
	return &Diff{
		Commit: NewCommit(repoName, commitID),
		Shard:  shard,
	}
}

// CreateRepo creates a new Repo object in pfs with the given name. Repos are
// the top level data object in pfs and should be used to store data of a
// similar type. For example rather than having a single Repo for an entire
// project you might have seperate Repos for logs, metrics, database dumps etc.
func CreateRepo(apiClient APIClient, repoName string) error {
	_, err := apiClient.CreateRepo(
		context.Background(),
		&CreateRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

// InspectRepo returns info about a specific Repo.
func InspectRepo(apiClient APIClient, repoName string) (*RepoInfo, error) {
	repoInfo, err := apiClient.InspectRepo(
		context.Background(),
		&InspectRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

// ListRepo returns info about all Repos.
func ListRepo(apiClient APIClient) ([]*RepoInfo, error) {
	repoInfos, err := apiClient.ListRepo(
		context.Background(),
		&ListRepoRequest{},
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
func DeleteRepo(apiClient APIClient, repoName string) error {
	_, err := apiClient.DeleteRepo(
		context.Background(),
		&DeleteRepoRequest{
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
func StartCommit(apiClient APIClient, repoName string, parentCommit string, branch string) (*Commit, error) {
	commit, err := apiClient.StartCommit(
		context.Background(),
		&StartCommitRequest{
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
func FinishCommit(apiClient APIClient, repoName string, commitID string) error {
	_, err := apiClient.FinishCommit(
		context.Background(),
		&FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

// CancelCommit ends the process of committing data to a repo. It differs from
// FinishCommit in that the Commit will not be used as a source for downstream
// pipelines. CancelCommit is used primarily by PPS for the output commits of
// errant jobs.
func CancelCommit(apiClient APIClient, repoName string, commitID string) error {
	_, err := apiClient.FinishCommit(
		context.Background(),
		&FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
			Cancel: true,
		},
	)
	return err
}

// InspectCommit returns info about a specific Commit.
func InspectCommit(apiClient APIClient, repoName string, commitID string) (*CommitInfo, error) {
	commitInfo, err := apiClient.InspectCommit(
		context.Background(),
		&InspectCommitRequest{
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
// block, when set to true, will cause ListCommit to block until at least 1 new CommitInfo is available.
// Using fromCommitIDs and block you can get subscription semantics from ListCommit.
// all, when set to true, will cause ListCommit to return cancelled commits as well.
func ListCommit(apiClient APIClient, repoNames []string, fromCommitIDs []string, block bool, all bool) ([]*CommitInfo, error) {
	var repos []*Repo
	for _, repoName := range repoNames {
		repos = append(repos, &Repo{Name: repoName})
	}
	var fromCommits []*Commit
	for i, fromCommitID := range fromCommitIDs {
		fromCommits = append(fromCommits, &Commit{
			Repo: repos[i],
			ID:   fromCommitID,
		})
	}
	commitInfos, err := apiClient.ListCommit(
		context.Background(),
		&ListCommitRequest{
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
func ListBranch(apiClient APIClient, repoName string) ([]*CommitInfo, error) {
	commitInfos, err := apiClient.ListBranch(
		context.Background(),
		&ListBranchRequest{
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
func DeleteCommit(apiClient APIClient, repoName string, commitID string) error {
	_, err := apiClient.DeleteCommit(
		context.Background(),
		&DeleteCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

// PutBlock takes a reader and splits it chunks the data in it into blocks.
// Blocks are guaranteed to be new line delimited.
// Blocks are content addressed and are thus identified by hashes of the content.
// NOTE: this is lower level function that's used internally and might not be
// useful to users.
func PutBlock(apiClient BlockAPIClient, reader io.Reader) (*BlockRefs, error) {
	putBlockClient, err := apiClient.PutBlock(context.Background())
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
func GetBlock(apiClient BlockAPIClient, hash string, offset uint64, size uint64) (io.Reader, error) {
	apiGetBlockClient, err := apiClient.GetBlock(
		context.Background(),
		&GetBlockRequest{
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
func DeleteBlock(apiClient BlockAPIClient, block *Block) error {
	_, err := apiClient.DeleteBlock(
		context.Background(),
		&DeleteBlockRequest{
			Block: block,
		},
	)

	return err
}

// InspectBlock returns info about a specific Block.
func InspectBlock(apiClient BlockAPIClient, hash string) (*BlockInfo, error) {
	blockInfo, err := apiClient.InspectBlock(
		context.Background(),
		&InspectBlockRequest{
			Block: NewBlock(hash),
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

// ListBlock returns info about all Blocks.
func ListBlock(apiClient BlockAPIClient) ([]*BlockInfo, error) {
	blockInfos, err := apiClient.ListBlock(
		context.Background(),
		&ListBlockRequest{},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

// PutFileWriter writes a file to PFS.
// handle is used to perform multiple writes that are guaranteed wind up
// contiguous in the final file. It may be safely left empty and likely won't
// be needed in most use cases.
// NOTE: PutFileWriter returns an io.WriteCloser you must call Close on it when
// you are done writing.
func PutFileWriter(apiClient APIClient, repoName string, commitID string, path string, handle string) (io.WriteCloser, error) {
	return newPutFileWriteCloser(apiClient, repoName, commitID, path, handle)
}

// PutFile writes a file to PFS from a reader.
func PutFile(apiClient APIClient, repoName string, commitID string, path string, reader io.Reader) (_ int, retErr error) {
	writer, err := PutFileWriter(apiClient, repoName, commitID, path, "")
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
func GetFile(apiClient APIClient, repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *Shard, writer io.Writer) error {
	if size == 0 {
		size = math.MaxInt64
	}
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&GetFileRequest{
			File:        NewFile(repoName, commitID, path),
			Shard:       shard,
			OffsetBytes: offset,
			SizeBytes:   size,
			FromCommit:  newFromCommit(repoName, fromCommitID),
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
func InspectFile(apiClient APIClient, repoName string, commitID string, path string, fromCommitID string, shard *Shard) (*FileInfo, error) {
	fileInfo, err := apiClient.InspectFile(
		context.Background(),
		&InspectFileRequest{
			File:       NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: newFromCommit(repoName, fromCommitID),
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
func ListFile(apiClient APIClient, repoName string, commitID string, path string, fromCommitID string, shard *Shard, recurse bool) ([]*FileInfo, error) {
	fileInfos, err := apiClient.ListFile(
		context.Background(),
		&ListFileRequest{
			File:       NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: newFromCommit(repoName, fromCommitID),
			Recurse:    recurse,
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
func DeleteFile(apiClient APIClient, repoName string, commitID string, path string) error {
	_, err := apiClient.DeleteFile(
		context.Background(),
		&DeleteFileRequest{
			File: NewFile(repoName, commitID, path),
		},
	)
	return err
}

// MakeDirectory creates a directory in PFS.
// Note directories are created implicitly by PutFile, so you technically never
// need this function unless you want to create an empty directory.
func MakeDirectory(apiClient APIClient, repoName string, commitID string, path string) (retErr error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return putFileClient.Send(
		&PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: FileType_FILE_TYPE_DIR,
		},
	)
}

type putFileWriteCloser struct {
	request       *PutFileRequest
	putFileClient API_PutFileClient
}

func newPutFileWriteCloser(apiClient APIClient, repoName string, commitID string, path string, handle string) (*putFileWriteCloser, error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return nil, err
	}
	return &putFileWriteCloser{
		request: &PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: FileType_FILE_TYPE_REGULAR,
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

func newFromCommit(repoName string, fromCommitID string) *Commit {
	if fromCommitID != "" {
		return NewCommit(repoName, fromCommitID)
	}
	return nil
}

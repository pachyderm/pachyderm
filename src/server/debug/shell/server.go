//nolint:wrapcheck
package shell

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func NewDumpServer(filePath string, port uint16) *debugDump {
	ctx := pctx.Background("dumpserver")
	mock, err := testpachd.NewMockPachd(ctx, port)
	if err != nil {
		panic(err)
	}

	d := &debugDump{
		path: filePath,
		mock: mock,
	}

	mock.PFS.ListRepo.Use(d.listRepo)
	mock.PFS.ListCommit.Use(d.listCommit)
	mock.PFS.ListCommitSet.Use(d.listCommitSet)
	mock.PFS.ListBranch.Use(d.listBranch)
	mock.PFS.InspectRepo.Use(d.inspectRepo)
	mock.PFS.InspectCommit.Use(d.inspectCommit)
	mock.PFS.InspectCommitSet.Use(d.inspectCommitSet)
	mock.PFS.InspectBranch.Use(d.inspectBranch)

	mock.PPS.ListPipeline.Use(d.listPipeline)
	mock.PPS.ListJob.Use(d.listJob)
	mock.PPS.ListJobSet.Use(d.listJobSet)
	mock.PPS.InspectPipeline.Use(d.inspectPipeline)
	mock.PPS.InspectJob.Use(d.inspectJob)
	mock.PPS.InspectJobSet.Use(d.inspectJobSet)

	mock.PPS.GetLogs.Use(d.getLogs)
	mock.Version.GetVersion.Use(d.getVersion)

	// report that enterprise and auth are disabled, to support console
	mock.Enterprise.GetState.Use(func(_ context.Context, _ *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error) {
		return &enterprise.GetStateResponse{State: enterprise.State_NONE}, nil
	})
	mock.Auth.WhoAmI.Use(func(_ context.Context, _ *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
		return nil, auth.ErrNotActivated
	})
	mock.Admin.InspectCluster.Use(func(_ context.Context, _ *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{Id: "debug", VersionWarningsOk: true}, nil
	})

	return d
}

// TODO: caching?
type debugDump struct {
	path string
	mock *testpachd.MockPachd
}

func (d *debugDump) Address() string {
	return d.mock.Addr.String()
}

func (d *debugDump) globTar(glob string, cb func(string, io.Reader) error) error {
	g := globlib.MustCompile(glob, '/')
	info, err := os.Stat(d.path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := filepath.WalkDir(d.path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(d.path, path)
			if err != nil {
				return err
			}
			// ToSlash to support windows
			rel = filepath.ToSlash(rel)
			if !g.Match(rel) {
				return nil
			}
			contents, err := os.Open(path)
			if err != nil {
				return err
			}
			defer contents.Close()
			return cb(rel, contents)
		}); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
		return nil
	}

	file, err := os.Open(d.path)
	if err != nil {
		return err
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if !g.Match(hdr.Name) {
			continue
		}
		if err := cb(hdr.Name, tr); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				break
			}
			return err
		}
	}
	return nil
}

func (d *debugDump) globTarProtos(glob string, template proto.Message, onProto func() error) (bool, error) {
	var found bool
	err := d.globTar(glob, func(path string, r io.Reader) error {
		found = true
		dec := protoutil.NewProtoJSONDecoder(r, protojson.UnmarshalOptions{})
		for dec.More() {
			if err := dec.UnmarshalNext(template); err != nil {
				return err
			}
			if err := onProto(); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					break
				}
				return err
			}
		}
		return nil
	})
	return found, err
}

const commitPatternFormatString = `{{source,input}-repos,pipelines}/+(%s/)commits{,\.json}`
const jobPatternFormatString = `pipelines/+(%s/)jobs{,\.json}`
const pipelineSpecPatternFormatString = `pipelines/+(%s/)spec{,\.json}`

func (d *debugDump) listCommit(req *pfs.ListCommitRequest, srv pfs.API_ListCommitServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, req.Repo.Name)
	var ci pfs.CommitInfo
	if found, err := d.globTarProtos(glob, &ci, func() error {
		if req.All ||
			(req.OriginKind != pfs.OriginKind_ORIGIN_KIND_UNKNOWN && ci.Origin.Kind == req.OriginKind) {
			return srv.Send(&ci)
		}
		return nil
	}); err != nil {
		return err
	} else if !found {
		return pfsserver.ErrRepoNotFound{Repo: req.Repo}
	}
	return nil
}

func (d *debugDump) inspectCommit(_ context.Context, req *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	targetID, ancestors, err := ancestry.Parse(req.Commit.Id)
	var targetBranch string
	if err != nil {
		return nil, err
	}
	if targetID == "" {
		targetBranch = req.Commit.Branch.Name
	} else if !uuid.IsUUIDWithoutDashes(targetID) {
		targetBranch = targetID
	}
	glob := fmt.Sprintf(commitPatternFormatString, req.Commit.Repo.Name)
	var info pfs.CommitInfo
	var foundCommit bool
	if found, err := d.globTarProtos(glob, &info, func() error {
		// check for match and traverse ancestry
		var match bool
		if targetBranch != "" {
			// assume the first we see on this branch is the current head
			match = info.Commit.Branch.Name == targetBranch
		} else {
			match = info.Commit.Id == targetID
		}
		if match {
			if ancestors == 0 {
				foundCommit = true
				return errutil.ErrBreak
			} else {
				// this will fail for negative/future ancestry, but I doubt people use that anyway
				ancestors--
				targetBranch = ""
				targetID = info.ParentCommit.Id
			}
		}
		return nil
	}); err != nil {
		return nil, err
	} else if !found {
		return nil, pfsserver.ErrRepoNotFound{Repo: req.Commit.Repo}
	}
	if !foundCommit {
		return nil, pfsserver.ErrCommitNotFound{Commit: req.Commit}
	}
	return &info, nil
}

func pathnameToRepo(pathname string) *pfs.Repo {
	var (
		part                       string
		reversedProjects, projects []string
	)
	dir, _ := filepath.Split(pathname)
	dir = filepath.Dir(dir)
	dir, repoName := filepath.Split(dir)
	dir = filepath.Dir(dir)
	for dir != "." {
		dir, part = filepath.Split(dir)
		dir = filepath.Dir(dir)
		reversedProjects = append(reversedProjects, part)
	}
	reversedProjects = reversedProjects[:len(reversedProjects)-1]
	projects = make([]string, len(reversedProjects))
	for i := 0; i < len(reversedProjects); i++ {
		projects[len(projects)-1-i] = reversedProjects[i]
	}
	return client.NewRepo(path.Join(projects...), repoName)
}

func (d *debugDump) listRepo(req *pfs.ListRepoRequest, srv pfs.API_ListRepoServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, "*")
	return d.globTar(glob, func(pathname string, _ io.Reader) error {
		return srv.Send(&pfs.RepoInfo{
			Repo: pathnameToRepo(pathname),
		})
	})
}

func (d *debugDump) inspectRepo(_ context.Context, req *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	glob := fmt.Sprintf(commitPatternFormatString, req.Repo.Name)
	var info pfs.RepoInfo
	info.Repo = req.Repo
	seenBranches := make(map[string]struct{})
	var ci pfs.CommitInfo
	if found, err := d.globTarProtos(glob, &ci, func() error {
		if _, ok := seenBranches[ci.Commit.Branch.Name]; !ok {
			seenBranches[ci.Commit.Branch.Name] = struct{}{}
			info.Branches = append(info.Branches, ci.Commit.Branch)
		}
		// commits should be sorted in debug dump
		info.Created = ci.Started
		if info.SizeBytesUpperBound == 0 && ci.Commit.Branch.Name == "master" && ci.Finished != nil {
			info.SizeBytesUpperBound = ci.SizeBytesUpperBound
		}
		return nil
	}); err != nil {
		return nil, err
	} else if !found {
		return nil, pfsserver.ErrRepoNotFound{Repo: req.Repo}
	}
	return &info, nil
}

func (d *debugDump) listCommitSet(req *pfs.ListCommitSetRequest, srv pfs.API_ListCommitSetServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, "*")
	// TODO: streaming? presumably for a local tool memory use isn't really a concern
	var info pfs.CommitInfo
	commitMap := make(map[string][]*pfs.CommitInfo)
	latestMap := make(map[string]time.Time)
	if _, err := d.globTarProtos(glob, &info, func() error {
		commitMap[info.Commit.Id] = append(commitMap[info.Commit.Id], proto.Clone(&info).(*pfs.CommitInfo))
		asTime := info.Started.AsTime()
		if latestMap[info.Commit.Id].Before(asTime) {
			latestMap[info.Commit.Id] = asTime
		}
		return nil
	}); err != nil {
		return err
	}
	keys := make([]string, 0, len(commitMap))
	for id := range commitMap {
		keys = append(keys, id)
	}
	sort.Slice(keys, func(i, j int) bool {
		return latestMap[keys[i]].After(latestMap[keys[j]])
	})
	for _, id := range keys {
		if err := srv.Send(&pfs.CommitSetInfo{
			CommitSet: client.NewCommitSet(id),
			Commits:   commitMap[id],
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *debugDump) inspectCommitSet(req *pfs.InspectCommitSetRequest, srv pfs.API_InspectCommitSetServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, "*")
	var info pfs.CommitInfo
	_, err := d.globTarProtos(glob, &info, func() error {
		if info.Commit.Id != req.CommitSet.Id {
			return nil
		}
		return srv.Send(&info)
	})
	return err
}

func (d *debugDump) listBranch(req *pfs.ListBranchRequest, srv pfs.API_ListBranchServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, req.Repo.Name)
	var ci pfs.CommitInfo
	branchSet := make(map[string]struct{})
	// we don't currently dump branches, so iterate over the commits
	// and just return the first we find for each branch
	if found, err := d.globTarProtos(glob, &ci, func() error {
		if _, ok := branchSet[ci.Commit.Branch.Name]; ok {
			return nil
		}
		branchSet[ci.Commit.Branch.Name] = struct{}{}
		return srv.Send(&pfs.BranchInfo{
			Branch: ci.Commit.Branch,
			Head:   ci.Commit,
			//			DirectProvenance: ci.OldDirectProvenance,
		})
	}); err != nil {
		return err
	} else if !found {
		return pfsserver.ErrRepoNotFound{Repo: req.Repo}
	}
	return nil
}

func (d *debugDump) inspectBranch(_ context.Context, req *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	glob := fmt.Sprintf(commitPatternFormatString, req.Branch.Repo.Name)
	var ci pfs.CommitInfo
	var bi *pfs.BranchInfo
	if found, err := d.globTarProtos(glob, &ci, func() error {
		if ci.Commit.Branch.Name != req.Branch.Name {
			return nil
		}
		bi = &pfs.BranchInfo{
			Branch: ci.Commit.Branch,
			Head:   ci.Commit, // will be inaccurate if the head is moved to an old commit
			//			DirectProvenance: ci.OldDirectProvenance,
		}
		return errutil.ErrBreak
	}); err != nil {
		return nil, err
	} else if !found {
		return nil, pfsserver.ErrRepoNotFound{Repo: req.Branch.Repo}
	}
	if bi == nil {
		return nil, pfsserver.ErrBranchNotFound{Branch: req.Branch}
	}
	return bi, nil
}

func (d *debugDump) getLogs(req *pps.GetLogsRequest, srv pps.API_GetLogsServer) error {
	var glob string
	var plainText bool

	if req.Pipeline == nil && req.Job == nil {
		glob = "pachd/*/pachd/logs.txt"
		plainText = true
	} else {
		name := req.Pipeline.GetName()
		if name == "" {
			name = req.Job.GetPipeline().GetName()
		}
		if name == "" {
			return errors.New("must provide a pipeline name")
		}
		glob = fmt.Sprintf("pipelines/%s/pods/*/user/logs.txt", name)
	}
	return d.globTar(glob, func(_ string, r io.Reader) error {
		return ppsutil.FilterLogLines(req, r, plainText, srv.Send)
	})
}

func (d *debugDump) listJob(req *pps.ListJobRequest, srv pps.API_ListJobServer) error {
	glob := fmt.Sprintf(jobPatternFormatString, req.Pipeline.Name)
	var info pps.JobInfo
	if found, err := d.globTarProtos(glob, &info, func() error {
		return srv.Send(&info)
	}); err != nil {
		return err
	} else if !found {
		return ppsserver.ErrPipelineNotFound{Pipeline: req.Pipeline}
	}
	return nil
}

func (d *debugDump) inspectJob(_ context.Context, req *pps.InspectJobRequest) (*pps.JobInfo, error) {
	glob := fmt.Sprintf(jobPatternFormatString, req.Job.Pipeline.Name)
	var info pps.JobInfo
	var foundJob bool
	if found, err := d.globTarProtos(glob, &info, func() error {
		if proto.Equal(info.Job, req.Job) {
			foundJob = true
			return errutil.ErrBreak
		}
		return nil
	}); err != nil {
		return nil, err
	} else if !found {
		return nil, ppsserver.ErrPipelineNotFound{Pipeline: req.Job.Pipeline}
	}
	if !foundJob {
		return nil, fmt.Errorf("job %v not found", req.Job)
	}
	return &info, nil
}

func (d *debugDump) listPipeline(req *pps.ListPipelineRequest, srv pps.API_ListPipelineServer) error {
	target := "*"
	if req.Pipeline != nil {
		target = req.Pipeline.Name
	}
	glob := fmt.Sprintf(pipelineSpecPatternFormatString, target)
	var info pps.PipelineInfo
	var counter int
	var prevPipeline string
	_, err := d.globTarProtos(glob, &info, func() error {
		if info.Pipeline.Name != prevPipeline {
			prevPipeline = info.Pipeline.Name
			counter = 0
		}
		if err := srv.Send(&info); err != nil {
			return err
		}
		counter++
		if counter > int(req.History) && req.History >= 0 {
			return errutil.ErrBreak
		}
		return nil
	})
	return err
}

func (d *debugDump) inspectPipeline(_ context.Context, req *pps.InspectPipelineRequest) (*pps.PipelineInfo, error) {
	name, ancestors, err := ancestry.Parse(req.Pipeline.Name)
	if err != nil {
		return nil, err
	}
	glob := fmt.Sprintf(pipelineSpecPatternFormatString, name)
	var info pps.PipelineInfo
	if found, err := d.globTarProtos(glob, &info, func() error {
		if ancestors == 0 {
			return errutil.ErrBreak
		}
		ancestors--
		return nil
	}); err != nil {
		return nil, err
	} else if !found {
		return nil, ppsserver.ErrPipelineNotFound{Pipeline: req.Pipeline}
	}
	return &info, nil
}

func (d *debugDump) listJobSet(req *pps.ListJobSetRequest, srv pps.API_ListJobSetServer) error {
	glob := fmt.Sprintf(jobPatternFormatString, "*")
	// TODO: streaming? presumably for a local tool memory use isn't really a concern
	var info pps.JobInfo
	jobMap := make(map[string][]*pps.JobInfo)
	latestMap := make(map[string]time.Time)
	if _, err := d.globTarProtos(glob, &info, func() error {
		jobMap[info.Job.Id] = append(jobMap[info.Job.Id], proto.Clone(&info).(*pps.JobInfo))
		asTime := info.Started.AsTime()
		if latestMap[info.Job.Id].Before(asTime) {
			latestMap[info.Job.Id] = asTime
		}
		return nil
	}); err != nil {
		return err
	}
	keys := make([]string, 0, len(jobMap))
	for id := range jobMap {
		keys = append(keys, id)
	}
	sort.Slice(keys, func(i, j int) bool {
		return latestMap[keys[i]].After(latestMap[keys[j]])
	})
	for _, id := range keys {
		if err := srv.Send(&pps.JobSetInfo{
			JobSet: client.NewJobSet(id),
			Jobs:   jobMap[id],
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *debugDump) inspectJobSet(req *pps.InspectJobSetRequest, srv pps.API_InspectJobSetServer) error {
	glob := fmt.Sprintf(jobPatternFormatString, "*")
	var info pps.JobInfo
	_, err := d.globTarProtos(glob, &info, func() error {
		if info.Job.Id != req.JobSet.Id {
			return nil
		}
		return srv.Send(&info)
	})
	return err
}

func (d *debugDump) getVersion(context.Context, *emptypb.Empty) (*versionpb.Version, error) {
	var version *versionpb.Version
	err := d.globTar("pachd/*/pachd/version.txt", func(_ string, r io.Reader) error {
		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		parts := strings.Split(strings.TrimSpace(string(b)), ".")
		if len(parts) != 3 {
			return fmt.Errorf("bad version string %s", b)
		}
		last := strings.Split(parts[2], "-")
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		micro, err := strconv.Atoi(last[0])
		if err != nil {
			return err
		}
		version = &versionpb.Version{
			Major: uint32(major),
			Minor: uint32(minor),
			Micro: uint32(micro),
		}
		if len(last) > 1 {
			version.Additional = last[1]
		}
		return errutil.ErrBreak
	})
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, fmt.Errorf("no version file captured")
	}
	return version, nil
}

package shell

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

func NewDumpServer(filePath string) *debugDump {
	mock, err := testpachd.NewMockPachd(context.TODO())
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

	return d
}

// TODO: caching?
type debugDump struct {
	path string
	mock *testpachd.MockPachd
}

func (d *debugDump) globTar(glob string, cb func(string, io.Reader) error) error {
	g := globlib.MustCompile(glob, '/')
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
		if g.Match(hdr.Name) {
			if err := cb(hdr.Name, tr); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					break
				}
				return err
			}
		}
	}
	return nil
}

func (d *debugDump) globTarProtos(glob string, template proto.Message, onFile func(string) error, onProto func() error) error {
	return d.globTar(glob, func(path string, r io.Reader) error {
		if err := onFile(path); err != nil {
			return err
		}
		dec := json.NewDecoder(r)
		for dec.More() {
			if err := jsonpb.UnmarshalNext(dec, template); err != nil {
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
}

const commitPatternFormatString = `{{source,input}-repos,pipelines}/%s/commits{,\.json}`
const jobPatternFormatString = `pipelines/%s/jobs{,\.json}`
const pipelineSpecPatternFormatString = `pipelines/%s/spec{,\.json}`

func (d *debugDump) listCommit(req *pfs.ListCommitRequest, srv pfs.API_ListCommitServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, req.Repo.Name)
	var found bool
	var ci pfs.CommitInfo
	if err := d.globTarProtos(glob, &ci, func(_ string) error {
		found = true
		return nil
	}, func() error {
		if req.All ||
			(req.OriginKind != pfs.OriginKind_ORIGIN_KIND_UNKNOWN && ci.Origin.Kind == req.OriginKind) ||
			ci.Origin.Kind != pfs.OriginKind_ALIAS {
			return srv.Send(&ci)
		}
		return nil
	}); err != nil {
		return err
	}
	if !found {
		return pfsserver.ErrRepoNotFound{Repo: req.Repo}
	}
	return nil
}

func (d *debugDump) inspectCommit(_ context.Context, req *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	targetID, ancestors, err := ancestry.Parse(req.Commit.ID)
	var targetBranch string
	if err != nil {
		return nil, err
	}
	if targetID == "" {
		targetBranch = req.Commit.Branch.Name
	} else if !uuid.IsUUIDWithoutDashes(targetID) {
		targetBranch = targetID
	}
	glob := fmt.Sprintf(commitPatternFormatString, req.Commit.Branch.Repo.Name)
	var info pfs.CommitInfo
	var found bool
	if err := d.globTarProtos(glob, &info, func(_ string) error {
		found = true
		return nil
	}, func() error {
		// check for match and traverse ancestry
		var match bool
		if targetBranch != "" {
			// assume the first we see on this branch is the current head
			match = info.Commit.Branch.Name == targetBranch
		} else {
			match = info.Commit.ID == targetID
		}
		if match {
			if ancestors == 0 {
				return errutil.ErrBreak
			} else {
				// this will fail for negative/future ancestry, but I doubt people use that anyway
				ancestors--
				targetBranch = ""
				targetID = info.ParentCommit.ID
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if !found {
		return nil, pfsserver.ErrRepoNotFound{Repo: req.Commit.Branch.Repo}
	}
	if info.Commit == nil {
		return nil, pfsserver.ErrCommitNotFound{Commit: req.Commit}
	}
	return &info, nil
}

func (d *debugDump) listRepo(req *pfs.ListRepoRequest, srv pfs.API_ListRepoServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, "*")
	return d.globTar(glob, func(path string, _ io.Reader) error {
		parts := strings.Split(path, "/")
		return srv.Send(&pfs.RepoInfo{
			Repo: client.NewRepo(parts[1]),
		})
	})
}

func (d *debugDump) inspectRepo(_ context.Context, req *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	glob := fmt.Sprintf(commitPatternFormatString, req.Repo.Name)
	var info pfs.RepoInfo
	info.Repo = req.Repo
	seenBranches := make(map[string]struct{})
	var found bool
	var ci pfs.CommitInfo
	if err := d.globTarProtos(glob, &ci, func(_ string) error {
		found = true
		return nil
	}, func() error {
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
	}
	if !found {
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
	if err := d.globTarProtos(glob, &info, func(_ string) error { return nil }, func() error {
		commitMap[info.Commit.ID] = append(commitMap[info.Commit.ID], proto.Clone(&info).(*pfs.CommitInfo))
		asTime, _ := types.TimestampFromProto(info.Started)
		if latestMap[info.Commit.ID].Before(asTime) {
			latestMap[info.Commit.ID] = asTime
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
	return d.globTarProtos(glob, &info, func(_ string) error { return nil }, func() error {
		if info.Commit.ID != req.CommitSet.ID {
			return nil
		}
		return srv.Send(&info)
	})
}

func (d *debugDump) listBranch(req *pfs.ListBranchRequest, srv pfs.API_ListBranchServer) error {
	glob := fmt.Sprintf(commitPatternFormatString, req.Repo.Name)
	var ci pfs.CommitInfo
	var found bool
	branchSet := make(map[string]bool)
	err := d.globTarProtos(glob, &ci, func(_ string) error {
		found = true
		return nil
	}, func() error {
		if branchSet[ci.Commit.Branch.Name] {
			return nil
		}
		branchSet[ci.Commit.Branch.Name] = true
		return srv.Send(&pfs.BranchInfo{
			Branch:           ci.Commit.Branch,
			Head:             ci.Commit,
			DirectProvenance: ci.DirectProvenance,
		})
	})
	if err != nil {

	}
	if !found {
		return pfsserver.ErrRepoNotFound{Repo: req.Repo}
	}
	return nil
}

func (d *debugDump) inspectBranch(_ context.Context, req *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	glob := fmt.Sprintf(commitPatternFormatString, req.Branch.Repo.Name)
	var ci pfs.CommitInfo
	var bi *pfs.BranchInfo
	var found bool
	err := d.globTarProtos(glob, &ci, func(_ string) error {
		found = true
		return nil
	}, func() error {
		if ci.Commit.Branch.Name != req.Branch.Name {
			return nil
		}
		bi = &pfs.BranchInfo{
			Branch:           ci.Commit.Branch,
			Head:             ci.Commit, // will be inaccurate if the head is moved to an old commit
			DirectProvenance: ci.DirectProvenance,
		}
		return errutil.ErrBreak
	})
	if err != nil {
		return nil, err
	}
	if !found {
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
		glob = fmt.Sprintf("pipelines/%s/pods/*/user/logs.txt", req.Pipeline.Name)
	}
	return d.globTar(glob, func(_ string, r io.Reader) error {
		return ppsutil.FilterLogLines(req, r, plainText, srv.Send)
	})
}

func (d *debugDump) listJob(req *pps.ListJobRequest, srv pps.API_ListJobServer) error {
	glob := fmt.Sprintf(jobPatternFormatString, req.Pipeline.Name)
	var found bool
	var info pps.JobInfo
	if err := d.globTarProtos(glob, &info, func(_ string) error {
		found = true
		return nil
	}, func() error {
		return srv.Send(&info)
	}); err != nil {
		return err
	}
	if !found {
		return ppsserver.ErrPipelineNotFound{Pipeline: req.Pipeline}
	}
	return nil
}

func (d *debugDump) inspectJob(_ context.Context, req *pps.InspectJobRequest) (*pps.JobInfo, error) {
	glob := fmt.Sprintf(jobPatternFormatString, req.Job.Pipeline.Name)
	var info pps.JobInfo
	var found bool
	err := d.globTarProtos(glob, &info, func(_ string) error {
		found = true
		return nil
	}, func() error {
		if proto.Equal(info.Job, req.Job) {
			return errutil.ErrBreak
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ppsserver.ErrPipelineNotFound{Pipeline: req.Job.Pipeline}
	}
	if info.Job == nil {
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
	return d.globTarProtos(glob, &info, func(_ string) error {
		counter = 0
		return nil
	}, func() error {
		if err := srv.Send(&info); err != nil {
			return err
		}
		counter++
		if counter > int(req.History) && req.History >= 0 {
			return errutil.ErrBreak
		}
		return nil
	})
}

func (d *debugDump) inspectPipeline(_ context.Context, req *pps.InspectPipelineRequest) (*pps.PipelineInfo, error) {
	name, ancestors, err := ancestry.Parse(req.Pipeline.Name)
	if err != nil {
		return nil, err
	}
	glob := fmt.Sprintf(pipelineSpecPatternFormatString, name)
	var info pps.PipelineInfo
	var found bool
	if err := d.globTarProtos(glob, &info, func(_ string) error {
		found = true
		return nil
	}, func() error {
		if ancestors == 0 {
			return errutil.ErrBreak
		}
		ancestors--
		return nil
	}); err != nil {
		return nil, err
	}
	if !found {
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
	if err := d.globTarProtos(glob, &info, func(_ string) error { return nil }, func() error {
		jobMap[info.Job.ID] = append(jobMap[info.Job.ID], proto.Clone(&info).(*pps.JobInfo))
		asTime, _ := types.TimestampFromProto(info.Started)
		if latestMap[info.Job.ID].Before(asTime) {
			latestMap[info.Job.ID] = asTime
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
	return d.globTarProtos(glob, &info, func(_ string) error { return nil }, func() error {
		if info.Job.ID != req.JobSet.ID {
			return nil
		}
		return srv.Send(&info)
	})
}

func (d *debugDump) getVersion(context.Context, *types.Empty) (*versionpb.Version, error) {
	var version *versionpb.Version
	err := d.globTar("pachd/*/pachd/version.txt", func(_ string, r io.Reader) error {
		b, err := ioutil.ReadAll(r)
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

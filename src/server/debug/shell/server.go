package shell

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

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

	//mock.PPS.ListPipeline
	//mock.PPS.ListJob
	//mock.PPS.ListJobSet
	//mock.PPS.InspectPipeline
	//mock.PPS.InspectJob
	//mock.PPS.InspectJobSet

	mock.PPS.GetLogs.Use(d.getLogs)

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
	glob := fmt.Sprintf(commitPatternFormatString, req.Commit.Branch.Repo.Name)
	var info pfs.CommitInfo
	var found bool
	err := d.globTarProtos(glob, &info, func(_ string) error {
		found = true
		return nil
	}, func() error {
		if proto.Equal(info.Commit, req.Commit) {
			return errutil.ErrBreak
		}
		return nil
	})
	if err != nil {
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
	return d.globTar(glob, func(path string, r io.Reader) error {
		var ci pfs.CommitInfo
		dec := json.NewDecoder(r)
		for dec.More() {
			if err := jsonpb.UnmarshalNext(dec, &ci); err != nil {
				return err
			}
			if ci.Commit.ID != req.CommitSet.ID {
				continue
			}
			if err := srv.Send(&ci); err != nil {
				return err
			}
			break // ignore other branches
		}
		return nil
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
		glob = fmt.Sprintf("pipelines/%s/pods/*/logs.txt", req.Pipeline.Name)
	}
	return d.globTar(glob, func(_ string, r io.Reader) error {
		return ppsutil.FilterLogLines(req, r, plainText, srv.Send)
	})
}

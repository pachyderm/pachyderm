package server

import (
	"github.com/pachyderm/pachyderm/src/client/admin"
	pfs1_10 "github.com/pachyderm/pachyderm/src/client/admin/v1_10/pfs"
	pps1_10 "github.com/pachyderm/pachyderm/src/client/admin/v1_10/pps"
	pfs1_11 "github.com/pachyderm/pachyderm/src/client/admin/v1_11/pfs"
	pps1_11 "github.com/pachyderm/pachyderm/src/client/admin/v1_11/pps"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

func convert1_10Repo(r *pfs1_10.Repo) *pfs1_11.Repo {
	if r == nil {
		return nil
	}
	return &pfs1_11.Repo{
		Name: r.Name,
	}
}

func convert1_10Commit(c *pfs1_10.Commit) *pfs1_11.Commit {
	if c == nil {
		return nil
	}
	return &pfs1_11.Commit{
		Repo: convert1_10Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_10Provenance(provenance *pfs1_10.CommitProvenance) *pfs1_11.CommitProvenance {
	if provenance == nil {
		return nil
	}
	return &pfs1_11.CommitProvenance{
		Commit: convert1_10Commit(provenance.Commit),
		Branch: convert1_10Branch(provenance.Branch),
	}
}

func convert1_10Provenances(provenances []*pfs1_10.CommitProvenance) []*pfs1_11.CommitProvenance {
	if provenances == nil {
		return nil
	}
	result := make([]*pfs1_11.CommitProvenance, 0, len(provenances))
	for _, p := range provenances {
		result = append(result, convert1_10Provenance(p))
	}
	return result
}

func convert1_10Job(j *pps1_10.CreateJobRequest) *pps1_11.CreateJobRequest {
	if j == nil {
		return nil
	}
	return &pps1_11.CreateJobRequest{
		Pipeline:      convert1_10Pipeline(j.Pipeline),
		OutputCommit:  convert1_10Commit(j.OutputCommit),
		Restart:       j.Restart,
		DataProcessed: j.DataProcessed,
		DataSkipped:   j.DataSkipped,
		DataTotal:     j.DataTotal,
		DataFailed:    j.DataFailed,
		DataRecovered: j.DataRecovered,
		Stats:         convert1_10Stats(j.Stats),
		StatsCommit:   convert1_10Commit(j.StatsCommit),
		State:         pps1_11.JobState(j.State),
		Reason:        j.Reason,
		Started:       j.Started,
		Finished:      j.Finished,
	}
}

func convert1_10Stats(s *pps1_10.ProcessStats) *pps1_11.ProcessStats {
	if s == nil {
		return nil
	}
	return &pps1_11.ProcessStats{
		DownloadTime:  s.DownloadTime,
		ProcessTime:   s.ProcessTime,
		UploadTime:    s.UploadTime,
		DownloadBytes: s.DownloadBytes,
		UploadBytes:   s.UploadBytes,
	}
}

func convert1_10CreateObject(o *pfs1_10.CreateObjectRequest) *pfs1_11.CreateObjectRequest {
	if o == nil {
		return nil
	}
	return &pfs1_11.CreateObjectRequest{
		Object:   convert1_10Object(o.Object),
		BlockRef: convert1_10BlockRef(o.BlockRef),
	}
}

func convert1_10Object(o *pfs1_10.Object) *pfs1_11.Object {
	if o == nil {
		return nil
	}
	return &pfs1_11.Object{
		Hash: o.Hash,
	}
}

func convert1_10BlockRef(b *pfs1_10.BlockRef) *pfs1_11.BlockRef {
	if b == nil {
		return nil
	}
	return &pfs1_11.BlockRef{
		Block: &pfs1_11.Block{
			Hash: b.Block.Hash,
		},
		Range: &pfs1_11.ByteRange{
			Lower: b.Range.Lower,
			Upper: b.Range.Upper,
		},
	}
}

func convert1_10Objects(objects []*pfs1_10.Object) []*pfs1_11.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs1_11.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_10Object(o))
	}
	return result
}

func convert1_10Tag(tag *pfs1_10.Tag) *pfs1_11.Tag {
	if tag == nil {
		return nil
	}
	return &pfs1_11.Tag{
		Name: tag.Name,
	}
}

func convert1_10Tags(tags []*pfs1_10.Tag) []*pfs1_11.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs1_11.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_10Tag(t))
	}
	return result
}

func convert1_10Branch(b *pfs1_10.Branch) *pfs1_11.Branch {
	if b == nil {
		return nil
	}
	return &pfs1_11.Branch{
		Repo: convert1_10Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_10Branches(branches []*pfs1_10.Branch) []*pfs1_11.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs1_11.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_10Branch(b))
	}
	return result
}

func convert1_10Pipeline(p *pps1_10.Pipeline) *pps1_11.Pipeline {
	if p == nil {
		return nil
	}
	return &pps1_11.Pipeline{
		Name: p.Name,
	}
}

func convert1_10SecretMount(s *pps1_10.SecretMount) *pps1_11.SecretMount {
	if s == nil {
		return nil
	}
	return &pps1_11.SecretMount{
		Name:      s.Name,
		Key:       s.Key,
		MountPath: s.MountPath,
		EnvVar:    s.EnvVar,
	}
}

func convert1_10SecretMounts(secrets []*pps1_10.SecretMount) []*pps1_11.SecretMount {
	if secrets == nil {
		return nil
	}
	result := make([]*pps1_11.SecretMount, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_10SecretMount(s))
	}
	return result
}

func convert1_10Transform(t *pps1_10.Transform) *pps1_11.Transform {
	if t == nil {
		return nil
	}
	return &pps1_11.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		ErrCmd:           t.ErrCmd,
		Env:              t.Env,
		Secrets:          convert1_10SecretMounts(t.Secrets),
		ImagePullSecrets: t.ImagePullSecrets,
		Stdin:            t.Stdin,
		ErrStdin:         t.ErrStdin,
		AcceptReturnCode: t.AcceptReturnCode,
		Debug:            t.Debug,
		User:             t.User,
		WorkingDir:       t.WorkingDir,
		Dockerfile:       t.Dockerfile,
	}
}

func convert1_10ParallelismSpec(s *pps1_10.ParallelismSpec) *pps1_11.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps1_11.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_10HashtreeSpec(h *pps1_10.HashtreeSpec) *pps1_11.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps1_11.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_10Egress(e *pps1_10.Egress) *pps1_11.Egress {
	if e == nil {
		return nil
	}
	return &pps1_11.Egress{
		URL: e.URL,
	}
}

func convert1_10GPUSpec(g *pps1_10.GPUSpec) *pps1_11.GPUSpec {
	if g == nil {
		return nil
	}
	return &pps1_11.GPUSpec{
		Type:   g.Type,
		Number: g.Number,
	}
}

func convert1_10ResourceSpec(r *pps1_10.ResourceSpec) *pps1_11.ResourceSpec {
	if r == nil {
		return nil
	}
	return &pps1_11.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Gpu:    convert1_10GPUSpec(r.Gpu),
		Disk:   r.Disk,
	}
}

func convert1_10PFSInput(p *pps1_10.PFSInput) *pps1_11.PFSInput {
	if p == nil {
		return nil
	}
	return &pps1_11.PFSInput{
		Name:       p.Name,
		Repo:       p.Repo,
		Branch:     p.Branch,
		Commit:     p.Commit,
		Glob:       p.Glob,
		JoinOn:     p.JoinOn,
		Lazy:       p.Lazy,
		EmptyFiles: p.EmptyFiles,
	}
}

func convert1_10CronInput(i *pps1_10.CronInput) *pps1_11.CronInput {
	if i == nil {
		return nil
	}
	return &pps1_11.CronInput{
		Name:      i.Name,
		Repo:      i.Repo,
		Commit:    i.Commit,
		Spec:      i.Spec,
		Overwrite: i.Overwrite,
		Start:     i.Start,
	}
}

func convert1_10GitInput(i *pps1_10.GitInput) *pps1_11.GitInput {
	if i == nil {
		return nil
	}
	return &pps1_11.GitInput{
		Name:   i.Name,
		URL:    i.URL,
		Branch: i.Branch,
		Commit: i.Commit,
	}
}

func convert1_10Input(i *pps1_10.Input) *pps1_11.Input {
	if i == nil {
		return nil
	}
	return &pps1_11.Input{
		Pfs:   convert1_10PFSInput(i.Pfs),
		Cross: convert1_10Inputs(i.Cross),
		Union: convert1_10Inputs(i.Union),
		Cron:  convert1_10CronInput(i.Cron),
		Git:   convert1_10GitInput(i.Git),
	}
}

func convert1_10Inputs(inputs []*pps1_10.Input) []*pps1_11.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps1_11.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_10Input(i))
	}
	return result
}

func convert1_10Service(s *pps1_10.Service) *pps1_11.Service {
	if s == nil {
		return nil
	}
	return &pps1_11.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
		Type:         s.Type,
	}
}

func convert1_10Spout(s *pps1_10.Spout) *pps1_11.Spout {
	if s == nil {
		return nil
	}
	return &pps1_11.Spout{
		Overwrite: s.Overwrite,
		Service:   convert1_10Service(s.Service),
		Marker:    s.Marker,
	}
}

func convert1_10Metadata(s *pps1_10.Metadata) *pps1_11.Metadata {
	if s == nil {
		return nil
	}
	if s.Annotations == nil {
		return nil
	}
	return &pps1_11.Metadata{
		Annotations: s.Annotations,
	}
}

func convert1_10ChunkSpec(c *pps1_10.ChunkSpec) *pps1_11.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps1_11.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_10SchedulingSpec(s *pps1_10.SchedulingSpec) *pps1_11.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps1_11.SchedulingSpec{
		NodeSelector:      s.NodeSelector,
		PriorityClassName: s.PriorityClassName,
	}
}

func convert1_10Op(op *admin.Op1_10) (*admin.Op1_11, error) {
	switch {
	case op.CreateObject != nil:
		return &admin.Op1_11{
			CreateObject: convert1_10CreateObject(op.CreateObject),
		}, nil
	case op.Job != nil:
		return &admin.Op1_11{
			Job: convert1_10Job(op.Job),
		}, nil
	case op.Tag != nil:
		if !objHashRE.MatchString(op.Tag.Object.Hash) {
			return nil, errors.Errorf("invalid object hash in op: %q", op)
		}
		return &admin.Op1_11{
			Tag: &pfs1_11.TagObjectRequest{
				Object: convert1_10Object(op.Tag.Object),
				Tags:   convert1_10Tags(op.Tag.Tags),
			},
		}, nil
	case op.Repo != nil:
		return &admin.Op1_11{
			Repo: &pfs1_11.CreateRepoRequest{
				Repo:        convert1_10Repo(op.Repo.Repo),
				Description: op.Repo.Description,
			},
		}, nil
	case op.Commit != nil:
		return &admin.Op1_11{
			Commit: &pfs1_11.BuildCommitRequest{
				Parent:     convert1_10Commit(op.Commit.Parent),
				Branch:     op.Commit.Branch,
				Provenance: convert1_10Provenances(op.Commit.Provenance),
				Tree:       convert1_10Object(op.Commit.Tree),
				Trees:      convert1_10Objects(op.Commit.Trees),
				Datums:     convert1_10Object(op.Commit.Datums),
				ID:         op.Commit.ID,
				SizeBytes:  op.Commit.SizeBytes,
			},
		}, nil
	case op.Branch != nil:
		newOp := &admin.Op1_11{
			Branch: &pfs1_11.CreateBranchRequest{
				Head:       convert1_10Commit(op.Branch.Head),
				Branch:     convert1_10Branch(op.Branch.Branch),
				Provenance: convert1_10Branches(op.Branch.Provenance),
			},
		}
		if newOp.Branch.Branch == nil {
			newOp.Branch.Branch = &pfs1_11.Branch{
				Repo: convert1_10Repo(op.Branch.Head.Repo),
				Name: op.Branch.SBranch,
			}
		}
		return newOp, nil
	case op.Pipeline != nil:
		return &admin.Op1_11{
			Pipeline: &pps1_11.CreatePipelineRequest{
				Pipeline:         convert1_10Pipeline(op.Pipeline.Pipeline),
				Transform:        convert1_10Transform(op.Pipeline.Transform),
				ParallelismSpec:  convert1_10ParallelismSpec(op.Pipeline.ParallelismSpec),
				HashtreeSpec:     convert1_10HashtreeSpec(op.Pipeline.HashtreeSpec),
				Egress:           convert1_10Egress(op.Pipeline.Egress),
				Update:           op.Pipeline.Update,
				OutputBranch:     op.Pipeline.OutputBranch,
				ResourceRequests: convert1_10ResourceSpec(op.Pipeline.ResourceRequests),
				ResourceLimits:   convert1_10ResourceSpec(op.Pipeline.ResourceLimits),
				Input:            convert1_10Input(op.Pipeline.Input),
				Description:      op.Pipeline.Description,
				CacheSize:        op.Pipeline.CacheSize,
				EnableStats:      op.Pipeline.EnableStats,
				Reprocess:        op.Pipeline.Reprocess,
				MaxQueueSize:     op.Pipeline.MaxQueueSize,
				Service:          convert1_10Service(op.Pipeline.Service),
				Spout:            convert1_10Spout(op.Pipeline.Spout),
				ChunkSpec:        convert1_10ChunkSpec(op.Pipeline.ChunkSpec),
				DatumTimeout:     op.Pipeline.DatumTimeout,
				JobTimeout:       op.Pipeline.JobTimeout,
				Salt:             op.Pipeline.Salt,
				Standby:          op.Pipeline.Standby,
				DatumTries:       op.Pipeline.DatumTries,
				SchedulingSpec:   convert1_10SchedulingSpec(op.Pipeline.SchedulingSpec),
				PodSpec:          op.Pipeline.PodSpec,
				PodPatch:         op.Pipeline.PodPatch,
				SpecCommit:       convert1_10Commit(op.Pipeline.SpecCommit),
				Metadata:         convert1_10Metadata(op.Pipeline.Metadata),
			},
		}, nil
	default:
		return nil, errors.Errorf("unrecognized 1.9 op type:\n%+v", op)
	}
	return nil, errors.Errorf("internal error: convert1_10Op() didn't return a 1.9 op for:\n%+v", op)
}

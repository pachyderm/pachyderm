package server

import (
	"github.com/pachyderm/pachyderm/src/client/admin"
	pfs1_10 "github.com/pachyderm/pachyderm/src/client/admin/v1_10/pfs"
	pps1_10 "github.com/pachyderm/pachyderm/src/client/admin/v1_10/pps"
	pfs1_9 "github.com/pachyderm/pachyderm/src/client/admin/v1_9/pfs"
	pps1_9 "github.com/pachyderm/pachyderm/src/client/admin/v1_9/pps"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

func convert1_9Repo(r *pfs1_9.Repo) *pfs1_10.Repo {
	if r == nil {
		return nil
	}
	return &pfs1_10.Repo{
		Name: r.Name,
	}
}

func convert1_9Commit(c *pfs1_9.Commit) *pfs1_10.Commit {
	if c == nil {
		return nil
	}
	return &pfs1_10.Commit{
		Repo: convert1_9Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_9Provenance(provenance *pfs1_9.CommitProvenance) *pfs1_10.CommitProvenance {
	if provenance == nil {
		return nil
	}
	return &pfs1_10.CommitProvenance{
		Commit: convert1_9Commit(provenance.Commit),
		Branch: convert1_9Branch(provenance.Branch),
	}
}

func convert1_9Provenances(provenances []*pfs1_9.CommitProvenance) []*pfs1_10.CommitProvenance {
	if provenances == nil {
		return nil
	}
	result := make([]*pfs1_10.CommitProvenance, 0, len(provenances))
	for _, p := range provenances {
		result = append(result, convert1_9Provenance(p))
	}
	return result
}

func convert1_9Job(j *pps1_9.CreateJobRequest) *pps1_10.CreateJobRequest {
	if j == nil {
		return nil
	}
	return &pps1_10.CreateJobRequest{
		Pipeline:      convert1_9Pipeline(j.Pipeline),
		OutputCommit:  convert1_9Commit(j.OutputCommit),
		Restart:       j.Restart,
		DataProcessed: j.DataProcessed,
		DataSkipped:   j.DataSkipped,
		DataTotal:     j.DataTotal,
		DataFailed:    j.DataFailed,
		DataRecovered: j.DataRecovered,
		Stats:         convert1_9Stats(j.Stats),
		StatsCommit:   convert1_9Commit(j.StatsCommit),
		State:         pps1_10.JobState(j.State),
		Reason:        j.Reason,
		Started:       j.Started,
		Finished:      j.Finished,
	}
}

func convert1_9Stats(s *pps1_9.ProcessStats) *pps1_10.ProcessStats {
	if s == nil {
		return nil
	}
	return &pps1_10.ProcessStats{
		DownloadTime:  s.DownloadTime,
		ProcessTime:   s.ProcessTime,
		UploadTime:    s.UploadTime,
		DownloadBytes: s.DownloadBytes,
		UploadBytes:   s.UploadBytes,
	}
}

func convert1_9CreateObject(o *pfs1_9.CreateObjectRequest) *pfs1_10.CreateObjectRequest {
	if o == nil {
		return nil
	}
	return &pfs1_10.CreateObjectRequest{
		Object:   convert1_9Object(o.Object),
		BlockRef: convert1_9BlockRef(o.BlockRef),
	}
}

func convert1_9Object(o *pfs1_9.Object) *pfs1_10.Object {
	if o == nil {
		return nil
	}
	return &pfs1_10.Object{
		Hash: o.Hash,
	}
}

func convert1_9BlockRef(b *pfs1_9.BlockRef) *pfs1_10.BlockRef {
	if b == nil {
		return nil
	}
	return &pfs1_10.BlockRef{
		Block: &pfs1_10.Block{
			Hash: b.Block.Hash,
		},
		Range: &pfs1_10.ByteRange{
			Lower: b.Range.Lower,
			Upper: b.Range.Upper,
		},
	}
}

func convert1_9Objects(objects []*pfs1_9.Object) []*pfs1_10.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs1_10.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_9Object(o))
	}
	return result
}

func convert1_9Tag(tag *pfs1_9.Tag) *pfs1_10.Tag {
	if tag == nil {
		return nil
	}
	return &pfs1_10.Tag{
		Name: tag.Name,
	}
}

func convert1_9Tags(tags []*pfs1_9.Tag) []*pfs1_10.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs1_10.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_9Tag(t))
	}
	return result
}

func convert1_9Branch(b *pfs1_9.Branch) *pfs1_10.Branch {
	if b == nil {
		return nil
	}
	return &pfs1_10.Branch{
		Repo: convert1_9Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_9Branches(branches []*pfs1_9.Branch) []*pfs1_10.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs1_10.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_9Branch(b))
	}
	return result
}

func convert1_9Pipeline(p *pps1_9.Pipeline) *pps1_10.Pipeline {
	if p == nil {
		return nil
	}
	return &pps1_10.Pipeline{
		Name: p.Name,
	}
}

func convert1_9Secret(s *pps1_9.Secret) *pps1_10.SecretMount {
	if s == nil {
		return nil
	}
	return &pps1_10.SecretMount{
		Name:      s.Name,
		Key:       s.Key,
		MountPath: s.MountPath,
		EnvVar:    s.EnvVar,
	}
}

func convert1_9Secrets(secrets []*pps1_9.Secret) []*pps1_10.SecretMount {
	if secrets == nil {
		return nil
	}
	result := make([]*pps1_10.SecretMount, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_9Secret(s))
	}
	return result
}

func convert1_9Transform(t *pps1_9.Transform) *pps1_10.Transform {
	if t == nil {
		return nil
	}
	return &pps1_10.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		ErrCmd:           t.ErrCmd,
		Env:              t.Env,
		Secrets:          convert1_9Secrets(t.Secrets),
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

func convert1_9ParallelismSpec(s *pps1_9.ParallelismSpec) *pps1_10.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps1_10.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_9HashtreeSpec(h *pps1_9.HashtreeSpec) *pps1_10.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps1_10.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_9Egress(e *pps1_9.Egress) *pps1_10.Egress {
	if e == nil {
		return nil
	}
	return &pps1_10.Egress{
		URL: e.URL,
	}
}

func convert1_9GPUSpec(g *pps1_9.GPUSpec) *pps1_10.GPUSpec {
	if g == nil {
		return nil
	}
	return &pps1_10.GPUSpec{
		Type:   g.Type,
		Number: g.Number,
	}
}

func convert1_9ResourceSpec(r *pps1_9.ResourceSpec) *pps1_10.ResourceSpec {
	if r == nil {
		return nil
	}
	return &pps1_10.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Gpu:    convert1_9GPUSpec(r.Gpu),
		Disk:   r.Disk,
	}
}

func convert1_9PFSInput(p *pps1_9.PFSInput) *pps1_10.PFSInput {
	if p == nil {
		return nil
	}
	return &pps1_10.PFSInput{
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

func convert1_9CronInput(i *pps1_9.CronInput) *pps1_10.CronInput {
	if i == nil {
		return nil
	}
	return &pps1_10.CronInput{
		Name:      i.Name,
		Repo:      i.Repo,
		Commit:    i.Commit,
		Spec:      i.Spec,
		Overwrite: i.Overwrite,
		Start:     i.Start,
	}
}

func convert1_9GitInput(i *pps1_9.GitInput) *pps1_10.GitInput {
	if i == nil {
		return nil
	}
	return &pps1_10.GitInput{
		Name:   i.Name,
		URL:    i.URL,
		Branch: i.Branch,
		Commit: i.Commit,
	}
}

func convert1_9Input(i *pps1_9.Input) *pps1_10.Input {
	if i == nil {
		return nil
	}
	return &pps1_10.Input{
		Pfs:   convert1_9PFSInput(i.Pfs),
		Cross: convert1_9Inputs(i.Cross),
		Union: convert1_9Inputs(i.Union),
		Cron:  convert1_9CronInput(i.Cron),
		Git:   convert1_9GitInput(i.Git),
	}
}

func convert1_9Inputs(inputs []*pps1_9.Input) []*pps1_10.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps1_10.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_9Input(i))
	}
	return result
}

func convert1_9Service(s *pps1_9.Service) *pps1_10.Service {
	if s == nil {
		return nil
	}
	return &pps1_10.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
		Type:         s.Type,
	}
}

func convert1_9Spout(s *pps1_9.Spout) *pps1_10.Spout {
	if s == nil {
		return nil
	}
	return &pps1_10.Spout{
		Overwrite: s.Overwrite,
		Service:   convert1_9Service(s.Service),
		Marker:    s.Marker,
	}
}

func convert1_9Metadata(s *pps1_9.Service) *pps1_10.Metadata {
	if s == nil {
		return nil
	}
	if s.Annotations == nil {
		return nil
	}
	return &pps1_10.Metadata{
		Annotations: s.Annotations,
	}
}

func convert1_9ChunkSpec(c *pps1_9.ChunkSpec) *pps1_10.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps1_10.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_9SchedulingSpec(s *pps1_9.SchedulingSpec) *pps1_10.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps1_10.SchedulingSpec{
		NodeSelector:      s.NodeSelector,
		PriorityClassName: s.PriorityClassName,
	}
}

func convert1_9Op(op *admin.Op1_9) (*admin.Op1_10, error) {
	switch {
	case op.CreateObject != nil:
		return &admin.Op1_10{
			CreateObject: convert1_9CreateObject(op.CreateObject),
		}, nil
	case op.Job != nil:
		return &admin.Op1_10{
			Job: convert1_9Job(op.Job),
		}, nil
	case op.Tag != nil:
		if !objHashRE.MatchString(op.Tag.Object.Hash) {
			return nil, errors.Errorf("invalid object hash in op: %q", op)
		}
		return &admin.Op1_10{
			Tag: &pfs1_10.TagObjectRequest{
				Object: convert1_9Object(op.Tag.Object),
				Tags:   convert1_9Tags(op.Tag.Tags),
			},
		}, nil
	case op.Repo != nil:
		return &admin.Op1_10{
			Repo: &pfs1_10.CreateRepoRequest{
				Repo:        convert1_9Repo(op.Repo.Repo),
				Description: op.Repo.Description,
			},
		}, nil
	case op.Commit != nil:
		return &admin.Op1_10{
			Commit: &pfs1_10.BuildCommitRequest{
				Parent:     convert1_9Commit(op.Commit.Parent),
				Branch:     op.Commit.Branch,
				Provenance: convert1_9Provenances(op.Commit.Provenance),
				Tree:       convert1_9Object(op.Commit.Tree),
				Trees:      convert1_9Objects(op.Commit.Trees),
				Datums:     convert1_9Object(op.Commit.Datums),
				ID:         op.Commit.ID,
				SizeBytes:  op.Commit.SizeBytes,
			},
		}, nil
	case op.Branch != nil:
		newOp := &admin.Op1_10{
			Branch: &pfs1_10.CreateBranchRequest{
				Head:       convert1_9Commit(op.Branch.Head),
				Branch:     convert1_9Branch(op.Branch.Branch),
				Provenance: convert1_9Branches(op.Branch.Provenance),
			},
		}
		if newOp.Branch.Branch == nil {
			newOp.Branch.Branch = &pfs1_10.Branch{
				Repo: convert1_9Repo(op.Branch.Head.Repo),
				Name: op.Branch.SBranch,
			}
		}
		return newOp, nil
	case op.Pipeline != nil:
		return &admin.Op1_10{
			Pipeline: &pps1_10.CreatePipelineRequest{
				Pipeline:         convert1_9Pipeline(op.Pipeline.Pipeline),
				Transform:        convert1_9Transform(op.Pipeline.Transform),
				ParallelismSpec:  convert1_9ParallelismSpec(op.Pipeline.ParallelismSpec),
				HashtreeSpec:     convert1_9HashtreeSpec(op.Pipeline.HashtreeSpec),
				Egress:           convert1_9Egress(op.Pipeline.Egress),
				Update:           op.Pipeline.Update,
				OutputBranch:     op.Pipeline.OutputBranch,
				ResourceRequests: convert1_9ResourceSpec(op.Pipeline.ResourceRequests),
				ResourceLimits:   convert1_9ResourceSpec(op.Pipeline.ResourceLimits),
				Input:            convert1_9Input(op.Pipeline.Input),
				Description:      op.Pipeline.Description,
				CacheSize:        op.Pipeline.CacheSize,
				EnableStats:      op.Pipeline.EnableStats,
				Reprocess:        op.Pipeline.Reprocess,
				MaxQueueSize:     op.Pipeline.MaxQueueSize,
				Service:          convert1_9Service(op.Pipeline.Service),
				Spout:            convert1_9Spout(op.Pipeline.Spout),
				ChunkSpec:        convert1_9ChunkSpec(op.Pipeline.ChunkSpec),
				DatumTimeout:     op.Pipeline.DatumTimeout,
				JobTimeout:       op.Pipeline.JobTimeout,
				Salt:             op.Pipeline.Salt,
				Standby:          op.Pipeline.Standby,
				DatumTries:       op.Pipeline.DatumTries,
				SchedulingSpec:   convert1_9SchedulingSpec(op.Pipeline.SchedulingSpec),
				PodSpec:          op.Pipeline.PodSpec,
				PodPatch:         op.Pipeline.PodPatch,
				SpecCommit:       convert1_9Commit(op.Pipeline.SpecCommit),
				Metadata:         convert1_9Metadata(op.Pipeline.Service),
			},
		}, nil
	default:
		return nil, errors.Errorf("unrecognized 1.9 op type:\n%+v", op)
	}
	return nil, errors.Errorf("internal error: convert1_9Op() didn't return a 1.9 op for:\n%+v", op)
}

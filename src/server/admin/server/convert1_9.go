package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/admin"
	pfs1_9 "github.com/pachyderm/pachyderm/src/client/admin/v1_9/pfs"
	pps1_9 "github.com/pachyderm/pachyderm/src/client/admin/v1_9/pps"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

func convert1_9Repo(r *pfs1_9.Repo) *pfs.Repo {
	if r == nil {
		return nil
	}
	return &pfs.Repo{
		Name: r.Name,
	}
}

func convert1_9Commit(c *pfs1_9.Commit) *pfs.Commit {
	if c == nil {
		return nil
	}
	return &pfs.Commit{
		Repo: convert1_9Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_9Job(j *pps1_9.CreateJobRequest) *pps.CreateJobRequest {
	if j == nil {
		return nil
	}
	return &pps.CreateJobRequest{
		Pipeline:      convert1_9Pipeline(j.Pipeline),
		OutputCommit:  convert1_9Commit(j.OutputCommit),
		Restart:       j.Restart,
		DataProcessed: j.DataProcessed,
		DataSkipped:   j.DataSkipped,
		DataTotal:     j.DataTotal,
		DataRecovered: j.DataRecovered,
		Stats:         convert1_9Stats(j.Stats),
		StatsCommit:   convert1_9Commit(j.StatsCommit),
		State:         pps.JobState(j.State),
		Reason:        j.Reason,
		Started:       j.Started,
		Finished:      j.Finished,
	}
}

func convert1_9Stats(s *pps1_9.ProcessStats) *pps.ProcessStats {
	if s == nil {
		return nil
	}
	return &pps.ProcessStats{
		DownloadTime:  s.DownloadTime,
		ProcessTime:   s.ProcessTime,
		UploadTime:    s.UploadTime,
		DownloadBytes: s.DownloadBytes,
		UploadBytes:   s.UploadBytes,
	}
}

func convert1_9CreateObject(o *pfs1_9.CreateObjectRequest) *pfs.CreateObjectRequest {
	if o == nil {
		return nil
	}
	return &pfs.CreateObjectRequest{
		Object:   convert1_9Object(o.Object),
		BlockRef: convert1_9BlockRef(o.BlockRef),
	}
}

func convert1_9Object(o *pfs1_9.Object) *pfs.Object {
	if o == nil {
		return nil
	}
	return &pfs.Object{
		Hash: o.Hash,
	}
}

func convert1_9BlockRef(b *pfs1_9.BlockRef) *pfs.BlockRef {
	if b == nil {
		return nil
	}
	return &pfs.BlockRef{
		Block: &pfs.Block{
			Hash: b.Block.Hash,
		},
		Range: &pfs.ByteRange{
			Lower: b.Range.Lower,
			Upper: b.Range.Upper,
		},
	}
}

func convert1_9Objects(objects []*pfs1_9.Object) []*pfs.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_9Object(o))
	}
	return result
}

func convert1_9Tag(tag *pfs1_9.Tag) *pfs.Tag {
	if tag == nil {
		return nil
	}
	return &pfs.Tag{
		Name: tag.Name,
	}
}

func convert1_9Tags(tags []*pfs1_9.Tag) []*pfs.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_9Tag(t))
	}
	return result
}

func convert1_9Branch(b *pfs1_9.Branch) *pfs.Branch {
	if b == nil {
		return nil
	}
	return &pfs.Branch{
		Repo: convert1_9Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_9Branches(branches []*pfs1_9.Branch) []*pfs.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_9Branch(b))
	}
	return result
}

func convert1_9Pipeline(p *pps1_9.Pipeline) *pps.Pipeline {
	if p == nil {
		return nil
	}
	return &pps.Pipeline{
		Name: p.Name,
	}
}

func convert1_9Secret(s *pps1_9.Secret) *pps.SecretMount {
	if s == nil {
		return nil
	}
	return &pps.SecretMount{
		Name: s.Name,
	}
}

func convert1_9Secrets(secrets []*pps1_9.Secret) []*pps.SecretMount {
	if secrets == nil {
		return nil
	}
	result := make([]*pps.SecretMount, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_9Secret(s))
	}
	return result
}

func convert1_9Transform(t *pps1_9.Transform) *pps.Transform {
	if t == nil {
		return nil
	}
	return &pps.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		ErrCmd:           nil,
		Env:              t.Env,
		Secrets:          convert1_9Secrets(t.Secrets),
		ImagePullSecrets: t.ImagePullSecrets,
		Stdin:            t.Stdin,
		ErrStdin:         nil,
		AcceptReturnCode: t.AcceptReturnCode,
		Debug:            t.Debug,
		User:             t.User,
		WorkingDir:       t.WorkingDir,
		Dockerfile:       t.Dockerfile,
	}
}

func convert1_9ParallelismSpec(s *pps1_9.ParallelismSpec) *pps.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_9HashtreeSpec(h *pps1_9.HashtreeSpec) *pps.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_9Egress(e *pps1_9.Egress) *pps.Egress {
	if e == nil {
		return nil
	}
	return &pps.Egress{
		URL: e.URL,
	}
}

func convert1_9GPUSpec(g *pps1_9.GPUSpec) *pps.GPUSpec {
	if g == nil {
		return nil
	}
	return &pps.GPUSpec{
		Type:   g.Type,
		Number: g.Number,
	}
}

func convert1_9ResourceSpec(r *pps1_9.ResourceSpec) *pps.ResourceSpec {
	if r == nil {
		return nil
	}
	return &pps.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Gpu:    convert1_9GPUSpec(r.Gpu),
		Disk:   r.Disk,
	}
}

func convert1_9PFSInput(p *pps1_9.PFSInput) *pps.PFSInput {
	if p == nil {
		return nil
	}
	return &pps.PFSInput{
		Name:       p.Name,
		Repo:       p.Repo,
		Branch:     p.Branch,
		Commit:     p.Commit,
		Glob:       p.Glob,
		Lazy:       p.Lazy,
		EmptyFiles: p.EmptyFiles,
	}
}

func convert1_9CronInput(i *pps1_9.CronInput) *pps.CronInput {
	if i == nil {
		return nil
	}
	return &pps.CronInput{
		Name:      i.Name,
		Repo:      i.Repo,
		Commit:    i.Commit,
		Spec:      i.Spec,
		Overwrite: i.Overwrite,
		Start:     i.Start,
	}
}

func convert1_9GitInput(i *pps1_9.GitInput) *pps.GitInput {
	if i == nil {
		return nil
	}
	return &pps.GitInput{
		Name:   i.Name,
		URL:    i.URL,
		Branch: i.Branch,
		Commit: i.Commit,
	}
}

func convert1_9Input(i *pps1_9.Input) *pps.Input {
	if i == nil {
		return nil
	}
	return &pps.Input{
		// Note: this is deprecated and replaced by `PfsInput`
		Pfs:   convert1_9PFSInput(i.Pfs),
		Cross: convert1_9Inputs(i.Cross),
		Union: convert1_9Inputs(i.Union),
		Cron:  convert1_9CronInput(i.Cron),
		Git:   convert1_9GitInput(i.Git),
	}
}

func convert1_9Inputs(inputs []*pps1_9.Input) []*pps.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_9Input(i))
	}
	return result
}

func convert1_9Service(s *pps1_9.Service) *pps.Service {
	if s == nil {
		return nil
	}
	return &pps.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
	}
}

func convert1_9Metadata(s *pps1_9.Service) *pps.Metadata {
	if s == nil {
		return nil
	}
	if s.Annotations == nil {
		return nil
	}
	return &pps.Metadata{
		Annotations: s.Annotations,
	}
}

func convert1_9ChunkSpec(c *pps1_9.ChunkSpec) *pps.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_9SchedulingSpec(s *pps1_9.SchedulingSpec) *pps.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps.SchedulingSpec{
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
			return nil, fmt.Errorf("invalid object hash in op: %q", op)
		}
		return &admin.Op1_10{
			Tag: &pfs.TagObjectRequest{
				Object: convert1_9Object(op.Tag.Object),
				Tags:   convert1_9Tags(op.Tag.Tags),
			},
		}, nil
	case op.Repo != nil:
		return &admin.Op1_10{
			Repo: &pfs.CreateRepoRequest{
				Repo:        convert1_9Repo(op.Repo.Repo),
				Description: op.Repo.Description,
			},
		}, nil
	case op.Commit != nil:
		return &admin.Op1_10{
			Commit: &pfs.BuildCommitRequest{
				Parent: convert1_9Commit(op.Commit.Parent),
				Branch: op.Commit.Branch,
				// Skip 'Provenance', as we only migrate input commits, so we can
				// rebuild the output commits
				Tree: convert1_9Object(op.Commit.Tree),
				ID:   op.Commit.ID,
			},
		}, nil
	case op.Branch != nil:
		newOp := &admin.Op1_10{
			Branch: &pfs.CreateBranchRequest{
				Head:       convert1_9Commit(op.Branch.Head),
				Branch:     convert1_9Branch(op.Branch.Branch),
				Provenance: convert1_9Branches(op.Branch.Provenance),
			},
		}
		if newOp.Branch.Branch == nil {
			newOp.Branch.Branch = &pfs.Branch{
				Repo: convert1_9Repo(op.Branch.Head.Repo),
				Name: op.Branch.SBranch,
			}
		}
		return newOp, nil
	case op.Pipeline != nil:
		return &admin.Op1_10{
			Pipeline: &pps.CreatePipelineRequest{
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
				Metadata:         convert1_9Metadata(op.Pipeline.Service),
				ChunkSpec:        convert1_9ChunkSpec(op.Pipeline.ChunkSpec),
				DatumTimeout:     op.Pipeline.DatumTimeout,
				JobTimeout:       op.Pipeline.JobTimeout,
				Standby:          op.Pipeline.Standby,
				DatumTries:       op.Pipeline.DatumTries,
				SchedulingSpec:   convert1_9SchedulingSpec(op.Pipeline.SchedulingSpec),
				PodSpec:          op.Pipeline.PodSpec,
				Salt:             op.Pipeline.Salt,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized 1.9 op type:\n%+v", op)
	}
	return nil, fmt.Errorf("internal error: convert1_9Op() didn't return a 1.9 op for:\n%+v", op)
}

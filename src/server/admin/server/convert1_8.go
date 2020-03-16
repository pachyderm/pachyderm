package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/admin"
	pfs1_8 "github.com/pachyderm/pachyderm/src/client/admin/v1_8/pfs"
	pps1_8 "github.com/pachyderm/pachyderm/src/client/admin/v1_8/pps"
	pfs1_9 "github.com/pachyderm/pachyderm/src/client/admin/v1_9/pfs"
	pps1_9 "github.com/pachyderm/pachyderm/src/client/admin/v1_9/pps"
)

func convert1_8Repo(r *pfs1_8.Repo) *pfs1_9.Repo {
	if r == nil {
		return nil
	}
	return &pfs1_9.Repo{
		Name: r.Name,
	}
}

func convert1_8Commit(c *pfs1_8.Commit) *pfs1_9.Commit {
	if c == nil {
		return nil
	}
	return &pfs1_9.Commit{
		Repo: convert1_8Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_8Object(o *pfs1_8.Object) *pfs1_9.Object {
	if o == nil {
		return nil
	}
	return &pfs1_9.Object{
		Hash: o.Hash,
	}
}

func convert1_8Objects(objects []*pfs1_8.Object) []*pfs1_9.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs1_9.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_8Object(o))
	}
	return result
}

func convert1_8Tag(tag *pfs1_8.Tag) *pfs1_9.Tag {
	if tag == nil {
		return nil
	}
	return &pfs1_9.Tag{
		Name: tag.Name,
	}
}

func convert1_8Tags(tags []*pfs1_8.Tag) []*pfs1_9.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs1_9.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_8Tag(t))
	}
	return result
}

func convert1_8Branch(b *pfs1_8.Branch) *pfs1_9.Branch {
	if b == nil {
		return nil
	}
	return &pfs1_9.Branch{
		Repo: convert1_8Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_8Branches(branches []*pfs1_8.Branch) []*pfs1_9.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs1_9.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_8Branch(b))
	}
	return result
}

func convert1_8Pipeline(p *pps1_8.Pipeline) *pps1_9.Pipeline {
	if p == nil {
		return nil
	}
	return &pps1_9.Pipeline{
		Name: p.Name,
	}
}

func convert1_8Secret(s *pps1_8.Secret) *pps1_9.Secret {
	if s == nil {
		return nil
	}
	return &pps1_9.Secret{
		Name:      s.Name,
		Key:       s.Key,
		MountPath: s.MountPath,
		EnvVar:    s.EnvVar,
	}
}

func convert1_8Secrets(secrets []*pps1_8.Secret) []*pps1_9.Secret {
	if secrets == nil {
		return nil
	}
	result := make([]*pps1_9.Secret, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_8Secret(s))
	}
	return result
}

func convert1_8Transform(t *pps1_8.Transform) *pps1_9.Transform {
	if t == nil {
		return nil
	}
	return &pps1_9.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		ErrCmd:           nil,
		Env:              t.Env,
		Secrets:          convert1_8Secrets(t.Secrets),
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

func convert1_8ParallelismSpec(s *pps1_8.ParallelismSpec) *pps1_9.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps1_9.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_8HashtreeSpec(h *pps1_8.HashtreeSpec) *pps1_9.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps1_9.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_8Egress(e *pps1_8.Egress) *pps1_9.Egress {
	if e == nil {
		return nil
	}
	return &pps1_9.Egress{
		URL: e.URL,
	}
}

func convert1_8GPUSpec(g *pps1_8.GPUSpec) *pps1_9.GPUSpec {
	if g == nil {
		return nil
	}
	return &pps1_9.GPUSpec{
		Type:   g.Type,
		Number: g.Number,
	}
}

func convert1_8ResourceSpec(r *pps1_8.ResourceSpec) *pps1_9.ResourceSpec {
	if r == nil {
		return nil
	}
	return &pps1_9.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Gpu:    convert1_8GPUSpec(r.Gpu),
		Disk:   r.Disk,
	}
}

func convert1_8PFSInput(a *pps1_8.AtomInput, p *pps1_8.PFSInput) *pps1_9.PFSInput {
	if a != nil {
		return &pps1_9.PFSInput{
			Name:       a.Name,
			Repo:       a.Repo,
			Branch:     a.Branch,
			Commit:     a.Commit,
			Glob:       a.Glob,
			Lazy:       a.Lazy,
			EmptyFiles: a.EmptyFiles,
		}
	} else if p != nil {
		return &pps1_9.PFSInput{
			Name:       p.Name,
			Repo:       p.Repo,
			Branch:     p.Branch,
			Commit:     p.Commit,
			Glob:       p.Glob,
			Lazy:       p.Lazy,
			EmptyFiles: p.EmptyFiles,
		}
	} else {
		return nil
	}
}

func convert1_8CronInput(i *pps1_8.CronInput) *pps1_9.CronInput {
	if i == nil {
		return nil
	}
	return &pps1_9.CronInput{
		Name:      i.Name,
		Repo:      i.Repo,
		Commit:    i.Commit,
		Spec:      i.Spec,
		Overwrite: i.Overwrite,
		Start:     i.Start,
	}
}

func convert1_8GitInput(i *pps1_8.GitInput) *pps1_9.GitInput {
	if i == nil {
		return nil
	}
	return &pps1_9.GitInput{
		Name:   i.Name,
		URL:    i.URL,
		Branch: i.Branch,
		Commit: i.Commit,
	}
}

func convert1_8Input(i *pps1_8.Input) *pps1_9.Input {
	if i == nil {
		return nil
	}
	return &pps1_9.Input{
		// Note: this is deprecated and replaced by `PfsInput`
		Pfs:   convert1_8PFSInput(i.Atom, i.Pfs),
		Cross: convert1_8Inputs(i.Cross),
		Union: convert1_8Inputs(i.Union),
		Cron:  convert1_8CronInput(i.Cron),
		Git:   convert1_8GitInput(i.Git),
	}
}

func convert1_8Inputs(inputs []*pps1_8.Input) []*pps1_9.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps1_9.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_8Input(i))
	}
	return result
}

func convert1_8Service(s *pps1_8.Service) *pps1_9.Service {
	if s == nil {
		return nil
	}
	return &pps1_9.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
	}
}

func convert1_8ChunkSpec(c *pps1_8.ChunkSpec) *pps1_9.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps1_9.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_8SchedulingSpec(s *pps1_8.SchedulingSpec) *pps1_9.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps1_9.SchedulingSpec{
		NodeSelector:      s.NodeSelector,
		PriorityClassName: s.PriorityClassName,
	}
}

func convert1_8Op(op *admin.Op1_8) (*admin.Op1_9, error) {
	switch {
	case op.Tag != nil:
		if !objHashRE.MatchString(op.Tag.Object.Hash) {
			return nil, fmt.Errorf("invalid object hash in op: %q", op)
		}
		return &admin.Op1_9{
			Tag: &pfs1_9.TagObjectRequest{
				Object: convert1_8Object(op.Tag.Object),
				Tags:   convert1_8Tags(op.Tag.Tags),
			},
		}, nil
	case op.Repo != nil:
		return &admin.Op1_9{
			Repo: &pfs1_9.CreateRepoRequest{
				Repo:        convert1_8Repo(op.Repo.Repo),
				Description: op.Repo.Description,
			},
		}, nil
	case op.Commit != nil:
		return &admin.Op1_9{
			Commit: &pfs1_9.BuildCommitRequest{
				Parent: convert1_8Commit(op.Commit.Parent),
				Branch: op.Commit.Branch,
				// Skip 'Provenance', as we only migrate input commits, so we can
				// rebuild the output commits
				Tree: convert1_8Object(op.Commit.Tree),
				ID:   op.Commit.ID,
			},
		}, nil
	case op.Branch != nil:
		newOp := &admin.Op1_9{
			Branch: &pfs1_9.CreateBranchRequest{
				Head:       convert1_8Commit(op.Branch.Head),
				Branch:     convert1_8Branch(op.Branch.Branch),
				Provenance: convert1_8Branches(op.Branch.Provenance),
			},
		}
		if newOp.Branch.Branch == nil {
			newOp.Branch.Branch = &pfs1_9.Branch{
				Repo: convert1_8Repo(op.Branch.Head.Repo),
				Name: op.Branch.SBranch,
			}
		}
		return newOp, nil
	case op.Pipeline != nil:
		return &admin.Op1_9{
			Pipeline: &pps1_9.CreatePipelineRequest{
				Pipeline:         convert1_8Pipeline(op.Pipeline.Pipeline),
				Transform:        convert1_8Transform(op.Pipeline.Transform),
				ParallelismSpec:  convert1_8ParallelismSpec(op.Pipeline.ParallelismSpec),
				HashtreeSpec:     convert1_8HashtreeSpec(op.Pipeline.HashtreeSpec),
				Egress:           convert1_8Egress(op.Pipeline.Egress),
				Update:           op.Pipeline.Update,
				OutputBranch:     op.Pipeline.OutputBranch,
				ResourceRequests: convert1_8ResourceSpec(op.Pipeline.ResourceRequests),
				ResourceLimits:   convert1_8ResourceSpec(op.Pipeline.ResourceLimits),
				Input:            convert1_8Input(op.Pipeline.Input),
				Description:      op.Pipeline.Description,
				CacheSize:        op.Pipeline.CacheSize,
				EnableStats:      op.Pipeline.EnableStats,
				Reprocess:        op.Pipeline.Reprocess,
				MaxQueueSize:     op.Pipeline.MaxQueueSize,
				Service:          convert1_8Service(op.Pipeline.Service),
				ChunkSpec:        convert1_8ChunkSpec(op.Pipeline.ChunkSpec),
				DatumTimeout:     op.Pipeline.DatumTimeout,
				JobTimeout:       op.Pipeline.JobTimeout,
				Standby:          op.Pipeline.Standby,
				DatumTries:       op.Pipeline.DatumTries,
				SchedulingSpec:   convert1_8SchedulingSpec(op.Pipeline.SchedulingSpec),
				PodSpec:          op.Pipeline.PodSpec,
				Salt:             op.Pipeline.Salt,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized 1.8 op type:\n%+v", op)
	}
	return nil, fmt.Errorf("internal error: convert1_8Op() didn't return a 1.9 op for:\n%+v", op)
}

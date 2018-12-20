package server

import (
	pfs_1_7 "github.com/pachyderm/pachyderm/src/client/admin/1_7/pfs"
	pps_1_7 "github.com/pachyderm/pachyderm/src/client/admin/1_7/pps"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

func convert1_7Repo(r *pfs_1_7.Repo) *pfs.Repo {
	return &pfs.Repo{
		Name: r.Name,
	}
}

func convert1_7Commit(c *pfs_1_7.Commit) *pfs.Commit {
	return &pfs.Commit{
		Repo: convert1_7Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_7Commits(commits []*pfs_1_7.Commit) []*pfs.Commit {
	result := make([]*pfs.Commit, 0, len(commits))
	for _, c := range commits {
		result = append(result, convert1_7Commit(c))
	}
}

func convert1_7Branch(b *pfs_1_7.Branch) *pfs.Branch {
	return &pfs.Branch{
		Repo: convert1_7Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_7Object(o *pfs_1_7.Object) *pfs.Object {
	return &pfs.Object{Hash: o.Hash}
}

func convert1_7Objects(objects []*pfs_1_7.Object) []*pfs.Object {
	result := make([]*pfs.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_7Object(o))
	}
	return result
}

func convert1_7Tag(tag *pfs_1_7.Tag) *pfs.Tag {
	return &pfs.Tag{Name: tag.Name}
}

func convert1_7Tags(tags []*pfs_1_7.Tag) []*pfs.Tag {
	result := make([]*pfs.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_7Tag(t))
	}
	return result
}

func convert1_7Branch(b *pfs_1_7.Branch) *pfs.Branch {
	return &pfs.Branch{
		Repo: &pfs.Repo{Name: b.Repo.Name},
		Name: b.Name,
	}
}

func convert1_7Branches(branches []*pfs_1_7.Branch) []*pfs.Branch {
	result := make([]*pfs.Branch, 0, len(branches))
	for _, b := range branches {
		result := append(result, convert1_7Branch(b))
	}
	return result
}

func convert1_7Pipeline(p *pps_1_7.Pipeline) *pps.Pipeline {
	return &pps.Pipeline{
		Name: p.Name,
	}
}

func convert1_7Transform(t *pps_1_7.Transform) *pps.Transform {
	return &pps.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		Env:              t.Env,
		Secrets:          t.Secrets,
		ImagePullSecrets: t.ImagePullSecrets,
		Stdin:            t.Stdin,
		AcceptReturnCode: t.AcceptReturnCode,
		Debug:            t.Debug,
		User:             t.User,
		WorkingDir:       t.WorkingDir,
	}
}

func convert1_7ParallelismSpec(s *pps_1_7.ParallelismSpec) *pps.ParallelismSpec {
	return &pps.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_7HashtreeSpec(h *pps_1_7.HashtreeSpec) *pps.HashtreeSpec {
	return &pps.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_7Egress(e *pps_1_7.Egress) *pps.Egress {
	return &pps.Egress{
		URL: e.URL,
	}
}

func convert1_7ResourceSpec(r *pps_1_7.ResourceSpec) *pps.ResourceSpec {
	return &pps.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Gpu:    r.Gpu,
		Disk:   r.Disk,
	}
}

func convert1_7Input(i *pps_1_7.Input) *pps.Input {
	convert1_7AtomInput := func(i *pps_1_7.AtomInput) *pps.AtomInput {
		return &pps.PFSInput{
			Name:       i.Name,
			Repo:       i.Repo,
			Branch:     i.Branch,
			Commit:     i.Commit,
			Glob:       i.Glob,
			Lazy:       i.Lazy,
			EmptyFiles: i.EmptyFiles,
		}
	}
	convert1_7PFSInput := func(i *pps_1_7.PFSInput) *pps.PFSInput {
		return &pps.PFSInput{
			Name:       i.Name,
			Repo:       i.Repo,
			Branch:     i.Branch,
			Commit:     i.Commit,
			Glob:       i.Glob,
			Lazy:       i.Lazy,
			EmptyFiles: i.EmptyFiles,
		}
	}
	convert1_7CronInput := func(i *pps_1_7.CronInput) *pps.CronInput {
		return &pps.CronInput{
			Name:   i.Name,
			Repo:   i.Repo,
			Commit: i.Commit,
			Spec:   i.Spec,
			Start:  i.Start,
		}
	}
	convert1_7GitInput := func(i *pps_1_7.GitInput) *pps.GitInput {
		return &pps.GitInput{
			Name:   i.Name,
			URL:    i.URL,
			Branch: i.Branch,
			Commit: i.Commit,
		}
	}
	return &pps.Input{
		// Note: this is deprecated and replaced by `PfsInput`
		Atom:  convert1_7AtomInput(i.Atom),
		Pfs:   convert1_7PFSInput(i.Pfs),
		Cross: convert1_7Inputs(i.Cross),
		Union: convert1_7Inputs(i.Union),
		Cron:  convert1_7CronInput(i.Cron),
		Git:   convert1_7GitInput(i.Git),
	}

}

func convert1_7Inputs(inputs []*pps_1_7.Input) []*pps.Input {
	result := make([]*pfs.Input, 0, len(inputs))
	for _, i := range branches {
		result := append(result, convert1_7Input(i))
	}
	return result
}

func convert1_7Service(s *pps_1_7.Service) *pps.Service {
	return &pps.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
	}
}

func convert1_7ChunkSpec(c *pps_1_7.ChunkSpec) *pps.ChunkSpec {
	return &pps.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_7SchedulingSpec(s *pps_1_7.SchedulingSpec) *pps.SchedulingSpec {
	return &pps.SchedulingSpec{
		NodeSelector:      s.NodeSelector,
		PriorityClassName: s.PriorityClassName,
	}
}

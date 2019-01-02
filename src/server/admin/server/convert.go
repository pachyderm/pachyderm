package server

import (
	"fmt"
	pathlib "path"
	"strings"

	hashtree_1_7 "github.com/pachyderm/pachyderm/src/client/admin/1_7/hashtree"
	pfs_1_7 "github.com/pachyderm/pachyderm/src/client/admin/1_7/pfs"
	pps_1_7 "github.com/pachyderm/pachyderm/src/client/admin/1_7/pps"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
)

func convert1_7Repo(r *pfs_1_7.Repo) *pfs.Repo {
	if r == nil {
		return nil
	}
	return &pfs.Repo{
		Name: r.Name,
	}
}

func convert1_7Commit(c *pfs_1_7.Commit) *pfs.Commit {
	if c == nil {
		return nil
	}
	return &pfs.Commit{
		Repo: convert1_7Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_7Commits(commits []*pfs_1_7.Commit) []*pfs.Commit {
	if commits == nil {
		return nil
	}
	result := make([]*pfs.Commit, 0, len(commits))
	for _, c := range commits {
		result = append(result, convert1_7Commit(c))
	}
	return result
}

func convert1_7Object(o *pfs_1_7.Object) *pfs.Object {
	if o == nil {
		return nil
	}
	return &pfs.Object{Hash: o.Hash}
}

func convert1_7Objects(objects []*pfs_1_7.Object) []*pfs.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_7Object(o))
	}
	return result
}

func convert1_7Tag(tag *pfs_1_7.Tag) *pfs.Tag {
	if tag == nil {
		return nil
	}
	return &pfs.Tag{Name: tag.Name}
}

func convert1_7Tags(tags []*pfs_1_7.Tag) []*pfs.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_7Tag(t))
	}
	return result
}

func convert1_7Branch(b *pfs_1_7.Branch) *pfs.Branch {
	if b == nil {
		return nil
	}
	return &pfs.Branch{
		Repo: convert1_7Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_7Branches(branches []*pfs_1_7.Branch) []*pfs.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_7Branch(b))
	}
	return result
}

func convert1_7Pipeline(p *pps_1_7.Pipeline) *pps.Pipeline {
	if p == nil {
		return nil
	}
	return &pps.Pipeline{
		Name: p.Name,
	}
}

func convert1_7Secret(s *pps_1_7.Secret) *pps.Secret {
	if s == nil {
		return nil
	}
	return &pps.Secret{
		Name:      s.Name,
		Key:       s.Key,
		MountPath: s.MountPath,
		EnvVar:    s.EnvVar,
	}
}

func convert1_7Secrets(secrets []*pps_1_7.Secret) []*pps.Secret {
	if secrets == nil {
		return nil
	}
	result := make([]*pps.Secret, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_7Secret(s))
	}
	return result
}

func convert1_7Transform(t *pps_1_7.Transform) *pps.Transform {
	if t == nil {
		return nil
	}
	return &pps.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		Env:              t.Env,
		Secrets:          convert1_7Secrets(t.Secrets),
		ImagePullSecrets: t.ImagePullSecrets,
		Stdin:            t.Stdin,
		AcceptReturnCode: t.AcceptReturnCode,
		Debug:            t.Debug,
		User:             t.User,
		WorkingDir:       t.WorkingDir,
	}
}

func convert1_7ParallelismSpec(s *pps_1_7.ParallelismSpec) *pps.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_7HashtreeSpec(h *pps_1_7.HashtreeSpec) *pps.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_7Egress(e *pps_1_7.Egress) *pps.Egress {
	if e == nil {
		return nil
	}
	return &pps.Egress{
		URL: e.URL,
	}
}

func convert1_7ResourceSpec(r *pps_1_7.ResourceSpec) *pps.ResourceSpec {
	if r == nil {
		return nil
	}
	result := &pps.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Disk:   r.Disk,
	}
	if r.Gpu != 0 {
		result.Gpu = &pps.GPUSpec{
			Type:   "nvidia.com/gpu", // What most existing customers are using
			Number: r.Gpu,
		}
	}
	return result
}

func convert1_7Input(i *pps_1_7.Input) *pps.Input {
	if i == nil {
		return nil
	}
	convert1_7AtomInput := func(i *pps_1_7.AtomInput) *pps.PFSInput {
		if i == nil {
			return nil
		}
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
		if i == nil {
			return nil
		}
		return &pps.CronInput{
			Name:   i.Name,
			Repo:   i.Repo,
			Commit: i.Commit,
			Spec:   i.Spec,
			Start:  i.Start,
		}
	}
	convert1_7GitInput := func(i *pps_1_7.GitInput) *pps.GitInput {
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
	return &pps.Input{
		// Note: this is deprecated and replaced by `PfsInput`
		Pfs:   convert1_7AtomInput(i.Atom),
		Cross: convert1_7Inputs(i.Cross),
		Union: convert1_7Inputs(i.Union),
		Cron:  convert1_7CronInput(i.Cron),
		Git:   convert1_7GitInput(i.Git),
	}
}

func convert1_7Inputs(inputs []*pps_1_7.Input) []*pps.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_7Input(i))
	}
	return result
}

func convert1_7Service(s *pps_1_7.Service) *pps.Service {
	if s == nil {
		return nil
	}
	return &pps.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
	}
}

func convert1_7ChunkSpec(c *pps_1_7.ChunkSpec) *pps.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_7SchedulingSpec(s *pps_1_7.SchedulingSpec) *pps.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps.SchedulingSpec{
		NodeSelector:      s.NodeSelector,
		PriorityClassName: s.PriorityClassName,
	}
}

func default1_7HashtreeRoot(s string) string {
	if s == "/" || s == "." {
		return ""
	}
	return s
}

// clean canonicalizes 'path' for a Pachyderm 1.7 hashtree
func clean1_7HashtreePath(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return default1_7HashtreeRoot(pathlib.Clean(p))
}

func convert1_7HashTreeNode(old *hashtree_1_7.HashTreeProto, t hashtree.HashTree, path string) error {
	path = clean1_7HashtreePath(path)
	node, ok := old.Fs[path]
	if !ok {
		return fmt.Errorf("expected node at %q, but found none", path)
	}
	switch {
	case node.FileNode != nil:
		if err := t.PutFile(
			path,
			convert1_7Objects(node.FileNode.Objects),
			node.SubtreeSize,
		); err != nil {
			return fmt.Errorf("could not convert file %q: %v", path, err)
		}
	case node.DirNode != nil:
		if err := t.PutDir(path); err != nil {
			return fmt.Errorf("could not convert directory %q: %v", path, err)
		}
		for _, child := range node.DirNode.Children {
			if err := convert1_7HashTreeNode(old, t, pathlib.Join(path, child)); err != nil {
				return err
			}
		}
	}
	return nil
}
func convert1_7HashTree(storageRoot string, old *hashtree_1_7.HashTreeProto) (hashtree.HashTree, error) {
	t, err := hashtree.NewDBHashTree(storageRoot)
	if err != nil {
		return nil, fmt.Errorf("could not create converted hashtree: %v", err)
	}
	if err := convert1_7HashTreeNode(old, t, "/"); err != nil {
		return nil, err
	}
	t.Hash() // canonicalize 't'
	return t, nil
}

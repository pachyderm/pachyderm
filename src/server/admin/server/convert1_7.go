package server

import (
	"bytes"
	pathlib "path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/admin"
	hashtree1_7 "github.com/pachyderm/pachyderm/src/client/admin/v1_7/hashtree"
	pfs1_7 "github.com/pachyderm/pachyderm/src/client/admin/v1_7/pfs"
	pps1_7 "github.com/pachyderm/pachyderm/src/client/admin/v1_7/pps"
	pfs1_8 "github.com/pachyderm/pachyderm/src/client/admin/v1_8/pfs"
	pps1_8 "github.com/pachyderm/pachyderm/src/client/admin/v1_8/pps"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
)

func convert1_7Repo(r *pfs1_7.Repo) *pfs1_8.Repo {
	if r == nil {
		return nil
	}
	return &pfs1_8.Repo{
		Name: r.Name,
	}
}

func convert1_7Commit(c *pfs1_7.Commit) *pfs1_8.Commit {
	if c == nil {
		return nil
	}
	return &pfs1_8.Commit{
		Repo: convert1_7Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_7Object(o *pfs1_7.Object) *pfs1_8.Object {
	if o == nil {
		return nil
	}
	return &pfs1_8.Object{Hash: o.Hash}
}

func convert1_7Objects(objects []*pfs1_7.Object) []*pfs1_8.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs1_8.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_7Object(o))
	}
	return result
}

func convert1_7Tag(tag *pfs1_7.Tag) *pfs1_8.Tag {
	if tag == nil {
		return nil
	}
	return &pfs1_8.Tag{Name: tag.Name}
}

func convert1_7Tags(tags []*pfs1_7.Tag) []*pfs1_8.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs1_8.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_7Tag(t))
	}
	return result
}

func convert1_7Branch(b *pfs1_7.Branch) *pfs1_8.Branch {
	if b == nil {
		return nil
	}
	return &pfs1_8.Branch{
		Repo: convert1_7Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_7Branches(branches []*pfs1_7.Branch) []*pfs1_8.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs1_8.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_7Branch(b))
	}
	return result
}

func convert1_7Pipeline(p *pps1_7.Pipeline) *pps1_8.Pipeline {
	if p == nil {
		return nil
	}
	return &pps1_8.Pipeline{
		Name: p.Name,
	}
}

func convert1_7Secret(s *pps1_7.Secret) *pps1_8.Secret {
	if s == nil {
		return nil
	}
	return &pps1_8.Secret{
		Name:      s.Name,
		Key:       s.Key,
		MountPath: s.MountPath,
		EnvVar:    s.EnvVar,
	}
}

func convert1_7Secrets(secrets []*pps1_7.Secret) []*pps1_8.Secret {
	if secrets == nil {
		return nil
	}
	result := make([]*pps1_8.Secret, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_7Secret(s))
	}
	return result
}

func convert1_7Transform(t *pps1_7.Transform) *pps1_8.Transform {
	if t == nil {
		return nil
	}
	return &pps1_8.Transform{
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

func convert1_7ParallelismSpec(s *pps1_7.ParallelismSpec) *pps1_8.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps1_8.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_7HashtreeSpec(h *pps1_7.HashtreeSpec) *pps1_8.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps1_8.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_7Egress(e *pps1_7.Egress) *pps1_8.Egress {
	if e == nil {
		return nil
	}
	return &pps1_8.Egress{
		URL: e.URL,
	}
}

func convert1_7ResourceSpec(r *pps1_7.ResourceSpec) *pps1_8.ResourceSpec {
	if r == nil {
		return nil
	}
	result := &pps1_8.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Disk:   r.Disk,
	}
	if r.Gpu != 0 {
		result.Gpu = &pps1_8.GPUSpec{
			Type:   "nvidia.com/gpu", // What most existing customers are using
			Number: r.Gpu,
		}
	}
	return result
}

func convert1_7Input(i *pps1_7.Input) *pps1_8.Input {
	if i == nil {
		return nil
	}
	convert1_7AtomInput := func(i *pps1_7.AtomInput) *pps1_8.PFSInput {
		if i == nil {
			return nil
		}
		return &pps1_8.PFSInput{
			Name:       i.Name,
			Repo:       i.Repo,
			Branch:     i.Branch,
			Commit:     i.Commit,
			Glob:       i.Glob,
			Lazy:       i.Lazy,
			EmptyFiles: i.EmptyFiles,
		}
	}
	convert1_7CronInput := func(i *pps1_7.CronInput) *pps1_8.CronInput {
		if i == nil {
			return nil
		}
		return &pps1_8.CronInput{
			Name:   i.Name,
			Repo:   i.Repo,
			Commit: i.Commit,
			Spec:   i.Spec,
			Start:  i.Start,
		}
	}
	convert1_7GitInput := func(i *pps1_7.GitInput) *pps1_8.GitInput {
		if i == nil {
			return nil
		}
		return &pps1_8.GitInput{
			Name:   i.Name,
			URL:    i.URL,
			Branch: i.Branch,
			Commit: i.Commit,
		}
	}
	return &pps1_8.Input{
		// Note: this is deprecated and replaced by `PfsInput`
		Pfs:   convert1_7AtomInput(i.Atom),
		Cross: convert1_7Inputs(i.Cross),
		Union: convert1_7Inputs(i.Union),
		Cron:  convert1_7CronInput(i.Cron),
		Git:   convert1_7GitInput(i.Git),
	}
}

func convert1_7Inputs(inputs []*pps1_7.Input) []*pps1_8.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps1_8.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_7Input(i))
	}
	return result
}

func convert1_7Service(s *pps1_7.Service) *pps1_8.Service {
	if s == nil {
		return nil
	}
	return &pps1_8.Service{
		InternalPort: s.InternalPort,
		ExternalPort: s.ExternalPort,
		IP:           s.IP,
	}
}

func convert1_7ChunkSpec(c *pps1_7.ChunkSpec) *pps1_8.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps1_8.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_7SchedulingSpec(s *pps1_7.SchedulingSpec) *pps1_8.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps1_8.SchedulingSpec{
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

func convert1_7HashTreeNode(old *hashtree1_7.HashTreeProto, t hashtree.HashTree, path string) error {
	path = clean1_7HashtreePath(path)
	node, ok := old.Fs[path]
	if !ok {
		return errors.Errorf("expected node at %q, but found none", path)
	}
	switch {
	case node.FileNode != nil:
		if err := t.PutFile(
			path,
			convert1_10Objects(convert1_9Objects(convert1_8Objects(convert1_7Objects(node.FileNode.Objects)))),
			node.SubtreeSize,
		); err != nil {
			return errors.Wrapf(err, "could not convert file %q", path)
		}
	case node.DirNode != nil:
		if err := t.PutDir(path); err != nil {
			return errors.Wrapf(err, "could not convert directory %q", path)
		}
		for _, child := range node.DirNode.Children {
			if err := convert1_7HashTreeNode(old, t, pathlib.Join(path, child)); err != nil {
				return err
			}
		}
	}
	return nil
}

func convert1_7HashTree(storageRoot string, old *hashtree1_7.HashTreeProto) (hashtree.HashTree, error) {
	t, err := hashtree.NewDBHashTree(storageRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create converted hashtree")
	}
	if err := convert1_7HashTreeNode(old, t, "/"); err != nil {
		return nil, err
	}
	t.Hash() // canonicalize 't'
	return t, nil
}

func convert1_7Op(pachClient *client.APIClient, storageRoot string, op *admin.Op1_7) (*admin.Op1_8, error) {
	switch {
	case op.Tag != nil:
		if !objHashRE.MatchString(op.Tag.Object.Hash) {
			return nil, errors.Errorf("invalid object hash in op: %q", op)
		}
		return &admin.Op1_8{
			Tag: &pfs1_8.TagObjectRequest{
				Object: convert1_7Object(op.Tag.Object),
				Tags:   convert1_7Tags(op.Tag.Tags),
			},
		}, nil
	case op.Repo != nil:
		return &admin.Op1_8{
			Repo: &pfs1_8.CreateRepoRequest{
				Repo:        convert1_7Repo(op.Repo.Repo),
				Description: op.Repo.Description,
			},
		}, nil
	case op.Commit != nil:
		// update hashtree
		var buf bytes.Buffer
		if err := pachClient.GetObject(op.Commit.Tree.Hash, &buf); err != nil {
			return nil, err
		}
		var oldTree hashtree1_7.HashTreeProto
		oldTree.Unmarshal(buf.Bytes())
		newTree, err := convert1_7HashTree(storageRoot, &oldTree)
		if err != nil {
			return nil, err
		}
		defer newTree.Destroy()

		// write new hashtree as an object
		w, err := pachClient.PutObjectAsync(nil)
		if err != nil {
			return nil, errors.Wrapf(err, "could not put new hashtree for commit %q", op.Commit.ID)
		}
		newTree.Serialize(w)
		if err := w.Close(); err != nil {
			return nil, errors.Wrapf(err, "could finish object containing new hashtree for commit %q", op.Commit.ID)
		}
		newTreeObj, err := w.Object()
		if err != nil {
			return nil, errors.Wrapf(err, "could retrieve object reference to new hashtree for commit %q", op.Commit.ID)
		}

		// Set op's object to new hashtree & finish building commit
		return &admin.Op1_8{
			Commit: &pfs1_8.BuildCommitRequest{
				Parent: convert1_7Commit(op.Commit.Parent),
				Branch: op.Commit.Branch,
				Tree: &pfs1_8.Object{
					Hash: newTreeObj.Hash,
				},
				ID: op.Commit.ID,
			},
		}, nil
		// TODO(msteffen): Should we delete the old tree object?
	case op.Branch != nil:
		newOp := &admin.Op1_8{
			Branch: &pfs1_8.CreateBranchRequest{
				Head:       convert1_7Commit(op.Branch.Head),
				Branch:     convert1_7Branch(op.Branch.Branch),
				Provenance: convert1_7Branches(op.Branch.Provenance),
			},
		}
		if newOp.Branch.Branch == nil {
			newOp.Branch.Branch = &pfs1_8.Branch{
				Repo: &pfs1_8.Repo{Name: op.Branch.Head.Repo.Name},
				Name: op.Branch.SBranch,
			}
		}
		return newOp, nil
	case op.Pipeline != nil:
		return &admin.Op1_8{
			Pipeline: &pps1_8.CreatePipelineRequest{
				Pipeline:           convert1_7Pipeline(op.Pipeline.Pipeline),
				Transform:          convert1_7Transform(op.Pipeline.Transform),
				ParallelismSpec:    convert1_7ParallelismSpec(op.Pipeline.ParallelismSpec),
				HashtreeSpec:       convert1_7HashtreeSpec(op.Pipeline.HashtreeSpec),
				Egress:             convert1_7Egress(op.Pipeline.Egress),
				Update:             op.Pipeline.Update,
				OutputBranch:       op.Pipeline.OutputBranch,
				ScaleDownThreshold: op.Pipeline.ScaleDownThreshold,
				ResourceRequests:   convert1_7ResourceSpec(op.Pipeline.ResourceRequests),
				ResourceLimits:     convert1_7ResourceSpec(op.Pipeline.ResourceLimits),
				Input:              convert1_7Input(op.Pipeline.Input),
				Description:        op.Pipeline.Description,
				CacheSize:          op.Pipeline.CacheSize,
				EnableStats:        op.Pipeline.EnableStats,
				Reprocess:          op.Pipeline.Reprocess,
				Batch:              op.Pipeline.Batch,
				MaxQueueSize:       op.Pipeline.MaxQueueSize,
				Service:            convert1_7Service(op.Pipeline.Service),
				ChunkSpec:          convert1_7ChunkSpec(op.Pipeline.ChunkSpec),
				DatumTimeout:       op.Pipeline.DatumTimeout,
				JobTimeout:         op.Pipeline.JobTimeout,
				Standby:            op.Pipeline.Standby,
				DatumTries:         op.Pipeline.DatumTries,
				SchedulingSpec:     convert1_7SchedulingSpec(op.Pipeline.SchedulingSpec),
				PodSpec:            op.Pipeline.PodSpec,
				// Note - don't set Salt, so we don't re-use old datum hashtrees
			},
		}, nil
	default:
		return nil, errors.Errorf("unrecognized 1.7 op type:\n%+v", op)
	}
	return nil, errors.Errorf("internal error: convert1.7Op() didn't return a 1.8 op for:\n%+v", op)
}

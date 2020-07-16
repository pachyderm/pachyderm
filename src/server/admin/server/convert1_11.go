package server

import (
    "github.com/pachyderm/pachyderm/src/client/admin"
    pfs1_11 "github.com/pachyderm/pachyderm/src/client/admin/v1_11/pfs"
    pps1_11 "github.com/pachyderm/pachyderm/src/client/admin/v1_11/pps"
    "github.com/pachyderm/pachyderm/src/client/pfs"
    "github.com/pachyderm/pachyderm/src/client/pkg/errors"
    "github.com/pachyderm/pachyderm/src/client/pps"
)

func convert1_11Repo(r *pfs1_11.Repo) *pfs.Repo {
    if r == nil {
        return nil
    }
    return &pfs.Repo{
        Name: r.Name,
    }
}

func convert1_11Commit(c *pfs1_11.Commit) *pfs.Commit {
    if c == nil {
        return nil
    }
    return &pfs.Commit{
        Repo: convert1_11Repo(c.Repo),
        ID:   c.ID,
    }
}

func convert1_11Provenance(provenance *pfs1_11.CommitProvenance) *pfs.CommitProvenance {
    if provenance == nil {
        return nil
    }
    return &pfs.CommitProvenance{
        Commit: convert1_11Commit(provenance.Commit),
        Branch: convert1_11Branch(provenance.Branch),
    }
}

func convert1_11Provenances(provenances []*pfs1_11.CommitProvenance) []*pfs.CommitProvenance {
    if provenances == nil {
        return nil
    }
    result := make([]*pfs.CommitProvenance, 0, len(provenances))
    for _, p := range provenances {
        result = append(result, convert1_11Provenance(p))
    }
    return result
}

func convert1_11Job(j *pps1_11.CreateJobRequest) *pps.CreateJobRequest {
    if j == nil {
        return nil
    }
    return &pps.CreateJobRequest{
        Pipeline:      convert1_11Pipeline(j.Pipeline),
        OutputCommit:  convert1_11Commit(j.OutputCommit),
        Restart:       j.Restart,
        DataProcessed: j.DataProcessed,
        DataSkipped:   j.DataSkipped,
        DataTotal:     j.DataTotal,
        DataFailed:    j.DataFailed,
        DataRecovered: j.DataRecovered,
        Stats:         convert1_11Stats(j.Stats),
        StatsCommit:   convert1_11Commit(j.StatsCommit),
        State:         pps.JobState(j.State),
        Reason:        j.Reason,
        Started:       j.Started,
        Finished:      j.Finished,
    }
}

func convert1_11Stats(s *pps1_11.ProcessStats) *pps.ProcessStats {
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

func convert1_11CreateObject(o *pfs1_11.CreateObjectRequest) *pfs.CreateObjectRequest {
    if o == nil {
        return nil
    }
    return &pfs.CreateObjectRequest{
        Object:   convert1_11Object(o.Object),
        BlockRef: convert1_11BlockRef(o.BlockRef),
    }
}

func convert1_11Object(o *pfs1_11.Object) *pfs.Object {
    if o == nil {
        return nil
    }
    return &pfs.Object{
        Hash: o.Hash,
    }
}

func convert1_11BlockRef(b *pfs1_11.BlockRef) *pfs.BlockRef {
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

func convert1_11Objects(objects []*pfs1_11.Object) []*pfs.Object {
    if objects == nil {
        return nil
    }
    result := make([]*pfs.Object, 0, len(objects))
    for _, o := range objects {
        result = append(result, convert1_11Object(o))
    }
    return result
}

func convert1_11Tag(tag *pfs1_11.Tag) *pfs.Tag {
    if tag == nil {
        return nil
    }
    return &pfs.Tag{
        Name: tag.Name,
    }
}

func convert1_11Tags(tags []*pfs1_11.Tag) []*pfs.Tag {
    if tags == nil {
        return nil
    }
    result := make([]*pfs.Tag, 0, len(tags))
    for _, t := range tags {
        result = append(result, convert1_11Tag(t))
    }
    return result
}

func convert1_11Branch(b *pfs1_11.Branch) *pfs.Branch {
    if b == nil {
        return nil
    }
    return &pfs.Branch{
        Repo: convert1_11Repo(b.Repo),
        Name: b.Name,
    }
}

func convert1_11Branches(branches []*pfs1_11.Branch) []*pfs.Branch {
    if branches == nil {
        return nil
    }
    result := make([]*pfs.Branch, 0, len(branches))
    for _, b := range branches {
        result = append(result, convert1_11Branch(b))
    }
    return result
}

func convert1_11Pipeline(p *pps1_11.Pipeline) *pps.Pipeline {
    if p == nil {
        return nil
    }
    return &pps.Pipeline{
        Name: p.Name,
    }
}

func convert1_11SecretMount(s *pps1_11.SecretMount) *pps.SecretMount {
    if s == nil {
        return nil
    }
    return &pps.SecretMount{
        Name:      s.Name,
        Key:       s.Key,
        MountPath: s.MountPath,
        EnvVar:    s.EnvVar,
    }
}

func convert1_11SecretMounts(secrets []*pps1_11.SecretMount) []*pps.SecretMount {
    if secrets == nil {
        return nil
    }
    result := make([]*pps.SecretMount, 0, len(secrets))
    for _, s := range secrets {
        result = append(result, convert1_11SecretMount(s))
    }
    return result
}

func convert1_11Transform(t *pps1_11.Transform) *pps.Transform {
    if t == nil {
        return nil
    }
    return &pps.Transform{
        Image:            t.Image,
        Cmd:              t.Cmd,
        ErrCmd:           t.ErrCmd,
        Env:              t.Env,
        Secrets:          convert1_11SecretMounts(t.Secrets),
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

func convert1_11ParallelismSpec(s *pps1_11.ParallelismSpec) *pps.ParallelismSpec {
    if s == nil {
        return nil
    }
    return &pps.ParallelismSpec{
        Constant:    s.Constant,
        Coefficient: s.Coefficient,
    }
}

func convert1_11HashtreeSpec(h *pps1_11.HashtreeSpec) *pps.HashtreeSpec {
    if h == nil {
        return nil
    }
    return &pps.HashtreeSpec{
        Constant: h.Constant,
    }
}

func convert1_11Egress(e *pps1_11.Egress) *pps.Egress {
    if e == nil {
        return nil
    }
    return &pps.Egress{
        URL: e.URL,
    }
}

func convert1_11GPUSpec(g *pps1_11.GPUSpec) *pps.GPUSpec {
    if g == nil {
        return nil
    }
    return &pps.GPUSpec{
        Type:   g.Type,
        Number: g.Number,
    }
}

func convert1_11ResourceSpec(r *pps1_11.ResourceSpec) *pps.ResourceSpec {
    if r == nil {
        return nil
    }
    return &pps.ResourceSpec{
        Cpu:    r.Cpu,
        Memory: r.Memory,
        Gpu:    convert1_11GPUSpec(r.Gpu),
        Disk:   r.Disk,
    }
}

func convert1_11PFSInput(p *pps1_11.PFSInput) *pps.PFSInput {
    if p == nil {
        return nil
    }
    return &pps.PFSInput{
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

func convert1_11CronInput(i *pps1_11.CronInput) *pps.CronInput {
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

func convert1_11GitInput(i *pps1_11.GitInput) *pps.GitInput {
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

func convert1_11Input(i *pps1_11.Input) *pps.Input {
    if i == nil {
        return nil
    }
    return &pps.Input{
        Pfs:   convert1_11PFSInput(i.Pfs),
        Cross: convert1_11Inputs(i.Cross),
        Union: convert1_11Inputs(i.Union),
        Cron:  convert1_11CronInput(i.Cron),
        Git:   convert1_11GitInput(i.Git),
    }
}

func convert1_11Inputs(inputs []*pps1_11.Input) []*pps.Input {
    if inputs == nil {
        return nil
    }
    result := make([]*pps.Input, 0, len(inputs))
    for _, i := range inputs {
        result = append(result, convert1_11Input(i))
    }
    return result
}

func convert1_11Service(s *pps1_11.Service) *pps.Service {
    if s == nil {
        return nil
    }
    return &pps.Service{
        InternalPort: s.InternalPort,
        ExternalPort: s.ExternalPort,
        IP:           s.IP,
        Type:         s.Type,
    }
}

func convert1_11Spout(s *pps1_11.Spout) *pps.Spout {
    if s == nil {
        return nil
    }
    return &pps.Spout{
        Overwrite: s.Overwrite,
        Service:   convert1_11Service(s.Service),
        Marker:    s.Marker,
    }
}

func convert1_11Metadata(s *pps1_11.Metadata) *pps.Metadata {
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

func convert1_11ChunkSpec(c *pps1_11.ChunkSpec) *pps.ChunkSpec {
    if c == nil {
        return nil
    }
    return &pps.ChunkSpec{
        Number:    c.Number,
        SizeBytes: c.SizeBytes,
    }
}

func convert1_11SchedulingSpec(s *pps1_11.SchedulingSpec) *pps.SchedulingSpec {
    if s == nil {
        return nil
    }
    return &pps.SchedulingSpec{
        NodeSelector:      s.NodeSelector,
        PriorityClassName: s.PriorityClassName,
    }
}

func convert1_11Op(op *admin.Op1_11) (*admin.Op1_12, error) {
    switch {
    case op.CreateObject != nil:
        return &admin.Op1_12{
            CreateObject: convert1_11CreateObject(op.CreateObject),
        }, nil
    case op.Job != nil:
        return &admin.Op1_12{
            Job: convert1_11Job(op.Job),
        }, nil
    case op.Tag != nil:
        if !objHashRE.MatchString(op.Tag.Object.Hash) {
            return nil, errors.Errorf("invalid object hash in op: %q", op)
        }
        return &admin.Op1_12{
            Tag: &pfs.TagObjectRequest{
                Object: convert1_11Object(op.Tag.Object),
                Tags:   convert1_11Tags(op.Tag.Tags),
            },
        }, nil
    case op.Repo != nil:
        return &admin.Op1_12{
            Repo: &pfs.CreateRepoRequest{
                Repo:        convert1_11Repo(op.Repo.Repo),
                Description: op.Repo.Description,
            },
        }, nil
    case op.Commit != nil:
        return &admin.Op1_12{
            Commit: &pfs.BuildCommitRequest{
                Parent:     convert1_11Commit(op.Commit.Parent),
                Branch:     op.Commit.Branch,
                Provenance: convert1_11Provenances(op.Commit.Provenance),
                Tree:       convert1_11Object(op.Commit.Tree),
                Trees:      convert1_11Objects(op.Commit.Trees),
                Datums:     convert1_11Object(op.Commit.Datums),
                ID:         op.Commit.ID,
                SizeBytes:  op.Commit.SizeBytes,
            },
        }, nil
    case op.Branch != nil:
        newOp := &admin.Op1_12{
            Branch: &pfs.CreateBranchRequest{
                Head:       convert1_11Commit(op.Branch.Head),
                Branch:     convert1_11Branch(op.Branch.Branch),
                Provenance: convert1_11Branches(op.Branch.Provenance),
            },
        }
        if newOp.Branch.Branch == nil {
            newOp.Branch.Branch = &pfs.Branch{
                Repo: convert1_11Repo(op.Branch.Head.Repo),
                Name: op.Branch.SBranch,
            }
        }
        return newOp, nil
    case op.Pipeline != nil:
        return &admin.Op1_12{
            Pipeline: &pps.CreatePipelineRequest{
                Pipeline:         convert1_11Pipeline(op.Pipeline.Pipeline),
                Transform:        convert1_11Transform(op.Pipeline.Transform),
                ParallelismSpec:  convert1_11ParallelismSpec(op.Pipeline.ParallelismSpec),
                HashtreeSpec:     convert1_11HashtreeSpec(op.Pipeline.HashtreeSpec),
                Egress:           convert1_11Egress(op.Pipeline.Egress),
                Update:           op.Pipeline.Update,
                OutputBranch:     op.Pipeline.OutputBranch,
                ResourceRequests: convert1_11ResourceSpec(op.Pipeline.ResourceRequests),
                ResourceLimits:   convert1_11ResourceSpec(op.Pipeline.ResourceLimits),
                Input:            convert1_11Input(op.Pipeline.Input),
                Description:      op.Pipeline.Description,
                CacheSize:        op.Pipeline.CacheSize,
                EnableStats:      op.Pipeline.EnableStats,
                Reprocess:        op.Pipeline.Reprocess,
                MaxQueueSize:     op.Pipeline.MaxQueueSize,
                Service:          convert1_11Service(op.Pipeline.Service),
                Spout:            convert1_11Spout(op.Pipeline.Spout),
                ChunkSpec:        convert1_11ChunkSpec(op.Pipeline.ChunkSpec),
                DatumTimeout:     op.Pipeline.DatumTimeout,
                JobTimeout:       op.Pipeline.JobTimeout,
                Salt:             op.Pipeline.Salt,
                Standby:          op.Pipeline.Standby,
                DatumTries:       op.Pipeline.DatumTries,
                SchedulingSpec:   convert1_11SchedulingSpec(op.Pipeline.SchedulingSpec),
                PodSpec:          op.Pipeline.PodSpec,
                PodPatch:         op.Pipeline.PodPatch,
                SpecCommit:       convert1_11Commit(op.Pipeline.SpecCommit),
                Metadata:         convert1_11Metadata(op.Pipeline.Metadata),
            },
        }, nil
    default:
        return nil, errors.Errorf("unrecognized 1.9 op type:\n%+v", op)
    }
    return nil, errors.Errorf("internal error: convert1_11Op() didn't return a 1.9 op for:\n%+v", op)
}

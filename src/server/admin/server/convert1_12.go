package server

import (
	"github.com/pachyderm/pachyderm/src/client/admin"
	auth1_12 "github.com/pachyderm/pachyderm/src/client/admin/v1_12/auth"
	pfs1_12 "github.com/pachyderm/pachyderm/src/client/admin/v1_12/pfs"
	pps1_12 "github.com/pachyderm/pachyderm/src/client/admin/v1_12/pps"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

func convert1_12Repo(r *pfs1_12.Repo) *pfs.Repo {
	if r == nil {
		return nil
	}
	return &pfs.Repo{
		Name: r.Name,
	}
}

func convert1_12Commit(c *pfs1_12.Commit) *pfs.Commit {
	if c == nil {
		return nil
	}
	return &pfs.Commit{
		Repo: convert1_12Repo(c.Repo),
		ID:   c.ID,
	}
}

func convert1_12Provenance(provenance *pfs1_12.CommitProvenance) *pfs.CommitProvenance {
	if provenance == nil {
		return nil
	}
	return &pfs.CommitProvenance{
		Commit: convert1_12Commit(provenance.Commit),
		Branch: convert1_12Branch(provenance.Branch),
	}
}

func convert1_12Provenances(provenances []*pfs1_12.CommitProvenance) []*pfs.CommitProvenance {
	if provenances == nil {
		return nil
	}
	result := make([]*pfs.CommitProvenance, 0, len(provenances))
	for _, p := range provenances {
		result = append(result, convert1_12Provenance(p))
	}
	return result
}

func convert1_12Job(j *pps1_12.CreateJobRequest) *pps.CreateJobRequest {
	if j == nil {
		return nil
	}
	return &pps.CreateJobRequest{
		Pipeline:      convert1_12Pipeline(j.Pipeline),
		OutputCommit:  convert1_12Commit(j.OutputCommit),
		Restart:       j.Restart,
		DataProcessed: j.DataProcessed,
		DataSkipped:   j.DataSkipped,
		DataTotal:     j.DataTotal,
		DataFailed:    j.DataFailed,
		DataRecovered: j.DataRecovered,
		Stats:         convert1_12Stats(j.Stats),
		StatsCommit:   convert1_12Commit(j.StatsCommit),
		State:         pps.JobState(j.State),
		Reason:        j.Reason,
		Started:       j.Started,
		Finished:      j.Finished,
	}
}

func convert1_12Stats(s *pps1_12.ProcessStats) *pps.ProcessStats {
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

func convert1_12CreateObject(o *pfs1_12.CreateObjectRequest) *pfs.CreateObjectRequest {
	if o == nil {
		return nil
	}
	return &pfs.CreateObjectRequest{
		Object:   convert1_12Object(o.Object),
		BlockRef: convert1_12BlockRef(o.BlockRef),
	}
}

func convert1_12Object(o *pfs1_12.Object) *pfs.Object {
	if o == nil {
		return nil
	}
	return &pfs.Object{
		Hash: o.Hash,
	}
}

func convert1_12BlockRef(b *pfs1_12.BlockRef) *pfs.BlockRef {
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

func convert1_12Objects(objects []*pfs1_12.Object) []*pfs.Object {
	if objects == nil {
		return nil
	}
	result := make([]*pfs.Object, 0, len(objects))
	for _, o := range objects {
		result = append(result, convert1_12Object(o))
	}
	return result
}

func convert1_12Tag(tag *pfs1_12.Tag) *pfs.Tag {
	if tag == nil {
		return nil
	}
	return &pfs.Tag{
		Name: tag.Name,
	}
}

func convert1_12Tags(tags []*pfs1_12.Tag) []*pfs.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*pfs.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, convert1_12Tag(t))
	}
	return result
}

func convert1_12Branch(b *pfs1_12.Branch) *pfs.Branch {
	if b == nil {
		return nil
	}
	return &pfs.Branch{
		Repo: convert1_12Repo(b.Repo),
		Name: b.Name,
	}
}

func convert1_12Branches(branches []*pfs1_12.Branch) []*pfs.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*pfs.Branch, 0, len(branches))
	for _, b := range branches {
		result = append(result, convert1_12Branch(b))
	}
	return result
}

func convert1_12Pipeline(p *pps1_12.Pipeline) *pps.Pipeline {
	if p == nil {
		return nil
	}
	return &pps.Pipeline{
		Name: p.Name,
	}
}

func convert1_12SecretMount(s *pps1_12.SecretMount) *pps.SecretMount {
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

func convert1_12SecretMounts(secrets []*pps1_12.SecretMount) []*pps.SecretMount {
	if secrets == nil {
		return nil
	}
	result := make([]*pps.SecretMount, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, convert1_12SecretMount(s))
	}
	return result
}

func convert1_12Transform(t *pps1_12.Transform) *pps.Transform {
	if t == nil {
		return nil
	}
	return &pps.Transform{
		Image:            t.Image,
		Cmd:              t.Cmd,
		ErrCmd:           t.ErrCmd,
		Env:              t.Env,
		Secrets:          convert1_12SecretMounts(t.Secrets),
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

func convert1_12ParallelismSpec(s *pps1_12.ParallelismSpec) *pps.ParallelismSpec {
	if s == nil {
		return nil
	}
	return &pps.ParallelismSpec{
		Constant:    s.Constant,
		Coefficient: s.Coefficient,
	}
}

func convert1_12HashtreeSpec(h *pps1_12.HashtreeSpec) *pps.HashtreeSpec {
	if h == nil {
		return nil
	}
	return &pps.HashtreeSpec{
		Constant: h.Constant,
	}
}

func convert1_12Egress(e *pps1_12.Egress) *pps.Egress {
	if e == nil {
		return nil
	}
	return &pps.Egress{
		URL: e.URL,
	}
}

func convert1_12GPUSpec(g *pps1_12.GPUSpec) *pps.GPUSpec {
	if g == nil {
		return nil
	}
	return &pps.GPUSpec{
		Type:   g.Type,
		Number: g.Number,
	}
}

func convert1_12ResourceSpec(r *pps1_12.ResourceSpec) *pps.ResourceSpec {
	if r == nil {
		return nil
	}
	return &pps.ResourceSpec{
		Cpu:    r.Cpu,
		Memory: r.Memory,
		Gpu:    convert1_12GPUSpec(r.Gpu),
		Disk:   r.Disk,
	}
}

func convert1_12PFSInput(p *pps1_12.PFSInput) *pps.PFSInput {
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

func convert1_12CronInput(i *pps1_12.CronInput) *pps.CronInput {
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

func convert1_12GitInput(i *pps1_12.GitInput) *pps.GitInput {
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

func convert1_12Input(i *pps1_12.Input) *pps.Input {
	if i == nil {
		return nil
	}
	return &pps.Input{
		Pfs:   convert1_12PFSInput(i.Pfs),
		Cross: convert1_12Inputs(i.Cross),
		Union: convert1_12Inputs(i.Union),
		Cron:  convert1_12CronInput(i.Cron),
		Git:   convert1_12GitInput(i.Git),
	}
}

func convert1_12Inputs(inputs []*pps1_12.Input) []*pps.Input {
	if inputs == nil {
		return nil
	}
	result := make([]*pps.Input, 0, len(inputs))
	for _, i := range inputs {
		result = append(result, convert1_12Input(i))
	}
	return result
}

func convert1_12Service(s *pps1_12.Service) *pps.Service {
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

func convert1_12Spout(s *pps1_12.Spout) *pps.Spout {
	if s == nil {
		return nil
	}
	return &pps.Spout{
		Overwrite: s.Overwrite,
		Service:   convert1_12Service(s.Service),
		Marker:    s.Marker,
	}
}

func convert1_12Metadata(s *pps1_12.Metadata) *pps.Metadata {
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

func convert1_12ChunkSpec(c *pps1_12.ChunkSpec) *pps.ChunkSpec {
	if c == nil {
		return nil
	}
	return &pps.ChunkSpec{
		Number:    c.Number,
		SizeBytes: c.SizeBytes,
	}
}

func convert1_12SchedulingSpec(s *pps1_12.SchedulingSpec) *pps.SchedulingSpec {
	if s == nil {
		return nil
	}
	return &pps.SchedulingSpec{
		NodeSelector:      s.NodeSelector,
		PriorityClassName: s.PriorityClassName,
	}
}

func convert1_12Acl(acl *auth1_12.SetACLRequest) *auth.SetACLRequest {
	req := &auth.SetACLRequest{
		Repo:    acl.Repo,
		Entries: make([]*auth.ACLEntry, len(acl.Entries)),
	}

	for i, entry := range acl.Entries {
		req.Entries[i] = &auth.ACLEntry{
			Username: entry.Username,
			Scope:    auth.Scope(entry.Scope),
		}
	}
	return req
}

func convert1_12ClusterRoleBinding(bindings *auth1_12.ModifyClusterRoleBindingRequest) *auth.ModifyClusterRoleBindingRequest {
	req := &auth.ModifyClusterRoleBindingRequest{
		Principal: bindings.Principal,
		Roles: &auth.ClusterRoles{
			Roles: make([]auth.ClusterRole, len(bindings.Roles.Roles)),
		},
	}

	for i, role := range bindings.Roles.Roles {
		req.Roles.Roles[i] = auth.ClusterRole(role)
	}
	return req
}

func convert1_12AuthConfig(config *auth1_12.SetConfigurationRequest) *auth.SetConfigurationRequest {
	req := &auth.SetConfigurationRequest{
		Configuration: &auth.AuthConfig{
			LiveConfigVersion: config.Configuration.LiveConfigVersion,
		},
	}
	if config.Configuration.SAMLServiceOptions != nil {
		req.Configuration.SAMLServiceOptions = &auth.AuthConfig_SAMLServiceOptions{
			ACSURL:          config.Configuration.SAMLServiceOptions.ACSURL,
			MetadataURL:     config.Configuration.SAMLServiceOptions.MetadataURL,
			DashURL:         config.Configuration.SAMLServiceOptions.DashURL,
			SessionDuration: config.Configuration.SAMLServiceOptions.SessionDuration,
			DebugLogging:    config.Configuration.SAMLServiceOptions.DebugLogging,
		}
	}
	return req
}

func convert1_12Op(op *admin.Op1_12) (*admin.Op1_13, error) {
	switch {
	case op.CreateObject != nil:
		return &admin.Op1_13{
			CreateObject: convert1_12CreateObject(op.CreateObject),
		}, nil
	case op.Job != nil:
		return &admin.Op1_13{
			Job: convert1_12Job(op.Job),
		}, nil
	case op.Tag != nil:
		if !objHashRE.MatchString(op.Tag.Object.Hash) {
			return nil, errors.Errorf("invalid object hash in op: %q", op)
		}
		return &admin.Op1_13{
			Tag: &pfs.TagObjectRequest{
				Object: convert1_12Object(op.Tag.Object),
				Tags:   convert1_12Tags(op.Tag.Tags),
			},
		}, nil
	case op.Repo != nil:
		return &admin.Op1_13{
			Repo: &pfs.CreateRepoRequest{
				Repo:        convert1_12Repo(op.Repo.Repo),
				Description: op.Repo.Description,
			},
		}, nil
	case op.Commit != nil:
		return &admin.Op1_13{
			Commit: &pfs.BuildCommitRequest{
				Parent:     convert1_12Commit(op.Commit.Parent),
				Branch:     op.Commit.Branch,
				Provenance: convert1_12Provenances(op.Commit.Provenance),
				Tree:       convert1_12Object(op.Commit.Tree),
				Trees:      convert1_12Objects(op.Commit.Trees),
				Datums:     convert1_12Object(op.Commit.Datums),
				ID:         op.Commit.ID,
				SizeBytes:  op.Commit.SizeBytes,
			},
		}, nil
	case op.Branch != nil:
		newOp := &admin.Op1_13{
			Branch: &pfs.CreateBranchRequest{
				Head:       convert1_12Commit(op.Branch.Head),
				Branch:     convert1_12Branch(op.Branch.Branch),
				Provenance: convert1_12Branches(op.Branch.Provenance),
			},
		}
		if newOp.Branch.Branch == nil {
			newOp.Branch.Branch = &pfs.Branch{
				Repo: convert1_12Repo(op.Branch.Head.Repo),
				Name: op.Branch.SBranch,
			}
		}
		return newOp, nil
	case op.Pipeline != nil:
		return &admin.Op1_13{
			Pipeline: &pps.CreatePipelineRequest{
				Pipeline:         convert1_12Pipeline(op.Pipeline.Pipeline),
				Transform:        convert1_12Transform(op.Pipeline.Transform),
				ParallelismSpec:  convert1_12ParallelismSpec(op.Pipeline.ParallelismSpec),
				HashtreeSpec:     convert1_12HashtreeSpec(op.Pipeline.HashtreeSpec),
				Egress:           convert1_12Egress(op.Pipeline.Egress),
				Update:           op.Pipeline.Update,
				OutputBranch:     op.Pipeline.OutputBranch,
				ResourceRequests: convert1_12ResourceSpec(op.Pipeline.ResourceRequests),
				ResourceLimits:   convert1_12ResourceSpec(op.Pipeline.ResourceLimits),
				Input:            convert1_12Input(op.Pipeline.Input),
				Description:      op.Pipeline.Description,
				CacheSize:        op.Pipeline.CacheSize,
				EnableStats:      op.Pipeline.EnableStats,
				Reprocess:        op.Pipeline.Reprocess,
				MaxQueueSize:     op.Pipeline.MaxQueueSize,
				Service:          convert1_12Service(op.Pipeline.Service),
				Spout:            convert1_12Spout(op.Pipeline.Spout),
				ChunkSpec:        convert1_12ChunkSpec(op.Pipeline.ChunkSpec),
				DatumTimeout:     op.Pipeline.DatumTimeout,
				JobTimeout:       op.Pipeline.JobTimeout,
				Salt:             op.Pipeline.Salt,
				Standby:          op.Pipeline.Standby,
				DatumTries:       op.Pipeline.DatumTries,
				SchedulingSpec:   convert1_12SchedulingSpec(op.Pipeline.SchedulingSpec),
				PodSpec:          op.Pipeline.PodSpec,
				PodPatch:         op.Pipeline.PodPatch,
				SpecCommit:       convert1_12Commit(op.Pipeline.SpecCommit),
				Metadata:         convert1_12Metadata(op.Pipeline.Metadata),
			},
		}, nil
	case op.SetAcl != nil:
		return &admin.Op1_13{
			SetAcl: convert1_12Acl(op.SetAcl),
		}, nil
	case op.SetClusterRoleBinding != nil:
		return &admin.Op1_13{
			SetClusterRoleBinding: convert1_12ClusterRoleBinding(op.SetClusterRoleBinding),
		}, nil
	case op.SetAuthConfig != nil:
		return &admin.Op1_13{
			SetAuthConfig: convert1_12AuthConfig(op.SetAuthConfig),
		}, nil
	case op.ActivateAuth != nil:
		return &admin.Op1_13{
			ActivateAuth: &auth.ActivateRequest{},
		}, nil
	case op.RestoreAuthToken != nil:
		return &admin.Op1_13{
			RestoreAuthToken: &auth.RestoreAuthTokenRequest{
				Token: &auth.HashedAuthToken{
					HashedToken: op.RestoreAuthToken.Token.HashedToken,
					Expiration:  op.RestoreAuthToken.Token.Expiration,
					TokenInfo: &auth.TokenInfo{
						Subject: op.RestoreAuthToken.Token.TokenInfo.Subject,
						Source:  auth.TokenInfo_TokenSource(op.RestoreAuthToken.Token.TokenInfo.Source),
					},
				},
			},
		}, nil
	case op.ActivateEnterprise != nil:
		return &admin.Op1_13{
			ActivateEnterprise: &enterprise.ActivateRequest{
				ActivationCode: op.ActivateEnterprise.ActivationCode,
			},
		}, nil
	case op.CheckAuthToken != nil:
		return &admin.Op1_13{
			CheckAuthToken: &admin.CheckAuthToken{},
		}, nil
	default:
		return nil, errors.Errorf("unrecognized 1.9 op type:\n%+v", op)
	}
	return nil, errors.Errorf("internal error: convert1_12Op() didn't return a 1.9 op for:\n%+v", op)
}

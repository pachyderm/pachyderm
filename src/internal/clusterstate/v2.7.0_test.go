package clusterstate

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	v2_7_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.7.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Create sample test data "a la collections", i.e. insert data into collections.* tables
func setupTestData(t *testing.T, ctx context.Context, db *sqlx.DB) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// OpenCV example
	// Create project
	projectInfo := pfs.ProjectInfo{Project: &pfs.Project{Name: "opencv"}, Description: "OpenCV project", CreatedAt: timestamppb.Now()}
	b, err := proto.Marshal(&projectInfo)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `INSERT INTO collections.projects(key, proto) VALUES($1, $2)`, v2_7_0.ProjectKey(projectInfo.Project), b)
	require.NoError(t, err)
	// Create repos
	repos := map[string]*pfs.RepoInfo{
		"images":       {Repo: &pfs.Repo{Name: "images", Type: pfs.UserRepoType, Project: &pfs.Project{Name: "opencv"}}},
		"edges":        {Repo: &pfs.Repo{Name: "edges", Type: pfs.UserRepoType, Project: &pfs.Project{Name: "opencv"}}},
		"edges.spec":   {Repo: &pfs.Repo{Name: "edges", Type: pfs.SpecRepoType, Project: &pfs.Project{Name: "opencv"}}},
		"edges.meta":   {Repo: &pfs.Repo{Name: "edges", Type: pfs.MetaRepoType, Project: &pfs.Project{Name: "opencv"}}},
		"montage":      {Repo: &pfs.Repo{Name: "montage", Type: pfs.UserRepoType, Project: &pfs.Project{Name: "opencv"}}},
		"montage.spec": {Repo: &pfs.Repo{Name: "montage", Type: pfs.SpecRepoType, Project: &pfs.Project{Name: "opencv"}}},
		"montage.meta": {Repo: &pfs.Repo{Name: "montage", Type: pfs.MetaRepoType, Project: &pfs.Project{Name: "opencv"}}},
	}
	for _, repoInfo := range repos {
		b, err := proto.Marshal(repoInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.repos(key, proto) VALUES($1, $2)`, v2_7_0.RepoKey(repoInfo.Repo), b)
		require.NoError(t, err)
	}
	// Create commits and commit provenance relationships
	commits := map[string]*pfs.CommitInfo{
		"images@12439bfdb10b4408aa7797efda44be24":       {Commit: &pfs.Commit{Repo: repos["images"].Repo, Id: "12439bfdb10b4408aa7797efda44be24"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"edges@12439bfdb10b4408aa7797efda44be24":        {Commit: &pfs.Commit{Repo: repos["edges"].Repo, Id: "12439bfdb10b4408aa7797efda44be24"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"edges.spec@12439bfdb10b4408aa7797efda44be24":   {Commit: &pfs.Commit{Repo: repos["edges.spec"].Repo, Id: "12439bfdb10b4408aa7797efda44be24"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_USER}},
		"edges.meta@12439bfdb10b4408aa7797efda44be24":   {Commit: &pfs.Commit{Repo: repos["edges.meta"].Repo, Id: "12439bfdb10b4408aa7797efda44be24"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"montage@1062b21221174d7984e1a7ece488e1ca":      {Commit: &pfs.Commit{Repo: repos["montage"].Repo, Id: "1062b21221174d7984e1a7ece488e1ca"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"montage.spec@1062b21221174d7984e1a7ece488e1ca": {Commit: &pfs.Commit{Repo: repos["montage.spec"].Repo, Id: "1062b21221174d7984e1a7ece488e1ca"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_USER}},
		"montage.meta@1062b21221174d7984e1a7ece488e1ca": {Commit: &pfs.Commit{Repo: repos["montage.meta"].Repo, Id: "1062b21221174d7984e1a7ece488e1ca"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"montage@f71d04d35df74baeb8cc4d6ef8a14da6":      {Commit: &pfs.Commit{Repo: repos["montage"].Repo, Id: "f71d04d35df74baeb8cc4d6ef8a14da6"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"montage.meta@f71d04d35df74baeb8cc4d6ef8a14da6": {Commit: &pfs.Commit{Repo: repos["montage.meta"].Repo, Id: "f71d04d35df74baeb8cc4d6ef8a14da6"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"images@f71d04d35df74baeb8cc4d6ef8a14da6":       {Commit: &pfs.Commit{Repo: repos["images"].Repo, Id: "f71d04d35df74baeb8cc4d6ef8a14da6"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_USER}},
		"edges@f71d04d35df74baeb8cc4d6ef8a14da6":        {Commit: &pfs.Commit{Repo: repos["edges"].Repo, Id: "f71d04d35df74baeb8cc4d6ef8a14da6"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"edges.meta@f71d04d35df74baeb8cc4d6ef8a14da6":   {Commit: &pfs.Commit{Repo: repos["edges.meta"].Repo, Id: "f71d04d35df74baeb8cc4d6ef8a14da6"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"montage@a91f6f92b145435396af700be4bb5533":      {Commit: &pfs.Commit{Repo: repos["montage"].Repo, Id: "a91f6f92b145435396af700be4bb5533"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"montage.meta@a91f6f92b145435396af700be4bb5533": {Commit: &pfs.Commit{Repo: repos["montage.meta"].Repo, Id: "a91f6f92b145435396af700be4bb5533"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"images@a91f6f92b145435396af700be4bb5533":       {Commit: &pfs.Commit{Repo: repos["images"].Repo, Id: "a91f6f92b145435396af700be4bb5533"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_USER}},
		"edges@a91f6f92b145435396af700be4bb5533":        {Commit: &pfs.Commit{Repo: repos["edges"].Repo, Id: "a91f6f92b145435396af700be4bb5533"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
		"edges.meta@a91f6f92b145435396af700be4bb5533":   {Commit: &pfs.Commit{Repo: repos["edges.meta"].Repo, Id: "a91f6f92b145435396af700be4bb5533"}, Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO}},
	}
	// images
	commits["images@a91f6f92b145435396af700be4bb5533"].ParentCommit = commits["images@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit
	commits["images@f71d04d35df74baeb8cc4d6ef8a14da6"].ParentCommit = commits["images@12439bfdb10b4408aa7797efda44be24"].Commit
	// edges
	commits["edges@a91f6f92b145435396af700be4bb5533"].ParentCommit = commits["edges@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit
	commits["edges@a91f6f92b145435396af700be4bb5533"].DirectProvenance = []*pfs.Commit{
		commits["images@a91f6f92b145435396af700be4bb5533"].Commit,
		commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit,
	}
	commits["edges@f71d04d35df74baeb8cc4d6ef8a14da6"].ParentCommit = commits["edges@12439bfdb10b4408aa7797efda44be24"].Commit
	commits["edges@f71d04d35df74baeb8cc4d6ef8a14da6"].DirectProvenance = []*pfs.Commit{
		commits["images@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit,
		commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit,
	}
	commits["edges@12439bfdb10b4408aa7797efda44be24"].DirectProvenance = []*pfs.Commit{
		commits["images@12439bfdb10b4408aa7797efda44be24"].Commit,
		commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit,
	}
	// edges.meta
	commits["edges.meta@a91f6f92b145435396af700be4bb5533"].ParentCommit = commits["edges.meta@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit
	commits["edges.meta@a91f6f92b145435396af700be4bb5533"].DirectProvenance = []*pfs.Commit{
		commits["images@a91f6f92b145435396af700be4bb5533"].Commit,
		commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit,
	}
	commits["edges.meta@f71d04d35df74baeb8cc4d6ef8a14da6"].ParentCommit = commits["edges.meta@12439bfdb10b4408aa7797efda44be24"].Commit
	commits["edges.meta@f71d04d35df74baeb8cc4d6ef8a14da6"].DirectProvenance = []*pfs.Commit{
		commits["images@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit,
		commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit,
	}
	commits["edges.meta@12439bfdb10b4408aa7797efda44be24"].DirectProvenance = []*pfs.Commit{
		commits["images@12439bfdb10b4408aa7797efda44be24"].Commit,
		commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit,
	}
	// montage
	commits["montage@a91f6f92b145435396af700be4bb5533"].ParentCommit = commits["montage@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit
	commits["montage@a91f6f92b145435396af700be4bb5533"].DirectProvenance = []*pfs.Commit{
		commits["images@a91f6f92b145435396af700be4bb5533"].Commit,
		commits["edges@a91f6f92b145435396af700be4bb5533"].Commit,
		commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit,
	}
	commits["montage@f71d04d35df74baeb8cc4d6ef8a14da6"].ParentCommit = commits["montage@1062b21221174d7984e1a7ece488e1ca"].Commit
	commits["montage@f71d04d35df74baeb8cc4d6ef8a14da6"].DirectProvenance = []*pfs.Commit{
		commits["images@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit,
		commits["edges@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit,
		commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit,
	}
	commits["montage@1062b21221174d7984e1a7ece488e1ca"].DirectProvenance = []*pfs.Commit{
		commits["images@12439bfdb10b4408aa7797efda44be24"].Commit,
		commits["edges@12439bfdb10b4408aa7797efda44be24"].Commit,
		commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit,
	}
	// montage.meta
	commits["montage.meta@a91f6f92b145435396af700be4bb5533"].ParentCommit = commits["montage@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit
	commits["montage.meta@a91f6f92b145435396af700be4bb5533"].DirectProvenance = []*pfs.Commit{
		commits["images@a91f6f92b145435396af700be4bb5533"].Commit,
		commits["edges@a91f6f92b145435396af700be4bb5533"].Commit,
		commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit,
	}
	commits["montage.meta@f71d04d35df74baeb8cc4d6ef8a14da6"].ParentCommit = commits["montage@1062b21221174d7984e1a7ece488e1ca"].Commit
	commits["montage.meta@f71d04d35df74baeb8cc4d6ef8a14da6"].DirectProvenance = []*pfs.Commit{
		commits["images@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit,
		commits["edges@f71d04d35df74baeb8cc4d6ef8a14da6"].Commit,
		commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit,
	}
	commits["montage.meta@1062b21221174d7984e1a7ece488e1ca"].DirectProvenance = []*pfs.Commit{
		commits["images@12439bfdb10b4408aa7797efda44be24"].Commit,
		commits["edges@12439bfdb10b4408aa7797efda44be24"].Commit,
		commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit,
	}
	for _, commitInfo := range commits {
		b, err := proto.Marshal(commitInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.commits(key, proto) VALUES($1, $2)`, v2_7_0.CommitKey(commitInfo.Commit), b)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO pfs.commits(commit_id, commit_set_id) VALUES($1, $2)`, v2_7_0.CommitKey(commitInfo.Commit), commitInfo.Commit.Id)
		require.NoError(t, err)

	}
	// Create branches and branch provenance relationships
	// TODO branch triggers
	branches := map[string]*pfs.BranchInfo{
		"images@master":       {Branch: &pfs.Branch{Repo: repos["images"].Repo, Name: "master"}, Head: commits["images@a91f6f92b145435396af700be4bb5533"].Commit},
		"edges@master":        {Branch: &pfs.Branch{Repo: repos["edges"].Repo, Name: "master"}, Head: commits["edges@a91f6f92b145435396af700be4bb5533"].Commit},
		"edges.spec@master":   {Branch: &pfs.Branch{Repo: repos["edges.spec"].Repo, Name: "master"}, Head: commits["edges.spec@12439bfdb10b4408aa7797efda44be24"].Commit},
		"edges.meta@master":   {Branch: &pfs.Branch{Repo: repos["edges.meta"].Repo, Name: "master"}, Head: commits["edges.meta@a91f6f92b145435396af700be4bb5533"].Commit},
		"montage@master":      {Branch: &pfs.Branch{Repo: repos["montage"].Repo, Name: "master"}, Head: commits["montage@a91f6f92b145435396af700be4bb5533"].Commit},
		"montage.spec@master": {Branch: &pfs.Branch{Repo: repos["montage.spec"].Repo, Name: "master"}, Head: commits["montage.spec@1062b21221174d7984e1a7ece488e1ca"].Commit},
		"montage.meta@master": {Branch: &pfs.Branch{Repo: repos["montage.meta"].Repo, Name: "master"}, Head: commits["montage.meta@a91f6f92b145435396af700be4bb5533"].Commit},
	}
	branches["edges@master"].DirectProvenance = []*pfs.Branch{
		branches["images@master"].Branch,
		branches["edges.spec@master"].Branch,
	}
	branches["edges.meta@master"].DirectProvenance = []*pfs.Branch{
		branches["images@master"].Branch,
		branches["edges.spec@master"].Branch,
	}
	branches["montage@master"].DirectProvenance = []*pfs.Branch{
		branches["images@master"].Branch,
		branches["edges@master"].Branch,
		branches["montage.spec@master"].Branch,
	}
	branches["montage.meta@master"].DirectProvenance = []*pfs.Branch{
		branches["images@master"].Branch,
		branches["edges@master"].Branch,
		branches["montage.spec@master"].Branch,
	}
	for _, branchInfo := range branches {
		b, err := proto.Marshal(branchInfo)
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, `INSERT INTO collections.branches(key, proto, idx_repo) VALUES($1, $2, $3)`, v2_7_0.BranchKey(branchInfo.Branch), b, v2_7_0.RepoKey(branchInfo.Branch.Repo))
		require.NoError(t, err)
	}

	// TODO create pipelines and jobs

	require.NoError(t, tx.Commit())
}

func Test_v2_7_0_ClusterState(t *testing.T) {
	ctx := pctx.TestContext(t)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, t)
	defer db.Close()
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}

	// Pre-migration
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_6_0))
	setupTestData(t, ctx, db)

	// Get all existing projects in collections.projects including the default project
	expectedProjects, err := v2_7_0.ListProjectsFromCollection(ctx, db)
	require.NoError(t, err)

	// Apply 2.7.0 migration
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, state_2_7_0))
	require.NoError(t, migrations.BlockUntil(ctx, db, state_2_7_0))

	// Verify Projects
	// Check whether all the data is migrated to core.projects table
	var gotProjects []v2_7_0.Project
	require.NoError(t, db.SelectContext(ctx, &gotProjects, `SELECT id, name, description, created_at, updated_at FROM core.projects ORDER BY id`))
	require.Equal(t, len(expectedProjects), len(gotProjects))
	if diff := cmp.Diff(expectedProjects, gotProjects, cmp.Comparer(func(t1, t2 time.Time) bool {
		// Ignore sub-microsecond differences becaues postgres stores timestamps with microsecond precision
		return t1.Sub(t2) < time.Microsecond

	})); diff != "" {
		t.Errorf("projects differ: (-want +got)\n%s", diff)
	}
	// Test project names can only be 51 characters long
	_, err = db.ExecContext(ctx, `INSERT INTO core.projects(name, description) VALUES($1, $2)`, strings.Repeat("A", 52), "project name with 52 characters")
	require.ErrorContains(t, err, "value too long for type character varying(51)")
}

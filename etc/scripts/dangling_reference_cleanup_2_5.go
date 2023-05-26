package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pkg/errors"
)

// pachyderm clusters that have not yet successfully migrated to 2.6.x, this script will detect dangling commit references that could otherwise block the migration
// and print them to stdout. a --fix flag can be passed as the second argument to cleanup those dangling references from the database
//
// run as ./dangling_reference_cleanup postgresql://<DB_NAME>@<HOST>:<PORT>/<USER> --fix
func main() {
	args := os.Args[1:]
	if len(args) < 1 || len(args) > 2 {
		panic("must provide postgres's DSN as a first argument; optionally pass in '--fix' as a second argument to force a fix")
	}
	var dsn string = args[0]
	var fix bool
	if len(args) == 2 {
		if args[1] == "--fix" {
			fix = true
		} else {
			panic("if provided, the only valid second argument is '--fix'")
		}
	}
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		panic(errors.Wrap(err, "connect to database"))
	}
	defer db.Close()
	tx, err := db.Beginx()
	if err != nil {
		panic(errors.Wrap(err, "start transaction"))
	}
	defer tx.Commit()
	cs, err := listReferencedCommits(tx)
	if err != nil {
		panic(errors.Wrap(err, "list referenced commits"))
	}
	rs, err := listRepoKeys(tx)
	if err != nil {
		panic(errors.Wrap(err, "list repos"))
	}
	var dangCommitKeys []string
	for _, c := range cs {
		if _, ok := rs[repoKey(c.Repo)]; !ok {
			dangCommitKeys = append(dangCommitKeys, commitKey_2_5(c))
		}
	}
	fmt.Printf("commits with dangling references %v\n", dangCommitKeys)
	if fix {
		ctx := context.Background()
		for _, id := range dangCommitKeys {
			if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commit_totals WHERE commit_id = $1`, id); err != nil {
				panic(errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_totals", id))
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, id); err != nil {
				panic(errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_diffs", id))
			}
		}
	}
}

func listReferencedCommits(sqlTx *sqlx.Tx) (map[string]*pfs.Commit, error) {
	cs := make(map[string]*pfs.Commit)
	var ids []string
	if err := sqlTx.Select(&ids, `SELECT commit_id from  pfs.commit_totals`); err != nil {
		return nil, errors.Wrap(err, "select commit ids from pfs.commit_totals")
	}
	for _, id := range ids {
		cs[id] = parseCommit_2_5(id)
	}
	ids = make([]string, 0)
	if err := sqlTx.Select(&ids, `SELECT commit_id from  pfs.commit_diffs`); err != nil {
		return nil, errors.Wrap(err, "select commit ids from pfs.commit_diffs")
	}
	for _, id := range ids {
		cs[id] = parseCommit_2_5(id)
	}
	return cs, nil
}

func listRepoKeys(tx *sqlx.Tx) (map[string]struct{}, error) {
	var keys []string
	if err := tx.Select(&keys, `SELECT key FROM collections.repos`); err != nil {
		return nil, errors.Wrap(err, "select keys from collections.repos")
	}
	rs := make(map[string]struct{})
	for _, k := range keys {
		rs[k] = struct{}{}
	}
	return rs, nil
}

func parseCommit_2_5(key string) *pfs.Commit {
	split := strings.Split(key, "=")
	if len(split) != 2 {
		panic(errors.Errorf("parsing commit key with 2.6.x+ structure %q", key))
	}
	b := parseBranch(split[0])
	return &pfs.Commit{
		Repo:   b.Repo,
		Branch: b,
		ID:     split[1],
	}
}

func parseBranch(key string) *pfs.Branch {
	split := strings.Split(key, "@")
	return &pfs.Branch{
		Repo: parseRepo(split[0]),
		Name: split[1],
	}
}

func parseRepo(key string) *pfs.Repo {
	slashSplit := strings.Split(key, "/")
	dotSplit := strings.Split(slashSplit[1], ".")
	return &pfs.Repo{
		Project: &pfs.Project{Name: slashSplit[0]},
		Name:    dotSplit[0],
		Type:    dotSplit[1],
	}
}

func commitKey_2_5(c *pfs.Commit) string {
	return branchKey(c.Branch) + "=" + c.ID
}

func branchKey(b *pfs.Branch) string {
	return repoKey(b.Repo) + "@" + b.Name
}

func repoKey(r *pfs.Repo) string {
	return r.Project.Name + "/" + r.Name + "." + r.Type
}

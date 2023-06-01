package cmds

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/archiveserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap"
)

func Cmds(ctx context.Context) []*cobra.Command {
	var commands []*cobra.Command

	var d net.Dialer
	setupControl(&d)

	r := new(net.Resolver)
	d.Resolver = r
	r.Dial = func(ctx context.Context, network, address string) (_ net.Conn, retErr error) {
		defer log.Span(ctx, "DialDNS", zap.String("network", network), zap.String("address", address))(log.Errorp(&retErr))
		return d.DialContext(ctx, network, address) //nolint:wrapcheck
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = d.DialContext
	t.TLSClientConfig.VerifyConnection = func(cs tls.ConnectionState) error {
		for _, cert := range cs.PeerCertificates {
			fmt.Printf("tls: cert: %v\n", cert.Subject.String())
		}
		fmt.Printf("tls: server name: %v\n", cs.ServerName)
		fmt.Printf("tls: negotiated protocol: %v\n", cs.NegotiatedProtocol)
		fmt.Println()
		return nil
	}
	t.TLSClientConfig.InsecureSkipVerify = true
	hc := &http.Client{
		Transport: promutil.InstrumentRoundTripper("pachctl", t),
	}

	dnsLookup := &cobra.Command{
		Use:   "{{alias}} <hostname>",
		Short: "Do a DNS lookup on a hostname.",
		Long:  "Do a DNS lookup on a hostname.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			var errs error
			if result, err := r.LookupHost(ctx, args[0]); err != nil {
				errors.JoinInto(&errs, errors.Wrap(err, "lookup host"))
			} else {
				fmt.Printf("A: %v\n", result)
			}
			if result, err := r.LookupCNAME(ctx, args[0]); err != nil {
				errors.JoinInto(&errs, errors.Wrap(err, "lookup cname"))
			} else {
				fmt.Printf("CNAME: %v\n", result)
			}
			if result, err := r.LookupAddr(ctx, args[0]); err != nil {
				errors.JoinInto(&errs, errors.Wrap(err, "lookup reverse address"))
			} else {
				fmt.Printf("IP: %v\n", result)
			}
			if errs != nil {
				fmt.Fprintf(os.Stderr, "some lookups not successful, this is normally fine:\n%v", errs)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dnsLookup, "misc dns-lookup"))

	httpHead := &cobra.Command{
		Use:   "{{alias}} <url>",
		Short: "Make an HTTP HEAD request.",
		Long:  "Make an HTTP HEAD request.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			ctx := pctx.Background("")
			req, err := http.NewRequestWithContext(ctx, "HEAD", args[0], nil)
			if err != nil {
				return errors.Wrap(err, "NewRequest")
			}
			if content, err := httputil.DumpRequestOut(req, false); err == nil {
				fmt.Printf("%s", content)
			}
			res, err := hc.Do(req)
			if res != nil {
				defer res.Body.Close()
				if content, err := httputil.DumpResponse(res, true); err == nil {
					fmt.Printf("%s", content)
				}
			}
			if err != nil {
				return errors.Wrap(err, "Do")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(httpHead, "misc http-head"))

	dial := &cobra.Command{
		Use:   "{{alias}} <network(tcp|udp)> <address>",
		Short: "Dials a network server and then disconnects.",
		Long:  "Dials a network server and then disconnects.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			ctx := pctx.Background("")
			conn, err := d.DialContext(ctx, args[0], args[1])
			if err != nil {
				return errors.Wrap(err, "Dial")
			}
			fmt.Printf("OK: %s -> %s\n", conn.LocalAddr(), conn.RemoteAddr())
			return errors.Wrap(conn.Close(), "Close")
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dial, "misc dial"))

	// NOTE(jonathan): This can move out of misc when we add OAuth support to the download
	// endpoint, and tell pachd what its externally-accessible URL is (so the link works when
	// you click it).
	generateURL := &cobra.Command{
		Use:   "{{alias}} project/repo@branch_or_commit:/file_or_directory ...",
		Short: "Generates the encoded part of an archive download URL.",
		Long:  "Generates the encoded part of an archive download URL.",
		Run: cmdutil.Run(func(args []string) error {
			u, err := archiveserver.EncodeV1(args)
			if err != nil {
				return errors.Wrap(err, "encode")
			}
			fmt.Println(u)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(generateURL, "misc generate-download-url"))

	decodeURL := &cobra.Command{
		Use:   "{{alias}} <url>",
		Short: "Decodes the encoded part of an archive download URL.",
		Long:  "Decodes the encoded part of an archive download URL.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			u, err := url.Parse(args[0])
			if err != nil {
				return errors.Wrap(err, "url.Parse")
			}
			if !strings.HasPrefix(u.Path, "/archive/") {
				u.Path = "/archive/" + u.Path
			}
			if !strings.HasSuffix(u.Path, ".zip") {
				u.Path = u.Path + ".zip"
			}
			req, err := archiveserver.ArchiveFromURL(u)
			if err != nil {
				return errors.Wrap(err, "ArchiveFromURL")
			}
			if err := req.ForEachPath(func(path string) error {
				fmt.Println(path)
				return nil
			}); err != nil {
				return errors.Wrap(err, "ForEachPath")
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(decodeURL, "misc decode-download-url"))

	testMigrations := &cobra.Command{
		Use:   "{{alias}} <postgres dsn>",
		Short: "Runs the database migrations against the supplied database, then rolls them back.",
		Long:  "Runs the database migrations against the supplied database, then rolls them back.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			ctx, c := signal.NotifyContext(pctx.Background(""), os.Interrupt)
			defer c()

			dsn := args[0]
			db, err := sqlx.Open("pgx", dsn)
			if err != nil {
				return errors.Wrap(err, "open database")
			}
			if err := dbutil.WaitUntilReady(ctx, db); err != nil {
				return errors.Wrap(err, "wait for database ready")
			}

			// Create test dirs for etcd data
			dir, err := os.MkdirTemp("", "test-migrations")
			if err != nil {
				return errors.Wrap(err, "create etcd server tmpdir")
			}
			defer os.RemoveAll(dir)

			etcdConfig := embed.NewConfig()
			etcdConfig.MaxTxnOps = 10000
			etcdConfig.Dir = filepath.Join(dir, "dir")
			etcdConfig.WalDir = filepath.Join(dir, "wal")
			etcdConfig.InitialElectionTickAdvance = false
			etcdConfig.TickMs = 10
			etcdConfig.ElectionMs = 50
			etcdConfig.ListenPeerUrls = []url.URL{}
			etcdConfig.ListenClientUrls = []url.URL{{
				Scheme: "http",
				Host:   "localhost:7777",
			}}
			log.AddLoggerToEtcdServer(ctx, etcdConfig)
			etcd, err := embed.StartEtcd(etcdConfig)
			if err != nil {
				return errors.Wrap(err, "start etcd")
			}
			defer etcd.Close()

			etcdCfg := log.GetEtcdClientConfig(ctx)
			etcdCfg.Endpoints = []string{"http://localhost:7777"}
			etcdCfg.DialOptions = client.DefaultDialOptions()
			etcdClient, err := clientv3.New(etcdCfg)
			if err != nil {
				return errors.Wrap(err, "connect to etcd")
			}
			defer etcdClient.Close()

			txx, err := db.BeginTxx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
			})
			if err != nil {
				return errors.Wrap(err, "start tx")
			}
			defer func() {
				if err := txx.Rollback(); err != nil {
					errors.JoinInto(&retErr, errors.Wrap(err, "rollback"))
				}
			}()
			states := migrations.CollectStates(nil, clusterstate.DesiredClusterState)
			env := migrations.MakeEnv(nil, etcdClient)
			env.Tx = txx
			var errs error
			for _, s := range states {
				if err := migrations.ApplyMigrationTx(ctx, env, s); err != nil {
					log.Error(ctx, "migration did not apply; continuing", zap.Error(err))
					errors.JoinInto(&retErr, errors.Wrapf(err, "migration %v", s.Number()))
				}
			}
			log.Info(ctx, "done applying migrations")
			return errs
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(testMigrations, "misc test-migrations"))

	var fix bool
	danglingCommitRefs := &cobra.Command{
		Use:   "{{alias}} <postgres dsn>",
		Short: "Detects dangling commit references in the database. If the --fix flag is provided, it will delete them.",
		Long:  "Detects dangling commit references in the database. If the --fix flag is provided, it will delete them.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			parseRepo := func(key string) *pfs.Repo {
				slashSplit := strings.Split(key, "/")
				dotSplit := strings.Split(slashSplit[1], ".")
				return &pfs.Repo{
					Project: &pfs.Project{Name: slashSplit[0]},
					Name:    dotSplit[0],
					Type:    dotSplit[1],
				}
			}
			parseBranch := func(key string) *pfs.Branch {
				split := strings.Split(key, "@")
				return &pfs.Branch{
					Repo: parseRepo(split[0]),
					Name: split[1],
				}
			}
			repoKey := func(r *pfs.Repo) string {
				return r.Project.Name + "/" + r.Name + "." + r.Type
			}
			branchKey := func(b *pfs.Branch) string {
				return repoKey(b.Repo) + "@" + b.Name
			}
			commitKey_2_5 := func(c *pfs.Commit) string {
				return branchKey(c.Branch) + "=" + c.ID
			}
			listRepoKeys := func(tx *sqlx.Tx) (map[string]struct{}, error) {
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
			parseCommit_2_5 := func(key string) (*pfs.Commit, error) {
				split := strings.Split(key, "=")
				if len(split) != 2 {
					return nil, errors.Errorf("parsing commit key with 2.6.x+ structure %q", key)
				}
				b := parseBranch(split[0])
				return &pfs.Commit{
					Repo:   b.Repo,
					Branch: b,
					ID:     split[1],
				}, nil
			}
			listReferencedCommits := func(sqlTx *sqlx.Tx) (map[string]*pfs.Commit, error) {
				cs := make(map[string]*pfs.Commit)
				var err error
				var ids []string
				if err := sqlTx.Select(&ids, `SELECT commit_id from  pfs.commit_totals`); err != nil {
					return nil, errors.Wrap(err, "select commit ids from pfs.commit_totals")
				}
				for _, id := range ids {
					cs[id], err = parseCommit_2_5(id)
					if err != nil {
						return nil, err
					}
				}
				ids = make([]string, 0)
				if err := sqlTx.Select(&ids, `SELECT commit_id from  pfs.commit_diffs`); err != nil {
					return nil, errors.Wrap(err, "select commit ids from pfs.commit_diffs")
				}
				for _, id := range ids {
					cs[id], err = parseCommit_2_5(id)
					if err != nil {
						return nil, err
					}
				}
				return cs, nil
			}
			dsn := args[0]
			db, err := sqlx.Open("pgx", dsn)
			if err != nil {
				return errors.Wrap(err, "connect to database")
			}
			defer db.Close()
			tx, err := db.Beginx()
			if err != nil {
				return errors.Wrap(err, "start transaction")
			}
			defer func() {
				if err := tx.Commit(); err != nil {
					errors.JoinInto(&retErr, errors.Wrap(err, "commit transaction"))
				}
			}()
			cs, err := listReferencedCommits(tx)
			if err != nil {
				return errors.Wrap(err, "list referenced commits")
			}
			rs, err := listRepoKeys(tx)
			if err != nil {
				return errors.Wrap(err, "list repos")
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
						return errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_totals", id)
					}
					if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, id); err != nil {
						return errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_diffs", id)
					}
				}
			}
			return nil
		}),
	}
	danglingCommitRefs.Flags().BoolVar(&fix, "fix", false, "Whether to delete dangling references.")
	commands = append(commands, cmdutil.CreateAlias(danglingCommitRefs, "misc dangling-commit-refs"))

	misc := &cobra.Command{
		Short:  "Miscellaneous utilities unrelated to Pachyderm itself.",
		Long:   "Miscellaneous utilities unrelated to Pachyderm itself.  These utilities can be removed or changed in minor releases; do not rely on them.",
		Hidden: true,
	}
	commands = append(commands, cmdutil.CreateAlias(misc, "misc"))
	return commands
}

package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"io"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/zeebo/blake3"
	"go.uber.org/zap"
)

var (
	source = flag.String("source", "", "debug dump archive to import")
	v      = flag.Bool("v", false, "verbose logs")
	dsn    = flag.String("dsn", "host=localhost port=5432 database=pachydermlogs user=postgres pool_max_conns=128", "postgres dsn")
)

var nerr int
var nrows, ndups, nlines, njson, nfiles, nbinary, nbytesin, nbytesout, ngoro, nconnacq atomic.Int64

func addFile(rctx context.Context, r *bytes.Buffer, name string, db *pgxpool.Conn) (retErr error) {
	ctx, done := log.SpanContext(rctx, "addFile", zap.String("filename", name))
	defer done(log.Errorp(&retErr))
	nfiles.Add(1)
	ngoro.Add(1)
	defer ngoro.Add(-1)

	slop := int64(r.Len())
	defer func() {
		nbytesout.Add(slop)
	}()

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Bytes()
		if !utf8.Valid(line) {
			nbinary.Add(1)
			continue
		}
		nlines.Add(1)
		nbytesout.Add(int64(len(line)))
		slop -= int64(len(line)) // account for unknown number of newlines

		hash := blake3.Sum256(line)
		hashBytes := make([]byte, 32)
		copy(hashBytes, hash[:])

		var parsed map[string]any
		var parseError sql.NullString
		if err := json.Unmarshal(s.Bytes(), &parsed); err != nil {
			parseError.Valid = true
			parseError.String = err.Error()
		} else {
			njson.Add(1)
		}

		var ts sql.NullTime
		for _, key := range []string{"time", "ts", "timestamp"} {
			t, err := DefaultTimeParser(parsed[key])
			if err != nil {
				continue
			}
			delete(parsed, key)
			ts.Valid = true
			ts.Time = t
			break
		}

		tag, err := db.Exec(ctx, "insert into logs(hash, dumpname, filename, time, line, parsed, parseerror) values ($1, $2, $3, $4, $5, $6, $7) on conflict do nothing", hashBytes, *source, name, ts, line, parsed, parseError)
		if err != nil {
			if err := context.Cause(ctx); err != nil {
				return err //nolint:wrapcheck
			}
			log.Info(ctx, "inserting line failed", zap.Error(err))
		}
		switch tag.RowsAffected() {
		case 0:
			ndups.Add(1)
		default:
			nrows.Add(tag.RowsAffected())
		}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "scan")
	}
	return nil
}

func printStats(ctx context.Context) {
	stats := func(m string) {
		log.Info(ctx, m, zap.Int64("duplicates", ndups.Load()), zap.Int64("rows", nrows.Load()), zap.Int64("json", njson.Load()), zap.Int64("lines", nlines.Load()), zap.Int64("files", nfiles.Load()), zap.Int64("binary", nbinary.Load()), zap.Int64("bytes_read", nbytesin.Load()), zap.Int64("bytes_processed", nbytesout.Load()), zap.Int64("bytes_in_flight", nbytesin.Load()-nbytesout.Load()), zap.Int64("running", ngoro.Load()), zap.Int64("awaiting_db", nconnacq.Load()))
	}
	for {
		select {
		case <-ctx.Done():
			stats("final stats")
			return
		case <-time.After(time.Second):
			stats("stats")
		}

	}
}

func main() {
	flag.Parse()
	done := log.InitBatchLogger("")
	defer func() {
		var err error
		if nerr > 0 {
			err = errors.New("dump errored")
		}
		done(err)
	}()

	if *v {
		log.SetLevel(log.DebugLevel)
	}

	ctx, c := pctx.Interactive()

	pool, err := pgxpool.Connect(ctx, *dsn)
	if err != nil {
		log.Exit(ctx, "problem connecting to database", zap.Error(err))
	}

	fr, err := os.Open(*source)
	if err != nil {
		log.Exit(ctx, "problem opening source", zap.Error(err))
	}
	gr, err := gzip.NewReader(fr)
	if err != nil {
		log.Exit(ctx, "problem creating gzip reader", zap.Error(err))
	}

	go printStats(ctx)

	wg := new(sync.WaitGroup)
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// but wait for everything else to finish
			log.Error(ctx, "problem reading dump", zap.Error(err))
			nerr++
			break
		}

		var buf bytes.Buffer
		n, err := io.Copy(&buf, tr)
		if err != nil {
			log.Debug(ctx, "read error", zap.String("filename", h.Name), zap.Error(err))
			continue
		}
		nbytesin.Add(n)

		nconnacq.Add(1)
		c, err := pool.Acquire(ctx)
		nconnacq.Add(-1)
		if err != nil {
			log.Error(ctx, "cannot acquire db conn", zap.Error(err))
			nerr++
			break
		}

		wg.Add(1)
		go func() {
			addFile(ctx, &buf, h.Name, c) //nolint:errcheck
			c.Release()
			wg.Done()
		}()
	}
	log.Info(ctx, "finishing up processing")
	wg.Wait()
	c()
}

// copied from jlog:
// DefaultTimeParser treats numbers as seconds since the Unix epoch and strings as RFC3339 timestamps.
func DefaultTimeParser(in interface{}) (time.Time, error) {
	float64AsTime := func(x float64) time.Time {
		return time.Unix(int64(math.Floor(x)), int64(1_000_000_000*(x-math.Floor(x))))
	}
	toInt := func(m map[string]interface{}, k string) (int64, bool) {
		v, ok := m[k]
		if !ok {
			return 0, false
		}
		floatVal, ok := v.(float64)
		if !ok {
			return 0, false
		}
		return int64(math.Floor(floatVal)), true
	}

	switch x := in.(type) {
	case int:
		return time.Unix(int64(x), 0), nil
	case int64:
		return time.Unix(x, 0), nil
	case float64:
		return float64AsTime(x), nil
	case string:
		t, err := time.Parse(time.RFC3339, x)
		if err != nil {
			return time.Time{}, errors.Errorf("interpreting string timestamp as RFC3339: %v", err)
		}
		return t, nil
	case map[string]interface{}: // logrus -> joonix Stackdriver format
		sec, sok := toInt(x, "seconds")
		nsec, nsok := toInt(x, "nanos")
		if !(sok && nsok) {
			return time.Time{}, errors.Errorf("map[string]interface{}%v not in stackdriver format", x)
		}
		return time.Unix(sec, nsec), nil
	default:
		return time.Time{}, errors.Errorf("invalid time format %T(%v)", x, x)
	}
}

const (
	//nolint:unused
	schema = `
create extension pg_trgm;
create table logs (hash bytea not null primary key, dumpname text not null, filename text not null, time timestamptz null, line text not null, parsed jsonb null, parseerror text null);
create index logs_text on logs using gin(line gin_trgm_ops);
create index logs_json on logs using gin(parsed);
`
)

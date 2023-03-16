// Command lokitool uses our code to query Loki.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

var (
	addr       = flag.String("loki", "http://localhost:3100", "where loki is")
	from       = flag.Duration("from", -30*24*time.Hour, "the start time of the query, relative to now")
	to         = flag.Duration("to", 0, "the end time of the query, relative to now")
	query      = flag.String("query", `{pipelineName="edges"}`, "the query string to send to loki")
	printLines = flag.Bool("printLines", false, "whether to print the returned lines")
)

func main() {
	flag.Parse()
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx := pctx.Background("")
	l := &loki.Client{
		Address: *addr,
	}
	var n int
	if err := lokiutil.QueryRange(ctx, l, *query, time.Now().Add(*from), time.Now().Add(*to), false, func(t time.Time, line string) error {
		n++
		if *printLines {
			fmt.Printf("%v: %s\n", t.Local().Format(time.RFC3339), line)
		}
		return nil
	}); err != nil {
		log.Error(ctx, "problem querying loki", zap.Error(err))
	}
	fmt.Fprintf(os.Stderr, "%d line(s)\n", n)
}

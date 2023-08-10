package v2_5_0

import (
	"context"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

type postgresBatcher struct {
	ctx   context.Context
	tx    *pachsql.Tx
	stmts []string
	max   int
	num   int
}

func newPostgresBatcher(ctx context.Context, tx *pachsql.Tx) *postgresBatcher {
	return &postgresBatcher{
		ctx: ctx,
		tx:  tx,
		max: 100,
	}
}

func (pb *postgresBatcher) Add(stmt string) error {
	pb.stmts = append(pb.stmts, stmt)
	if len(pb.stmts) < pb.max {
		return nil
	}
	return pb.execute()
}

func (pb *postgresBatcher) execute() error {
	if len(pb.stmts) == 0 {
		return nil
	}
	log.Info(pb.ctx, "executing postgres statement batch",
		zap.String("number", strconv.Itoa(pb.num)),
		zap.String("size", strconv.Itoa(len(pb.stmts))),
	)
	pb.num++
	if _, err := pb.tx.ExecContext(pb.ctx, strings.Join(pb.stmts, ";")); err != nil {
		return errors.EnsureStack(err)
	}
	pb.stmts = nil
	return nil
}

func (pb *postgresBatcher) Close() error {
	return pb.execute()
}

package pjs

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"testing"

	"google.golang.org/grpc"
)

func NewTestClient(t testing.TB, db *pachsql.DB) pjs.APIClient {
	srv := NewAPIServer(Env{
		DB: db,
	})
	gc := grpcutil.NewTestClient(t, func(s *grpc.Server) {
		pjs.RegisterAPIServer(s, srv)
	})
	return pjs.NewAPIClient(gc)
}

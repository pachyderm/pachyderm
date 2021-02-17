package server

import (
	"context"

	"github.com/blevesearch/bleve"
	"github.com/pachyderm/pachyderm/v2/src/catalog"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

var _ catalog.APIServer = &apiServer{}

type apiServer struct {
	env   *serviceenv.ServiceEnv
	index bleve.Index
}

func newAPIServer(env *serviceenv.ServiceEnv) (*apiServer, error) {
	go func() { env.GetPachClient(context.Background()) }() // Begin dialing connection on startup
	index, err := bleve.New("/pachyderm.bleve", bleve.NewIndexMapping())
	if err != nil {
		return nil, err
	}
	return &apiServer{
		env:   env,
		index: index,
	}, nil
}

func (a *apiServer) Query(ctx context.Context, req *catalog.QueryRequest) (*catalog.QueryResponse, error) {
	searchResult, err := a.index.Search(bleve.NewSearchRequest(bleve.NewQueryStringQuery(req.Query)))
	if err != nil {
		return nil, err
	}
	result := &catalog.QueryResponse{}
	for _, doc := range searchResult.Hits {
		result.Results = append(result.Results, doc.ID)
	}
	return result, nil
}

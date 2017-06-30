package server

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/auth"
)

type apiServer struct {
	etcdClient *etcd.Client
}

func (a *apiServer) PutActivationCode(context.Context, *auth.ActivationCode) (*types.Empty, error) {
	return &types.Empty{}, nil
}
func (a *apiServer) GetActivationCode(context.Context, *types.Empty) (*auth.ActivationCode, error) {
	return &auth.ActivationCode{}, nil
}

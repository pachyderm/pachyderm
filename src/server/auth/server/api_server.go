package server

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client/auth"
)

const (
	etcdActivationCode = "activation-code"
	publicKey          = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAoaPoEfv5RcVUbCuWNnOB
WtLHzcyQSe4SbtGGQom/X27iq/7s8dcebSsCd2cwYoyKihEQ5OlaghrhcxTTV5AN
39O6S0YnWjt/+4PWQQP3NpcEhqWj8RLPJtYq+JNrqlyjxBlca7vDcFSTa6iCqXay
iVD2OyTbWrD6KZ/YTSmSY8mY2qdYvHyp3Ue5ueH3rSkKRUjo4Jyjf59PntZD884P
yb9kC+weh/1KlbDQ4aV0U9p6DSBkW7dinOQj7a1/ikDoA9Nebnrkb1FF9Hr2+utO
We4e4yOViDzAP9hhQiBhOVR0F6wJF5i+NfuLit4tk5ViboogEZqIyuakTD6abSFg
UPqBTDDG0UsVqjnU5ysJ1DKQqALnOrxEKZoVXtH80/m7kgmeY3VDHCFt+WCSdaSq
1w8SoIpJAZPJpKlDjMxe+NqsX2qUODQ2KNkqfEqFtyUNZzfS9o9pEg/KJzDuDclM
oMQr1BG8vc3msX4UiGQPkohznwlCSGWf62IkSS6P8hQRCBKGRS5yGjmT3J+/chZw
Je46y8zNLV7t2pOL6UemdmDjTaMCt0YBc1FmG2eUipAWcHJWEHgQm2Yz6QjtBgvt
jFqnYeiDwdxU7CQD3oF9H+uVHqz8Jmmf9BxY9PhlMSUGPUsTpZ717ysL0UrBhQhW
xYp8vpeQ3by9WxPBE/WrxN8CAwEAAQ==
-----END PUBLIC KEY-----
`
)

type apiServer struct {
	etcdClient *etcd.Client
}

func (a *apiServer) PutActivationCode(ctx context.Context, request *auth.ActivationCode) (*types.Empty, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to pem decode public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: %+v", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key isn't an RSA key")
	}
	hashed := sha256.Sum256([]byte(request.Token))
	if err := rsa.VerifyPSS(rsaPub, crypto.SHA256, hashed[:], []byte(request.Signature), nil); err != nil {
		return nil, err
	}
	marshaler := &jsonpb.Marshaler{}
	activationCode, err := marshaler.MarshalToString(request)
	if err != nil {
		return nil, err
	}
	if _, err := a.etcdClient.Put(ctx, etcdActivationCode, activationCode); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}
func (a *apiServer) GetActivationCode(ctx context.Context, _ *types.Empty) (*auth.ActivationCode, error) {
	resp, err := a.etcdClient.Get(ctx, etcdActivationCode)
	if err != nil {
		return nil, err
	}
	result := &auth.ActivationCode{}
	jsonpb.UnmarshalString(string(resp.Kvs[0].Value), result)
	return result, nil
}

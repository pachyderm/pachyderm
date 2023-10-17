package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	det "github.com/determined-ai/determined/proto/pkg/apiv1"
	userv1 "github.com/determined-ai/determined/proto/pkg/userv1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	mlc "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type provisioner interface {
	FindUser(ctx context.Context, name string) (*User, error)
	CreateUser(context.Context, *User) (*User, error)
	FindGroup(ctx context.Context, name string) (*Group, error)
	CreateGroup(context.Context, *Group) (*Group, error)
	SetUserGroups(context.Context, *User, []*Group) error
	Close() error
}

type User struct {
	id         int32
	name       string
	currGroups []*Group
}

type Group struct {
	id   int32
	name string
}

type determinedProvisioner struct {
	conn *grpc.ClientConn
	dc   det.DeterminedClient
}

type DeterminedConfig struct {
	MasterURL string
	Username  string
	Password  string
	TLS       bool
}

type errNotFound struct{}

func (e errNotFound) Error() string {
	return fmt.Sprintf("not found")
}

func NewDeterminedProvisioner(ctx context.Context, config DeterminedConfig) (provisioner, error) {
	tlsOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
	if config.TLS {
		tlsOpt = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		}))
	}
	determinedURL, err := url.Parse(config.MasterURL)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing determined url %q", config.MasterURL)
	}
	conn, err := grpc.DialContext(ctx, determinedURL.Host, tlsOpt, grpc.WithStreamInterceptor(mlc.LogStream), grpc.WithUnaryInterceptor(mlc.LogUnary))
	if err != nil {
		return nil, errors.Wrapf(err, "dialing determined at %q", determinedURL.Host)
	}
	defer conn.Close()
	dc := det.NewDeterminedClient(conn)
	tok, err := mintDeterminedToken(ctx, dc, config.Username, config.Password)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-user-token", fmt.Sprintf("Bearer %s", tok))
	d := &determinedProvisioner{
		conn: conn,
		dc:   dc,
	}
	return d, nil
}

func mintDeterminedToken(ctx context.Context, dc det.DeterminedClient, username, password string) (string, error) {
	loginResp, err := dc.Login(ctx, &det.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", errors.Wrap(err, "login as determined user")
	}
	return loginResp.Token, nil
}

func (d *determinedProvisioner) FindUser(ctx context.Context, name string) (*User, error) {
	u, err := d.dc.GetUserByUsername(ctx, &det.GetUserByUsernameRequest{Username: name})
	if err != nil {
		return nil, errors.Wrapf(err, "get determined user %q", name)
	}
	return &User{name: u.GetUser().GetUsername()}, nil
}

func (d *determinedProvisioner) CreateUser(ctx context.Context, user *User) (*User, error) {
	u, err := d.dc.PostUser(ctx, &det.PostUserRequest{User: &userv1.User{
		Username: user.name,
		Active:   true,
		Remote:   true,
	}})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to provision determiend user %v", user.name)
	}
	return &User{id: u.User.Id, name: u.User.Username}, err
}

func (d *determinedProvisioner) FindGroup(ctx context.Context, name string) (*Group, error) {
	g, err := d.dc.GetGroups(ctx, &det.GetGroupsRequest{Name: name})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find determiend group %v", name)
	}
	if len(g.Groups) == 0 {
		return nil, &errNotFound{}
	}
	return &Group{id: g.Groups[0].Group.GroupId, name: name}, nil
}

func (d *determinedProvisioner) CreateGroup(ctx context.Context, group *Group) (*Group, error) {
	g, err := d.dc.CreateGroup(ctx, &det.CreateGroupRequest{Name: group.name})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create determiend group %v", group.name)
	}
	return &Group{id: g.Group.GroupId, name: g.Group.Name}, nil
}

func (d *determinedProvisioner) SetUserGroups(ctx context.Context, user *User, groups []*Group) error {
	for _, g := range groups {
		if _, err := d.dc.UpdateGroup(ctx, &det.UpdateGroupRequest{
			GroupId:  g.id,
			AddUsers: []int32{user.id},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *determinedProvisioner) Close() error {
	return d.conn.Close()
}

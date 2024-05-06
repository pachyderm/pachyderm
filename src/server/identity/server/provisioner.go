package server

import (
	"context"
	"strings"

	det "github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/userv1"

	detutil "github.com/pachyderm/pachyderm/v2/src/internal/determined"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type provisioner interface {
	findUser(ctx context.Context, name string) (*user, error)
	createUser(context.Context, *user) (*user, error)
	findGroup(ctx context.Context, name string) (*group, error)
	createGroup(context.Context, *group) (*group, error)
	setUserGroups(context.Context, *user, []*group) error
	close() error
}

type user struct {
	id         int32
	name       string
	prevGroups []*group
}

type group struct {
	id   int32
	name string
}

type determinedProvisioner struct {
	dc    det.DeterminedClient
	cf    func() error
	token string
}

type errNotFound struct {
	err error
}

func (e errNotFound) Error() string {
	return e.err.Error()
}

func (e errNotFound) Unwrap() error {
	return e.err
}

func newDeterminedProvisioner(ctx context.Context, config detutil.Config) (provisioner, error) {
	dc, cf, err := detutil.NewClient(ctx, config.MasterURL, config.TLS)
	if err != nil {
		return nil, errors.Wrapf(err, "build determined client with endpoint %q", config.MasterURL)
	}
	token, err := detutil.MintToken(ctx, dc, config.Username, config.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "mint determined token for user", config.Username)
	}
	d := &determinedProvisioner{
		dc:    dc,
		cf:    cf,
		token: token,
	}
	return d, nil
}

func (d *determinedProvisioner) findUser(ctx context.Context, name string) (*user, error) {
	ctx = detutil.WithToken(ctx, d.token)
	u, err := d.dc.GetUserByUsername(ctx, &det.GetUserByUsernameRequest{Username: name})
	if err != nil {
		err = errors.Wrapf(err, "get determined user %q", name)
		if strings.Contains(err.Error(), "not found") {
			return nil, errNotFound{err: err}
		}
		return nil, err
	}
	gs, err := d.dc.GetGroups(ctx, &det.GetGroupsRequest{UserId: u.User.GetId(), Limit: 500})
	if err != nil {
		err = errors.Wrapf(err, "get groups for determined user %q", name)
		if strings.Contains(err.Error(), "not found") {
			return nil, errNotFound{err: err}
		}
		return nil, err
	}
	var prevGrps []*group
	for _, g := range gs.Groups {
		prevGrps = append(prevGrps, &group{
			id:   g.Group.GetGroupId(),
			name: g.Group.GetName(),
		})
	}
	return &user{
		id:         u.User.GetId(),
		name:       u.GetUser().GetUsername(),
		prevGroups: prevGrps,
	}, nil
}

func (d *determinedProvisioner) createUser(ctx context.Context, usr *user) (*user, error) {
	ctx = detutil.WithToken(ctx, d.token)
	u, err := d.dc.PostUser(ctx, &det.PostUserRequest{User: &userv1.User{
		Username:    usr.name,
		Active:      true,
		Remote:      true,
		Admin:       false,
		DisplayName: usr.name,
	}})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to provision determiend user %v", usr.name)
	}
	return &user{id: u.User.Id, name: u.User.Username}, nil
}

func (d *determinedProvisioner) findGroup(ctx context.Context, name string) (*group, error) {
	ctx = detutil.WithToken(ctx, d.token)
	g, err := d.dc.GetGroups(ctx, &det.GetGroupsRequest{Name: name, Limit: 500})
	if err != nil {
		return nil, errNotFound{err: errors.Wrapf(err, "failed to find determiend group %v", name)}
	}
	if len(g.Groups) == 0 {
		return nil, &errNotFound{}
	}
	return &group{id: g.Groups[0].Group.GroupId, name: name}, nil
}

func (d *determinedProvisioner) createGroup(ctx context.Context, grp *group) (*group, error) {
	ctx = detutil.WithToken(ctx, d.token)
	g, err := d.dc.CreateGroup(ctx, &det.CreateGroupRequest{Name: grp.name})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create determiend group %v", grp.name)
	}
	return &group{id: g.Group.GroupId, name: g.Group.Name}, nil
}

func (d *determinedProvisioner) setUserGroups(ctx context.Context, usr *user, groups []*group) error {
	ctx = detutil.WithToken(ctx, d.token)
	gIds := make(map[int32]struct{})
	for _, g := range groups {
		gIds[g.id] = struct{}{}
		if _, err := d.dc.UpdateGroup(ctx, &det.UpdateGroupRequest{
			GroupId:  g.id,
			AddUsers: []int32{usr.id},
		}); err != nil {
			return errors.Wrapf(err, "set user group %q for user %q", g.name, usr.name)
		}
	}
	for _, g := range usr.prevGroups {
		if _, ok := gIds[g.id]; !ok {
			if _, err := d.dc.UpdateGroup(ctx, &det.UpdateGroupRequest{
				RemoveUsers: []int32{usr.id},
			}); err != nil {
				return errors.Wrapf(err, "remove determined user %q from group %q", usr.name, g)
			}
		}
	}
	return nil
}

func (d *determinedProvisioner) close() error {
	return errors.Wrap(d.cf(), "close the determined client's grpc connection")
}

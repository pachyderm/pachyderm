package auth

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcify returns an error e such that e.Error() is similar to what grpc
// errors emit (though this doesn't convert 'error' to an actual GRPC error)
func grpcify(err error) error {
	return status.Convert(err).Err()
}

func TestIsErrNotActivated(t *testing.T) {
	require.False(t, IsErrNotActivated(nil))
	require.True(t, IsErrNotActivated(ErrNotActivated))
	require.True(t, IsErrNotActivated(grpcify(ErrNotActivated)))
}

func TestIsErrAlreadyActivated(t *testing.T) {
	require.False(t, IsErrAlreadyActivated(nil))
	require.True(t, IsErrAlreadyActivated(ErrAlreadyActivated))
	require.True(t, IsErrAlreadyActivated(grpcify(ErrAlreadyActivated)))
}

func TestIsErrNotSignedIn(t *testing.T) {
	require.False(t, IsErrNotSignedIn(nil))
	require.True(t, IsErrNotSignedIn(ErrNotSignedIn))
	require.True(t, IsErrNotSignedIn(grpcify(ErrNotSignedIn)))
}

func TestIsErrNoMetadata(t *testing.T) {
	require.False(t, IsErrNoMetadata(nil))
	require.True(t, IsErrNoMetadata(ErrNoMetadata))
	require.True(t, IsErrNoMetadata(grpcify(ErrNoMetadata)))
}

func TestIsErrBadToken(t *testing.T) {
	require.False(t, IsErrBadToken(nil))
	require.True(t, IsErrBadToken(ErrBadToken))
	require.True(t, IsErrBadToken(grpcify(ErrBadToken)))
}

func TestIsErrNotAuthorized(t *testing.T) {
	require.False(t, IsErrNotAuthorized(nil))
	require.True(t, IsErrNotAuthorized(&ErrNotAuthorized{
		Subject:  "alice",
		Resource: &Resource{Type: ResourceType_REPO, Name: "data"},
		Required: []Permission{},
	}))
	require.True(t, IsErrNotAuthorized(grpcify(&ErrNotAuthorized{
		Subject:  "alice",
		Resource: &Resource{Type: ResourceType_REPO, Name: "data"},
		Required: []Permission{},
	})))
	require.True(t, IsErrNotAuthorized(&ErrNotAuthorized{
		Subject:  "alice",
		Resource: &Resource{Type: ResourceType_CLUSTER},
		Required: []Permission{},
	}))
	require.True(t, IsErrNotAuthorized(grpcify(&ErrNotAuthorized{
		Subject:  "alice",
		Resource: &Resource{Type: ResourceType_CLUSTER},
		Required: []Permission{},
	})))
	s, ok := status.FromError(grpcify(&ErrNotAuthorized{
		Subject:  "alice",
		Resource: &Resource{Type: ResourceType_CLUSTER},
		Required: []Permission{},
	}))
	require.True(t, ok, "ErrNotAuthorized should be a gRPC status")
	require.Equal(t, s.Code(), codes.PermissionDenied, "ErrNotAuthorized should be a gRPC PermissionDenied")
}

func TestErrNoRoleBinding(t *testing.T) {
	require.True(t, IsErrNoRoleBinding(&ErrNoRoleBinding{
		&Resource{Type: ResourceType_REPO, Name: "test"},
	}))
	require.True(t, IsErrNoRoleBinding(grpcify(&ErrNoRoleBinding{
		&Resource{Type: ResourceType_REPO, Name: "test"},
	})))
}

func TestIsErrInvalidPrincipal(t *testing.T) {
	require.False(t, IsErrInvalidPrincipal(nil))
	require.True(t, IsErrInvalidPrincipal(&ErrInvalidPrincipal{
		Principal: "alice",
	}))
	require.True(t, IsErrInvalidPrincipal(grpcify(&ErrInvalidPrincipal{
		Principal: "alice",
	})))
}

func TestIsErrTooShortTTL(t *testing.T) {
	require.False(t, IsErrTooShortTTL(nil))
	require.True(t, IsErrTooShortTTL(ErrTooShortTTL{
		RequestTTL:  1234,
		ExistingTTL: 2345,
	}))
	require.True(t, IsErrTooShortTTL(grpcify(ErrTooShortTTL{
		RequestTTL:  1234,
		ExistingTTL: 2345,
	})))
}

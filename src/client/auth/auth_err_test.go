package auth

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// grpcify returns an error e such that e.Error() is similar to what grpc
// errors emit (though this doesn't convert 'error' to an actual GRPC error)
func grpcify(err error) error {
	return errors.Errorf("rpc error: code = Unknown desc = %s", err.Error())
}

func TestIsErrNotActivated(t *testing.T) {
	require.False(t, IsErrNotActivated(nil))
	require.True(t, IsErrNotActivated(ErrNotActivated))
	require.True(t, IsErrNotActivated(grpcify(ErrNotActivated)))
}

func TestIsErrPartiallyActivated(t *testing.T) {
	require.False(t, IsErrPartiallyActivated(nil))
	require.True(t, IsErrPartiallyActivated(ErrPartiallyActivated))
	require.True(t, IsErrPartiallyActivated(grpcify(ErrPartiallyActivated)))
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
		Repo:     "data",
		Required: Scope_WRITER,
	}))
	require.True(t, IsErrNotAuthorized(grpcify(&ErrNotAuthorized{
		Subject:  "alice",
		Repo:     "data",
		Required: Scope_WRITER,
	})))
	require.True(t, IsErrNotAuthorized(&ErrNotAuthorized{
		Subject: "alice",
		AdminOp: "GetAuthToken on another user's token",
	}))
	require.True(t, IsErrNotAuthorized(grpcify(&ErrNotAuthorized{
		Subject: "alice",
		AdminOp: "GetAuthToken on another user's token",
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

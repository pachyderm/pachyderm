package s2

import (
	"net/http"
)

// AuthController is an interface defining authentication
type AuthController interface {
	// SecretKey is called when a request is made using AWS' auth V4 or V2. If
	// the given access key exists, a non-nil secret key should be returned.
	// Otherwise nil should be returned.
	SecretKey(r *http.Request, accessKey string, region *string) (*string, error)
	// CustomAuth handles requests that are not using AWS' auth V4 or V2. You
	// can use this to implement custom auth algorithms. Return true if the
	// request passes the auth check.
	CustomAuth(r *http.Request) (bool, error)
}

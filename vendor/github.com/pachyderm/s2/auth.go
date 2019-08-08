package s2

import (
	"net/http"
)

// AuthController is an interface defining authentication
type AuthController interface {
	SecretKey(r *http.Request, accessKey string, region *string) (secretKey *string, err error)
	CustomAuth(r *http.Request) (passed bool, err error)
}

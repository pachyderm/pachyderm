// Package license implements license validation.
package license

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	publicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtJnDuD05fJZVsWDvN/un
m5xbG7jcmxUsSOQZfvMaafZjV6iG/z6Wst2uhcMGAMrLHBxFiRYiVVM3kbUhbfbw
3nVzALDLh4l/QzovCcF12FzVY8fB5Q6VQFfnup1aKimyJX7/au0ihvv//olQ1xrL
XRaG7h/hnCbmjLhsaGA6nqB4gtRI+HI3tBvQBicaN0P5pcfJlT49BSgJq6pnbZPY
SmXeL5m/o1sWZzjzlkmXuxxptG8WTDU3cYF2wmGNMDV/e7u7TuvnFLEz+xf8MUcq
LrDaDj1OuQVwftfz+jqZunQifx4pq6Sxk3ecQll2OhHE1LHrDdE+KSYumUVr0h5i
OVro2tqn4CUmwWrDb4O3TxowrNHylXWAWsLukXQCxguYPRRdIlpu8QPYvsdjU0xT
F7sRv8juuBMSOwRnEZE0M0E/XeLiJo9ROzVxHbRga2AHgDtt0rVHrUrlKmJFJyU2
DACvluEWcjXKXRJJkeieSQopITTQtBSYVu0fr1HG1pLOs1ZakPRPUi/xnSnDb2zK
XinORcb47IsWIHXtwHcwY1C7kV0IK3DxJrJZsSib171vAwi6q/HSOSkWxCURsOtK
x90hW9XbejJCpAiOYfPEOq0lT8fy1Ve0qBen1y4mcxtnXANrgQyYCCBftoc7Ctkk
m5MuBYYSa4PH/uIZktTYOkMCAwEAAQ==
-----END PUBLIC KEY-----
`
)

// ActivationCode is the outer JSON structure of an enterprise license
type ActivationCode struct {
	Token     string
	Signature string
}

// Token is the inner JSON structure of an enterprise license. These
// claims are signed and must be kept in sync with the license generation tool.
type Token struct {
	Expiry string
}

// Validate checks the validity of an enterprise license code
func Validate(code string) (expiration time.Time, err error) {
	// Parse the public key.  If these steps fail, something is seriously
	// wrong and we should crash the service by panicking.
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return time.Time{}, errors.Errorf("failed to pem decode public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "failed to parse DER encoded public key")
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return time.Time{}, errors.Errorf("public key isn't an RSA key")
	}

	activationCode, err := Unmarshal(code)
	if err != nil {
		return time.Time{}, err
	}

	// Decode the signature
	decodedSignature, err := base64.StdEncoding.DecodeString(activationCode.Signature)
	if err != nil {
		return time.Time{}, errors.Errorf("signature is not base64 encoded")
	}

	// Compute the sha256 checksum of the token
	hashedToken := sha256.Sum256([]byte(activationCode.Token))

	// Verify that the signature is valid
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hashedToken[:], decodedSignature); err != nil {
		return time.Time{}, errors.Errorf("invalid signature in activation code")
	}

	// Unmarshal the token
	token := Token{}
	if err := json.Unmarshal([]byte(activationCode.Token), &token); err != nil {
		return time.Time{}, errors.Errorf("token is not valid JSON")
	}

	// Parse the expiration. Note that this string is generated by Date.toJSON()
	// running in node, so Go's definition of RFC 3339 timestamps (which is
	// incomplete) must be compatible with the strings that node generates. So far
	// it seems to work.
	expiration, err = time.Parse(time.RFC3339, token.Expiry)
	if err != nil {
		return time.Time{}, errors.Errorf("expiration is not valid ISO 8601 string")
	}
	// Check that the activation code has not expired
	if time.Now().After(expiration) {
		return time.Time{}, errors.Errorf("the activation code has expired")
	}
	return expiration, nil
}

// Unmarshal deserializes the outer base64-encoded JSON payload of an enterprise license
func Unmarshal(code string) (*ActivationCode, error) {
	// Decode the base64-encoded activation code
	decodedActivationCode, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return nil, errors.Errorf("activation code is not base64 encoded")
	}
	activationCode := &ActivationCode{}
	if err := json.Unmarshal(decodedActivationCode, &activationCode); err != nil {
		return nil, errors.Errorf("activation code is not valid JSON")
	}

	return activationCode, nil
}

package main

import (
	"context"
	"fmt"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

const (
	vaultAddress = "http://127.0.0.1:8200"
	pachdAddress = "127.0.0.1:30650"
	pluginName   = "pachyderm"
)

func configurePlugin(t *testing.T, v *vault.Client) {

	c, err := client.NewFromAddress(pachdAddress)
	if err != nil {
		t.Errorf(err.Error())
	}
	resp, err := c.Authenticate(
		context.Background(),
		&auth.AuthenticateRequest{GitHubUsername: "admin", GitHubToken: "y"})

	if err != nil {
		t.Errorf(err.Error())
	}

	vl := v.Logical()
	config := make(map[string]interface{})
	config["admin_token"] = resp.PachToken
	config["pachd_address"] = pachdAddress
	_, err = vl.Write(
		fmt.Sprintf("/%v/config", pluginName),
		config,
	)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestLogin(t *testing.T) {
	// Negative control:
	//     Before we have a valid pach token, we should not
	// be able to list admins
	c, err := client.NewFromAddress(pachdAddress)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = c.AuthAPIClient.GetAdmins(context.Background(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Errorf("client could list admins before using auth token. this is likely a bug")
	}

	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Errorf(err.Error())
	}
	v.SetToken("root")
	// Hit login before admin token is set, expect err

	configurePlugin(t, v)

	// Now hit login endpoint w invalid vault token, expect err
	config := make(map[string]interface{})
	config["username"] = "daffyduck"
	vl := v.Logical()
	secret, err := vl.Write(
		fmt.Sprintf("/%v/login", pluginName),
		config,
	)

	if err != nil {
		t.Errorf(err.Error())
	}

	pachToken, ok := secret.Auth.Metadata["user_token"]
	if !ok {
		t.Errorf("vault login response did not contain user token")
	}
	reportedPachdAddress, ok := secret.Auth.Metadata["pachd_address"]
	if !ok {
		t.Errorf("vault login response did not contain pachd address")
	}

	// Now do the actual test:
	// Try and list admins w a client w a valid pach token
	c, err = client.NewFromAddress(reportedPachdAddress)
	if err != nil {
		t.Errorf(err.Error())
	}
	c.SetAuthToken(pachToken)

	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestLoginTTL(t *testing.T) {
	// Same as above, validate that pach token expires after given TTL
}

func TestRenew(t *testing.T) {
	// Does login, issues renew request before TTL expires

	// Does login, issues renew request after TTL expires (expect err)
}

func TestRevoke(t *testing.T) {
	// Do normal login
	// Use user token to connect
	// Issue revoke
	// Now renewal should fail ... but that token should still work? AFAICT
}

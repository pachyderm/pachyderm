package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

const (
	vaultAddress = "http://127.0.0.1:8200"
	pluginName   = "pachyderm"
)

var pachClient *client.APIClient
var getPachClientOnce sync.Once

func getPachClient(t testing.TB) *client.APIClient {
	getPachClientOnce.Do(func() {
		var err error
		if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewOnUserMachine(false, "user")
		}
		if err != nil {
			t.Fatalf(err.Error())
		}
	})
	return pachClient
}

func configurePlugin(t *testing.T, v *vault.Client, ttl string) error {
	c := getPachClient(t)
	resp, err := c.Authenticate(
		context.Background(),
		&auth.AuthenticateRequest{GitHubToken: "admin"})

	if err != nil {
		return err
	}

	return configurePluginHelper(c, v, resp.PachToken, c.GetAddress(), ttl)
}

func configurePluginHelper(pachClient *client.APIClient, v *vault.Client, testPachToken string, testPachdAddress string, ttl string) error {
	vl := v.Logical()
	config := make(map[string]interface{})
	config["admin_token"] = testPachToken
	config["pachd_address"] = testPachdAddress
	if ttl != "" {
		config["ttl"] = ttl
	}
	secret, err := vl.Write(
		fmt.Sprintf("/%v/config", pluginName),
		config,
	)
	if err != nil {
		return err
	}
	if respErr, ok := secret.Data["error"]; ok {
		return fmt.Errorf("error in response: %v (%T)", respErr, respErr)
	}
	return nil
}

func TestBadConfig(t *testing.T) {
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	v.SetToken("root")
	c := getPachClient(t)
	resp, err := c.Authenticate(
		context.Background(),
		&auth.AuthenticateRequest{GitHubToken: "admin"})

	if err != nil {
		t.Fatalf(err.Error())
	}

	// make sure we get an error for missing admin_token
	err = configurePluginHelper(c, v, "", "", "")
	if err == nil {
		t.Fatalf("expected error: missing token in config (but got none)")
	}

	// make sure we get an error for missing pachd_address (not set in configurePluginHelper)
	err = configurePluginHelper(c, v, resp.PachToken, "", "")
	if err == nil {
		t.Fatalf("expected missing address in config to error")
	}

	// make sure that missing TTL is OK
	err = configurePluginHelper(c, v, resp.PachToken, c.GetAddress(), "")
	if err != nil {
		t.Fatalf(err.Error())
	}

	// make sure that malformed TTL is not OK
	err = configurePluginHelper(c, v, resp.PachToken, c.GetAddress(), "234....^^^")
	if err == nil {
		t.Fatalf("expected bad ttl in config to error")
	}
}

func TestMinimalConfig(t *testing.T) {
	c := getPachClient(t)
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	v.SetToken("root")

	// Test using just defaults
	// We'll see an error if the admin token / pachd address are not set
	err = configurePlugin(t, v, "")
	if err != nil {
		t.Fatalf(err.Error())
	}

	vl := v.Logical()
	secret, err := vl.Read(
		fmt.Sprintf("/%v/config", pluginName),
	)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if secret.Data["pachd_address"] != c.GetAddress() {
		t.Fatalf("pachd_address configured incorrectly")
	}
	if secret.Data["ttl"] == "0s" {
		t.Fatalf("ttl configured incorrectly")
	}
	ttlIface, ok := secret.Data["ttl"]
	if !ok {
		t.Fatalf("ttl wasn't set in config, but it should always have a default value")
	}
	ttlStr, ok := ttlIface.(string)
	if !ok {
		t.Fatalf("ttl has the wrong type (should be string but was %T)", ttlIface)
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		t.Fatalf("could not parse duration (%s) from config: %v", ttlStr, err)
	}
	if ttl < (30 * time.Second) {
		t.Fatalf("ttl configured incorrectly; should take default vaule but is actually (%v)", secret.Data["ttl"])
	}

}

func loginHelper(t *testing.T, ttl string) (*client.APIClient, *vault.Client, *vault.Secret) {
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	v.SetToken("root")

	err = configurePlugin(t, v, ttl)
	if err != nil {
		t.Fatalf(err.Error())
	}

	vl := v.Logical()
	params := make(map[string]interface{})
	if ttl != "" {
		params["ttl"] = ttl
	}
	secret, err := vl.Write(
		fmt.Sprintf("/%v/login/github:bogusgithubusername", pluginName), params)

	if err != nil {
		t.Fatalf(err.Error())
	}

	pachToken, ok := secret.Data["user_token"].(string)
	if !ok {
		t.Fatalf("vault login response did not contain user token")
	}
	reportedPachdAddress, ok := secret.Data["pachd_address"].(string)
	if !ok {
		t.Fatalf("vault login response did not contain pachd address")
	}

	c, err := client.NewFromAddress(reportedPachdAddress)
	if err != nil {
		t.Fatalf(err.Error())
	}
	c.SetAuthToken(pachToken)

	return c, v, secret
}

func TestLogin(t *testing.T) {
	// Negative control: before we have a valid pach token, we should not
	// be able to list admins
	c := getPachClient(t)
	_, err := c.AuthAPIClient.GetAdmins(context.Background(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Fatalf("client could list admins before using auth token. this is likely a bug")
	}

	c, _, _ = loginHelper(t, "")

	// Now do the actual test: try and list admins w a client w a valid pach token
	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}
}

// TestLoginExpires tests two features:
// 1. Returned Pachyderm tokens are revoked when their vault lease expires
// 2. If a TTL is set in the plugin config and not in the login request, then
//    the TTL from the config is applied to the login token
func TestLoginExpires(t *testing.T) {
	c, _, secret := loginHelper(t, "2s")

	// Make sure token is valid
	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Wait for TTL to expire and check that token is no longer valid
	time.Sleep(time.Duration(secret.LeaseDuration+1) * time.Second)
	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Fatalf("API call should fail, but token did not expire")
	}
}

// TestLoginTTLParam tests that if TTL is set in the request, make sure the
// returned token has that TTL
func TestLoginTTLParam(t *testing.T) {
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	v.SetToken("root")

	err = configurePlugin(t, v, "")
	if err != nil {
		t.Fatalf(err.Error())
	}

	vl := v.Logical()
	secret, err := vl.Write(
		fmt.Sprintf("/%v/login/github:bogusgithubusername", pluginName),
		map[string]interface{}{"ttl": "2s"},
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if secret.LeaseDuration != 2 {
		t.Fatalf("Expected pachyderm token with TTL=2s, but was %ds", secret.LeaseDuration)
	}
	pachToken := secret.Data["user_token"].(string)
	reportedPachdAddress := secret.Data["pachd_address"].(string)
	c, err := client.NewFromAddress(reportedPachdAddress)
	if err != nil {
		t.Fatalf(err.Error())
	}
	c.SetAuthToken(pachToken)
	// Make sure token is valid
	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Wait for TTL to expire and check that token is no longer valid
	time.Sleep(time.Duration(secret.LeaseDuration+1) * time.Second)
	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Fatalf("API call should fail, but token did not expire")
	}
}

// If TTL is not in the request OR the config, make sure a sensible default is
// used
func TestLoginDefaultTTL(t *testing.T) {
	c, _, secret := loginHelper(t, "")
	if secret.LeaseDuration < 600 {
		t.Fatalf("Expected Pachyderm token with duration at least 10m, but actual "+
			"duration was %ds", secret.LeaseDuration)
	}

	// Make sure token is valid
	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}
}

// TestLoginLongTTL tests that it's possible to get a Pachyderm token with a
// long TTL
func TestLoginLongTTL(t *testing.T) {
	// Get Vault client
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	v.SetToken("root")
	vl := v.Logical()

	// Get Pachyderm token with long TTL
	secret, err := vl.Write(
		fmt.Sprintf("/%v/login/github:bogusgithubusername", pluginName),
		map[string]interface{}{"ttl": "768h"},
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if secret.LeaseDuration < int((700 * time.Hour).Seconds()) {
		t.Fatalf("Expected pachyderm token with TTL=768h, but was %ds", secret.LeaseDuration)
	}

	// Make sure token is valid
	pachToken, ok := secret.Data["user_token"].(string)
	if !ok {
		t.Fatalf("vault login response did not contain user token")
	}
	reportedPachdAddress, ok := secret.Data["pachd_address"].(string)
	if !ok {
		t.Fatalf("vault login response did not contain pachd address")
	}
	c, err := client.NewFromAddress(reportedPachdAddress)
	if err != nil {
		t.Fatalf(err.Error())
	}
	c.SetAuthToken(pachToken)
	if _, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{}); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestRenewBeforeTTLExpires(t *testing.T) {
	ttl := 10
	c, v, secret := loginHelper(t, fmt.Sprintf("%vs", ttl))

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	renewer, err := v.NewRenewer(&vault.RenewerInput{
		Secret:    secret,
		Increment: ttl,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	time.Sleep(time.Duration(ttl/2) * time.Second)
	go renewer.Renew()
	defer renewer.Stop()

	select {
	case err := <-renewer.DoneCh():
		if err != nil {
			t.Fatalf(err.Error())
		}
	case <-renewer.RenewCh():
	}
	time.Sleep(time.Duration(ttl/2+1) * time.Second)

	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestRenewAfterTTLExpires(t *testing.T) {
	ttl := 2
	c, v, secret := loginHelper(t, fmt.Sprintf("%vs", ttl))

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	renewer, err := v.NewRenewer(&vault.RenewerInput{
		Secret:    secret,
		Increment: ttl,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	time.Sleep(time.Duration(ttl+1) * time.Second)
	go renewer.Renew()
	defer renewer.Stop()

	select {
	case err := <-renewer.DoneCh():
		if err == nil {
			t.Fatalf("Expected an error renewing but got none\n")
		}
	case <-renewer.RenewCh():
		t.Fatal("Expected failed renewal, but got successful renewal\n")
	}

	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Fatalf("Expected error using pach token after expiry, but got no error\n")
	}
}

func TestRevoke(t *testing.T) {
	c, v, secret := loginHelper(t, "")
	if secret.LeaseDuration < 60 {
		t.Fatalf("Expected Pachyderm token to have long lease duration, but was %d", secret.LeaseDuration)
	}

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	vl := v.Logical()
	_, err = vl.Write(
		fmt.Sprintf("/sys/leases/revoke"),
		map[string]interface{}{"lease_id": secret.LeaseID},
	)

	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Fatalf("expected error with revoked pach token, got none\n")
	}
}

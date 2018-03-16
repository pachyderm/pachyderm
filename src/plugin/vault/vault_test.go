package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	vault_plugin "github.com/pachyderm/pachyderm/src/plugin/vault/pachyderm"
)

const (
	vaultAddress = "http://127.0.0.1:8200"
	pachdAddress = "127.0.0.1:30650"
	pluginName   = "pachyderm"
)

func configurePlugin(v *vault.Client, ttl string) error {

	c, err := client.NewFromAddress(pachdAddress)
	if err != nil {
		return err
	}
	resp, err := c.Authenticate(
		context.Background(),
		&auth.AuthenticateRequest{GitHubToken: "admin"})

	if err != nil {
		return err
	}

	return configurePluginHelper(v, resp.PachToken, pachdAddress, ttl)
}

func configurePluginHelper(v *vault.Client, testPachToken string, testPachdAddress string, ttl string) error {

	vl := v.Logical()
	config := make(map[string]interface{})
	config["admin_token"] = testPachToken
	config["pachd_address"] = testPachdAddress
	if ttl != "" {
		config["ttl"] = ttl
	}
	_, err := vl.Write(
		fmt.Sprintf("/%v/config", pluginName),
		config,
	)

	if err != nil {
		return err
	}

	return nil
}

func TestBadConfig(t *testing.T) {
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Errorf(err.Error())
	}
	v.SetToken("root")
	c, err := client.NewFromAddress(pachdAddress)
	if err != nil {
		t.Errorf(err.Error())
	}
	resp, err := c.Authenticate(
		context.Background(),
		&auth.AuthenticateRequest{GitHubToken: "admin"})

	if err != nil {
		t.Errorf(err.Error())
	}

	err = configurePluginHelper(v, "", pachdAddress, "")
	if err == nil {
		t.Errorf("expected missing token in config to error")
	}
	err = configurePluginHelper(v, resp.PachToken, "", "")
	if err == nil {
		t.Errorf("expected missing address in config to error")
	}
	err = configurePluginHelper(v, resp.PachToken, pachdAddress, "234....^^^")
	if err == nil {
		t.Errorf("expected bad ttl in config to error")
	}

}

func TestMinimalConfig(t *testing.T) {
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Errorf(err.Error())
	}
	v.SetToken("root")

	// Test using just defaults
	// We'll see an error if the admin token / pachd address are not set
	err = configurePlugin(v, "")
	if err != nil {
		t.Errorf(err.Error())
	}

	vl := v.Logical()
	secret, err := vl.Read(
		fmt.Sprintf("/%v/config", pluginName),
	)

	if err != nil {
		t.Errorf(err.Error())
	}

	if secret.Data["pachd_address"] != pachdAddress {
		t.Errorf("pachd_address configured incorrectly")
	}
	if secret.Data["ttl"] == "0s" {
		t.Errorf("ttl configured incorrectly")
	}
	if secret.Data["ttl"] != vault_plugin.DefaultTTL {
		t.Errorf("ttl configured incorrectly, should be default (%v) but is actually (%v)", vault_plugin.DefaultTTL, secret.Data["ttl"])
	}

}

func loginHelper(t *testing.T, ttl string) (*client.APIClient, *vault.Client, *vault.Secret) {
	vaultClientConfig := vault.DefaultConfig()
	vaultClientConfig.Address = vaultAddress
	v, err := vault.NewClient(vaultClientConfig)
	if err != nil {
		t.Errorf(err.Error())
	}
	v.SetToken("root")

	err = configurePlugin(v, ttl)
	if err != nil {
		t.Errorf(err.Error())
	}

	params := make(map[string]interface{})
	params["username"] = "bogusgithubusername"
	vl := v.Logical()
	secret, err := vl.Write(
		fmt.Sprintf("/%v/login", pluginName),
		params,
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

	c, err := client.NewFromAddress(reportedPachdAddress)
	if err != nil {
		t.Errorf(err.Error())
	}
	c.SetAuthToken(pachToken)

	return c, v, secret
}

func TestLogin(t *testing.T) {
	// Negative control: before we have a valid pach token, we should not
	// be able to list admins
	c, err := client.NewFromAddress(pachdAddress)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = c.AuthAPIClient.GetAdmins(context.Background(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Errorf("client could list admins before using auth token. this is likely a bug")
	}

	c, _, _ = loginHelper(t, "")

	// Now do the actual test: try and list admins w a client w a valid pach token
	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestLoginExpires(t *testing.T) {
	c, _, secret := loginHelper(t, "2s")

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Errorf(err.Error())
	}

	time.Sleep(time.Duration(secret.Auth.LeaseDuration+1) * time.Second)
	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Errorf("API call should fail, but token did not expire")
	}
}

func TestRenewBeforeTTLExpires(t *testing.T) {
	ttl := 10
	c, vaultClient, secret := loginHelper(t, fmt.Sprintf("%vs", ttl))

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Errorf(err.Error())
	}

	renewer, err := vaultClient.NewRenewer(&vault.RenewerInput{
		Secret:    secret,
		Increment: ttl,
	})
	if err != nil {
		t.Errorf(err.Error())
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
		t.Errorf(err.Error())
	}
}

func TestRenewAfterTTLExpires(t *testing.T) {
	ttl := 2
	c, vaultClient, secret := loginHelper(t, fmt.Sprintf("%vs", ttl))

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Errorf(err.Error())
	}

	renewer, err := vaultClient.NewRenewer(&vault.RenewerInput{
		Secret:    secret,
		Increment: ttl,
	})
	if err != nil {
		t.Errorf(err.Error())
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
		t.Errorf("Expected error using pach token after expiry, but got no error\n")
	}
}

func TestRevoke(t *testing.T) {
	ttl := 2
	c, v, secret := loginHelper(t, fmt.Sprintf("%vs", ttl))

	_, err := c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err != nil {
		t.Errorf(err.Error())
	}

	params := make(map[string]interface{})
	params["user_token"] = secret.Auth.Metadata["user_token"]
	vl := v.Logical()
	_, err = vl.Write(
		fmt.Sprintf("/%v/revoke", pluginName),
		params,
	)

	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = c.AuthAPIClient.GetAdmins(c.Ctx(), &auth.GetAdminsRequest{})
	if err == nil {
		t.Errorf("expected error with revoked pach token, got none\n")
	}
}

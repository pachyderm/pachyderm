# Vault Secret Engine

Pachyderm supports Vault integration by providing a Vault Secret Engine.


## Deployment

Vault instructions for the admin deploying/configuring/managing vault

1) Get plugin binary

- Navigate to the Pachyderm Repo [on github]()
    - Go to the latest release page
    - Download the `vault` asset

2) Download / Install that binary on your vault server instance

On your vault server:

```
# Assuming the binary was downloaded to /tmp/vault-plugins/pachyderm
export SHASUM=$(shasum -a 256 "/tmp/vault-plugins/pachyderm" | cut -d " " -f1)
echo $SHASUM
vault write sys/plugins/catalog/pachyderm sha_256="$SHASUM" command="pachyderm"
vault secrets enable -path=$PLUGIN_PATH -plugin-name=pachyderm plugin
```

3) Configure the plugin

You'll need to provide the plugin with a few fields for it to work:

- `admin_token` : is the (machine user) pachyderm token the plugin will use to cut new credentials on behalf of users
- `pachd_address` : is the URL where the pachyderm cluster can be accessed
- `ttl` : is the max TTL a token can be issued (TODO)

```
vault write pachyderm/config \
    admin_token="${ADMIN_TOKEN}" \
	pachd_address="${ADDRESS:-127.0.0.1}:30650"
```

To get a machine user `admin_token` from pachyderm:

```
$ yes | pachctl auth activate -u admin
$ pachctl auth login admin
$ cat ~/.pachyderm/config.json | jq -r '.v1.session_token'
f7d120ee5084466b984376f34f3f06f6
```

4) Manage user tokens with `revoke`

```
$vault token revoke d2f1f95c-2445-65ab-6a8b-546825e4997a
Success! Revoked token (if it existed)
```

Which will revoke the vault token. But if you also want to manually revoke a pachyderm token, you can do so by issuing:

```
$vault write pachyderm/revoke user_token=xxx

```

TODO: The above needs to be implemented. It's unclear how the sidechannel/etc works for the revocation step.


## Usage

When your application needs to access pachyderm, you will first do the following:

1) Connect / login to vault

Depending on your language / deployment this can vary. [see the vault documentation]() for more details.

2) Anytime you are going to issue a request to a pachyderm cluster first:

- check to see if you have a valid pachyderm token
    - if you do not have a token, hit the `login` path as described below
    - if you have a token but it's TTL will expire soon (latter half of TTL is what's recommended), hit the `renew` path as described below
- then use the response token when constructing your client to talk to the pachyderm cluster

### login

Again, your client could be in any language. But as an example using the vault CLI:

```
$ vault write pachyderm/login username=bogusgithubusername
Key                         Value
---                         -----
token                       365686aa-d4e5-6b64-d557-aeb196ebd596
token_accessor              73afdb69-1654-0227-62ba-9f5aba741e11
token_duration              45s
token_renewable             true
token_policies              [default]
token_meta_pachd_address    127.0.0.1:30650
token_meta_user_token       bc1e9979f8f7410d8bee606f914a65fc

```

The response metadata contains the `user_token` that you need to use to connect to the pachyderm cluster,
    as well as the `pachd_address` and the `TTL`.

So, again if you wanted to use this on the command line:


```
$cat ~/.pachyderm/config.json | jq -r '.v1.session_token |= "bc1e9979f8f7410d8bee606f914a65fc"'
{
  "user_id": "ac6ff71a39d541149a1081a33b3b7af3",
  "v1": {
    "session_token": "bc1e9979f8f7410d8bee606f914a65fc"
  }
}
$ADDRESS=127.0.0.1:30650 pachctl list-repo
```

### renew

You should issue a `renew` request once the halfway mark of the TTL has elapsed. The only argument for renewal is the tokens UUID which you receive from the `login` endpoint.

```
$vault token renew 681f1d67-6865-aee0-994d-f119f7f118a0
Key                         Value
---                         -----
token                       681f1d67-6865-aee0-994d-f119f7f118a0
token_accessor              3e5e9716-9095-fc5d-a918-e74265f8dcf2
token_duration              2m5s
token_renewable             true
token_policies              [default]
token_meta_pachd_address    127.0.0.1:30650
token_meta_user_token       3af25f5e74134f16b3464d4df4f5c7c9
```


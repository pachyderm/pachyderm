# Infra Folder

This folder captures the infra as code required to run Jupyterlab/Jupyterhub. This is currently used to 
create preview environments of the current build in PRs.

Preview environments are built per branch/PR. They are keyed on the git branch name (which becomes part of the URL)
Each environment maps to a pulumi "stack" which is an instance of the code contained in `main.go` in this folder.
The pulumi stack is created (upserted) when a branch is pushed using the `preview` job in `.circleci.yml`
This is done in circle to coordinate with the existing build process, ensuring the environment is updated
after the build is complete. You can view these stacks in the pulumi console here: https://app.pulumi.com/pachyderm/jupyter-test
(Creds in 1pass in the engineering vault)

When a PR is closed, a github action is used to remove the resources in the stack. A Github Action is used here as it's 
difficult to trigger on "PR closed" in circle, but is trivial in Github.

Files

```
root.py - Passed to the helm chart, sets `SYS_ADMIN` capabilities on the Jupyterlab pods, which is required for fuse
main.go - The pulumi program, also contains a reference helm values section for Jupyterhub
Pulumi.yaml - Defines this folder as a pulumi program
```

# Troubleshooting

The pulumi commands can be run on your workstation. You'll need to download the pulumi CLI app and run `pulumi login` to login

Once that's complete, you can select your stack by running (from `infra` folder)

```
pulumi stack select -c <stack name>
```

You can then update your stack manually by running

```
pulumi up
```

You can also destroy your stack manually by running

```
pulumi destroy
```

## Renewing the certificate


```
certbot certonly --config-dir ~/letsencrypt/config --work-dir ~/letsencrypt/work --logs-dir ~/letsencrypt/logs -d *.clusters-ci.pachyderm.io --agree-tos --manual --preferred-challenges=dns --email buildbot@pachyderm.io --server https://acme-v02.api.letsencrypt.org/directory 
```

```
cd ~/letsencrypt/config/live/clusters-ci.pachyderm.io/
```

```
kubectl delete secret wildcard-tls
```

```
kubectl create secret tls wildcard-tls --key=privkey.pem --cert=fullchain.pem
````

```
kubectl edit secret wildcard-tls
```

Add annotation

```
replicator.v1.mittwald.de/replicate-to-matching: needs-ci-tls=true
```

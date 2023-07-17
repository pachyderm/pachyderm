# jonathan-install-scripts

This is how I develop Pachyderm locally. It's speedy and accurate.

You need to create some files before you can use this:

`enterprise-key-values.yaml`:

```yaml
pachd:
    enterpriseLicenseKey: "<the bytes of your license key>"
```

`hostname-values.yaml`:

```yaml
ingress:
    host: 192.168.1.70
```

It might work if you use `localhost`, but I forget, since I run my browser on a different machine
than what this is running on and have to use 192.168.1.70. Using `localhost` does work on the
command line even if `ingress.host` is some IP address; this is mostly for console and OAuth
clients.

To select which Helm chart and code is deployed, `cd` to the Pachyderm checkout that contains both.
If this repo is checked out to `../install`, then you can run:

```shell
../install/setup.sh
```

Pachyderm is now available at "localhost:80" (or just localhost in your browser). A pachyderm
context "kind" is also created that is connected and logged in as root.

The first install depends on your local Docker having a `pachyderm/pachd:local` image tagged.
`rebuild.sh` does this.

When you want to deploy a new version based on your working directory:

```shell
../install/rebuild.sh
```

If you just changed the Helm chart and don't need to rebuild executables:

```shell
../install/upgrade.sh
```

Upgrade takes args like "diff --show-secrets" if you have "helm diff" installed and want a diff.

Whenever you run `rebuild.sh`, the versions of the images are recorded in `./version.yaml`; so you
can `upgrade.sh` to change Helm values without changing the code that is running.

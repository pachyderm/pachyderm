# Using Skaffold for local dev

## Install Skaffold

See [here](https://skaffold.dev/docs/getting-started/#installing-skaffold)

## Up and running

Run

```
skaffold dev
```

## Notes / Gotchas
- Skaffold will automatically detect the `docker-for-desktop` / `docker-desktop` contexts and not attempt to push an image. This can be overridden with:
```
build:
  local:
    push: false
```
- To refresh the Kubernetes deployment files, run the following:
```
pachctl deploy local --dry-run -o yaml > k8s/local.yaml
```
Then remove the tag from pachd deployment image: (See issue [here](https://github.com/GoogleContainerTools/skaffold/issues/2547))
```
image: pachyderm/pachd:1.9.3 -> image: pachyderm/pachd
```

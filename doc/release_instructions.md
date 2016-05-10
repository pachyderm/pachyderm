# Release procedure


## Pachctl

1) Update the CLI docs in this repo by doing:

```shell
make doc
```

2) [Follow the release instructions for pachctl](https://github.com/pachyderm/homebrew-tap/blob/master/README.md)

## Pachd

- [find the tag/release on github you want to release](https://github.com/pachyderm/pachyderm/releases)
  - CI needs to pass
  - and it needs to be merged into master to generate a tag
- make sure on GH its marked as a release
- checkout the commit locally, then run:

```shell
    git fetch --tags
    make release-pachd
```

3) Use the image by tag in the manifest

To update the version, look in `etc/kube/pachyderm.json` and update any lines where you see `pachd` listed as the image. An example of a line that should be changed to reflect the latest version:

```shell
            "image": "pachyderm/pachd:v1.0.0-530",
```

To test / verify:

```shell
    git pull --tags
    cp etc/kube/pachyderm.json .
    tag=`git tag -l --points-at HEAD`
    docker_tag=`echo $tag | sed -e 's/[\(]/-/g' | sed -e 's/[\)]//g'`
    sed "s/pachyderm\/pachd/pachyderm\/pachd:$docker_tag/" pachyderm.json > tagged_pachyderm.json
    docker ps # check DM is connected
    kubectl get all # check k8s is up
    make clean-launch
    kubectl create -f tagged_pachyderm.json
    until timeout 5s pachctl list-repo 2>/dev/null >/dev/null; do sleep 5; done
```

Then do:

    $ pachctl version
    COMPONENT           VERSION             
    pachctl             1.0.494             
    pachd               1.0.498     

And independent of your pachctl version, you should see the version of pachd you just tagged/created.

If all of the above works, commit your changes. The commit message should say "Bumping manifest to use pachd image version X.Y.Z(ABC)"

4) Deploy the updated manifest

Once you've verified the image works, you'll need to update the image referenced in the manifest and update the live manifest.

To do that:

1. Clone the corpsite repository (www)
2. Update the `manifest.json` file by replacing it w a copy of the manifest from `etc/kube/pachyderm.json`
3. Commit your change on a branch
4. Run the make command to deploy
5. Repeat the verification steps above (from #3) using the manifest from the live url: `https://pachyderm.io/manifest.json`
6. Once its verified to be working, merge your branch onto master






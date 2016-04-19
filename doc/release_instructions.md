# Release procedure


## Pachctl

[Follow the release instructions for pachctl](https://github.com/pachyderm/homebrew-tap/blob/master/README.md)

## Pachd

- [find the tag/release on github you want to release](https://github.com/pachyderm/pachyderm/releases)
  - CI needs to pass
  - and it needs to be merged into master to generate a tag
- make sure on GH its marked as a release
- checkout the commit locally, then run:

    git fetch --tags
    make release-pachd

To test / verify:

    git pull --tags
    cp etc/kube/pachyderm.json .
    tag=`git tag -l --points-at HEAD`
    sed "s/pachyderm\/pachd/pachyderm\/pachd:$tag/" pachyderm.json > tagged_pachyderm.json
    kubectl create -f tagged_pachyderm.json
    until timeout 5s $GOPATH/bin/pachctl list-repo 2>/dev/null >/dev/null; do sleep 5; done

Then do:

    $ pachctl version
    COMPONENT           VERSION             
    pachctl             1.0.494             
    pachd               1.0.498     

And independent of your pachctl version, you should see the version of pachd you just tagged/created.
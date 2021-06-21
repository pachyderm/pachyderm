>![pach_logo](../img/pach_logo.svg) INFO Pachyderm 2.0 introduces profound architectual changes to the product. As a result, our examples pre and post 2.0 are kept in two separate branches:
> - Branch Master: Examples using Pachyderm 2.0 and later versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13 and older versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples

## Inititialization
```shell
make clean
make init
```

## globalID - An unique ID to trace all of the commits and jobs that resulted from an initial change
With GlobalId, all logically-dependent commits share the same ID. 
-  Let our initial change be a `put file` in our initial images repository
    ```shell
    pachctl put file images@master -i data/images.txt
    ```	
    and notice that 2 jobs (one per pipeline) have been created with the same id.
    ```shell
    pachctl list job
    ```
    ```shell
    ID                               PIPELINE STARTED        DURATION           RESTART PROGRESS  DL       UL       STATE
    bd48546ed5534b50bbf71d0b3c6dbd28 edges    24 seconds ago 5 seconds          0       2 + 1 / 3 181.1KiB 111.4KiB success
    bd48546ed5534b50bbf71d0b3c6dbd28 montage  21 seconds ago 6 seconds          0       1 + 0 / 1 371.9KiB 1.292MiB success
    ```

- list the commits in the `images` repo and notice that a commit ID of the same value exists in that repo
    ```shell
    pachcl list commit images@master
    ```
    ```
    REPO   BRANCH COMMIT                           FINISHED      SIZE DESCRIPTION
    images master bd48546ed5534b50bbf71d0b3c6dbd28 5 minutes ago 0B   
    ```
- list the commits in the `edges` and `montage` repos, the same commit ID is listed.
    ```shell
    REPO    BRANCH COMMIT                           FINISHED           SIZE DESCRIPTION
    montage master bd48546ed5534b50bbf71d0b3c6dbd28 About a minute ago 0B
    ```
    ```shell
    REPO  BRANCH COMMIT                           FINISHED           SIZE DESCRIPTION
    edges master bd48546ed5534b50bbf71d0b3c6dbd28 59 seconds ago     0B
    ```
- Inspect the commit of said ID in the `images` repo - The repo in which our change (`put file`) has originated:
    ```shell
    pachctl inspect commit images@bd48546ed5534b50bbf71d0b3c6dbd28 --raw
    ```
    Note that the original commit is marked as empty.
    ```json
    "origin": {

    },
    ```
- Inspect the following commits produced in the output repos of the edges and montages pipelines:
    ```shell
    pachctl inspect commit edges@bd48546ed5534b50bbf71d0b3c6dbd28 --raw
    ```

    Note that the origin of this commit is of kind **`AUTO`** as it has been trigerred by the arrival of a commit in the upstream repo images
    ```json
    "origin": {
        "kind": "AUTO"
    },
    ```
    as mentionned here:
    ```json
    "directProvenance": [
        {
        "repo": {
            "name": "edges",
            "type": "spec"
        },
        "name": "master"
        },
        {
        "repo": {
            "name": "images",
            "type": "user"
        },
        "name": "master"
        }
    ]
    ```
- Repeat the same command on the montage repo and note the same origin of the commmit.

- Now list the repo of all types to identify the edges and montages spec system repos.
```shell
pachctl list repo --all
```
```shell
NAME         CREATED     SIZE (MASTER) ACCESS LEVEL
montage      4 hours ago 1.664MiB      [repoOwner]  Output repo for pipeline montage.
montage.spec 4 hours ago 413B          [repoOwner]  Spec repo for pipeline montage.
montage.meta 4 hours ago 2.107MiB      [repoOwner]  Meta repo formontage
edges.meta   4 hours ago 373.9KiB      [repoOwner]  Meta repo foredges
edges.spec   4 hours ago 252B          [repoOwner]  Spec repo for pipeline edges.
edges        4 hours ago 133.6KiB      [repoOwner]  Output repo for pipeline edges.
images       4 hours ago 238.3KiB      [repoOwner]
```
- Run an inspect commit on `edges.spec` and `montage.spec`:
```shell
pachctl inspect commit edges.spec@bd48546ed5534b50bbf71d0b3c6dbd28 --raw
pachctl inspect commit montage.spec@bd48546ed5534b50bbf71d0b3c6dbd28 --raw
```
Note that in this case, the commits are **`ALIAS`**. 




The execution is tagged - GlobalID keep a snapshot of the data and the pipelines versions at the time of the processing of a change (create, update pipeline. Put files, squash commit...)






Let's change the glog pattern in our edges.json from `/*' to `/` and update our pipeline:
```json
  "input": {
    "pfs": {
      "glob": "/",
      "repo": "images"
    }
  },
```
```shell
pachctl update pipeline -f pipelines/edges.json --reprocess
```

```shell
pachctl list job
```
```
ID                               PIPELINE STARTED        DURATION   RESTART PROGRESS  DL       UL       STATE
be97b64f110643389f171eb64697d4e1 edges    48 seconds ago 5 seconds  0       1 + 0 / 1 238.3KiB 133.6KiB success
be97b64f110643389f171eb64697d4e1 montage  58 seconds ago 17 seconds 0       0 + 1 / 1 0B       0B       success
```
```shell
pachctl inspect commit edges.spec@be97b64f110643389f171eb64697d4e1 --raw
```
This time, the original commit happens in the esges.repo by updating the pipeline.

As a result, downstream commits in edges and montage are automatically created (note that their kind is `Auto` in their respective json):
```shell
pachctl inspect commit edges@be97b64f110643389f171eb64697d4e1 --raw
pachctl inspect commit edges@be97b64f110643389f171eb64697d4e1 --raw
```
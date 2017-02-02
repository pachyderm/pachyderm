
Setup
-----

- Run `./setup-demo.sh`
- You'll want to do that, and run through this script before any demo (to cache the opencv image)
- Then do a `pachctl delete-all` and re-create the images repo

Demo
----

## Exposition:

Local setup
    - minikube
    - local VM running k8s

kubectl get all
    - running all on my machine
    - from a VM
pachctl list-repo
    - see images repo
    - version control data
    - all terms from git

## Step 1 -- Add a file

```shell
pachctl put-file images master -c -i doc/examples/opencv/images.txt
```

    - show them the mount
    - locally view the data
    - can see repo images
    - now there is an image in it
    - because everything is snapshotted (commit 0)
    - statue of liberty

## Step 2 -- Add new images
   
```shell
pachctl put-file images master -c -i doc/examples/opencv/images2.txt
```

    - refresh
    - new commit -- master/1
    - overlays data based on diffs / analgous to git diffs
    - first one is there
    - two more on top of it

## Step 3 -- Run a pipeline

```shell
pachctl create-pipeline -f edges.json
pachctl list-job
```

    - the pipeline uses openCV
    - show powerpoint from dropbox:
        - show them the pipeline code (or show in vim)
        - take intput / output repo
        - multiple inputs
        - subscribe to new data coming in on data repo
        - describe w json manifest
    - dont know opencv
        - just pull in a library as a processing step
        - plug and play different approaches / steps of analysis
    - NOTE!!! create pipeline will take 5 min now if you haven't cached the opencv image
        - so do it once before, then do delete-all
    - see output of pipeline
    - output commits correspond to input structure

## Step 4 -- add more data

```shell
pachctl put-file images master -c -i doc/examples/opencv/images3.txt
```
    - kicks off the pipeline
    - see new commit in the edges repo

Common questions
---

- provenance
  - look at input commits for one of the outputs
- no reprocessing!
  - only processed the new images per commit
- if you delete minikube has to pull container again
  - you can just kill pachyderm, thats ok

Gotchas
---

VM network req
broken putfile / do delete all to recover



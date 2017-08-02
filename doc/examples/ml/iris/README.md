# ML pipeline for Iris Classification - R, Python, or Julia

![alt tag](pipeline.png)

This machine learning pipeline implements a "hello world" ML pipeline that trains a model to predict the species of Iris flowers (based on measurements of those flowers) and then utilizes that trained model to perform predictions. The pipeline can be deployed with R-based components, Python-based components, or Julia-based components.  In fact, you can even deploy an R-based pipeline, for example, and then switch out the R pipeline stages with Julia or Python pipeline stages.  This illustrates the language agnostic nature of Pachyderm's containerized pipelines.

1. [Make sure Pachyderm is running](README.md#1-make-sure-pachyderm-is-running)
2. [Create the input "data repositories"](README.md#2-create-the-input-data-repositories)
3. [Commit the training data set into Pachyderm](README.md#3-commit-the-training-data-set-into-pachyderm)
4. [Create the training pipeline](README.md#4-create-the-training-pipeline)
5. [Commit input attributes](README.md#5-commit-input-attributes)
6. [Create the inference pipeline](README.md#6-create-the-inference-pipeline)
7. [Examine the results](README.md#7-examine-the-results)

Bonus:

8. [Parallelize the inference](README.md#8-parallelize-the-inference)
9. [Update the model training](README.md#9-update-the-model-training)
10. [Update the training data set](README.md#10-update-the-training-data-set)
11. [Examine pipeline provenance](README.md#11-examine-pipeline-provenance)

Finally, we provide some [Resources](README.md#resources) for you for further exploration.

## Getting Started

- Clone this repo.
- Install/deploy Pachyderm (See the [Pachyderm docs](http://docs.pachyderm.io/en/latest/) for details. In particular, the [local installation](http://docs.pachyderm.io/en/latest/getting_started/local_installation.html) is a super easy way to experiment with Pachyderm).

## 1. Make sure Pachyderm is running

You should be able to connect to your Pachyderm cluster via the `pachctl` CLI.  To verify that everything is running correctly on your machine, you should be able to run the following with similar output:

```
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.8
pachd               1.4.8
```

## 2. Create the input data repositories

On the Pachyderm cluster running in your remote machine, we will need to create the two input data repositories (for our training data and input iris attributes).  To do this run:

```
$ pachctl create-repo training
$ pachctl create-repo attributes
```

As a sanity check, we can list out the current repos, and you should see the two repos you just created:

```
$ pachctl list-repo
NAME                CREATED             SIZE
attributes          5 seconds ago       0 B
training            8 seconds ago       0 B
```

## 3. Commit the training data set into pachyderm

We have our training data repository, but we haven't put our training data set into this repository yet.  The training data set, `iris.csv`, is included here in the [data](data) directory.

To get this data into Pachyderm, navigate to this directory and run:

```
$ cd data
$ pachctl put-file training master -c -f iris.csv
```

Then, you should be able to see the following:

```
$ pachctl list-repo
NAME                CREATED             SIZE
training            3 minutes ago       4.444 KiB
attributes          3 minutes ago       0 B
$ pachctl list-file training master
NAME                TYPE                SIZE
iris.csv            file                4.444 KiB
```

## 4. Create the training pipeline

Next, we can create the `model` pipeline stage to process the data in the training repository. To do this, we just need to provide Pachyderm with a JSON pipeline specification that tells Pachyderm how to process the data. This `model` pipeline can be generated to train a model with R, Python, or Julia and with a variety of types of models.  The following Docker images are available for the training:

- `pachyderm/iris-train:python-svm` - Python-based SVM implemented in [python/iris-train-python-svm/pytrain.py](python/iris-train-python-svm/pytrain.py)
- `pachyderm/iris-train:python-lda` - Python-based LDA implemented in [python/iris-train-python-lda/pytrain.py](python/iris-train-python-lda/pytrain.py)
- `pachyderm/iris-train:rstats-svm` - R-based SVM implemented in [rstats/iris-train-r-svm/train.R](rstats/iris-train-r-svm/train.R)
- `pachyderm/iris-train:rstats-lda` - R-based LDA implemented in [rstats/iris-train-r-lda/train.R](rstats/iris-train-r-lda/train.R)
- `pachyderm/iris-train:julia-tree` - Julia-based decision tree implemented in [julia/iris-train-julia-tree/train.jl](julia/iris-train-julia-tree/train.jl)
- `pachyderm/iris-train:julia-forest` - Julia-based random forest implemented in [julia/iris-train-julia-forest/train.jl](julia/iris-train-julia-forest/train.jl)

You can utilize any one of these images in your model training by using specification corresponding to the language of interest, `<julia, python, rstats>_train.json`, and making sure that the particular image is specified in the `image` field.  For example, if we wanted to train a random forest model with Julia, we would use [julia_train.json](julia_train.json) and make sure that the `image` field read as follows:

```
  "transform": {
    "image": "pachyderm/iris-train:julia-forest",
    ...
```

Once you have specified your choice of modeling in the pipeline spec (the below output was generated with the Julia images, but you would see similar output with the Python/R equivalents), create the training pipeline:

```
$ cd ..
$ pachctl create-pipeline -f <julia, python, rstats>_train.json
```

Immediately you will notice that Pachyderm has kicked off a job to perform the model training:

```
$ pachctl list-job
ID                                   OUTPUT COMMIT STARTED        DURATION RESTART PROGRESS STATE
a0d78926-ce2a-491a-b926-90043bce7371 model/-       12 seconds ago -        0       0 / 1    running
```

This job should run for about 1-2 minutes (it actually runs faster after this, but we have to pull the Docker image on the first run).  After your model has successfully been trained, you should see:

```
$ pachctl list-job
ID                                   OUTPUT COMMIT                          STARTED       DURATION       RESTART PROGRESS STATE
a0d78926-ce2a-491a-b926-90043bce7371 model/98e55f3bccc6444a888b1adbed4bba8b 2 minutes ago About a minute 0       1 / 1    success
$ pachctl list-repo
NAME                CREATED             SIZE
model               2 minutes ago       43.67 KiB
training            8 minutes ago       4.444 KiB
attributes          7 minutes ago       0 B
$ pachctl list-file model master
NAME                TYPE                SIZE
model.jld           file                43.67 KiB
```

## 5. Commit input attributes

Great! We now have a trained model that will infer the species of iris flowers.  Let's commit some iris attributes into Pachyderm that we would like to run through the inference.  We have a couple examples under [test](data/test).  Feel free to use these, find your own, or even create your own.  To commit our samples (assuming you have cloned this repo on the remote machine), you can run:

```
$ cd /home/pachrat/julia-workshop/data/test/
$ pachctl put-file attributes master -c -r -f .
```

You should then see:

```
$ pachctl list-file attributes master
NAME                TYPE                SIZE
1.csv               file                16 B
2.csv               file                96 B
```

## 6. Create the inference pipeline

We have another JSON blob, `<julia, python, rstats>_infer.json`, that will tell Pachyderm how to perform the processing for the inference stage.  This is similar to our last JSON specification except, in this case, we have two input repositories (the `attributes` and the `model`) and we are using a different Docker image.  Similar to the training pipeline stage, this can be created in R, Python, or Julia.  However, you should create it in the language that was used for training (because the model output formats aren't standardized across the languages). The available docker images are as follows:

- `pachyderm/iris-infer:python` - Python-based inference implemented in [python/iris-infer-python/infer.py](python/iris-infer-python/pyinfer.py)
- `pachyderm/iris-infer:rstats` - R-based inferenced implemented in [rstats/iris-infer-rstats/infer.R](rstats/iris-infer-r/infer.R)
- `pachyderm/iris-infer:julia` - Julia-based inference implemented in [julia/iris-infer-julia/infer.jl](julia/iris-infer-julia/infer.jl)

Then, to create the inference stage, we simply run:

```
$ cd ../../
$ pachctl create-pipeline -f <julia, python, rstats>_infer.json
```

where `<julia, python, rstats>` is replaced by the language you are using.  This will immediately kick off an inference job, because we have committed unprocessed reviews into the `reviews` repo.  The results will then be versioned in a corresponding `inference` data repository:

```
$ pachctl list-job
ID                                   OUTPUT COMMIT                          STARTED        DURATION       RESTART PROGRESS STATE
21552ae0-b0a9-4089-bfa5-d74a4a9befd7 inference/-                            33 seconds ago -              0       0 / 2    running
a0d78926-ce2a-491a-b926-90043bce7371 model/98e55f3bccc6444a888b1adbed4bba8b 7 minutes ago  About a minute 0       1 / 1    success
$ pachctl list-job
ID                                   OUTPUT COMMIT                              STARTED            DURATION       RESTART PROGRESS STATE
21552ae0-b0a9-4089-bfa5-d74a4a9befd7 inference/c4f6b269ad0349469effee39cc9ee8fb About a minute ago About a minute 0       2 / 2    success
a0d78926-ce2a-491a-b926-90043bce7371 model/98e55f3bccc6444a888b1adbed4bba8b     8 minutes ago      About a minute 0       1 / 1    success
$ pachctl list-repo
NAME                CREATED              SIZE
inference           About a minute ago   100 B
attributes          13 minutes ago       112 B
model               8 minutes ago        43.67 KiB
training            13 minutes ago       4.444 KiB
```

## 7. Examine the results

We have created results from the inference, but how do we examine those results?  There are multiple ways, but an easy way is to just "get" the specific files out of Pachyderm's data versioning:

```
$ pachctl list-file inference master
NAME                TYPE                SIZE
1.csv               file                15 B
2.csv               file                85 B
$ pachctl get-file inference master 1.csv
Iris-virginica
$ pachctl get-file inference master 2.csv
Iris-versicolor
Iris-virginica
Iris-virginica
Iris-virginica
Iris-setosa
Iris-setosa
```

Here we can see that each result file contains a predicted iris flower species corresponding to each set of input attributes.

## Bonus exercises

### 8. Parallelize the inference

You may have noticed that our pipeline specs included a `parallelism_spec` field.  This tells Pachyderm how to parallelize a particular pipeline stage.  Let's say that in production we start receiving a huge number of attribute files, and we need to keep up with our inference.  In particular, let's say we want to spin up 10 inference workers to perform inference in parallel.

This actually doesn't require any change to our code.  We can simply change our `parallelism_spec` in `<julia, python, rstats>_infer.json` to:

```
  "parallelism_spec": {
    "constant": "10"
  },
```

Pachyderm will then spin up 10 inference workers, each running our same script, to perform inference in parallel.  This can be confirmed by updating our pipeline and then examining the cluster:

```
$ vim infer.json
$ pachctl update-pipeline -f julia_infer.json
$ kubectl get all
NAME                             READY     STATUS        RESTARTS   AGE
po/etcd-4197107720-906b7         1/1       Running       0          52m
po/pachd-3548222380-cm1ts        1/1       Running       0          52m
po/pipeline-inference-v1-vsq8x   2/2       Terminating   0          6m
po/pipeline-inference-v2-0w438   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-1tdm7   0/2       Pending       0          5s
po/pipeline-inference-v2-2tqtl   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-6x917   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-cc5jz   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-cphcd   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-d5rc0   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-lhpcv   0/2       Init:0/1      0          5s
po/pipeline-inference-v2-mpzwf   0/2       Pending       0          5s
po/pipeline-inference-v2-p753f   0/2       Init:0/1      0          5s
po/pipeline-model-v1-1gqv2       2/2       Running       0          13m

NAME                       DESIRED   CURRENT   READY     AGE
rc/pipeline-inference-v2   10        10        0         5s
rc/pipeline-model-v1       1         1         1         13m

NAME                        CLUSTER-IP       EXTERNAL-IP   PORT(S)                       AGE
svc/etcd                    10.97.253.64     <nodes>       2379:32379/TCP                52m
svc/kubernetes              10.96.0.1        <none>        443/TCP                       53m
svc/pachd                   10.108.55.75     <nodes>       650:30650/TCP,651:30651/TCP   52m
svc/pipeline-inference-v2   10.99.47.41      <none>        80/TCP                        5s
svc/pipeline-model-v1       10.109.198.229   <none>        80/TCP                        13m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/etcd    1         1         1            1           52m
deploy/pachd   1         1         1            1           52m

NAME                  DESIRED   CURRENT   READY     AGE
rs/etcd-4197107720    1         1         1         52m
rs/pachd-3548222380   1         1         1         52m
$ kubectl get all
NAME                             READY     STATUS    RESTARTS   AGE
po/etcd-4197107720-906b7         1/1       Running   0          53m
po/pachd-3548222380-cm1ts        1/1       Running   0          53m
po/pipeline-inference-v2-0w438   2/2       Running   0          40s
po/pipeline-inference-v2-1tdm7   2/2       Running   0          40s
po/pipeline-inference-v2-2tqtl   2/2       Running   0          40s
po/pipeline-inference-v2-6x917   2/2       Running   0          40s
po/pipeline-inference-v2-cc5jz   2/2       Running   0          40s
po/pipeline-inference-v2-cphcd   2/2       Running   0          40s
po/pipeline-inference-v2-d5rc0   2/2       Running   0          40s
po/pipeline-inference-v2-lhpcv   2/2       Running   0          40s
po/pipeline-inference-v2-mpzwf   2/2       Running   0          40s
po/pipeline-inference-v2-p753f   2/2       Running   0          40s
po/pipeline-model-v1-1gqv2       2/2       Running   0          14m

NAME                       DESIRED   CURRENT   READY     AGE
rc/pipeline-inference-v2   10        10        10        40s
rc/pipeline-model-v1       1         1         1         14m

NAME                        CLUSTER-IP       EXTERNAL-IP   PORT(S)                       AGE
svc/etcd                    10.97.253.64     <nodes>       2379:32379/TCP                53m
svc/kubernetes              10.96.0.1        <none>        443/TCP                       54m
svc/pachd                   10.108.55.75     <nodes>       650:30650/TCP,651:30651/TCP   53m
svc/pipeline-inference-v2   10.99.47.41      <none>        80/TCP                        40s
svc/pipeline-model-v1       10.109.198.229   <none>        80/TCP                        14m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/etcd    1         1         1            1           53m
deploy/pachd   1         1         1            1           53m

NAME                  DESIRED   CURRENT   READY     AGE
rs/etcd-4197107720    1         1         1         53m
rs/pachd-3548222380   1         1         1         53m
```

### 9. Update the model training

Let's now imagine that we want to update our model from random forest to decision tree, SVM to LDA, etc. To do this, modify the image tag in `train.json`.  For example, to run a Julia-based decision tree instead of a random forest:

```
"image": "pachyderm/iris-train:julia-tree",
```

Once you modify the spec, you can update the pipeline by running:

```
$ pachctl update-pipeline -f <julia, python, rstats>_train.json
```

Pachyderm will then automatically kick off a new job to retrain our model with the new model:

```
$ pachctl list-job
ID                                   OUTPUT COMMIT                              STARTED        DURATION       RESTART PROGRESS STATE
7d913835-2c0a-42a3-bfa2-c8a5941ceaa5 model/-                                    3 seconds ago  -              0       0 / 1    running
21552ae0-b0a9-4089-bfa5-d74a4a9befd7 inference/c4f6b269ad0349469effee39cc9ee8fb 11 minutes ago About a minute 0       2 / 2    success
a0d78926-ce2a-491a-b926-90043bce7371 model/98e55f3bccc6444a888b1adbed4bba8b     19 minutes ago About a minute 0       1 / 1    success
```

Not only that, once the model is retrained, Pachyderm see the new model and updates our inferences with the latest version of the model:

```
$ pachctl list-job
ID                                   OUTPUT COMMIT                              STARTED            DURATION       RESTART PROGRESS STATE
0477e755-79b4-4b14-ac04-5416d9a80cf3 inference/5dec44a330d24a1cb3822610c886489b 53 seconds ago     44 seconds     0       2 / 2    success
7d913835-2c0a-42a3-bfa2-c8a5941ceaa5 model/444b5950bcb642cfba5b087286640898     About a minute ago 56 seconds     0       1 / 1    success
21552ae0-b0a9-4089-bfa5-d74a4a9befd7 inference/c4f6b269ad0349469effee39cc9ee8fb 13 minutes ago     About a minute 0       2 / 2    success
a0d78926-ce2a-491a-b926-90043bce7371 model/98e55f3bccc6444a888b1adbed4bba8b     20 minutes ago     About a minute 0       1 / 1    success
```

### 10. Update the training data set

Let's say that one or more observations in our training data set were corrupt or unwanted.  Thus, we want to update our training data set.  To simulate this, go ahead and open up `iris.csv` (e.g., with `vim`) and remove a couple of the rows (non-header rows).  Then, let's replace our training set:

```
$ pachctl start-commit training master
9cc070dadc344150ac4ceef2f0758509
$ pachctl delete-file training 9cc070dadc344150ac4ceef2f0758509 iris.csv
$ pachctl put-file training 9cc070dadc344150ac4ceef2f0758509 -f iris.csv
$ pachctl finish-commit training 9cc070dadc344150ac4ceef2f0758509
```

Immediately, Pachyderm "knows" that the data has been updated, and it starts a new jobs to update the model and inferences:

### 11. Examine pipeline provenance

Let's say that we have updated our model or training set in one of the above scenarios (step 11 or 12).  Now we have multiple inferences that were made with different models and/or training data sets.  How can we know which results came from which specific models and/or training data sets?  This is called "provenance," and Pachyderm gives it to you out of the box.

Suppose we have run the following jobs:

```
$ pachctl list-job
ID                                   OUTPUT COMMIT                              STARTED        DURATION       RESTART PROGRESS STATE
0477e755-79b4-4b14-ac04-5416d9a80cf3 inference/5dec44a330d24a1cb3822610c886489b 6 minutes ago  44 seconds     0       2 / 2    success
7d913835-2c0a-42a3-bfa2-c8a5941ceaa5 model/444b5950bcb642cfba5b087286640898     7 minutes ago  56 seconds     0       1 / 1    success
2560f096-0515-4d68-be66-35e3b4f3e730 inference/db9e675de0274ce9a73d3fc9dd50fd51 13 minutes ago About a minute 1       2 / 2    success
21552ae0-b0a9-4089-bfa5-d74a4a9befd7 inference/c4f6b269ad0349469effee39cc9ee8fb 19 minutes ago About a minute 0       2 / 2    success
a0d78926-ce2a-491a-b926-90043bce7371 model/98e55f3bccc6444a888b1adbed4bba8b     26 minutes ago About a minute 0       1 / 1    success
```

If we want to know which model and training data set was used for the latest inference, commit id `5dec44a330d24a1cb3822610c886489b`, we just need to inspect the particular commit:

```
$ pachctl inspect-commit inference 5dec44a330d24a1cb3822610c886489b
$ pachctl inspect-commit inference 5dec44a330d24a1cb3822610c886489b
Commit: inference/5dec44a330d24a1cb3822610c886489b
Parent: db9e675de0274ce9a73d3fc9dd50fd51
Started: 6 minutes ago
Finished: 6 minutes ago
Size: 100 B
Provenance:  training/9881a078c14c47e9b71fcc626a86499f  attributes/f62907bda09d48cfa817476fd3e4f07f  model/444b5950bcb642cfba5b087286640898
```

The `Provenance` tells us exactly which model and training set was used (along with which commit to attributes triggered the inference).  For example, if we wanted to see the exact model used, we would just need to reference commit `444b5950bcb642cfba5b087286640898` to the `model` repo:

```
$ pachctl list-file model 444b5950bcb642cfba5b087286640898
NAME                TYPE                SIZE
model.jld           file                70.34 KiB
```

We could get this model to examine it, rerun it, revert to a different model, etc.

## Resources

- Join the [Pachyderm Slack team](http://slack.pachyderm.io/) to ask questions, get help, and talk about production deploys.
- Follow [Pachyderm on Twitter](https://twitter.com/pachydermIO),
- Find [Pachyderm on GitHub](https://github.com/pachyderm/pachyderm), and
- [Spin up Pachyderm](http://docs.pachyderm.io/en/latest/getting_started/getting_started.html) in just a few commands to try this and [other examples](http://docs.pachyderm.io/en/latest/examples/readme.html) locally.

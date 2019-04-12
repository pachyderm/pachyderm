## Machine Learning on the edge with Pachyderm
In this example we're going to build an object detection model that will be deployed to an android device. This example is based on [https://github.com/googlecodelabs/tensorflow-for-poets-2](https://github.com/googlecodelabs/tensorflow-for-poets-2)

### Prerequisites
- A running instance of Pachyderm: If this is your first example, checkout [Pachyderm Getting Started](https://pachyderm.readthedocs.io/en/latest/getting_started/getting_started.html)
- [Android Studio](https://developer.android.com/studio) or [Xcode](https://itunes.apple.com/us/app/xcode) installed
- A physical Android or IOS device with developer mode enabled. 

### Step 1 - Get the data
First thing we need to do is fetch the data

`curl <link> --create-dirs -o ./data/flowers.zip && unzip ./data/flower.zip`

Next, create our Pachyderm repo and "check-in" the data  

`pachctl create-repo flower_photos`
`cd data/flowers/ && pachctl put-file flower_photos master -r -f ./` 

Once that's done, lets go back to the root of the example.
`cd ../../`


### Step 2 - Create our first pipeline to train the model
Now that we have the `flower_photos` repo create and populated with the data we need to train the model, we can move on to training Mobile_Net to recognize some new objects.  

To create our pipeline simply run the following:
`pachctl create-pipeline -f pipeline-train.json`

While that's running (it can take a few minutes depending on your compute resources), lets take a closer look at the pipeline we're creating. 

```
{
  "pipeline": {
    "name": "Train"
  },
  "transform": {
    "image": "pachyderm/pachyderm_tflite_ex:v1",
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "python -m scripts.retrain --bottleneck_dir=/pfs/out/bottlenecks --how_many_training_steps=500 --model_dir=/pfs/out/models/ --summaries_dir=/pfs/out/training_summaries/mobilenet_0.50_$IMAGE_SIZE --output_graph=/pfs/out/retrained_graph.pb --output_labels=/pfs/out/retrained_labels.txt --architecture=mobilenet_0.50_$IMAGE_SIZE --image_dir=/pfs/flower_photos/"
    ],
    "env": {
      "IMAGE_SIZE": "224"
    }
  },
  "parallelism_spec": {
       "constant" : 1
  },
  "input": {
    "pfs": {
      "repo": "flower_photos",
      "glob": "/"
    }
  }
}
{
  "pipeline": {
    "name": "Tensorboard"
  },
  "transform": {
    "image": "pachyderm/pachyderm_tflite_ex:v1",
    "cmd": [ "/bin/bash" ],
    "stdin": [ "tensorboard --logdir /pfs/Mobile_Object_Train/training_summaries" ]
  },
  "parallelism_spec": {
       "constant" : 1
  },
  "service": {
    "internal_port": 6006,
    "external_port": 30006
  },
  "input": {
    "pfs": {
      "repo": "Train",
      "glob": "/"
    }
  }
}
```

This particular pipeline is doing a few things that are interesting. The first thing you should notice is that this is actually two pipeline steps in one. The first one is "Train" and it's running our pipeline code to retrain MobileNet and using our repo `flower_photos` as input. The second pipeline step is running Tensorboard so we can see how well the model trained. We're leveraging the "service" capability in Pachyderm to make sure that once the pod is deployed in k8s, we can access it via port 30006. However, before we can view Tensorboard directly, we have to do a couple of things. 

To get Tensorboard up, we have to edit the service type from `NodePort` to `LoadBalancer`. Thankfully, this is actually really simple. Just run:

`kubectl edit service pipeline-tensorboard-v1-user`

Then go towards the bottom of the file and replace `type: NodePort` to `type: LoadBalancer`. It should look something like this:

```
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2019-04-12T00:47:00Z"
  labels:
    app: pipeline-tensorboard-v1
    component: worker
    pipelineName: Tensorboard
    suite: pachyderm
    version: 1.8.6
  name: pipeline-tensorboard-v1-user
  namespace: default
  resourceVersion: "135796"
  selfLink: /api/v1/namespaces/default/services/pipeline-tensorboard-v1-user
  uid: 7b19cfbf-5cbc-11e9-9ac2-080027949e53
spec:
  clusterIP: 10.96.79.174
  externalTrafficPolicy: Cluster
  ports:
  - name: user-port
    nodePort: 30006
    port: 30006
    protocol: TCP
    targetPort: 6006
  selector:
    app: pipeline-tensorboard-v1
    component: worker
    pipelineName: Tensorboard
    suite: pachyderm
    version: 1.8.6
  sessionAffinity: None
  type: LoadBalancer
status:
  loadBalancer: {}
```

Lastly, and if you're running minikube, open a separate terminal window and run:

`minikube tunnel`

You should now be able to access Tensorboard at `<yourvmip>:30006`

The last step in this section is to check and make sure that our model ended up in our next repo:

```
$ pachctl list-repo
NAME              CREATED        SIZE (MASTER)
Tensorboard       8 minutes ago  0B
Train             8 minutes ago  45.29MiB
flower_photos     27 minutes ago 85.14MiB
``` 

Looks like we're good to go.

### Step 3: Convert our Model for TFlite

Next stop on our Object Dection on the edge pipeline is to convert our newly retrained model so that it can be used by our mobile app. We'll use Pachyderm to do it: 

Simply run:
`pachctl create-pipeline -f pipeline-convert.json`

This should run pretty quickly and if we check our output repo, we should our re-trained model:

```
$ pachctl list-file Convert master
COMMIT                           NAME                  TYPE COMMITTED      SIZE
5f43ec065bc7413a99df1d095341dab3 /optimized_graph.lite file 54 seconds ago 5.081MiB
```

### Step 4: Run our App
Now that we've got our model training and model converting steps setup and automated all we have to do now is copy these files to our app directory and run the app. 

First thing we need to do is "get" the files out of Pachyderm:

`pachctl get-file Convert master optimized_graph.lite > ./output/fresh/optimized_graph.lite`
`pachctl get-file Train master retrained_labels.txt > ./output/fresh/retrained_labels.txt`

our next steps depend on your app of choice...

For Android:
`cp output/fresh/optimized_graph.lite android/tflite/app/src/main/assets/graph.lite`
`cp output/fresh/retrained_labels.txt android/tflite/app/src/main/assets/labels.txt`

For IOS:
`cp output/fresh/optimized_graph.lite ios/tflite/data/graph.lite`
`cp output/fresh/retrained_labels.txt ios/tflite/data/labels.txt`

Now, we just build our app. 

<screenshot for android gradle sync>


## Now lets witness the importance of Provenance
In just a few short minutes you were able to retrain a object detection model and deploy it to a device on the edge (sort of). But what we want to do is build ML pipelines that are production grade. The second half of this example is to show you how important provenance is. 

What we need to do first is simulate "bad data". To do that run the following:

`curl <link> --create-dirs -o ./data/orchids.zip && unzip ./data/orchids.zip`

Then lets check-in our "bad data" 

`cd data/ && pachctl put-file flower_photos master -r -f orchids`

If you haven't done so already, take a look at our "orchids" directory. It's not actually orchids, it's tulips.

And because we've added new data to our flower_photos repo, our pipeline is executing all the steps above automatically. Only thing we have to do is get the new model and labels out of pachyderm and into our app directory. 

`pachctl get-file Convert master optimized_graph.lite > ./output/retrain/optimized_graph.lite`
`pachctl get-file Train master retrained_labels.txt > ./output/retrain/retrained_labels.txt`

cp those files to the app directory:

For Android:
`cp output/retrain/optimized_graph.lite android/tflite/app/src/main/assets/graph.lite`
`cp output/retrain/retrained_labels.txt android/tflite/app/src/main/assets/labels.txt`

For IOS:
`cp output/retrain/optimized_graph.lite ios/tflite/data/graph.lite`
`cp output/retrain/retrained_labels.txt ios/tflite/data/labels.txt`

then build our app and point it at a tulip. What? it's classified it as an orchid.... How did that happen? Thankfully, Pachyderm makes it incredibly simple to find out exactly what data was used to train a model. To view the provenance of that model lets run 

```
$ pachctl list-file Convert master
COMMIT                           NAME                  TYPE COMMITTED     SIZE
03af2c900b9a41f2a10cb92a4a7f50e8 /optimized_graph.lite file 5 minutes ago 5.085MiB
```

Lets inspect the actual commit

```
$ pachctl inspect-commit Convert 03af2c900b9a41f2a10cb92a4a7f50e8
Commit: Convert/03af2c900b9a41f2a10cb92a4a7f50e8
Parent: 5f43ec065bc7413a99df1d095341dab3
Started: 7 minutes ago
Finished: 5 minutes ago
Size: 5.085MiB
Provenance:  Train/4b3b03c2ba57476690f4cb3a9b4e799a  __spec__/241d833468a24df99edcfe3f04f57be1  __spec__/656d972b52904364a2788165adfd0261  flower_photos/64a918e38fa145e6bae7c1f4a6bf48f5
```

That last field is pretty important as it provides the trail of breadcrumbs we need to figure out exactly where things went wrong. In this case it looks like this model was trained from the commit `64a918e38fa145e6bae7c1f4a6bf48f5` in `flower_photos` lets see what that actually was. 

```
$ pachctl list-file flower_photos master --history 1
COMMIT                           NAME        TYPE COMMITTED         SIZE
cb58ab567842418f86fe8d0158ddf8dd /.DS_Store  file About an hour ago 8.004KiB
cb58ab567842418f86fe8d0158ddf8dd /daisy      dir  About an hour ago 32.73MiB
64a918e38fa145e6bae7c1f4a6bf48f5 /orchids    dir  10 minutes ago    52.56MiB
cb58ab567842418f86fe8d0158ddf8dd /sunflowers dir  About an hour ago 52.4MiB
``` 

We can see that commit `64a918e38fa145e6bae7c1f4a6bf48f5` was the commit id of our orchids. If we then looked at that data closer, we would see that they're actually tulips and if we remove that commit, change the name of the directory to tulips, add it back in, we'll be all set. 




------ Extra ----------

Looks like we're good to go. 

Steps: 

1. Create repo flower_photos: `pachctl create-repo flower_photos`
2. Create training pipeline: `pachctl create-pipeline -f pipeline-train.json`
	- edit service to `Loadbalancer`
3. Create convert pipeline: `pachctl create-pipeline -f pipeline-convert.json`
4. export files 
`pachctl get-file Mobile_Object_TFlite master optimized_graph.lite > ./output/optimized_graph.lite`
`pachctl get-file Mobile_Object_Train master retrained_labels.txt > ./output/retrained_labels.txt`

`cp output/fresh/optimized_graph.lite android/tflite/app/src/main/assets/graph.lite`
`cp output/fresh/retrained_labels.txt android/tflite/app/src/main/assets/labels.txt`

5. Android deploy


android/tflite/app/src/main/assets

pachctl put-file flower_photos new orchids -f ./orchids/ -r


Fresh
$ pachctl inspect-commit Mobile_Object_TFlite 103b698e23a3427eb25fbc1634fb57e3
Commit: Mobile_Object_TFlite/103b698e23a3427eb25fbc1634fb57e3
Started: 32 minutes ago
Finished: 32 minutes ago
Size: 5.088MiB
Provenance:  Mobile_Object_Train/f1998793b5f64d6db8793602d007dbd8  __spec__/aaf3d6709df9433f9ba37dea4ba5880f  __spec__/35d7b8c318434472b55c593031893aad  flower_photos/c0a204f03541423e86cf2f3b0e655a67



Train_round 1: 
Repo: flower_photos
- daisy
- dandelion
- roses
- sunflowers

Glob pattern: `/`

Train_round 2: 
Repo: flower_photos
- daisy
- dandelion
- roses
- sunflowers
* tulips





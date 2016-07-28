# Quick Start Guide: OpenCV Dominant Color Example
In this guide you're going to create a Pachyderm pipeline to run a dominant color algorithm using the popular computer vision library OpenCV. The algorithm is written in C++, compiled and statically linked in order for easy distribution throughout the containers that run the Pachyderm pipeline. If you are interested in the implementation of the [k-means clustering](https://en.wikipedia.org/wiki/K-means_clustering) based algorithm, please refer to [the OpenCV source code](https://github.com/SoheilSalehian/CVSearchEngine/tree/master/src/cpp). We are also going to run the algorithm on 1000 images from the Flikr dataset which are stored in an S3 bucket.
 
## Setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster.  [Detailed setup instructions can be found here](http://pachyderm.readthedocs.io/en/latest/deploying_setup.html).

As mentioned in the introduction, having a statically linked native opencv binary simplifies the dev environment and reduces the size of docker images required. Note: an approach that was taken for the purposes of this guide was to build the binary (called `dominantColor`) in a docker images such as [this one](https://github.com/SoheilSalehian/Docker-OpenCV) and then copy over the resulting binary into the containers that run the pipeline.

In the case that you are running your own version of build with a Dockerfile, a [Makefile]([examples/opencv_dominant_color/Makefile) is provided in order to download the static binary file:

```shell
$ make
```

This guide also assumes access to an S3 bucket for the dataset images which will be downloaded at the beginning of the pipeline. You should download `images.zip` via our [google drive link](https://drive.google.com/file/d/0B351HZYtYt77XzlscG0wQkxIZmM/view?usp=sharing), unzip the images, and upload to an S3 bucket for the purposes of this example. In our case, the bucket is named `opencv-example`. 

## Create a Repo

A `Repo` is the highest level primitive in `pfs`. Like all primitives in pfs, they share
their name with a primitive in Git and are designed to behave analogously.
Generally, a `repo` should be dedicated to a single source of data such as log
messages from a particular service. Repos are dirt cheap so don't be shy about
making them very specific.

For this demo we'll simply create a `repo` called
“images” to hold a list of the name of images in the S3 bucket.

```shell
$ pachctl create-repo images
$ pachctl list-repo
images
```


## Start a Commit
Now that we’ve created a `repo` we’ve got a place to add data.
If you try writing to the `repo` right away though, it will fail because you can't write directly to a
`Repo`. In Pachyderm, you write data to an explicit `commit`. Commits are
immutable snapshots of your data which give Pachyderm its version control for
data properties. Unlike Git though, commits in Pachyderm must be explicitly
started and finished.

Let's start a new commit in the “images” repo:
```shell
$ pachctl start-commit images
e1b57763af0b41e5ad4ada262a392de1
```

This returns a brand new commit id. Yours should be different from mine.
Now if we take a look inside our repo, we’ve created a directory for the new commit:
```shell
$ pachctl list-commit images
e1b57763af0b41e5ad4ada262a392de1
```

A new directory has been created for our commit and now we can start adding
data. Data for this example is just a single file with a list of image names. The sample file `images.txt` includes the name of the 1000 images that the pipeline will be processing. Remember to unzip `images.zip` and upload to S3 prior to running the pipeline.
We're going to write that data as a file called “images” in pfs.

```shell
# Write sample data into pfs
$ cat examples/opencv_dominant_color/images.txt | pachctl put-file images e1b57763af0b41e5ad4ada262a392de1 images
```

## Finish a Commit

Pachyderm won't let you process data from a commit that isn't `finished`.
This prevents reads from racing with writes. Furthermore, every write
to pfs is atomic. So now let's finish the commit:

```shell
$ pachctl finish-commit images e1b57763af0b41e5ad4ada262a392de1
```

Now we can view the head of the file:

```shell
$ pachctl get-file images e1b57763af0b41e5ad4ada262a392de1 images | head
im1.jpg
im10.jpg
im100.jpg
im1000.jpg
im101.jpg
im102.jpg
im103.jpg
im104.jpg
im105.jpg
```
However, we've lost the ability to write to this `commit` since finished
commits are immutable. In Pachyderm, a `commit` is always either _write-only_
when it's been started and files are being added, or _read-only_ after it's
finished. But now we can to process the images!

## Create a Pipeline

Now that we've got some data in our `repo` it's time to do something with it.
Pipelines are the core primitive for Pachyderm's processing system (pps) and
they're specified with a JSON encoding. We're going to create a pipeline that simply scrapes each of the web pages in “urls.”

```
+---------------------------+     +------------------------------+
| image file names (images) | --> |opencv_dominant_color pipeline|
+---------------------------+     +------------------------------+
```

The `pipeline` we're creating can be found at [examples/opencv_dominant_color/opencv-pipeline.json](opencv-pipeline.json).
```
{
  "pipeline": {
    "name": "opencv"
  },
  "transform": {
    "image": "soheilsalehian/opencv_example:20170722-00",
    "cmd": ["sh"],
    "stdin": [
      "export AWS_ACCESS_KEY_ID=[your aws access key id]",
      "export AWS_SECRET_ACCESS_KEY=[your aws secret access key]",
      "./dominant_color -images_file /pfs/images/images -aws_bucket_name [your aws s3 bucket of images] >> /pfs/out/opencv"
    ]
  },
  "parallelism": "1",
  "inputs": [
    {
      "repo": {
        "name": "images"
      },
      "method": "map"
    }
  ]
}
```
As the pipeline uses S3 for image input storage, please add the aws access id and aws secret access key environment variables along with the name of the s3 bucket (in this case it is 'opencv-example').

In this pipeline, we’re using the already built docker image that includes the dominant color algorithm. For a closer look at this image please take a look at [examples/opencv_dominant_color/Dockerfile](Dockerfile) which can be found in this example directory. S3 credentials are passed in explicitly to keep the scope of the example (Pachyderm also supports passing S3 creds through AWS secret passing via kubernetes manifests).

Finally, the interesting part of the pipeline: the `dominant_color` go binary (based on [examples/opencv_dominant_color/dominant_color.go](dominant_color.go) which is responsible for downloading the images from s3 and calling the c++ algorithm binary) is run with flags for the images file (which we have added as a commit to the images repo), the name of the AWS bucket that holds the images.

Another important section to notice is that we read data
from `/pfs/images/images` (/pfs/[input_repo_name]) and write data to `/pfs/out/opencv`.

Now let's create the pipeline in Pachyderm:

```shell
$ pachctl create-pipeline -f examples/opencv_dominant_color/opencv-pipeline.json
```

## What Happens When You Create a Pipeline
Creating a `pipeline` tells Pachyderm to run your code on *every* finished
`commit` in a `repo` as well as *all future commits* that happen after the pipeline is
created. Our `repo` already had a `commit` with the file “images” in it so Pachyderm will automatically
launch a `job` to scrape those webpages.

You can view the job with:

```shell
$ pachctl list-job
ID                                 OUTPUT                                    STATE
09a7eb68995c43979cba2b0d29432073   opencv/2b43def9b52b4fdfadd95a70215e90c9   JOB_STATE_RUNNING
```

Depending on how quickly the job is ran, you may see `running` or
`successful`. For 1000 images and no parallelism, this may take up to 15 minutes.

Pachyderm `job`s are implemented as Kubernetes jobs, so you can also see your job with:

```shell
$ kubectl get job
JOB                                CONTAINER(S)   IMAGE(S)             SELECTOR                                                         SUCCESSFUL
09a7eb68995c43979cba2b0d29432073   user           pachyderm/job-shim   app in (09a7eb68995c43979cba2b0d29432073),suite in (pachyderm)   1
```

A good way to check on the pipeline

Every `pipeline` creates a corresponding `repo` with the same
name where it stores its output results. In our example, the pipeline was named “opencv” so it created a `repo` called “opencv” which contains the final output.


## Reading the Output
There are a couple of different ways to retrieve the output. We can read a single output file from the “opencv” `repo` in the same fashion that we read the input data:

```shell
$ pachctl list-file opencv 09a7eb68995c43979cba2b0d29432073 opencv
$ pachctl get-file opencv 09a7eb68995c43979cba2b0d29432073 opencv
```

Using `get-file` is good if you know exactly what file you’re looking for, but for this example we want to just see the top 5 most dominant color for each image. One great way to do this is to mount the distributed file system locally and then just poke around.

## Mount the Filesystem
First create the mount point:

```shell
$ mkdir ~/pfs
```

And then mount it:

```shell
# We background this process because it blocks.
$ pachctl mount ~/pfs &
```

This will mount pfs on `~/pfs` you can inspect the filesystem like you would any
other local filesystem. Try:

```shell
$ ls ~/pfs
images opencv
```
You should see the images repo that we created or the `opencv` repo that pachyderm created as part of the pipeline.

Now you can simply `ls` and `cd` around the file system. You can read the top 5 dominant colors in HSV space (from right to left in this case).
Looking at the results:
```shell
$ head ~/pfs/opencv/09a7eb68995c43979cba2b0d29432073/opencv
im1.jpg |  20.948992   32.863491   216.865738  | 35.246891   60.671085   54.082661  | 97.668373   206.915863   55.697308  | 120.866508   25.080130   146.936523  | 19.569460   56.620155   137.499222  |

im10.jpg |  19.392403   17.160366   165.163666  | 149.043121   34.640427   91.192322  | 92.382164   203.258102   13.145894  | 130.047241   11.899960   194.779739  | 13.647389   55.731003   78.900078  |

im100.jpg |  32.702194   186.271851   159.835724  | 24.227354   79.743240   167.419662  | 38.063137   200.102142   215.346375  | 132.114319   68.211082   128.087021  | 38.833736   151.232437   107.771217  |

im1000.jpg |  7.989898   25.358656   153.719635  | 139.577148   11.809133   101.354256  | 13.005986   58.725502   223.807602  | 11.071401   19.911144   39.587875  | 109.780624   40.455315   50.042118  |

im101.jpg |  160.830017   46.397697   69.901993  | 28.465757   205.267059   25.863764  | 0.361928   5.317534   5.269302  | 13.312026   121.914291   85.093811  | 27.073235   83.940109   202.059784  |
```

As you can see we have ran the algorithm which gives us the top 5 most dominant colors in the HSV color space.



## Processing More Data

Pipelines can be triggered manually, but also will automatically process the data from new commits as they are
created. Think of pipelines as being subscribed to any new commits that are
finished on their input repo(s).

If we want to run the algorithm on more images, we can add new commits and Pachyderm takes care of running the algorithm on new images.


## Next Steps
You've now got a working Pachyderm cluster with data and a pipelines!

In the next part of this guide, we will take a look at how to search for images that have a distinct dominant color such as pink. Some of the ideas on how to do this include:
- Parsing the `opencv` repo output and evaluating the results.
- Dumping the results of the algorithm into Elastic Search and then make queries to it.

We'd love to help and see what you come up with so submit any issues/questions you come across or email at info@pachyderm.io if you want to show off anything nifty you've created with OpenCV and Pachyderm!

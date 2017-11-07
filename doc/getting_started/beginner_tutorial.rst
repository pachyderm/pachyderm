Beginner Tutorial
=================
Welcome to the beginner tutorial for Pachyderm! If you've already got Pachyderm installed, this guide should take about 15 minutes, and it will introduce you to the basic concepts of Pachyderm.

Image processing with OpenCV
----------------------------

In this guide we're going to create a Pachyderm pipeline to do some simple `edge detection <https://en.wikipedia.org/wiki/Edge_detection>`_ on a few images. Thanks to Pachyderm's processing system, we'll be able to run the pipeline in a distributed, streaming fashion. As new data is added, the pipeline will automatically process it and output the results.

If you hit any errors not covered in this guide, get help in our `public commity Slack <http://slack.pachyderm.io>`_, submit an issue on `GitHub <https://github.com/pachyderm/pachyderm>`_, or email us at `support@pachyderm.io <mailto:support@pachyderm.io>`_. We are more than happy to help!

Prerequisites
^^^^^^^^^^^^^
This guide assumes that you already have Pachyderm running locally. Check out our :doc:`local_installation` instructions if haven't done that yet and then come back here to continue.

Create a Repo
^^^^^^^^^^^^^

A ``repo`` is the highest level data primitive in Pachyderm. Like many things in Pachyderm, it shares it's name with primitives in Git and is designed to behave analogously. Generally, repos should be dedicated to a single source of data such as log messages from a particular service, a users table, or training data for an ML model. Repos are dirt cheap so don't be shy about making tons of them.

For this demo, we'll simply create a repo called ``images`` to hold the data we want to process:

.. code-block:: shell

  $ pachctl create-repo images

  # See the repo we just created
  $ pachctl list-repo
  NAME                CREATED             SIZE
  images              2 minutes ago       0 B


Adding Data to Pachyderm
^^^^^^^^^^^^^^^^^^^^^^^^

Now that we've created a repo it's time to add some data. In Pachyderm, you write data to an explicit ``commit`` (again, similar to Git). Commits are immutable snapshots of your data which give Pachyderm its version control properties. ``Files`` can be added, removed, or updated in a given commit.

Let's start by just adding a file, in this case an image, to a new commit. We've provided some sample images for you that we host on Imgur. 

We'll use the ``put-file`` command along with two flags, ``-c`` and ``-f``. ``-f`` can take either a local file or a URL which it'll automatically scrape. In our case, we'll simply pass the URL.

Unlike Git though, commits in Pachyderm must be explicitly started and finished as they can contain huge amounts of data and we don't want that much "dirty" data hanging around in an unpersisted state. The ``-c`` flag specifies that we want to start a new commit, add data, and finish the commit in a convenient one-liner.

We also specify the repo name "images", the branch name "master", and what we want to name the file, "liberty.png".

.. code-block:: shell

	$ pachctl put-file images master liberty.png -c -f http://imgur.com/46Q8nDz.png

Finally, we check to make sure the data we just added is in Pachyderm.

.. code-block:: shell

  # If we list the repos, we can see that there is now data
  $ pachctl list-repo
  NAME                CREATED             SIZE
  images              5 minutes ago   57.27 KiB

  # We can view the commit we just created
  $ pachctl list-commit images
  REPO                ID                                 PARENT              STARTED            DURATION            SIZE
  images              7162f5301e494ec8820012576476326c   <none>              2 minutes ago      38 seconds          57.27 KiB
  
  # And view the file in that commit
  $ pachctl list-file images master
  NAME                TYPE                SIZE
  liberty.png         file                57.27 KiB

We can view the file we just added to Pachyderm. Since this is an image, we can't just print it out in the terminal, but the following commands will let you view it easily.

.. code-block:: shell
 
  # on OSX
  $ pachctl get-file images master liberty.png | open -f -a /Applications/Preview.app

  # on Linux
  $ pachctl get-file images master liberty.png | display

Create a Pipeline
^^^^^^^^^^^^^^^^^

Now that we've got some data in our repo, it's time to do something with it. ``Pipelines`` are the core processing primitive in Pachyderm and they're specified with a JSON encoding. For this example, we've already created the pipeline for you and you can find the `code on Github <https://github.com/pachyderm/pachyderm/blob/master/doc/examples/opencv>`_. 

When you want to create your own pipelines later, you can refer to the full :doc:`../reference/pipeline_spec` to use more advanced options. This includes building your own code into a container instead of the pre-built Docker image we'll be using here.

For now, we're going to create a single pipeline that takes in images and does some simple edge detection.

.. image:: opencv-liberty.png

Below is the pipeline spec and python code we're using. Let's walk through the details. 

.. code-block:: shell

  # edges.json
  {
    "pipeline": {
      "name": "edges"
    },
    "transform": {
      "cmd": [ "python3", "/edges.py" ],
      "image": "pachyderm/opencv"
    },
    "input": {
      "atom": {
        "repo": "images",
        "glob": "/*"
      }
    }
  }


Our pipeline spec contains a few simple sections. First is the pipeline ``name``, edges. Then we have the ``transform`` which specifies the docker image we want to use, ``pachyderm/opencv`` (defaults to Dockerhub as the registry), and the entry point ``edges.py``. Lastly, we specify the input.  Here we only have one "atom" input, our images repo with a particular glob pattern. 

The glob pattern defines how the input data can be broken up if we wanted to distribute our computation. ``/*`` means that each file can be processed individually, which makes sense for images. Glob patterns are one of the most powerful features of Pachyderm so when you start creating your own pipelines, check out the :doc:`../reference/pipeline_spec`.

.. code-block:: python

  # edges.py
  import cv2
  import numpy as np
  from matplotlib import pyplot as plt
  import os
  
  # make_edges reads an image from /pfs/images and outputs the result of running
  # edge detection on that image to /pfs/out. Note that /pfs/images and
  # /pfs/out are special directories that Pachyderm injects into the container.
  def make_edges(image):
     img = cv2.imread(image)
     tail = os.path.split(image)[1]
     edges = cv2.Canny(img,100,200)
     plt.imsave(os.path.join("/pfs/out", os.path.splitext(tail)[0]+'.png'), edges, cmap = 'gray')

  # walk /pfs/images and call make_edges on every file found
  for dirpath, dirs, files in os.walk("/pfs/images"):
     for file in files:
         make_edges(os.path.join(dirpath, file))

Our python code is really straight forward. We're simply walking over all the images in ``/pfs/images``, do our edge detection and write to ``/pfs/out``. 

``/pfs/images`` and ``/pfs/out`` are special local directories that Pachyderm creates within the container for you. All the input data for a pipeline will be found in ``/pfs/[input_repo_name]`` and your code should always write to ``/pfs/out``.

Now let's create the pipeline in Pachyderm:

.. code-block:: shell

  $ pachctl create-pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/doc/examples/opencv/edges.json



What Happens When You Create a Pipeline
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Creating a pipeline tells Pachyderm to run your code on **every** finished commit in a repo as well as **all future commits** that happen after the pipeline is created. Our repo already had a commit, so Pachyderm automatically launched a ``job`` to process that data. 

This first time it runs a pipeline it needs to download the image from DockerHub so this might take a minute. Every subsequent run will be much faster. 

You can view the job with:

.. code-block:: shell

  $ pachctl list-job
  ID                                     OUTPUT COMMIT                            STARTED             DURATION            STATE
  a6c70aa5-9f0c-4e36-b30a-4387fac54eac   edges/1a9c76a2cd154e6e90f200fb80c46d2f   2 minutes ago      About a minute      success

Every pipeline creates a corresponding repo with the same name where it stores its output results. In our example, the "edges" pipeline created a repo called "edges" to store the results. 

.. code-block:: shell

  $ pachctl list-repo
  NAME                CREATED            SIZE
  edges               2 minutes ago      22.22 KiB
  images              10 minutes ago     57.27 KiB


Reading the Output
^^^^^^^^^^^^^^^^^^

We can view the output data from the "edges" repo in the same fashion that we viewed the input data.

.. code-block:: shell
 
  # on OSX
  $ pachctl get-file edges master liberty.png | open -f -a /Applications/Preview.app

  # on Linux
  $ pachctl get-file edges master liberty.png | display

The output should look similar to:

.. image:: edges-screenshot.png

Processing More Data
^^^^^^^^^^^^^^^^^^^^

Pipelines will also automatically process the data from new commits as they are created. Think of pipelines as being subscribed to any new commits on their input repo(s). Also similar to Git, commits have a parental structure that tracks which files have changed. In this case we're going to be adding more images.

Let's create two new commits in a parental structure. To do this we will simply do two more ``put-file`` commands with ``-c`` and by specifying ``master`` as the branch, it'll automatically parent our commits onto each other. Branch names are just references to a particular HEAD commit.

.. code-block:: shell

  $ pachctl put-file images master AT-AT.png -c -f http://imgur.com/8MN9Kg0.png

  $ pachctl put-file images master kitten.png -c -f http://imgur.com/g2QnNqa.png

Adding a new commit of data will automatically trigger the pipeline to run on the new data we've added. We'll see corresponding jobs get started and commits to the output "edges" repo. Let's also view our new outputs. 

.. code-block:: shell

  # view the jobs that were kicked off
  $ pachctl list-job
  ID                                     OUTPUT COMMIT                            STARTED             DURATION             STATE
  7395c7c9-df0e-4ea8-8202-ec846970b982   edges/8848e11056c04518a8d128b6939d9985   2 minutes ago      Less than a second   success
  b90afeb1-c12b-4ca5-a4f4-50c50efb20bb   edges/da51395708cb4812bc8695bb151b69e3   2 minutes ago      1 seconds            success
  9182d65e-ea36-4b98-bb07-ebf40fefcce5   edges/4dd2459531414d80936814b13b1a3442   5 minutes ago      3 seconds            success

.. code-block:: shell

  # View the output data

  # on OSX
  $ pachctl get-file edges master AT-AT.png | open -f -a /Applications/Preview.app

  $ pachctl get-file edges master kitten.png | open -f -a /Applications/Preview.app

  # on Linux
  $ pachctl get-file edges master AT-AT.png | display

  $ pachctl get-file edges master kitten.png | display

Adding Another Pipeline
^^^^^^^^^^^^^^^^^^^^^^^

We have succesfully deployed and utilized a single stage Pachyderm pipeline, but now let's add a processing stage to illustrate a multi-stage Pachyderm pipeline. Specifically, let's add a ``montage`` pipeline that take our original and edge detected images and arranges them into a single montage of images:

.. image:: opencv-liberty-montage.png

Below is the pipeline spec for this new pipeline:

.. code-block:: shell

  # montage.json
  {
    "pipeline": {
      "name": "montage"
    },
    "input": {
      "cross": [ {
        "atom": {
          "glob": "/",
          "repo": "images"
        }
      },
      {
        "atom": {
          "glob": "/",
          "repo": "edges"
        }
      } ]
    },
    "transform": {
      "cmd": [ "sh" ],
      "image": "v4tech/imagemagick",
      "stdin": [ "montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png" ]
    }
  }

This pipeline spec is very similar to our ``edges`` pipeline except, for ``montage``: (1) we are using a different Docker image that has imagemagick installed, (2) we are executing a ``sh`` command with ``stdin`` instead of a python script, and (3) we have multiple input data repositories.  

In this case we are combining our multiple input data repositories using a ``cross`` pattern.  There are multiple interesting ways to combine data in Pachyderm, which are further discussed `here <http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#input-required>`_ and `here <http://pachyderm.readthedocs.io/en/latest/cookbook/combining.html>`_.  For the purposes of this example, suffice it to say that this ``cross`` pattern creates a single pairing of our input images with our edge detected images.

We create this next pipeline as before, with ``pachctl``:

.. code-block:: shell

  $ pachctl create-pipeline -f montage.json

This will automatically trigger jobs that generate montages for all of the commits to our input repo:

.. code-block:: shell

  $ pachctl list-job
  ID                                   OUTPUT COMMIT                            STARTED        DURATION           RESTART PROGRESS  DL       UL       STATE            
  3bddeb20-1b49-4980-85ab-8423ab902322 montage/bff0999b4cd84f0d96a923f92f933c1e 2 seconds ago  1 second           0       1 + 0 / 1 371.9KiB 1.284MiB success 
  246377b0-0710-437b-ad9f-9dafd326efed montage/8956fa782b894ef68ef9ae0a3213334d 3 seconds ago  Less than a second 0       1 + 0 / 1 195.3KiB 809.2KiB success 
  3c03465c-3d96-4d57-ac18-40b761f943cd montage/6cdb6ca9455341d193555cd9e523bc5b 3 seconds ago  Less than a second 0       1 + 0 / 1 79.49KiB 378.6KiB success 
  80651695-98ef-44dd-a382-f476b9bc75ea edges/8fa0be775d66479daadb71af82d6ed49   38 seconds ago Less than a second 0       1 + 2 / 3 102.4KiB 74.21KiB success 
  d97bdaad-e651-4ee5-a3c4-5b6c7697ccb6 edges/e3e40dee57174bacbd359575f2e4a87f   44 seconds ago Less than a second 0       1 + 1 / 2 78.7KiB  37.15KiB success 
  28d0be41-69cf-46d3-b3e0-ebe2a66540d1 edges/61f5e8fbf4b444219805e49584a42d5c   51 seconds ago Less than a second 0       1 + 0 / 1 57.27KiB 22.22KiB success

And you can view the generated montage image via:

.. code-block:: shell

  # on OSX
  $ pachctl get-file montage master montage.png | open -f -a /Applications/Preview.app

  # on Linux
  $ pachctl get-file montage master montage.png | display

.. image:: montage-screenshot.png

Next Steps
^^^^^^^^^^
We've now got Pachyderm running locally with data and a pipeline! If you want to keep playing with Pachyderm locally, you can  use what you've learned to build on or change this pipeline. You can also start learning some of the more advanced topics to develop analysis in Pachyderm:

- :doc:`../deployment/deploy_intro` 
- :doc:`../fundamentals/getting_data_into_pachyderm`
- :doc:`../fundamentals/creating_analysis_pipelines`

We'd love to help and see what you come up with so submit any issues/questions you come across on `GitHub <https://github.com/pachyderm/pachyderm>`_ , `Slack <http://slack.pachyderm.io>`_ or email at support@pachyderm.io if you want to show off anything nifty you've created!

**Note**: This example has been tested on Pachyderm version 1.5.2. It needs to be updated for the latest versions of Pachyderm.

# Game of Thrones / Tensor Flow Example

Adapted from the tensor flow LSTM example [here](https://www.tensorflow.org/versions/r0.8/tutorials/recurrent/index.html#recurrent-neural-networks)

## Overview

In this example, you'll generate a new Game of Thrones script based on a bunch of previous GoT scripts.

To do so, we'll be adapting [this LSTM Neural Net example](https://www.tensorflow.org/versions/r0.8/tutorials/recurrent/index.html#recurrent-neural-networks) from Tensor Flow. We won't cover any LSTM or Neural Net theory in this example. For background we recommend reading that example and the resources they link to.

This guide assumes you already have a [working pachyderm setup](http://pachyderm.readthedocs.io/en/stable/getting_started/local_installation.html), and you have a basic grasp of Pachyderm repos and pipelines. If you don't, you may want to start with the [fruit stand](http://pachyderm.readthedocs.io/en/stable/getting_started/beginner_tutorial.html) example or our [cloud deployment guide](http://pachyderm.readthedocs.io/en/stable/deployment/deploying_on_the_cloud.html).

## How

Getting this neural net running on Pachyderm will require a few steps:

1. Creating the data repo, and initializing it with some GoT scripts
2. Creating the Docker image that includes the Tensor Flow library and our code
3. Creating the Pachyderm Pipeline that trains the neural net, and generates the new script

---

### Initializing the Data

#### Loading the data

Since this data set isn't tiny, we've included some helpers to create the repos we need and input the data. To initialize the data set, just run:

```shell
make input-data
```

This task does 2 things:

1. It grabs the data set in the form of a tarball from a URL, and extracts the data
2. It inputs this data into Pachyderm by:
    - creating a new repo `GoT_scripts`
    - starting a commit
    - mounting the [Pachyderm File System (PFS)](http://pachyderm.io/pfs.html) at `./mnt`
    - copying the data over to the `./mnt/GoT_scripts/{commitID}/` path
    - finishing the commit

The result is a new repo with all the data we need stored inside. To confirm the setup, you can do:

```shell
$ pachctl list-repo
NAME                CREATED             SIZE
GoT_scripts         15 seconds ago      2.625 MiB
4
$ pachctl list-commit GoT_scripts
BRANCH        REPO/ID       PARENT              STARTED             FINISHED            SIZE
              master/0      <none>              24 seconds ago      24 seconds ago      2.625 MiB
```


#### Understanding the Data

For our neural net, we collected a bunch of Game of Thrones scripts and
pre-processed them to normalize them. If you take a look at one of the files:

```
$ cat data/all.txt | head -n 10
 <open-exp> First scene opens with three Rangers riding through a tunnel , leaving the Wall , and going into the woods <eos> (Eerie music in background) One Ranger splits off and finds a campsite full of mutilated bodies , including a child hanging from a tree branch <eos> A birds-eye view shows the bodies arranged in a shield-like pattern <eos> The Ranger rides back to the other two <eos> <close-exp>
 <boname> WAYMAR_ROYCE <eoname> What d 'you expect <question> They 're savages <eos> One lot steals a goat from another lot and before you know it , they 're ripping each other to pieces <eos>
 <boname> WILL <eoname> I 've never seen wildlings do a thing like this <eos> I 've never seen a thing like this , not ever in my life <eos>
 <boname> WAYMAR_ROYCE <eoname> How close did you get <question>
 <boname> WILL <eoname> Close as any man would <eos>
 <boname> GARED <eoname> We should head back to the wall <eos>
 <boname> ROYCE <eoname> Do the dead frighten you <question>
 <boname> GARED <eoname> Our orders were to track the wildlings <eos> We tracked them <eos> They won 't trouble us no more <eos>
 <boname> ROYCE <eoname> You don 't think he 'll ask us how they died <question> Get back on your horse <eos>
 <open-exp> GARED grumbles <eos> <close-exp>
```

You'll notice a bunch of funny tokens. Since the raw scripts had different ways
of denoting structure (some used capitalization and colons to denote who was
speaking .. others didn't), we normalized them so the the punctuation and
structure was consistently represented. You'll also noticed open/closing tokens
for non speaking 'exposition' lines. Don't worry too much about these tokens
right now. Once you see the output, you'll appreciate how the neural net has learned some
of this structure.

---

### Creating the Transformation Image

Using Tensor Flow with Pachyderm is easy! Since [Pachyderm Pipeline System (PPS)](http://pachyderm.io/pps.html) allows you to use any Docker image, getting the code in place is straightforward. In fact, since Tensor Flow provides a docker image to work with, they've done most of the work for us!

To construct the image, we need to:

1. Make sure we use an image with the Tensor Flow library installed
2. Make sure the image includes our code
3. Build the image

If you take a look at the [Dockerfile](./Dockerfile) in this directory, you'll notice a couple things.

1) The top line specifies that we're basing our image off of Tensor Flow's image:

```
FROM tensorflow/tensorflow
```

2) On line 5 we're including the [code found locally](./code) in the Docker image:

```
ADD code /code
```

3) Build the image

Again, we have a helper. If you run:

```
make docker-build
```

It will compile the Docker image above and name it `tensor_flow_rnn_got`

#### A few things to note:

This represents one of two ways to construct a custom image to run on [Pachyderm Pipeline System (PPS)](http://pachyderm.io/pps.html). In this example, we use the `FROM` directive to base our image off of a 3rd party image of our choosing. This works well - we just need to add the dependencies that [Pachyderm Pipeline System (PPS)](http://pachyderm.io/pps.html) needs to run. This is the more common use case when constructing custom images with complicated dependencies -- in this case you probably have an image in mind that has the dependencies you need. [Here is the canonical Dockerfile that will always list the dependencies you need to add](https://github.com/pachyderm/pachyderm/blob/master/etc/user-job/Dockerfile))

The alternative is using the `FROM` directive to base the image off of Pachyderm's standard image. [You can see an example here](../word_count/Dockerfile). This usage pattern is more helpful for simple dependencies, or if you don't already have a Docker image you need to use.

If you're familiar with Dockerfiles, one thing that seems noticeably absent is an `ENTRYPOINT` command to tell the image what to run. In [Pachyderm Pipeline System (PPS)](http://pachyderm.io/pps.html) you specify this when creating a pipeline, as we'll see in a moment.


---

### Creating the Pipeline

Now that we have the data and the image ready, we can specify how we want to process the data. To do that, we use a pipeline manifest. Take a look at `pipeline.json` in this directory.

#### Training:

We've specified two pipelines here. The first is `GoT_train`, which represents the processing needed to train the neural net on the input data. You can see that it takes `GoT_scripts` as its input.

Note that we've specified the `tensor_flow_rnn_got` image, and we specify an entry point by providing a command and its stdin:

```
cd /code && python ptb_word_lm.py --data_path=/data --model=small --model_path_prefix=/data
```

This line tells [Pachyderm Pipeline System (PPS)](http://pachyderm.io/pps.html) how to run the job. Here you can see we specify how to run the code with the arguments we want.

The outputs are dictated by the script. In this case its really two types of files -- the script outputs the training model weights into a Tensor Flow `checkpoint` file, and also outputs two `.json` files containing a map of ID -> word and vice versa. These files will be output to the repo matching the pipeline name: `GoT_train`. We'll use these files in the next pipeline.


#### Generating:

This pipeline uses the model from the training stage to generate new text. It needs to consume the trained model as input, as well as the ID -> word and word -> ID json maps to output human readable text. To enable this, we specify the `GoT_train` repo as our input.

Notice that we specify the same image as above, just a different entrypoint via the `transform` property of the manifest. In this step, we're outputting the results to a file called `new_script.txt`


#### Running:

Now that we have all of the pieces in place, we can run the pipeline. Do so by running:

`make run`

This creates a pipeline from `pipeline.json`, and since a commit exists on the very first input repo `GoT_scripts`, the pipeline runs automatically. To see what's happening, you can run:

```
$ pachctl list-job
ID                                 OUTPUT                                       STARTED             DURATION            STATE
8225e745ef8e3d0c4dcf550c895634e3   GoT_train/dd2c024a5da041cb89e12e7984c81359   9 seconds ago       -                   running
```

and once the jobs have completed you'll see the output commit on the `GoT_generate` repo:

```
$ pachctl list-job
ID                                 OUTPUT                                          STARTED             DURATION            STATE
8f137e20299c85d1f0326be6e8c1bca6   GoT_generate/dcc8ba9984d442ababc75ddff42a055b   4 minutes ago       16 seconds          success
8225e745ef8e3d0c4dcf550c895634e3   GoT_train/dd2c024a5da041cb89e12e7984c81359      6 minutes ago       2 minutes           success

$ pachctl list-commit GoT_generate
BRANCH              ID                                 PARENT              STARTED             FINISHED            SIZE
                    dcc8ba9984d442ababc75ddff42a055b   <none>              5 minutes ago       5 minutes ago       4.354 KiB
```

#### Results:

By default, we've set the size of the model to train to `test`. You can see this on each of the entrypoints specified in the `pipeline.json` file. For this size model, the training pipeline should run for a couple minutes, and the generate step much quicker than that. Let's take a look at the output.

Once `pachctl list-commit GoT_generate` shows a single commit, we can take a look at the output. To do so, you can run:

```
pachctl get-file GoT_generate {the commit id from the above command} new_script.txt
```

And you should see some text!

Keep in mind, the model we just trained was very simplistic. Doing the 'test' model suffices as a proof of concept that we can train and use a neural net using Tensor Flow on Pachyderm. That said, the test model is dumb, and the output won't be that readable.

Actually, you can see how 'dumb' the model is. If you [read through the Tensor Flow example](https://www.tensorflow.org/versions/r0.8/tutorials/recurrent/index.html#run-the-code) they describe how 'perplexity' is used to measure how good this model will perform. Let's look at the perplexity of your model.

```
$pachctl list-job
$pachctl get-logs {job ID from GoT_train job}
0 | Epoch: 1 Learning rate: 1.000
0 | 0.002 perplexity: 8526.820 speed: 1558 wps
0 | 0.102 perplexity: 880.494 speed: 1598 wps
0 | 0.202 perplexity: 674.214 speed: 1604 wps
0 | 0.302 perplexity: 597.037 speed: 1604 wps
0 | 0.402 perplexity: 561.110 speed: 1604 wps
0 | 0.501 perplexity: 535.682 speed: 1599 wps
0 | 0.601 perplexity: 518.105 speed: 1601 wps
0 | 0.701 perplexity: 504.523 speed: 1600 wps
0 | 0.801 perplexity: 498.718 speed: 1601 wps
0 | 0.901 perplexity: 491.138 speed: 1604 wps
0 | Epoch: 1 Train Perplexity: 486.294
0 | Epoch: 1 Valid Perplexity: 458.257
0 | Test Perplexity: 433.468
...
```

(You can ignore any FUSE errors in the logs. Most of these are innocuous)

Remember, this file is the stdout from the python script while its training the model. It outputs the 'perplexity' at standard intervals on the training set, as well as outputs the 'perplexity' for the validation and test sets. You should see a perplexity of about 600 or less.

That's not great. As a next step, you can improve that measure, and the readability of the output script!

#### Next Iteration

As referenced in the [Tensor Flow example](https://www.tensorflow.org/versions/r0.8/tutorials/recurrent/index.html#run-the-code), a 'perplexity' of less than 100 starts to get pretty good / readable. Running the 'small' model on the GoT data set should produce a perplexity lower than 100.

To do this, you'll need to tear down this pipeline and re-create it. Specifically:

1) Delete the existing data / pipeline by running:

```
pachctl delete-all
```

2) Change the entrypoint commands in `pipeline.json` under the transformation property to use the `small` model not the `test` model. Save your changes.

3) Re-initialize the data / compile the image / create the pipeline

Now you can do this in one step:

```
make all
```

The small model runs in about an hour. Once its complete, you can look at the output again. This time, the perplexity should be south of 100 and the output script should be semi-readable.

---

### Example Output

Here's an example script generated by the small model:

```
[ OLENNA ]:  Then where would she serve that Grand Maester and nine are men ? And soon , we need to rule that to a man who wore on the walls . .
[ CERSEI ]:  The world doesn 't make there . .
[ CERSEI ]:  Home . Oh . .
{{ MAESTER UNELLA takes the leaves , merchants pushes againt }} .
[ SAM ]:  Doesn 't make me wake 't asking a woman was better prison . .
{{ Something moving with Bronn and scans to another side , where Jon grabs his sword and is added on the actor 's main performance of life . Inside , as a number of moments cuts like that we 're thrown to Daenerys , who seems to stop . }} .
[ LADY_CRANE ]:  impression Ser later . Let him teach him . All I always have to banners south of my life , Meryn . .
{{ Stannis runs away and Melisandre walks through the lift door , no smiles) . }} .
[ TYRION ]:  I wanted him to be a spoiled Greyjoy again , that 's your niece . And you 're a terrible man . My little Give me A_VOICE . .
{{ The HIGH SPARROW sits down on the bucket , DORAN sentences to have waiting to the shoulder , wearing this building . She wargs off the woods and approaches The sound of the riders . BLACK WALDER reaches over his horse beside MARGAERY . }} .
[ EDMURE ]:  We 're among The Khaleesi . .
[ JORAH ]:  For far soon . .
{{ But let 's worry to my children ? }} .
[ SANSA_STARK ]:  He 's travelling a true vow to command until his father 's death . .
{{ The man cheer . JAIME walks up the stairs and looks to JON SNOW . }} .
[ JON ]:  People 're wrong enough of game for it . .
{{ Arya and Gilly reach YOUNG HODOR , then a bottle to meet the stones . punch out the basket of horse wrestles . After a moment , he walks away from it . Missandei looks up to see Ramsay turns his back to around Jon . }} .
{{ EXT . BRAAVOS - GATE }} .
{{ present by Maester Aemon , Myranda , , some having sex , and a few piece of commoners . A group of wights louder on the roof of the throne . Tyene is kneeling on a giant , and Tyrion . He catches her blade and looks at Ellaria shoots something else . DAVOS is seen walking aside . Viserys punches the out of a Harpy . The septa brings into their weapons , snapping to a man in the Yunkai and the Freys of Sand Motte . BRAN gasps on the marketplace and Bronn directly over a prop hill while parts of a carriage . BALON goes behind her }} .
[ CATELYN ]:  Who good one behind mammoths on those words ? .
[ TYRION ]:  I read you . .
[ PYCELLE ]:  And what are you in ? .
[ JORY_CASSEL ]:  Seven band . (Walking soup , Ser Hugh kneels . .
{{ Grey Worm still fired up . Animals goes off the hill . }} .
{{ INT . HOUSE OF BLACK AND WHITE }} .
{{ Sansa is a statue and Bronn examining a lady and tapping it . }} .
[ JEOR_MORMONT ]:  united the pace , Janos . .
{{ ARYA retrieves a couple of water . }} .
[ ARYA ]:  What you 're , I 've had a trap meeting , I lost . wolf puts you out . The greatest way work , for one of me , though . And if he never Speaking again . .
[ ARYA ]:  We were so happy . .
[ QYBURN ]:  I heard what he 'll turn . It was dark and "My 'ghar .
[ ARYA ]:  They did nothing for the old time . .
{{ LOTHAR turns to BRIENNE . }} .
[ LITTLEFINGER ]:  That your father is coming through Winterfell . For all of them were on the old city . He knows that 's someone what they came to us . .
[ DAARIO ]:  If I have fight in the North , the Targaryens , and beg it for your cock , falsely men is small . Cersei will last coming . .
[ BRIENNE ]:  Shouldn 't I be when she saw you ? .
[ SANSA ]:  Go on , no unrepentant . .
{{ The white guard . jumps down the corner and looks at BERIC . OLENNA and BRYNDEN shakes the horses in the courtyard . MARGAERY 's eyes rolls under them . }} .
[ DAVOS ]:  I don 't mind , Your Grace . I got to build your place and see rolled women in all of their fanatics And so I getting another conversation , watcher ships in the matters of the dead as gods of Westeros . I saw her serve once . I like easy , I can feel just if they 're out there . I 'd hurt a first hooded Dothraki chance , try to conquer . .
[ DAENERYS ]:  run ! .
[ TORMUND ]:  In the past what else was full of girls . .
[ ROBERT_BARATHEON ]:  Oh , perhaps . Who are you asking me ? .
[ JAIME ]:  I 'm not asking you to look to your father . .
{{ EXT . CASTLE BLACK }} .
{{ Jon walks through the Kingsroad , but BLACK LADY CRANE follows her on the table , while parts beside her . She finds another , WILDERNESS eye , and happens to him . He leaves . As Theon shoots out her grip , he turns to
```
#### Some interesting things to note

This neural net didn't know english, much less grammar, but it did pick up on some pieces of structure:

- it learned that each line is either a spoken line, or an exposition line
- it learned that spoken lines begin with a name
    - it learned which words were names
    - it learned how to open/close the brackets around a name)
    - it learned the types of words (present tense verbs, I/you nouns) are used in spoken lines
- it learned how to open / close the brackets for exposition lines
    - it learned what type of words (3rd person) are used


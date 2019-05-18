# ML Pipeline for Tweet Generation

In this example you'll machine learning pipeline that generates tweets
using a OpenAI's gpt-2 text generation model. This tutorial assumes that
you already have Pachyderm up and running and just focuses on the pipeline
creation. If that's not the case head on over to our [getting started
guide](link).

The pipeline we're making has 3 steps in it:

- tweet scraping
- model training
- tweet generation

At the top of our DAG is a repo that contains Twitter queries we'd like to
run to get our tweets to train on.

## Tweet scraping

The first step in our pipeline is scraping tweets off of twitter. This
step is called `tweets` and the code for it is in [tweets.py](link),
here's what it looks like:

```python
#!/usr/local/bin/python3
import os
import twitterscraper as t

for query in os.listdir("/pfs/queries/"):
    with open(os.path.join("/pfs/queries", query)) as f:
        for q in f:
            q = q.strip()  # clean whitespace
            with open(os.path.join("/pfs/out", query), "w+") as out:
                for tweet in t.query_tweets(q):
                    out.write("<|startoftext|> ")
                    out.write(tweet.text)
                    out.write(" <|endoftext|> ")
```

Most of this is fairly standard Pachyderm pipeline code, `"/pfs/queries"`
is the path where our input (a list of queries) is mounted. `query_tweets` is
where we actually send the query to twitter. And then we write the tweets
out to a file called `/pfs/out/<name-of-input-file>`. Notice that we
inject `"<|startoftext|>"` and `"<|endoftext|>"` at the beginning and end
of each tweet, these are special delimiters that gpt-2 has been trained on
and that we can use to generate one tweet at a time in our generation
step.

To deploy this as a Pachyderm pipeline we'll need a pipeline spec, you can
find it in [tweets.json](link), here's what it looks like:

```json
{
    "pipeline": {
        "name": "tweets"
    },
    "transform": {
        "image": "gpt-2-example",
        "cmd": ["/tweets.py"]
    },
    "input": {
        "pfs": {
            "repo": "queries",
            "glob": "/*"
        }
    }
}
```

Notice that we are taking the `"queries"` repo as input with a glob
pattern of `"/*"`, this means that our pipeline can run in parallel over
several queries. Before you can create this pipeline you'll need to create
its input repo:

```sh 
$ pachctl create repo queries
```

Now create the pipeline:

```sh
$ pachctl create pipeline -f tweets.json
```

The pipeline has now been created, let's test to see if it's working by giving
it a query:

```sh
$ echo "from:<username>" | pachctl put file queries@master:<username>
```

Note that the username should not contain the `@`, this is a fairly simple
query that just gets all the tweets from a single user. If you'd like to
construct a more complicated query, and aren't fluent in Twitter's query
language then [this page](https://twitter.com/search-advanced) will help you.
(Hit the search button and along the top of the page will be the query string.)

After you run that `put file` you will have a new commit in your `"queries"`
repo and a new output commit in `"tweets"`, along with a job that's scraping
the tweets. To see the job running do:

```sh
$ pachctl list job
```

Once it's finished you can view the scraped tweets with:

```sh
$ pachctl get file "tweets@master:/<username>"
```

Assuming those results look reasonable, let's move on to training a model.


# Model training

As mentioned we'll be using OpenAI's gpt-2 text generation model, actually
we'll be using a handy wrapper:
[gpt-2-simple](https://github.com/minimaxir/gpt-2-simple).

The code for this is in [train.py](link) and looks like this:

```python
#!/usr/local/bin/python3
import gpt_2_simple as gpt2
import os


tweets = [f for f in os.listdir("/pfs/tweets")]

# chdir so that the training process outputs to the right place
out = os.path.join("/pfs/out", tweets[0])
os.mkdir(out)
os.chdir(out)

model_name = "117M"
gpt2.download_gpt2(model_name=model_name)


sess = gpt2.start_tf_sess()
gpt2.finetune(sess,
              os.path.join("/pfs/tweets", tweets[0]),
              model_name=model_name,
              steps=25)   # steps is max number of training steps
```

Again most of this is standard Pachyderm pipeline code to grab our inputs (this
time our input is the `"tweets"`). We're also making a few choices in this
pipeline, we're using the 117M version of the model, for better results you can
use the 345M version of the model but expect it to take more time to train.
We're also limiting our training process to 25 steps, this was a more-or-less
arbitrary choice that seems to get good results without taking too long to run.

The pipeline spec for training the model is very similar to the one above for
scraping tweets:

```json
{
    "pipeline": {
        "name": "train"
    },
    "transform": {
        "image": "gpt-2-example",
        "cmd": ["/train.py"]
    },
    "input": {
        "pfs": {
            "repo": "tweets",
            "glob": "/*"
        }
    }
}
```

The only thing that's change is that we're taking `"tweets"` as input, rather
than `"queries"` and we're running a different script in our transform. You can
create this pipeline with:

```sh
$ pachctl create pipeline -f train.json
```

This will kick off a job immediately, because there are already inputs to be
processed. Expect this job to take a while to run, on a laptop with 4 Intel
Core i7-6600U CPU @ 2.60GHz and 32 GB of ram it took about an hour. While
that's running, let's setup the generate pipeline.

# Incrementality

Pachyderm performs computations in an incremental fashion, that is, rather
than computing a result all at once, it computes it in small pieces and
then stitches those pieces together. This allows Pachyderm to compute
things much more efficiently by reusing things that it's already computed
rather than computing them again. There are two forms of incrementality
that Pachyderm supports, this doc will cover them both and explain how to
leverage them to speed up your workload.

# Inter-datum Incrementality

Computations in Pachyderm are defined over a set of [datums](LINK). Each
of the input datums is processed in isolation, and the results of these
computations is combined to create the final result. Pachyderm will never
process the same datum twice (unless you update the pipeline with the
`--reprocess` flag). If you commit new data that leaves some of the datums
intact, then the results of processing that datum in a previous job will
be used. This is best illustrated with an example. Suppose a pipeline with
a single input that looks like this:

```json
{
  "atom": {
    "repo": "R",
    "glob": "*",
  }
}
```

Now, suppose you make a commit to `R` which adds a single file `F1`. Your
pipeline will run a job, and will find a single datum to process (`F1`).
This datum will be processed, because it's the first time the pipeline has
seen `F1` (it's the first time it's been run at all).

Now, suppose you make a second commit to `R` which adds another file `F2`.
Your pipeline will run a second job, this job will find two datums to
process (`F1` and `F2`). This is where it gets interesting, `F2` will be
processed, because it hasn't been seen before. However `F1` will not be
processed, because the output from processing it already exists within
PFS. Instead, that output from the previous job will be used to create the
output of this job. This effectively halves the amount of work necessary
to process the second commit.

Finally, suppose you make a third commit to `R` which modifies `F1`. Again
you'll have a job that sees two datums (the new `F1` and `F2`). This time
`F2` won't get processed, but the new `F1` will because it has different
content from the old `F1`.

Note that you, as a user, don't need to do anything to enable this
feature. It happens automatically, and should should be transparent from
your perspective. Notice, in the above example the result of processing
a commit never depends on the history of that commit, you always get the
same result you'd have gotten if you'd committed the same data in a single
commit. As of 1.5.1, `list-job` and `inspect-job` will tell you how many
datums the job processed and how many it skipped.

# Intra-datum Incrementality

Pachyderm supports another form of incrementality, which is useful when
the processing you're doing can be done
["online"](https://en.wikipedia.org/wiki/Online_algorithm). Because not
all computations can be done online you have to enable this form of
incrementality with the [`incremental`](LINK) field in pipelines.
A canonical example of such an operation is summing a set of numbers.
Again an example is instructive, suppose you have a pipeline like the
above. However, instead of each datum being a single file, it's
a directory which contains multiple files, each of which contains multiple
numbers, and the pipeline enables incrementality via the `incremental`
field.

Suppose the first commit you make adds `F1/1`, your pipeline will run
a job which sums up the files in `F1/1`. Just like it would if the
pipeline were not incremental.

Suppose the second commit now adds `F1/2`. Again, the second commit is
where things get interesting. You'll get another job which again processes
`F1`. However, the data it sees in `F1` will be different from what it
would be if the pipeline weren't `incremental`. Instead of seeing `F1/1`
and `F1/2` it will only see `F1/2`. Furthermore, unlike in non-incremental
jobs the output directory `/pfs/out` won't be empty, instead it will
contain the results of last time `F1` was processed. That is, it will
contain the sum of all the numbers in `F1/1. All your code needs to do is
sum the numbers in `F1/2` and add them to the results of summing `F1/1`
which are already contained in `/pfs/out`. Ultimately this process should
overwrite the value in `/pfs/out` with a new value.

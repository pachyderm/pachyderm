# Incremental Processing

Pachyderm performs computations in an incremental fashion. That is, rather than
computing a result all at once, it computes it in small pieces and then stitches
those pieces together to form results. This allows Pachyderm to reuse results
and compute things much more efficiently than traditional systems, which are
forced to compute everything from scratch during every job.

If you are new to the idea of Pachyderm "datums," you can learn more
[here](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/).

## Inter-datum Incrementality

Each of the input datums in a Pachyderm pipeline is processed in isolation, and
the results of these isolated computations are combined to create the final
result. Pachyderm will never process the same datum twice (unless you update a
pipeline with the `--reprocess` flag). If you commit new data in Pachyderm that
leaves some of the previously existing datums intact, the results of processing
those pre-existing datums in a previous job will also remain intact. That is,
the previous results for those pre-existing datums won't be recalculated.

This inter-datum incrementality is best illustrated with an example. Suppose we
have a pipeline with a single input that looks like this:

```json
{
    "pfs": {
        "repo": "R",
        "glob": "/*"
    }
}
```

Now, suppose you make a commit to `R` which adds a single file `F1`. Your
pipeline will run a job, and that job will find a single datum to process
(`F1`). This datum will be processed, because it's the first time the pipeline
has seen `F1`.

![alt tag](incrementality1.png)

If you then make a second commit to `R` adding another file `F2`, the pipeline
will run a second job. This job will find two datums to process (`F1` and `F2`).
`F2` will be processed, because it hasn't been seen before. However `F1` will
NOT be processed, because an output from processing it already exists in
Pachyderm.

Instead, the output from the previous job for `F1` will be combined with the new
result from processing `F2` to create the output of this second job. This reuse
of the result for `F1` effectively halves the amount of work necessary to
process the second commit.

![alt tag](incrementality2.png)

Finally, suppose you make a third commit to `R`, which modifies `F1`. Again
you'll have a job that sees two datums (the new `F1` and the already processed
`F2`). This time `F2` won't get processed, but the new `F1` will be processed
because it has different content as compared to the old `F1`.

![alt tag](incrementality3.png)

Note, you as a user don't need to do anything to enable this inter-datum
incrementality. It happens automatically, and it should should be transparent
from your perspective. In the above example, you get the same result you would
have gotten if you committed the same data in a single commit.

As of Pachyderm v1.5.1, `list job` and `inspect job` will tell you how many
datums the job processed and how many it skipped. Below is an example of a job
that had 5 datums, 3 that were processed and 2 that were skipped.

```
ID                                   OUTPUT COMMIT                             STARTED            DURATION           RESTART PROGRESS      DL       UL       STATE
54fbc366-3f11-41f6-9000-60fc8860fa55 pipeline/9c348deb64304d118101e5771e18c2af 13 seconds ago     10 seconds         0       3 + 2 / 5     0B       0B       success
```

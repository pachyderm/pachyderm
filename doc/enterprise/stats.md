# Advanced Statistics

To take advantage of the advanced statistics features in Pachyderm Enterprise Edition, you need to:

1. Run your pipelines on a Pachyderm cluster that has activated Enterprise features (see [Deploying Enterprise Edition](deployment.html) for more details).
2. Enable stats collection in your pipelines by including `"enable_stats": true` in your [pipeline specifications](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#enable-stats-optional).

You will then be able to access the following information for any jobs corresponding to your pipelines:

- The amount of data that was uploaded and downloaded during the job and on a per-datum level (see [here](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html#datums) for info about Pachyderm datums).
- The time spend uploading and downloading data on a per-datum level.
- The amount of data uploaded and downloaded on a per-datum level.
- The total time spend processing on a per-datum level.
- Success/failure information on a per-datum level.
- The directory structure of input data that was seen by the job.

The primary and recommended way to view this information is via the Pachyderm Enterprise dashboard, which can be deployed as detailed [here](deployment.html#deploying-the-pachyderm-enterprise-edition-dashboard). However, the same information is available through the `inspect-datum` and `list-datum` `pachctl` commands or through their language client equivalents.  

**Note** - We recommend enabling stats for all of your pipeline and only disabling the feature for very stable, long-running pipelines. In most cases, the debugging/maintainence benefits of the stats data will outweigh any disadvantages of storing the extra data associated with the stats. Also note, none of your data is duplicated in producing the stats.

## Enabling stats for a pipeline

As mentioned above, enabling stats collection for a pipeline is as simple as adding the `"enable_stats": true` field to a pipeline specification.  For example, to enable stats collection for our [OpenCV demo pipeline](http://pachyderm.readthedocs.io/en/latest/getting_started/beginner_tutorial.html#image-processing-with-opencv), we would modify the pipeline specification as follows:

```
{
  "pipeline": {
    "name": "edges"
  },
  "input": {
    "atom": {
      "glob": "/*",
      "repo": "images"
    }
  },
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv"
  },
  "enable_stats": true
}
```

Once the pipeline has been created and you have utilized it to process data, you can confirm that stats are being collected with `list-file`. There should now be stats data in the output repo of the pipeline under a branch called `stats`:

```
$ pachctl list-file edges stats
NAME                                                               TYPE                SIZE                
002f991aa9db9f0c44a92a30dff8ab22e788f86cc851bec80d5a74e05ad12868   dir                 342.7KiB            
0597f2df3f37f1bb5b9bcd6397841f30c62b2b009e79653f9a97f5f13432cf09   dir                 1.177MiB            
068fac9c3165421b4e54b358630acd2c29f23ebf293e04be5aa52c6750d3374e   dir                 270.3KiB            
0909461500ce508c330ca643f3103f964a383479097319dbf4954de99f92f9d9   dir                 109.6KiB
etc...
```

Don't worry too much about this view of the stats data.  It just confirms that stats are being collected.

## Accessing stats via the dashboard

Assuming that you have deployed and activated the Pachyderm Enterprise dashboard, you can explore your advanced statistics in just a few clicks. For example, if we navigate to our `edges` pipeline (specified above), we will see something similar to this:

![alt tag](stats1.png)

In this example case, we can see that the pipeline has had 1 recent successful job and 2 recent job failures.  Pachyderm advanced stats can be very helpful in debugging these job failures.  When we click on one of the job failures we will see the following general stats about the failed job (total time, total data upload/download, etc.):

![alt tag](stats2.png)

To get more granular per-datum stats (see [here](http://pachyderm.readthedocs.io/en/latest/fundamentals/distributed_computing.html#datums) for info on Pachyderm datums), we can click on the `41 datums total`, which will reveal the following:

![alt tag](stats3.png)

We can easily identify the exact datums that caused our pipeline to fail and the associated stats:

- Total time
- Time spent downloading data
- Time spent processing
- Time spent uploading data
- Amount of data downloaded
- Amount of data uploaded

If we need to, we can even go a level deeper and explore the exact details of a failed datum.  Clicking on one of the failed datums will reveal the logs corresponding to the datum processing failure along with the exact input files of the datum:

![alt tag](stats4.png)



# Processing Time-Windowed Data

If you are analyzing data that is changing over time, chances are that you will want to perform some sort of analysis on "the last two weeks of data," "January's data," or some other moving or static time window of data.  There are a few different ways of doing these types of analyses in Pachyderm, depending on your use case.  We recommend one of the following patterns:

1. [Fixed time window directory structures](#fixed-time-window-directory-structures) - for rigid, fixed time windows, such as months (Jan, Feb, etc.) or days (01-01-17, 01-02-17, etc.).

2. [Rolling time window directory structures](#rolling-time-window-directory-structures) - for rolling time windows of data, such as three day windows or two week windows. 

3. [Shuffled moving time windows](#shuffled-moving-time-windows) - for updating a constanly moving time-windowed data set.

## Fixed time window directory structures

As further discussed in [Creating Analysis Pipelines](http://docs.pachyderm.io/en/latest/fundamentals/creating_analysis_pipelines.html) and [Distributed Computing](http://docs.pachyderm.io/en/latest/fundamentals/distributed_computing.html), the basic unit of data partitioning in Pachyderm is a "datum" which is defined by a glob pattern. When analyzing data within fixed time windows (e.g., corresponding to fixed calendar times/dates), we recommend organizing your data repositories such that each of the time windows that you are going to analyze corresponds to a separate file or directory in your repository. By doing this, you will be able to:

- Analyze each time window in parallel.
- Only re-process data within a time window when that data, or a corresponding data pipeline, changes.

For example, if you have monthly time windows of CSV sales data that need to be analyzed, you could create a `sales` data repository and structure it like:

```
sales
├── January.csv
├── February.csv
├── March.csv
└── April.csv
```

When you run a pipeline with a glob pattern of `/*` on `sales`, each of the months of sales data is processed in parallel (if possible).  Further, when you add a new month (e.g., `May.csv`), only that new datum, `May.csv`, will be processed.

In some cases, you may also have more than one file associated with a particular time window or more than one timeframe that is relevant (e.g., both months and days).  In these cases, you would want to create a directory structure like:

```
sales
├── January
|   ├── online_sales.csv
|   └── in_store_sales.csv
├── February
|   ├── online_sales.csv
|   └── in_store_sales.csv
└── March
    ├── online_sales.csv
    └── in_store_sales.csv
```

or, in the case when you need two time scales (months AND days):

```
sales
├── January
|   ├── 01-01-17.csv
|   ├── 01-02-17.csv
|   └── etc...
├── February
|   ├── 01-01-17.csv
|   ├── 01-02-17.csv
|   └── etc...
└── March
    ├── 01-01-17.csv
    ├── 01-02-17.csv
    └── etc...
```

Now you can create:

- Pipelines that process data on a monthly basis via a `/*` glob pattern.
- Pipelines that only use a certain month's data via, e.g., a `/January/*` or `/January/` glob pattern.
- Pipelines that process data on a daily basis via a `/*/*` glob pattern.
- Any combination of the above.

## Rolling time window directory structures

In certain use cases, you need to run analyses for rolling time windows, even when those don't correspond to certain calendar months, days, etc.  For example, you may need to analyze the last three days of data, the three days of data prior to that, the three days of data prior to that, etc.  In other words, you need to run an analysis for every rolling length of time.

For rolling time windows, we recommend a two stage structuring and analyzing of your data in "bins" that correspond to each rolling time window:

- *Pipeline 1* - Read in data, determine which bins the data corresponds to, and write the data into those bins   

- *Pipeline 2* - Read in and analyze the binned data. 

Let's take the three day rolling time windows as an example, and let's say that we want to analyze three day rolling windows of sales data.  In a first repo, called `sales`, a first day's worth of sales data is committed:

```
sales
└── 01-01-17.csv
```

We then create a first pipeline to bin this into a repository directory corresponding to our first rolling time window from 01-01-17 to 01-03-17:

```
binned_sales
└── 01-01-17_to_01-03-17
    └── 01-01-17.csv
```

When our next day's worth of sales is committed,

```
sales
├── 01-01-17.csv
└── 01-02-17.csv
```

the first pipeline executes again to bin the 01-02-17 data into any relevant bins.  In this case, we would put it in the previously created bin for 01-01-17 to 01-03-17, but we would also put it into a bin starting on 01-02-17:

```
binned_sales
├── 01-01-17_to_01-03-17
|   ├── 01-01-17.csv
|   └── 01-02-17.csv
└── 01-02-17_to_01-04-17
    └── 01-02-17.csv
```

As more and more daily data is added, you will end up with a directory structure that looks like:

```
binned_sales
├── 01-01-17_to_01-03-17
|   ├── 01-01-17.csv
|   ├── 01-02-17.csv
|   └── 01-03-17.csv
├── 01-02-17_to_01-04-17
|   ├── 01-02-17.csv
|   ├── 01-03-17.csv
|   └── 01-04-17.csv
├── 01-03-17_to_01-05-17
|   ├── 01-03-17.csv
|   ├── 01-04-17.csv
|   └── 01-05-17.csv
└── etc...
```

Your second pipeline can then process these bins in parallel, via a glob pattern of `/*`, or in any other relevant way as discussed further in the ["Fixed time windows" section](#fixed-time-window-directory-structures).  Both your first and second pipelines can be easily parallelized.

**Note** - When looking at the above directory structure, it may seem like there is an uneccessary duplication of the data.  However, under the hood Pachyderm deduplicates all of these files and maintains a space efficient representation of your data.  The binning of the data is merely a structural re-arrangement to allow you to process these types of rolling time windows.  

**Note** - It might also seem as if there is unecessary data transfers over the network to perform the above binning.  However, the Pachyderm team is currently working on enhancements to ensure that performing these types of "shuffles" doesn't actually require transferring data over the network.

## Shuffled moving time windows

The advantage of the ["Rolling time window" pattern](#rolling-time-window-directory-structures) above is that both the binning and time-windowed processing can be easily parallelized.  However, you do need to put in some logic to maintain the binning directory structure.  However, there is another pattern for moving time windows that avoids the binning of the above approach and maintains an up-to-date version of a moving time-windowed data set.  The is also a two stage process:

- *Pipeline 1* - Read in data, determine which files belong in your moving time window, and write the relevant files into an updated version of the moving time-windowed data set.  

- *Pipeline 2* - Read in and analyze the moving time-windowed data set.

**Note** - The disadvantage of this method is that "Pipeline 1" does not parallelize.  You would need to use a glob pattern of `/` in this pipeline to be able to extract the relevant files (corresponding to the moving time window) from all of the files that are input.

Let's utilize our sales example again to see how this would work.  In the example, we want to keep a moving time window of the last three days worth of data.  Now say that our daily `sales` repo looks like the following:

```
sales
├── 01-01-17.csv
├── 01-02-17.csv
├── 01-03-17.csv
└── 01-04-17.csv
```

Our first pipeline would extract the last three days worth of data and put it into a repo corresponding to our moving time window (think of this as a "shuffle" step):

```
moving_sales_window
├── 01-02-17.csv
├── 01-03-17.csv
└── 01-04-17.csv
```

Then, when a new day's worth of data is added to our `sales` repo,

```
sales
├── 01-01-17.csv
├── 01-02-17.csv
├── 01-03-17.csv
├── 01-04-17.csv
└── 01-05-17.csv
```

the first pipeline would update the moving window:

```
moving_sales_window
├── 01-03-17.csv
├── 01-04-17.csv
└── 01-05-17.csv
```

Whatever analysis we need to run on the moving windowed data set in `moving_sales_window` can use a glob pattern of `/` or `/*` (depending on whether we need to process all of the time windowed files together or they can be processed in parallel).


# Structured Data

This is an ML pipeline for the Boston Housing Dataset. It utilizes pandas dataframes to perform the following:

1. Import the structured dataset
2. Perform exploratory data analysis (EDA)
3. Trains a regression model (with a light grid search) to predict the housing price
4. Generate a learning curve and performance metrics to estimate the quality of the model

For simplicity, we'll just train an of Gradient Boosting Regressor ensemble, but we could configure it to train many different types of models

## Housing prices datasets

The housing prices dataset used for this example is a reduced version of the original, focusing on a subset of the features.
|Feature| Description|
|---|---|
|RM |       Average number of rooms per dwelling|
|LSTAT |    Percent lower status of the population|
|PTRATIO |  Pupil-teacher ratio by town|
|MEDV |     Median value of owner-occupied homes in $1000's|

Sample:
|RM   |LSTAT|PTRATIO|MEDV|
|-----|----|----|--------|
|6.575|4.98|15.3|504000.0|
|6.421|9.14|17.8|453600.0|
|7.185|4.03|17.8|728700.0|
|6.998|2.94|18.7|701400.0|


## Getting Started
This example requires a running Pachyderm deployment. You can deploy a cluster on [PacHub](hub.pachyderm.com) or deploy locally as described here:

- [Pachyderm Getting Started](https://docs.pachyderm.com/latest/getting_started/)

Once everything is up, we can check the setup by running:

1. `kubectl get all` to ensure all the pods are up.
2. `pachctl version` which will show both the `pachctl` and `pachd` versions.

Next, clone this repo and follow the steps below.

## TLDR;

```bash
# Step 1: Create input data repository
pachctl create repo data

# Step 2: Create the regression pipeline
pachctl create pipeline -f regression.json

# Step 3: Add the housing dataset to the repo
cd data
pachctl put file data@master -f housing-simplified.csv

# Step 4: Download files once the pipeline has finished
pachctl get --recursive regression@master
```

### Step 1: Create input data repository

Once the pachyderm cluster is running, we will create a data repository where our dataset will go.

```bash
$ pachctl create repo data
$ pachctl list repo
NAME                CREATED             SIZE
data                3 seconds ago       0 B
```

### Step 2: Create the regression pipeline

We can now connect a pipeline to watch the data repo. Once this pipeline is created, it will be looking for any changes to its input, retraining if we modify the dataset.

```bash
$ pachctl create pipeline -f regression.json
```
The pipeline will write the output to a PFS repo (`/pfs/out/` in the pipeline json) created with the same name as the pipeline.

### Step 3: Add the housing dataset to the repo
Now we can add the data which will kick off the processing automatically. If we update the data with a new commit, then the pipeline will automatically re-run. 

```bash
$ cd data
$ pachctl put file data@master -f housing-simplified.csv
```

We can inspect that the data is in the repository by looking at the files in the repository.

```bash
$ pachctl list file data@master
NAME                    TYPE SIZE
/housing-simplified.csv file 12.14KiB
```

### Step 4: Download files once the pipeline has finished
Once the pipeline is completed, we can download the files that were created.

```bash
$ pachctl list file regression@master
NAME               TYPE SIZE
/corr_matrix.png   file 18.66KiB
/cv_reg_output.png file 62.19KiB
/final_model.sav   file 1.007KiB
/pairplot.png      file 207.5KiB

$ pachctl get file regression@master:/ --recursive --output .
```





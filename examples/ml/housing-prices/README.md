>![pach_logo](../../img/pach_logo.svg) INFO Each new minor version of Pachyderm introduces profound architectural changes to the product. For this reason, our examples are kept in separate branches:
> - Branch Master: Examples using Pachyderm 2.1.x versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 2.0.x: Examples using Pachyderm 2.0.x versions - https://github.com/pachyderm/pachyderm/tree/2.0.x/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13.x versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples

This example creates a simple machine learning pipeline in Pachyderm to train a regression model on the Boston Housing Dataset to predict the value of homes in Boston. The pipeline itself is written in Python, though a Pachyderm pipeline could be written in any language.

<p align="center">
  <img width="200" height="300" src="images/regression_pipeline.png">
</p>

The Pachyderm pipeline performs the following actions:

1. Imports the structured dataset (`.csv`) with `pandas`.
2. Performs data analysis with `scikit-learn`.
3. Trains a regression model to predict housing prices.
4. Generates a learning curve and performance metrics to estimate the quality of the model.

Table of Contents:

- [Housing Prices Dataset](#housing-prices-dataset)
- [Prerequisites](#prerequisites)
- [Python Code](#python-code)
  - [Data Analysis](#data-analysis)
  - [Train a regression model](#train-a-regression-model)
  - [Evaluate the model](#evaluate-the-model)
- [Pachyderm Pipeline](#pachyderm-pipeline)
  - [TLDR; Just give me the code](#tldr-just-give-me-the-code)
  - [Step 1: Create an input data repository](#step-1-create-an-input-data-repository)
  - [Step 2: Create the regression pipeline](#step-2-create-the-regression-pipeline)
  - [Step 3: Add the housing dataset to the repo](#step-3-add-the-housing-dataset-to-the-repo)
  - [Step 4: Download files once the pipeline has finished](#step-4-download-files-once-the-pipeline-has-finished)
  - [Step 5: Update Dataset](#step-5-update-dataset)
  - [Step 6: Inspect the Pipeline Lineage](#step-6-inspect-the-pipeline-lineage)

## Housing Prices Dataset

The housing prices dataset used for this example is a reduced version of the original [Boston Housing Datset](https://www.cs.toronto.edu/~delve/data/boston/bostonDetail.html), which was originally collected by the U.S. Census Service. We choose to focus on three features of the originally dataset (RM, LSTST, and PTRATIO) and the output, or target (MEDV) that we are learning to predict.
|Feature| Description|
|---|---|
|RM |       Average number of rooms per dwelling|
|LSTAT |    A measurement of the socioeconomic status of people living in the area|
|PTRATIO |  Pupil-teacher ratio by town - approximation of the local education system's quality|
|MEDV |     Median value of owner-occupied homes in $1000's|

Sample:
|RM   |LSTAT|PTRATIO|MEDV|
|-----|----|----|--------|
|6.575|4.98|15.3|504000.0|
|6.421|9.14|17.8|453600.0|
|7.185|4.03|17.8|728700.0|
|6.998|2.94|18.7|701400.0|

## Prerequisites

Before you can deploy this example you need to have the following components:

1. A clone of this Pachyderm repository on your local computer. (could potentially include those instructions)
2. A Pachyderm cluster - You can deploy locally as described [here](https://docs.pachyderm.com/latest/getting-started/).

Verify that your environment is accessible by running `pachctl version` which will show both the `pachctl` and `pachd` versions.
```shell
$ pachctl version
COMPONENT           VERSION
pachctl             2.7.0
pachd               2.7.0
```

## Python Code

The `regression.py` Python file contains the machine learning code for the example. We will give a brief description of it here, but full knowledge of it is not required for the example. 

```
$ python regression.py --help

usage: regression.py [-h] [--input INPUT] [--target-col TARGET_COL]
                     [--output DIR]

Structured data regression

optional arguments:
  -h, --help            show this help message and exit
  --input INPUT         csv file with all examples
  --target-col TARGET_COL
                        column with target values
  --output DIR          output directory
  ```

The regression code performs the following actions:

1. Analyses the data.
2. Trains a regressor.
3. Evaluates the model.

### Data Analysis
The first step in the pipeline creates a pairplot showing the relationship between features. By seeing what features are positively or negatively correlated to the target value (or each other), it can helps us understand what features may be valuable to the model.
<p align="center">
  <img width="500" height="400"  src="images/pairplot-1.png">
</p>

We can represent the same data in color form with a correlation matrix. The darker the color, the higher the correlation (+/-).

<p align="center">
  <img width="500" height="400"  src="images/corr_matrix-1.png">
</p>

### Train a regression model
To train the regression model using scikit-learn. In our case, we will train a Random Forest Regressor ensemble. After splitting the data into features and targets (`X` and `y`), we can fit the model to our parameters.  

### Evaluate the model
After the model is trained we output some visualizations to evaluate its effectiveness of it using the learning curve and other statistics.
<p align="center">
  <img src="images/cv_reg_output-1.png">
</p>


## Pachyderm Pipeline

Now we'll deploy this python code with Pachyderm.

### TLDR; Just give me the code

```shell
# Step 1: Create input data repository
pachctl create repo housing_data

# Step 2: Create the regression pipeline
pachctl create pipeline -f regression.json

# Step 3: Add the housing dataset to the repo
pachctl put file housing_data@master:housing-simplified.csv -f data/housing-simplified-1.csv

# Step 4: Download files once the pipeline has finished
pachctl get file regression@master:/ --recursive --output .

# Step 5: Update dataset with more data
pachctl put file housing_data@master:housing-simplified.csv -f data/housing-simplified-2.csv

# Step 6: Inspect the lineage of the pipeline
pachctl list commit regression@master
```

### Step 1: Create an input data repository

Once the Pachyderm cluster is running, create a data repository called `housing_data` where we will put our dataset.

```shell
$ pachctl create repo housing_data
$ pachctl list repo
NAME                CREATED             SIZE
housing_data        3 seconds ago       0 B
```

### Step 2: Create the regression pipeline

We can now connect a pipeline to watch the data repo. Pipelines are defined in `json` format. Here is the one that we'll be used for the regression pipeline:

```json
# regression.json
{
    "pipeline": {
        "name": "regression"
    },
    "description": "A pipeline that trains produces a regression model for housing prices.",
    "input": {
        "pfs": {
            "glob": "/*",
            "repo": "housing_data"
        }
    },
    "transform": {
        "cmd": [
            "python", "regression.py",
            "--input", "/pfs/housing_data/",
            "--target-col", "MEDV",
            "--output", "/pfs/out/"
        ],
        "image": "pachyderm/housing-prices:1.11.0"
    }
}
```

For the **input** field in the pipeline definition, we define input data repo(s) and a [glob pattern](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/glob-pattern/). A glob pattern tells the pipeline how to map data into a job, here we have it create a new job for each datum in the `housing_data` repository.

The **image** defines what Docker image will be used for the pipeline, and the **transform** is the command run once a pipeline job starts.

Once this pipeline is created, it watches for any changes to its input, and if detected, it starts a new job to train given the new dataset.

```shell
$ pachctl create pipeline -f regression.json
```

The pipeline writes the output to a PFS repo (`/pfs/out/` in the pipeline json) created with the same name as the pipeline.

### Step 3: Add the housing dataset to the repo
Now we can add the data, which will kick off the processing automatically. If we update the data with a new commit, then the pipeline will automatically re-run. 

```shell
$ pachctl put file housing_data@master:housing-simplified.csv -f data/housing-simplified-1.csv
```

We can inspect that the data is in the repository by looking at the files in the repository.

```shell
$ pachctl list file housing_data@master
NAME                    TYPE SIZE
/housing-simplified.csv file 12.14KiB
```

We can see that the pipeline is running by looking at the status of the job(s). 

```shell
$ pachctl list job
ID                               PIPELINE   STARTED        DURATION   RESTART PROGRESS  DL       UL      STATE
299b4f36535e47e399e7df7fc6ee2f7f regression 23 seconds ago 18 seconds 0       1 + 0 / 1 2.482KiB 1002KiB success
```

### Step 4: Download files once the pipeline has finished
Once the pipeline is completed, we can download the files that were created.

```shell
$ pachctl list file regression@master
NAME               TYPE SIZE
/housing-simplified_corr_matrix.png   file 18.66KiB
/housing-simplified_cv_reg_output.png file 62.19KiB
/housing-simplified_final_model.sav   file 1.007KiB
/housing-simplified_pairplot.png      file 207.5KiB

$ pachctl get file regression@master:/ --recursive --output .
```

When we inspect the learning curve, we can see that there is a large gap between the training score and the validation score. This typically indicates that our model could benefit from the addition of more data. 

<p align="center">
  <img width="400" height="350" src="images/learning_curve-1.png">
</p>

Now let's update our dataset with additional examples.

### Step 5: Update Dataset
Here's where Pachyderm truly starts to shine. To update our dataset we can run the following command (note that we could also append new examples to the existing file, but in this example we're simply overwriting our previous file to one with more data):

```shell
$ pachctl put file housing_data@master:housing-simplified.csv -f data/housing-simplified-2.csv
```

The new commit of data to the `housing_data` repository automatically kicks off a job on the `regression` pipeline without us having to do anything. 

When the job is complete we can download the new files and see that our model has improved, given the new learning curve.
<p align="center">
  <img src="images/cv_reg_output-2.png">
</p>

### Step 6: Inspect the Pipeline Lineage

Note that because versions all of our input and output data automatically, we can continue to iterate on our data and code and Pachyderm will track all of our experiments. For any given output commit, Pachyderm will tell us exactly which input commit of data was run. In our simple example we only have 2 experiments run so far, but this becomes incredibly important and valuable when we do many more iterations.

We can list out the commits to any repository by using the `list commit` commandand.

```shell
$ pachctl list commit housing_data@master
REPO         BRANCH COMMIT                           FINISHED       SIZE     PROGRESS DESCRIPTION
housing_data master a186886de0bf430ebf6fce4d538d4db7 3 minutes ago  12.14KiB ▇▇▇▇▇▇▇▇
housing_data master bbe5ce248aa44522a012f1967295ccdd 23 minutes ago 2.482KiB ▇▇▇▇▇▇▇▇

$ pachctl list commit regression@master
REPO       BRANCH COMMIT                           FINISHED       SIZE     PROGRESS DESCRIPTION
regression master f59a6663073b4e81a2d2ab3b4b7c68fc 2 minutes ago  4.028MiB -
regression master bc0ecea5a2cd43349a9db3e89933fb42 22 minutes ago 1001KiB  -
```

We can show exactly what version of the dataset and pipeline created the model by selecting the commmit ID and using the `inspect` command.

```shell
$ pachctl inspect commit regression@f59a6663073b4e81a2d2ab3b4b7c68fc
Commit: regression@f59a6663073b4e81a2d2ab3b4b7c68fc
Original Branch: master
Parent: bc0ecea5a2cd43349a9db3e89933fb42
Started: 7 minutes ago
Finished: 7 minutes ago
Size: 4.028MiB
Provenance:  __spec__@5b17c425a8d54026a6daaeaf8721707a (regression)  housing_data@a186886de0bf430ebf6fce4d538d4db7 (master)
```

Additionally, can also show the downstream provenance of a commit by using the `flush` command, showing us everything that was run and produced from a commit.

```shell
$ pachctl flush commit housing_data@bbe5ce248aa44522a012f1967295ccdd
REPO       BRANCH COMMIT                           FINISHED       SIZE    PROGRESS DESCRIPTION
regression master bc0ecea5a2cd43349a9db3e89933fb42 31 minutes ago 1001KiB -
```

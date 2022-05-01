# Spout Pipelines - An introductory example

>**COMING SOON: README UPDATE** We have adapted the source code of this example to Pachyderm 2.0.0 [new Map/Reduce Pattern](https://docs.pachyderm.com/2.0.x-beta/concepts/pipeline-concepts/datum/relationship-between-datums/). However, this README needs rewriting. Stay tuned.

## Intro
A spout is a type of pipeline that ingests 
streaming data (message queue, database transactions logs,
event notifications... ), 
acting as **a bridge
between an external stream of data and Pachyderm's repo**.

For those familiar with enterprise integration patterns,
a Pachyderm spout implements the
*[Polling Consumer](https://www.enterpriseintegrationpatterns.com/patterns/messaging/PollingConsumer.html)* 
(subscribes to a stream of data,
reads its published messages, 
then push them to -in our case- the spout's output repository).

For more information about spout pipelines,
we recommend to read the following page in our documentation:

- [Spout](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/pipeline/spout/) concept.
- [Spout](https://docs.pachyderm.com/latest/reference/pipeline_spec/#spout-optional) configuration 


In this example, we have emulated the reception 
of messages from a third-party messaging system
to focus on the specificities of the spout pipeline.

Note that we used [`python-pachyderm`](https://github.com/pachyderm/python-pachyderm)'s Client to connect to Pachyderm's API.
![](./img/spout101.png)
Feel free to explore more of our examples
to discover how we used spout to listen
for new S3 objects notifications via an Amazon™ SQS queue
or connected to an IMAP email account
to analyze the polarity of its emails.

## Getting ready
***Prerequisite***
- A running workspace. Install Pachyderm[locally](https://docs.pachyderm.com/latest/getting-started/local-installation/).
- [pachctl command-line ](https://docs.pachyderm.com/latest/getting-started/local-installation/#install-pachctl) installed, and your context created (i.e., you are logged in)

***Getting started***
- Clone this repo.
- Make sure Pachyderm is running. You should be able to connect to your Pachyderm cluster via the `pachctl` CLI. 
Run a quick:
```shell
$ pachctl version

COMPONENT           VERSION
pachctl             2.0.0
pachd               2.0.0
```
Ideally, have your pachctl and pachd versions match. At a minimum, you should always use the same major & minor versions of your pachctl and pachd. 

## Example - Spout 101 
***Goal***
In this example,
we will keep generating two random strings, 
one of 1KB and one of 2KB,
at intervals varying between 10s and 30s.
Our spout pipeline will actively receive
those events and commit them as text files
to the output repo 
using the **put file** command
of `pachyderm-python`'s library. 
A second pipeline will then process those commits
and log an entry in a separate log file
depending on their size.


1. **Pipeline input repository**: None

1. **Spout and processing pipelines**: [`spout.json`](./pipelines/spout.json) polls and commits to its output repo using `pachctl put file` from the pachyderm-python library.  [`processor.json`](./pipelines/processor.json) then reads the files from its spout input repo and log their content separately depending on their size.

    >![pach_logo](./img/pach_logo.svg) Have a quick look at the source code of our spout pipeline in [`./src/consumer/main.py`](./src/consumer/main.py) and notice that we used `client = python_pachyderm.Client()` to connect to pachd and `client.put_file_bytes` to write files to the spout output repo. 


1. **Pipeline output repository**: `spout` will contain one commit per set of received messages. Each message has been written to a txt file named after its hash (for uniqueness). `processor` will contain two files (1K.txt and 2K.txt) listing the messages received according to their size.


***Example walkthrough***

1.  We have a Docker Hub image of this example ready for you.
    However, you can choose to build your own and push it to your repository.
    
    In the `examples/spouts/spout101` directory,
    make sure to update the 
    `CONTAINER_TAG` in the `Makefile` accordingly
    as well as your pipelines' specifications,
    then run:
    ```shell
    $ make docker-image
    ```
    >![pach_logo](./img/pach_logo.svg) Need a refresher on building, tagging, pushing your image on Docker Hub? Take a look at this [how-to](https://docs.pachyderm.com/latest/how-tos/developer-workflow/working-with-pipelines/).

1. Let's deploy our spout and processing pipelines: 

    Update the `image` field of the `transform` attribute in your pipelines specifications `./pipelines/spout.json` and `./pipelines/processor.json`.

    In the `examples/spouts/spout101` directory, run:
    ```shell
	$ pachctl create pipeline -f ./pipelines/spout.json
	$ pachctl create pipeline -f ./pipelines/processor.json
    ```
    Or, run the following target: 
    ```shell
    $ make deploy
    ```
    Your pipelines  should all be running:
    ![pipelines](./img/pachctl_list_pipeline.png)  

1. Now that the spout pipeline is up, check its output repo once or twice. 
    You should be able to see that new commits are coming in.
    
    ```shell
    $ pachctl list file spout@master
    ```
    ![list_file_spout_master](./img/pachctl_list_file_spout_master.png)

    and some time later...

    ![list_file_spout_master](./img/pachctl_list_file_spout_master_later.png)
    
    Each of those commits triggers a job in the `processing` pipeline:
    ```shell
    $ pachctl list job
    ```

    ![list_job](./img/pachctl_list_job.png)

    and...

    ![list_job](./img/pachctl_list_job_later.png)

1. Take a look at the output repository of your second pipeline: `processor`:
    ```shell
    $ pachctl list file processor@master
    ```

    ![list_file_processor](./img/pachctl_list_file_processor_master.png)

    New entry keep being added to each of the two log files:
    
    ![list_file_processor](./img/pachctl_list_file_processor_master_later.png)

    Zoom into one of them:
    ```shell
    $ pachctl get file processor@master:/1K.txt
    ```
    ![get_file_processor_master_1K](./img/pachctl_get_file_processor_master_1K.png)

    ...

    ![get_file_processor_master_1K](./img/pachctl_get_file_processor_master_1K_later.png)   

    That is it. 

1. When you are done, think about deleting your pipelines.
Remember, a spout pipeline keeps running: 

    In the `examples/spouts/spout101` directory, run:.
    ```shell
    $ pachctl delete all
    ```
    You will be prompted to make sure the delete is intentional. Yes it is. 

     >![pach_logo](./img/pach_logo.svg) Hub users, try `pachctl delete pipeline --all` and `pachctl delete repo --all`.
   
   
    
    A final check at your pipelines: the list should be empty. You are good to go.
    ```shell
    $ pachctl list pipeline
    ```

# Spout Pipelines - An introductory example

>![pach_logo](./img/pach_logo.svg) INFO Pachyderm 2.0 introduces profound architectual changes to the product. As a result, our examples pre and post 2.0 are kept in two separate branches:
> - Branch Master: Examples using Pachyderm 2.0 and later versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13 and older versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples

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
we recommend reading the following pages:

- [Spout](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/pipeline/spout/) concept.
- [Spout](https://docs.pachyderm.com/latest/reference/pipeline_spec/#spout-optional) configuration 


In this example, we have emulated the reception 
of messages from a third-party messaging system
to focus on the specificities of the spout pipeline.

Note that we used [`python-pachyderm`](https://github.com/pachyderm/python-pachyderm)'s Client to connect to Pachyderm's API.
![Spout 101](./img/spout101.png)
Feel free to explore more of our examples
to discover how we used spout to listen
for new S3 objects notifications via an Amazon™ SQS queue
or connected to an IMAP email account
to analyze the polarity of its emails.

## Getting ready
***Prerequisite***
- A workspace on [Pachyderm Hub](https://docs.pachyderm.com/latest/hub/hub_getting_started/) (recommended) or Pachyderm running [locally](https://docs.pachyderm.com/latest/getting_started/local_installation/).
- [pachctl command-line ](https://docs.pachyderm.com/latest/getting_started/local_installation/#install-pachctl) installed, and your context created (i.e., you are logged in)

***Getting started***
- Clone this repo.
- Make sure Pachyderm is running. You should be able to connect to your Pachyderm cluster via the `pachctl` CLI. 
Run a quick:
```shell
$ pachctl version

COMPONENT           VERSION
pachctl             2.0.1
pachd               2.0.1
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


1. **Pipeline input repository**: None. 

1. **Spout and processing pipelines**: 
    - [`spout.json`](./pipelines/spout.json) polls and commits to its output repo using `pachctl put file` from the pachyderm-python library.  

        Then, following the 2 steps pattern described in our [datum processing documentation](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/), we will need 2 pipelines to:
    - read the files from its spout input repo and sort them into 2 directories, depending on their size: [`processor.json`](./pipelines/processor.json) 

    >![pach_logo](./img/pach_logo.svg) Have a quick look at the source code of our spout pipeline in [`./src/consumer/main.py`](./src/consumer/main.py) and notice that we used `client = python_pachyderm.Client()` to connect to pachd and `client.put_file_bytes` to write files to the spout output repo. 


1. **Pipeline output repository**: 
    - `spout` will contain one commit per set of received messages. Each message has been written to a txt file named after its hash (for uniqueness). 
    - `processor` will contain two directories `/1K.txt` and `2K.txt` where the received files will be sorted by size.


***Example walkthrough***

1.  We have a Docker Hub image of this example ready for you.
    However, you can choose to build your own and push it to your organization's registry.
    
    In the `examples/spouts/spout101` directory,
    make sure to update the `CONTAINER_TAG` and `DOCKER_ACCOUNT` in the `Makefile` accordingly,
    then run:

    ```shell
    make docker-image
    ```
    >![pach_logo](./img/pach_logo.svg) Need a refresher on building, tagging, pushing your image on Docker Hub? Take a look at this [how-to](https://docs.pachyderm.com/latest/how-tos/developer-workflow/working-with-pipelines/).

1. Let's deploy our spout and following pipelines: 

    If you have built your own image, update the `image` field of the `transform` attribute in your pipelines' specifications `./pipelines/spout.json` and `./pipelines/processor.json`, otherwise, we have set it up for you.

    In the `examples/spouts/spout101` directory, run:
    ```shell
	pachctl create pipeline -f ./pipelines/spout.json
	pachctl create pipeline -f ./pipelines/processor.json
    ```
    Or, run the following target: 
    ```shell
    make deploy
    ```
    Your 2 pipelines should be running (`pachctl list pipeline`).

1. Now that the spout pipeline is up, check its output repo once or twice. 
    You should be able to see that new commits are coming in.
    
    ```shell
    pachctl list file spout@master
    ```
    ![list_file_spout_master](./img/pachctl_list_file_spout_master.png)

    and some time later...

    ![list_file_spout_master](./img/pachctl_list_file_spout_master_later.png)
    
    Each of those commits triggers a job in the `processing` pipeline.

1. Take a look at the output repository of your second pipeline: `processor`:
    ```shell
    pachctl list file processor@master
    ```
    ```
    NAME TYPE SIZE
    /1K.txt/ dir  672B
    /2K.txt/ dir  672B
    ```
    Notice that new entries keep being added to each of the two directories:
    ```
    NAME TYPE SIZE
    /1K.txt/ file 768B
    /2K.txt/ file 768B
    ```

1. Now, look at the content of those directories:

    Their size keep changing as the logs are coming in.

    Check the content of one of them:
    ```shell
    pachctl list file processor@master:/1K.txt
    ```
    
    ```
    NAME                                                                         TYPE SIZE
    /1K.txt/0b71f006577668cca9c44054166d21c2db0a86bcae742e0f18631f03971a12d7.txt file 96B
    /1K.txt/3f97d1e05c1ce342896a56d44f03d523455bc3df8d8fbbd9cf3c1a71f08df172.txt file 96B
    /1K.txt/73f8f7d74f84e99fd5727d9199b603217fbe8fafa52b6ebdec658639c636fd6f.txt file 96B
    /1K.txt/b9bf82b90b9e1b4a3978af2779f72552601bbd4ba7cfec9d2b69d396eee00602.txt file 96B
    /1K.txt/ba11e93809dca73d23e5c5b6ef158e3336233c85c58857240533a9aa500cb813.txt file 96B
    /1K.txt/bbbae612df0a7d22661a00adc7506db97c72ebd1caf828059365bc8b357bd864.txt file 96B
    /1K.txt/d1b22f0e2c0a3304ba4db38497fd94206bfd18b3e673d89ec17c6f5bee4a0dc1.txt file 96B
    /1K.txt/e46f9f057eb22096dac85f31f11aca86bbe831d7dfa58e5359420598e0d1f905.txt file 96B
    ```




    That is it. You created your first spout pipeline.

1. When you are done, think about deleting your pipelines.
Remember, a spout pipeline keeps running: 

    In the `examples/spouts/spout101` directory, run:.
    ```shell
    pachctl delete pipeline processor
    pachctl delete pipeline spout
    ```
 
    
    A final check at your pipelines: your list should be empty of the 2 pipelines above. 
    ```shell
    pachctl list pipeline
    ```

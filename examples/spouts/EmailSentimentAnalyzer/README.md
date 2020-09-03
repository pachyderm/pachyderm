# Email Sentiment Analysis

## Background

This example connects to an IMAP mail account, 
collects all the incoming mail and analyzes it for positive or negative sentiment,
sorting the emails into directories in its output repo with scoring information added to the email header "X-Sentiment-Rating".

It is inspired by the [email sentiment analysis bot](https://github.com/shanglun/SentimentAnalyzer) documented in [this article](https://www.toptal.com/java/email-sentiment-analysis-bot) by Shanglung Wang, 

It uses [Python-based VADER](https://github.com/cjhutto/vaderSentiment) from CJ Hutto at Georgia Tech.

## Introduction
In this example, we will connect a spout called imap_spout to an email account using IMAP.
That spout's repo will be the input to a pipeline, sentimentalist,  which will score the email's positive, negative, neutral, and compound sentiment, 
adding a header to each with a detailed sentiment score and sorting them into two folders, 
positive and negative, 
in its output repo based on the compound score.

This demo will process emails from an account you configure, moving them from the Inbox to a mailbox called "Processed", 
which it will create if it doesn't exist.
The emails will be scored and then sorted.
You'll see them in the sentimentalist output repo by their unique identifier from the Inbox, 
which ensures they'll be unique.

## Setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster and `kubectl` to talk to Kubernetes.
[Installation instructions can be found here](http://pachyderm.readthedocs.io/en/stable/getting_started/local_installation.html).

1. Create an email account you want to use.  
   Keep the email addrees (which is usually the account name) and the password handy.

1. Enable IMAP on that account. 
   In Gmail, click the gear for "settings" and then click "Forwarding and POP/IMAP" to get to the IMAP settings. 
   In this example, we're assuming you're using Gmail.
   Look in the source code for [./imap_spout.py](imap_spout.py) for environment variables you may need to add to the pipeline spec for the spout to use another email service or other default IMAP folders.

1. The next few steps show you how to add a secret with the following two keys

   * `IMAP_LOGIN`
   * `IMAP_PASSWORD`

   First, we'll save some values to files. 
   The values `<your-password>` and `<account name>` are enclosed in single quotes to prevent the shell from interpreting them.
   
   ```sh
   $ echo -n '<account-name>' > IMAP_LOGIN ; chmod 600 IMAP_LOGIN
   $ echo -n '<your-password>' > IMAP_PASSWORD ; chmod 600 IMAP_PASSWORD
   ```
   
1. Confirm the values in these files are what you expect.

   ```sh
   $ cat IMAP_LOGIN
   $ cat IMAP_PASSWORD
   ```
   
   The output from those two commands should be `<account-name>` and `<your-password>`, respectively.
   
   Creating the secret will require different steps,
   depending on whether you have Kubernetes access or not.
   Pachyderm Hub users don't have access to Kubernetes.
   If you have Kubernetes access, 
   follow the two steps prefixed with "(Kubernetes)".
   If you don't have access to Kubernetes,
   follow the two steps labeled "(Pachyderm Hub)" 

1. (Kubernetes) If you have direct access to the Kubernetes cluster, you can create a secret using `kubectl`.
   
   ```sh
   $ kubectl create secret generic imap-credentials --from-file=./IMAP_LOGIN --from-file=./IMAP_PASSWORD
   ```
   
1. (Kubernetes) Confirm that the secrets got set correctly.
   You use `kubectl get secret` to output the secrets, and then decode them using `jq` to confirm they're correct.
   
   ```sh
   $ kubectl get secret imap-credentials -o json | jq '.data | map_values(@base64d)'
   {
       "IMAP_LOGIN": "<account-name>",
       "IMAP_PASSWORD": "<your-password>"
   }
   ```

   You will have to use pachctl if you're using Pachyderm Hub,
   or don't have access to the Kubernetes cluster.
   The next two steps show how to do that.

1. (Pachyderm Hub) Create a secrets file from the provided template.

   ```sh
   $ jq '.data["IMAP_LOGIN"]="'$(cat IMAP_LOGIN)'"|.data["IMAP_PASSWORD"]="'$(cat IMAP_PASSWORD)'"' imap-credentials-template.json > imap-credentials-secret.json
   $ chmod 600 imap-credentials-secret.json
   ```

1. (Pachyderm Hub) Generate a secret using pachctl

   ```sh
   $ pachctl create secret -f imap-credentials-secret.json
   ```

1. Build the docker image for the imap_spout. 
   Put your own docker account name in for`<docker-account-name>`.
   There is a prebuilt image in the Pachyderm DockerHub registry account, if you want to use it.
   
   ```sh
   $ docker login
   $ docker build -t <docker-account-name>/imap_spout:1.11 -f ./Dockerfile.imap_spout .
   $ docker push <docker-account-name>/imap_spout:1.11
   ```
   
1. Build the docker image for the sentimentalist. 
   Put your own docker account name in for`<docker-account-name>`.
   There is a prebuilt image in the Pachyderm DockerHub registry account, if you want to use it.
   
   ```sh
   $ docker build -t <docker-account-name>/sentimentalist:1.11 -f ./Dockerfile.sentimentalist .
   $ docker push <docker-account-name>/sentimentalist:1.11
   ```
   
1. Edit the pipeline definition files to refer to your own docker repo.
   Put your own docker account name in for `<docker-account-name>`.
   There are prebuilt images for both pipelines in the Pachyderm DockerHub registry account, if you want to use those.
   
   ```sh
   $ sed s/pachyderm/<docker-account-name>/g < sentimentalist.json > my_sentimentalist.json
   $ sed s/pachyderm/<docker-account-name>/g < imap_spout.json > my_imap_spout.json
   ```
   
1. Confirm the pipeline definition files are correct.

1. Create the pipelines

   ```sh
   pachctl create pipeline -f my_imap_spout.json
   pachctl create pipeline -f my_sentimentalist.json
   ```
   
1. Start sending plain-text emails to the account you created. 
   Every few seconds, the imap_spout pipeline will fetch emails from that account via IMAP and send them to its output repo, 
   where the sentimentalist pipeline will score them as positive or negative and sort them into output repos accordingly.
   Have fun! 
   Try tricking the VADER sentiment engine with vague and ironic statements.
   Try emojis!

## Pipelines

### imap_spout

The imap_spout pipeline is an implementation of a [Pachyderm spout](http://docs.pachyderm.com/en/latest/fundamentals/spouts.html) in Python. 
It's configurable with environment variables that can be populated by [Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/).

The spout connects to an IMAP account via SSL, 
creates a "Processed" mailbox for storing already-scored emails, 
and every five seconds checks for new emails.

It then puts each email as a separate file in the spout's output repo.

A couple of things to note, to expand on the [Pachyderm spout](http://docs.pachyderm.com/en/latest/fundamentals/spouts.html) documentation.

1. Look in the source code for [./imap_spout.py](imap_spout.py) for environment variables you may need to add to the pipeline spec for the spout to use another email service or other default IMAP folders.
1. The function `open_pipe` opens `/pfs/out`, 
   the named pipe that's the gateway to the spout's output repo. 
   Note that it must open that pipe as _write only_ and in _binary_ mode. 
   If you omit this, you're likely to see errors like `TypeError: a bytes-like object is required, not 'str'` in your `pachctl logs` for the pipeline.
1. The files are not written directly to the `/pfs/out`; 
   they're written as part of a `tarfile` object.  
   Pachyderm uses the Unix `tar` format to ensure that multiple files can be written to `/pfs/out` and appear correctly in the output repo of your spout.
1. In Python, the `tarfile.open()` command must use the `mode="w|"` argument,
   along with the named pipe's file object,
   to ensure that the `tarfile` object won't try to `seek` on the named pipe `/pfs/out`.
   If you forget this argument, you're likely to to see errors like `file stream is not seekable` in your `pachctl logs` for the pipeline.
1. Every time you `close()` `/pfs/out`, it's a commit.
1. Note that `open_pipe` backs off and attempts to open `/pfs/out` if any errors happen.
   Sometimes it'll take the spout a little bit of time to reopen`/pfs/out` after out code closes it for a commit;
   the backoff is insurance.
1. It saves each email in a file with the `mbox` extension, which is the standard extension for Unix emails. 
   `eml` is also commonly used, but is a slightly different format than what we use here.
   Each `mbox` file contains one email.

### sentimentalist

Sentimentalist is a thin wrapper around the [Python-based VADER](https://github.com/cjhutto/vaderSentiment) from CJ Hutto at Georgia Tech.

It looks in its input repo for individual email files, loads them into a Python email object, and extracts the body and subject as plain text for scoring.  

It uses the "compound" score to sort the emails into different directories, and adds a header to each email with detailed scoring information for use by subsequent pipelines.

## Citations
```
Hutto, C.J. & Gilbert, E.E. (2014). VADER: A Parsimonious Rule-based Model for
Sentiment Analysis of Social Media Text. Eighth International Conference on
Weblogs and Social Media (ICWSM-14). Ann Arbor, MI, June 2014.
```

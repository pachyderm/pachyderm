# Pachyderm Spouts Examples

[Spouts](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/pipeline/spout/)
are a way to get streaming data from any source into Pachyderm. To create a
spout, you need three things

1. A source of streaming data, such as Kafka, nifi, rabbitMQ, etc.
1. A containerized client for the streaming data that will pass it on to the
   spout.
1. A spout pipeline specification file that uses the container.

## Email Sentiment Analysis

[This example](https://github.com/pachyderm/pachyderm/tree/master/examples/spouts/EmailSentimentAnalyzer)
connects to an IMAP mail account, collects all the incoming mail and analyzes it
for positive or negative sentiment, sorting the emails into folders in its
output repo with scoring information added to a header
"X-VADER-Sentiment-Score".

It is inspired by the
[email sentiment analysis bot](https://github.com/shanglun/SentimentAnalyzer)
documented in
[this article](https://www.toptal.com/java/email-sentiment-analysis-bot) by
Shanglung Wang,

It uses [Python-based VADER](https://github.com/cjhutto/vaderSentiment) from CJ
Hutto at Georgia Tech.

```
Hutto, C.J. & Gilbert, E.E. (2014). VADER: A Parsimonious Rule-based Model for
Sentiment Analysis of Social Media Text. Eighth International Conference on
Weblogs and Social Media (ICWSM-14). Ann Arbor, MI, June 2014.
```

## Ingress data from an S3 bucket using SQS

[This example](https://github.com/pachyderm/pachyderm/tree/master/examples/spouts/SQS-S3)
shows how to use spouts to ingress data from an S3 bucket using an SQS queue for
notification of new items added to the bucket.

## Commit messages from a Kafka queue

A
[simple example](https://github.com/pachyderm/pachyderm/tree/master/examples/spouts/go-kafka-spout)
of using spouts with Kafka to process messages and write them to files.

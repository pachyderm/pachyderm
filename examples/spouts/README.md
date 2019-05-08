# Pachyderm Spouts Example

[Spouts](http://docs.pachyderm.com/fundamentals/spouts.md) are a way to get streaming data from any source into Pachyderm.
To create a spout, you need three things

1. A source of streaming data, such as Kafka, nifi, rabbitMQ, etc.
1. A containerized client for the streaming data that will pass it on to the spout.
1. A spout pipeline specification file that uses the container.

## Basic Kafka Spout in Python

[This example](./kafka/) contains a spout, written in python, that consumes messages from kafka and puts them in files in an output repo.

## Email Sentiment Analysis

[This example](./EmailSentimentAnalysis/) connects to an IMAP mail account, 
collects all the incoming mail and analyzes it for positive or negative sentiment,
sorting the emails into repos with scoring information attached.

It is based on the [email sentiment analysis bot](https://github.com/shanglun/SentimentAnalyzer) documented in [this article](https://www.toptal.com/java/email-sentiment-analysis-bot) by Shanglung Wang.


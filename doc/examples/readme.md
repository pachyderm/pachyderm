# Examples

## OpenCV Edge Detection

This example does edge detection using OpenCV. This is our canonical starter demo. If you haven't used Pachyderm before, start here. We'll get you started running Pachyderm locally in just a few minutes and processing sample log lines.

[Open CV](http://pachyderm.readthedocs.io/en/stable/getting_started/beginner_tutorial.html)

## Word Count (Map/Reduce)

Word count is basically the "hello world" of distributed computation. This example is great for benchmarking in distributed deployments on large swaths of text data.

[Word Count](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/word_count)

## Periodic Ingress from a Database

This example pipeline executes a query periodically against a MongoDB database outside of Pachyderm.  The results of the query are stored in a corresponding output repository.  This repository could be used to drive additional pipeline stages periodically based on the results of the query.

[Periodic Ingress from MongoDB](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/cron)

## Machine Learning

### Iris flower classification with R, Python, or Julia

The "hello world" of machine learning implemented in Pachyderm.  You can deploy this pipeline using R, Python, or Julia commponents, where the pipeline includes the trianing of a SVM, LDA, Decision Tree, or Random Forest model and the subsequent utilization of that model to perform inferences.

[R, Python, or Julia - Iris flower classification](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/ml/iris)

### Sentiment analysis with Neon

This example implements the machine learning template pipeline discussed in [this blog post](https://medium.com/pachyderm-data/sustainable-machine-learning-workflows-8c617dd5506d#.hhkbsj1dn).  It trains and utilizes a neural network (implemented in Python using Nervana Neon) to infer the sentiment of movie reviews based on data from IMDB. 

[Neon - Sentiment Analysis](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/ml/neon)

### pix2pix with TensorFlow

If you haven't seen pix2pix, check out [this great demo](https://affinelayer.com/pixsrv/).  In this example, we implement the training and image translation of the pix2pix model in Pachyderm, so you can generate cat images from edge drawings, day time photos from night time photos, etc.

[TensorFlow - pix2pix](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/ml/tensorflow)



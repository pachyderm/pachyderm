# Creating Machine Learning Workflows

Because Pachyderm is language/framework agnostic and because it easily distributes analyses over large data sets, data scientists can use whatever tooling they like for ML. Even if that tooling isn’t familiar to the rest of an engineering organization, data scientists can autonomously develop and deploy scalable solutions via containers. Moreover, Pachyderm’s pipelining logic paired with data versioning, allows any results to be exactly reproduced (e.g., for debugging or during the development of improvements to a model).

More specifically, we recommend combining model training processes, persisted models, and a model utilization processes (e.g., making inferences or generating results) into a single Pachyderm pipeline DAG (Directed Acyclic Graph). Such a pipeline allows us to:

- Keep a rigorous historical record of exactly what models were used on what data to produce which results.
- Automatically update online ML models when training data or parameterization changes.
- Easily revert to other versions of an ML model when a new model is not performing or when “bad data” is introduced into a training data set.

This sort of sustainable ML pipeline looks like this:

![alt tag](ml_workflow.png)

At any time, a data scientist or engineer could update the training dataset utilized by the model to trigger the creation of a newly persisted model in the versioned collection of data (aka, a Pachyderm data “repository”) called “Persisted model.” This training could utilize any language or framework (Spark, Tensorflow, scikit-learn, etc.) and output any format of persisted model (pickle, XML, POJO, etc.).

Regardless of framework, the model will be automatically updated and is versioned by Pachyderm. This process explicitly tracks what data was flowing into which model AND exactly what data was used to train that model.

In addition, when the model is updated, any new input data coming into the “Input data” repository will be processed with the updated model. Old predictions can be re-computed with the updated model, or new models could be backtested on previously input and versioned data. No more manual updates to historical results or worrying about how to swap out ML models in production!

## Examples

We have implemented this machine learning workflows in [some example pipelines](http://docs.pachyderm.io/en/latest/examples/readme.html#machine-learning) using a couple of different frameworks.  These examples are a great starting point if you are trying to implement ML in Pachyderm.  

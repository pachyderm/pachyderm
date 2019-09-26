# Create a Machine Learning Workflow

Because Pachyderm is a language and framework agnostic and
platform, and because it easily distributes analysis over
large data sets, data scientists can use any tooling for
creating machine learning workflows. Even if that tooling
is not familiar to the rest of an engineering organization,
data scientists can autonomously develop and deploy scalable
solutions by using containers. Moreover, Pachyderm’s
pipeline logic paired with data versioning make any results
reproducible for debugging purposes or during the development of
improvements to a model.

For maximum leverage of Pachyderm's built functionality, Pachyderm
recommends that you combine model training processes, persisted models,
and model utilization processes, such as making inferences or
generating results, into a single Pachyderm pipeline Directed Acyclic Graph
(DAG).

Such a pipeline enables you to achieve the following goals:

- Keep a rigorous historical record of which models were used
  on what data to produce which results.
- Automatically update online ML models when training data or
  parameterization changes.
- Easily revert to other versions of an ML model when a new model
  does not produce an expected result or when *bad data* is
  introduced into a training data set.

The following diagram demonstrates an ML pipeline:

![Example of a machine learning workflow](../assets/images/d_ml_workflow.svg)

You can update the training dataset at any time
to automatically train a new persisted model. Also, you can use
any language or framework, including Apache Spark™, Tensorflow™,
scikit-learn™, or other, and output any format of persisted model,
such as pickle, XML, POJO, or other. Regardless of the framework,
Pachyderm versions the model so that you can track the data that
was used to train each model.

Pachyderm processes new data coming into the input repository with the
updated model. Also, you can recompute old predictions with the updated model,
or test new models on previously input and versioned data. This feature
enables you to avoid manual updates to historical results or swapping
ML models in production.

For examples of ML workflows in Pachyderm see
[Machine Learning Examples](../examples/examples.md#machine-learning).

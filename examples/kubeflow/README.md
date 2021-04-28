# Pachyderm Kubeflow Examples

Pachyderm makes production data pipelines repeatable, scalable and provable.
Data scientists and engineers use Pachyderm pipelines to connect data acquisition, cleaning, processing, modeling, and analysis code,
while using Pachyderm's versioned data repositories to keep a complete history of all the data, models, parameters and code
that went into producing each result anywhere in their pipelines. 
This is called data provenance.

If you're currently using Kubeflow to manage your machine learning workloads,
Pachyderm can add value to your Kubeflow deployment in a couple of ways.
You can use Pachyderm's pipelines and containers to call Kubeflow API's to connect and orchestrate Kubeflow jobs.
You can use Pachyderm's versioned data repositories to provide data provenance to the data, models and parameters you use with your Kubeflow code.

This directory contains an example of integrating Pachyderm with Kubeflow.

## mnist with TFJob and Pachyderm

[This example](https://github.com/pachyderm/pachyderm/tree/master/examples/kubeflow/mnist) 
uses the canonical mnist dataset, Kubeflow, TFJobs, and Pachyderm to demonstrate an end-to-end machine learning workflow with data provenance.

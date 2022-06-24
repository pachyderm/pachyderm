>![pach_logo](../img/pach_logo.svg) INFO Each new minor version of Pachyderm introduces profound architectual changes to the product. For this reason, our examples are kept in separate branches:
> - Branch Master: Examples using Pachyderm 2.1.x versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 2.0.x: Examples using Pachyderm 2.0.x versions - https://github.com/pachyderm/pachyderm/tree/2.0.x/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13.x versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples

# Deferred Processing Examples

[Deferring processing](https://docs.pachyderm.com/latest/concepts/advanced-concepts/deferred-processing/) is a technique for controlling when data is processed by Pachyderm.
It allows you to commit data more often than it is processed.


## Deferred Processing Plus Transactions

[This example](./deferred-processing_plus_transactions) uses a simple DAG based on our [OpenCV example](https://github.com/pachyderm/pachyderm/tree/master/examples/opencv)
to illustrate two Pachyderm usage patterns for fine-grained job control.


## Automated Deferred Processing 

[This example](./automated_deferred_processing) can be used to automate the movement of branches when doing deferred processing.
It allows you to trigger a job based on the amount of time that's elapsed since the last commit. 
It has versions that work with authentication activated on your Pachyderm cluster and with authentication disabled.











# Deferred Processing Examples

[Deferring processing](https://docs.pachyderm.com/latest/concepts/advanced-concepts/deferred_processing/) is a technique for controlling when data is processed by Pachyderm.
It allows you to commit data more often than it is processed.


## Deferred Processing Plus Transactions

[This example](./deferred_processing_plus_transactions) uses a simple DAG based on our [OpenCV example](https://github.com/pachyderm/pachyderm/tree/master/examples/opencv)
to illustrate two Pachyderm usage patterns for fine-grained job control.


## Automated Deferred Processing 

[This example](./automated_deferred_processing) can be used to automate the movement of branches when doing deferred processing.
It allows you to trigger a job based on the amount of time that's elapsed since the last commit. 
It has versions that work with authentication activated on your Pachyderm cluster and with authentication disabled.











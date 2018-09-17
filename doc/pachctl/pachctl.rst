Pachctl Command Line Tool
=========================


Pachctl is the command line interface for Pachyderm. To install Pachctl, follow the :doc:`../getting_started/local_installation` instructions

Synopsis
--------

Access the Pachyderm API.

Environment variables:

  ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).


Options
-------

  --no-metrics   Don't report user metrics for this command
  -v, --verbose      Output verbose logs

.. toctree::
    :maxdepth: 2
    :caption: pachctl CLI
    
    pachctl auth 
    pachctl commit
    pachctl completion
    pachctl copy-file
    pachctl create-branch
    pachctl create-pipeline
    pachctl create-repo
    pachctl delete-all
    pachctl delete-branch
    pachctl delete-commit
    pachctl delete-file
    pachctl delete-job
    pachctl delete-pipeline
    pachctl delete-repo
    pachctl deploy
    pachctl diff-file
    pachctl edit-pipeline
    pachctl enterprise
    pachctl extract
    pachctl extract-pipeline
    pachctl file
    pachctl finish-commit
    pachctl flush-commit
    pachctl flush-job
    pachctl garbage-collect
    pachctl get-file
    pachctl get-logs
    pachctl get-object
    pachctl get-tag
    pachctl glob-file
    pachctl inspect-cluster
    pachctl inspect-commit
    pachctl inspect-datum
    pachctl inspect-file
    pachctl inspect-job
    pachctl inspect-pipeline
    pachctl inspect-repo
    pachctl job
    pachctl list-branch
    pachctl list-commit
    pachctl list-datum
    pachctl list-file
    pachctl list-job
    pachctl list-pipeline
    pachctl list-repo
    pachctl mount
    pachctl pipeline
    pachctl port-forward
    pachctl put-file
    pachctl repo
    pachctl restart-datum
    pachctl restore
    pachctl set-branch
    pachctl start-commit
    pachctl start-pipeline
    pachctl stop-job
    pachctl stop-pipeline
    pachctl subscribe-commit 
    pachctl undeploy
    pachctl unmount
    pachctl update-dash
    pachctl update-pipeline
    pachctl update-repo
    pachctl version

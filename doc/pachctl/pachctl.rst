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
    :maxdepth: 1
    :caption: pachctl CLI

    pachctl_copy-file
    pachctl_create-job
    pachctl_create-pipeline
    pachctl_create-repo
    pachctl_delete-all
    pachctl_delete-branch
    pachctl_delete-file
    pachctl_delete-job
    pachctl_delete-pipeline
    pachctl_delete-repo
    pachctl_deploy
    pachctl_finish-commit
    pachctl_flush-commit
    pachctl_garbage-collect
    pachctl_get-file
    pachctl_get-logs
    pachctl_get-object
    pachctl_get-tag
    pachctl_glob-file
    pachctl_inspect-commit
    pachctl_inspect-file
    pachctl_inspect-job
    pachctl_inspect-pipeline
    pachctl_inspect-repo
    pachctl_list-branch
    pachctl_list-commit
    pachctl_list-file
    pachctl_list-job
    pachctl_list-pipeline
    pachctl_list-repo
    pachctl_login
    pachctl_mount
    pachctl_port-forward
    pachctl_put-file
    pachctl_repo
    pachctl_run-pipeline
    pachctl_set-branch
    pachctl_start-commit
    pachctl_start-pipeline
    pachctl_stop-pipeline
    pachctl_undeploy
    pachctl_unmount
    pachctl_update-pipeline
    pachctl_version

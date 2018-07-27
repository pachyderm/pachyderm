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
    
    pachctl_auth 
    pachctl_commit
    pachctl_completion
    pachctl_copy-file
    pachctl_create-branch
    pachctl_create-pipeline
    pachctl_create-repo
    pachctl_delete-all
    pachctl_delete-branch
    pachctl_delete-commit
    pachctl_delete-file
    pachctl_delete-job
    pachctl_delete-pipeline
    pachctl_delete-repo
    pachctl_deploy
    pachctl_diff-file
    pachctl_edit-pipeline
    pachctl_enterprise
    pachctl_extract
    pachctl_extract-pipeline
    pachctl_file
    pachctl_finish-commit
    pachctl_flush-commit
    pachctl_flush-job
    pachctl_garbage-collect
    pachctl_get-file
    pachctl_get-logs
    pachctl_get-object
    pachctl_get-tag
    pachctl_glob-file
    pachctl_inspect-cluster
    pachctl_inspect-commit
    pachctl_inspect-datum
    pachctl_inspect-file
    pachctl_inspect-job
    pachctl_inspect-pipeline
    pachctl_inspect-repo
    pachctl_job
    pachctl_list-branch
    pachctl_list-commit
    pachctl_list-datum
    pachctl_list-file
    pachctl_list-job
    pachctl_list-pipeline
    pachctl_list-repo
    pachctl_mount
    pachctl_pipeline
    pachctl_port-forward
    pachctl_put-file
    pachctl_repo
    pachctl_restart-datum
    pachctl_restore
    pachctl_set-branch
    pachctl_start-commit
    pachctl_start-pipeline
    pachctl_stop-job
    pachctl_stop-pipeline
    pachctl_subscribe-commit 
    pachctl_undeploy
    pachctl_unmount
    pachctl_update-dash
    pachctl_update-pipeline
    pachctl_update-repo
    pachctl_version

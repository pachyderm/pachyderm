# Troubleshooting

This section describe troubleshooting guidelines that should
help you in troubleshooting your deployment and pipelines.

Pachyderm has a built-in logging system that collects
information about events in your Pachyderm environment at
pipeline, datum, and job level. See [pachctl logs](../reference/pachctl/pachctl_logs.md).

To troubleshoot the cluster itself, use the `kubectl` tool
troubleshooting tips. A few basic commands that you can use
include the following:

* Get the list of all Kubernetes objects:

  ```shell
  kubectl get all
  ```

* Get the information about a pod:

  ```shell
  kubectl describe pod <podname>
  ```

The sections below provide troubleshooting steps for specific
issues:

<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">General Troubleshooting &nbsp;&nbsp;&nbsp;<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Start troubleshooting your Pachyderm cluster
        by revieweing general troubleshooting guidelines.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="general-troubleshooting/" class="md-typeset md-link">
          General Troubleshooting
          </a>
          </li>
       </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Troubleshooting Deployments &nbsp;&nbsp;&nbsp;<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Learn how to resolve issues in your Pachyderm
        cluster.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="deploy-troubleshooting/" class="md-typeset md-link">
            Troubleshoot Deployments
          </a>
          </li>
        </ul>
       </div>
     </div>
  </div>
</div>
<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Pipeline Troubleshooting &nbsp;&nbsp;&nbsp;<i class="fa fa-book"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Learn about the troubleshooting steps to debug
        your pipeline.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
           <li><a href="pipeline-troubleshooting/" class="md-typeset md-link">
           Troubleshoot Pipelines
           </a>
           </li>
        </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
  </div>
  </div>
 <div>
<div>

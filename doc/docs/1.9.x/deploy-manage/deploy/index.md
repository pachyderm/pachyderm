# Overview

Pachyderm runs on [Kubernetes](http://kubernetes.io/) and
is backed by an object store of your choice. Because of that,
Pachyderm can run on any platform that supports Kubernetes
and an object store. This section covers common
deployment options and related topics:

<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Deploy for Testing &nbsp;&nbsp;&nbsp;<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Deploy in Pachyderm Hub or on your local
        computer to test basic Pachyderm functionalityi.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="../../pachub/pachub_getting_started/" class="md-typeset md-link">
          Getting Started with Pachyderm Hub
          </a>
          </li>
          <li><a href="../../getting_started/local_installation/" class="md-typeset md-link">
          Deploy Locally
          </a>
          </li>
        </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Production deployment  &nbsp;&nbsp;&nbsp;<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Deploy your production Pachyderm environment in
        one of the supported cloud platforms.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="google_cloud_platform/" class="md-typeset md-link">
          Deploy on GKE
          </a>
          </li>
          <li><a href="amazon_web_services/" class="md-typeset md-link">
          Deploy on AWS
          </a>
          </li>
          <li><a href="azure/" class="md-typeset md-link">
          Deploy on Azure
          </a>
          </li>
          <li><a href="openshift/" class="md-typeset md-link">
          Deploy on OpenShift
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
        <h4 class="mdl-card__title-text">Customize Deployments &nbsp;&nbsp;&nbsp;<i class="fa fa-book"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Learn how to create a customized deployment by
        using various deployment command options.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
           <li><a href="deploy_custom/" class="md-typeset md-link">
           Create a Custom Deployment
           </a>
           </li>
           <li><a href="namespaces/" class="md-typeset md-link">
           Deploy in a Custom Namespace
           </a>
           </li>
           <li><a href="non-cloud-object-stores/" class="md-typeset md-link">
           Deploy On-Premises With Non-Cloud Object Stores
           </a>
           </li>
           <li><a href="rbac/" class="md-typeset md-link">
           Configure RBAC
           </a>
           </li>
        </ul>
      </div>
    </div>
  </div>
<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Post-Deployment &nbsp;&nbsp;&nbsp;<i class="fa fa-flask"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Perform post-deployment tasks.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
           <li><a href="connect-to-cluster/" class="md-typeset md-link">
           Connect to a Pachyderm cluster
           </a>
           </li>
        </ul>
      </div>
    </div>
  </div>
</div>

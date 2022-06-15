# Overview

Pachyderm runs on [Kubernetes](https://kubernetes.io/){target=_blank},
is backed by an object store of your choice, and comes with a bundled version of [PostgreSQL](https://www.postgresql.org/){target=_blank} (metadata storage) by default. 
We recommended disabling the bundled PostgreSQL and using a **managed database instance** (such as RDS, CloudSQL, or PostgreSQL Server) for production environments.

This section covers common
deployment options and related topics:

<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Quick Start &nbsp;&nbsp;&nbsp;<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        To get started, install Pachyderm locally, or use our quickstart deployment instructions on the Cloud.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="../../getting-started/local-installation/" class="md-typeset md-link">
          Deploy Locally
          </a>
          </li>
          <li><a href="./quickstart/" class="md-typeset md-link">
          Quick Cloud Deployment
          </a>
          </li>
          <li><a href="../../getting-started/install-pachctl-completion/" class="md-typeset md-link">
          Install pachctl Autocompletion
          </a>
          </li>
          <li><a href="../manage/pachctl-shell/" class="md-typeset md-link">
          Use Pachyderm Shell - Our Auto Suggest Tool
          </a>
          </li>         
        </ul>
      </div>
    </div>
  </div>
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Production Deployments  &nbsp;&nbsp;&nbsp;<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Deploy Pachyderm in production on
        one of the supported cloud platforms.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./ingress/" class="md-typeset md-link">
          Architecture, Ingress, and LB
          </a>
          </li>
          <li><a href="google-cloud-platform/" class="md-typeset md-link">
          Deploy on GKE
          </a>
          </li>
          <li><a href="aws-deploy-pachyderm/" class="md-typeset md-link">
          Deploy on AWS
          </a>
          </li>
          <li><a href="azure/" class="md-typeset md-link">
          Deploy on Azure
          </a>
          </li>
          <li><a href="on-premises/" class="md-typeset md-link">
          Deploy On Premises
          </a>
          </li>
          <li><a href="console/" class="md-typeset md-link">
          Deploy Console
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
        <h4 class="mdl-card__title-text">Additional Customizations &nbsp;&nbsp;&nbsp;<i class="fa fa-book"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Customized deployment options.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
           <li><a href="import-kubernetes-context/" class="md-typeset md-link">
           Import a Kubernetes Context
           </a>
           </li>
           <li><a href="deploy-w-tls/" class="md-typeset md-link">
           Deploy Pachyderm with TLS
           </a>
           </li>
           <li><a href="namespaces/" class="md-typeset md-link">
           Deploy in a Custom Namespace
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
           <li><a href="tracing/" class="md-typeset md-link">
           Configure Tracing with Jaeger 
           </a>
           </li>
           <li><a href="loki/" class="md-typeset md-link">
           Enable logs aggregation with Loki
           </a>
           </li>
           <li><a href="prometheus/" class="md-typeset md-link">
           Monitor cluster metrics with Prometheus
           </a>
           </li>
           <li><a href="environment-variables/" class="md-typeset md-link">
           Configure Environment Variables
           </a>
           </li>
        </ul>
      </div>
    </div>
  </div>
</div>

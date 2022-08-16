# Environment Variables

You can set required configuration directly to your application using environment variables. 

## Variable Types

<div class="row">
  <div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Pachd Env Variables<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Define parameters for your Pachyderm daemon container.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./pachd" class="md-typeset md-link">
          Read this article
          </a>
          </li>        
        </ul>
      </div>
    </div>
  </div>
<div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Pipeline Worker Env Variables<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Define parameters for your Pachyderm pipeline workers.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./workers" class="md-typeset md-link">
          Read this article
          </a>
          </li>
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
        <h4 class="mdl-card__title-text">Gobal Env Variables<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Define global Pachyderm parameters.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./global" class="md-typeset md-link">
          Read this article
          </a>
          </li>        
        </ul>
      </div>
    </div>
  </div>
<div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">Enterprise Env Variables<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Define parameters specific to Enterprise Pachyderm.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./enterprise" class="md-typeset md-link">
          Read this article
          </a>
          </li>
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
        <h4 class="mdl-card__title-text">Storage Env Variables<i class="fa fa-rocket"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Define storage-specific parameters.
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./storage" class="md-typeset md-link">
          Read this article
          </a>
          </li>        
        </ul>
      </div>
    </div>
  </div>
<div class="column-2">
    <div class="card-square mdl-card mdl-shadow--2dp">
      <div class="mdl-card__title mdl-card--expand">
        <h4 class="mdl-card__title-text">OS Env Variables<i class="fa fa-cogs"></i></h4>
      </div>
      <div class="mdl-card__supporting-text">
        Define OS-level parameters. 
      </div>
      <div class="mdl-card__actions mdl-card--border">
        <ul>
          <li><a href="./os" class="md-typeset md-link">
          Read this article
          </a>
          </li>
          </li>
        </ul>
       </div>
     </div>
  </div>
</div>


### Additional Variables


Kubernetes injects additional variables for services running inside a cluster. While powerful, using them can result in multiple processing retries. Interaction with outside services must be idempotent to prevent
unexpected behavior.

!!! note "See Also"
    - [transform.env](../../../reference/pipeline-spec/#transform-required)

import {
  CronInput,
  Egress,
  Input,
  PFSInput,
  Pipeline,
  PipelineInfo,
  PipelineState,
  SchedulingSpec,
  Transform,
  JobState,
  ParallelismSpec,
  Service,
  Spout,
} from '@dash-backend/proto';
import {
  ObjectStorageEgress,
  Project,
  SQLDatabaseEgress,
} from '@dash-backend/proto/proto/pfs/pfs_pb';

import {DAGS} from './loadLimits';

// Need to define this up here, as the node selector
// map is a mutable set that can't be initialized with
// values
const schedulingSpec = new SchedulingSpec();
schedulingSpec.setPriorityClassName('high-priority');
schedulingSpec.getNodeSelectorMap().set('disktype', 'ssd');

// clone of customerTeam and tutorial
const defaultPipelines = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('likelihoods')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('samples')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('reference')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('models')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('training')
              .setProject('Data-Cleaning-Process'),
          ),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('joint_call')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_KILLED)
    .setState(PipelineState.PIPELINE_FAILURE)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('reference')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('likelihoods')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('split')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('raw_data')
              .setProject('Data-Cleaning-Process'),
          ),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('model')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_PAUSED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('split')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('parameters_pachyderm_version_alternate_replicant')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('test')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('split')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('select')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('test')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('detect_pachyderm_repo_version_alternate')
        .setProject(new Project().setName('default')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('images')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('montage')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('edges')
                .setProject('Solar-Panel-Data-Sorting'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('images')
                .setProject('Solar-Panel-Data-Sorting'),
            ),
          ]),
        )
        .setDescription('Not my favorite pipeline')
        .setOutputBranch('master')
        .setEgress(new Egress().setUrl('https://egress.com'))
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setReason(
      'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('edges')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('images')
              .setProject('Solar-Panel-Data-Sorting'),
          ),
        )
        .setDescription('Very cool edges description')
        .setOutputBranch('master')
        .setTransform(
          new Transform()
            .setCmdList(['python3', './edges.py'])
            .setImage('pachyderm/opencv'),
        ),
    )
    .setState(PipelineState.PIPELINE_RUNNING),
];

// clone of customerTeam with a different project name
const solarPricePredictionModal = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('likelihoods')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('samples')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('reference')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('models')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('training')
              .setProject('Solar-Price-Prediction-Modal'),
          ),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('joint_call')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_KILLED)
    .setState(PipelineState.PIPELINE_FAILURE)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('reference')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('likelihoods')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('split')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('raw_data')
              .setProject('Solar-Price-Prediction-Modal'),
          ),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('model')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_PAUSED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('split')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('parameters_pachyderm_version_alternate_replicant')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('test')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('split')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('select')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('test')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('detect_pachyderm_repo_version_alternate')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('images')
                .setProject('Solar-Price-Prediction-Modal'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),
];

const tutorial = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('montage')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('edges')
                .setProject('Solar-Panel-Data-Sorting'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('images')
                .setProject('Solar-Panel-Data-Sorting'),
            ),
          ]),
        )
        .setDescription('Not my favorite pipeline')
        .setOutputBranch('master')
        .setEgress(new Egress().setUrl('https://egress.com'))
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setReason(
      'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('edges')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('images')
              .setProject('Solar-Panel-Data-Sorting'),
          ),
        )
        .setDescription('Very cool edges description')
        .setOutputBranch('master')
        .setTransform(
          new Transform()
            .setCmdList(['python3', './edges.py'])
            .setImage('pachyderm/opencv'),
        ),
    )
    .setState(PipelineState.PIPELINE_RUNNING),
];

const egress = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('egress_s3')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput().setRepo('edges').setProject('Egress-Examples'),
            ),
            new Input().setPfs(
              new PFSInput().setRepo('images').setProject('Egress-Examples'),
            ),
          ]),
        )
        .setDescription('a pipeline with egress to an s3 bucket')
        .setOutputBranch('master')
        .setEgress(new Egress().setUrl('https://egress.com'))
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('egress_sql')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput().setRepo('edges').setProject('Egress-Examples'),
            ),
            new Input().setPfs(
              new PFSInput().setRepo('images').setProject('Egress-Examples'),
            ),
          ]),
        )
        .setDescription('a pipeline with egress to an sql database')
        .setOutputBranch('master')
        .setEgress(
          new Egress().setSqlDatabase(
            new SQLDatabaseEgress()
              .setUrl(
                'snowflake://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
              )
              .setFileFormat(
                new SQLDatabaseEgress.FileFormat().setType(
                  SQLDatabaseEgress.FileFormat.Type.CSV,
                ),
              ),
          ),
        )
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('egress_object')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput().setRepo('edges').setProject('Egress-Examples'),
            ),
            new Input().setPfs(
              new PFSInput().setRepo('images').setProject('Egress-Examples'),
            ),
          ]),
        )
        .setDescription('a pipeline with egress to object storage')
        .setOutputBranch('master')
        .setEgress(
          new Egress().setObjectStorage(
            new ObjectStorageEgress().setUrl(
              'object://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
            ),
          ),
        )
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('edges')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput().setRepo('images').setProject('Egress-Examples'),
          ),
        )
        .setDescription('Very cool edges description')
        .setOutputBranch('master')
        .setTransform(
          new Transform()
            .setCmdList(['python3', './edges.py'])
            .setImage('pachyderm/opencv'),
        ),
    )
    .setState(PipelineState.PIPELINE_RUNNING),
];

const customerTeam = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('likelihoods')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('samples')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('reference')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('models')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('training')
              .setProject('Data-Cleaning-Process'),
          ),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('joint_call')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_KILLED)
    .setState(PipelineState.PIPELINE_FAILURE)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('reference')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('likelihoods')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('split')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('raw_data')
              .setProject('Data-Cleaning-Process'),
          ),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('model')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_PAUSED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('split')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('parameters_pachyderm_version_alternate_replicant')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('test')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('split')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('select')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('test')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('detect_pachyderm_repo_version_alternate')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(
              new PFSInput()
                .setRepo('model')
                .setProject('Data-Cleaning-Process'),
            ),
            new Input().setPfs(
              new PFSInput()
                .setRepo('images')
                .setProject('Data-Cleaning-Process'),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),
];

const cron = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('processor')
        .setProject(
          new Project().setName('Solar-Power-Data-Logger-Team-Collab'),
        ),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCron(
            new CronInput()
              .setRepo('cron')
              .setProject('Solar-Power-Data-Logger-Team-Collab'),
          ),
        )
        .setOutputBranch('master'),
    ),
];

const traitDiscovery = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_orfs_blastdb')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('orfs').setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_refseqfasta')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('reference_sequences')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_search')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_orfs_blastdb')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_refseqfasta')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_candidates')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_search')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_atg_fasta')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('atgs').setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_completeness')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_candidates')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_candidate_fasta')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_candidates')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_group_candidate_bam')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('assembly_bam_files')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidates')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_clustering')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_atg_fasta')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidate_fasta')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_quality_downselect')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_group_candidate_bam')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_group_contig_candidates')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidates')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_atg_fasta')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_quality')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_quality_downselect')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_neighbors')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_group_contig_candidates')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_domainscan')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidate_fasta')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('inter_pro_scan')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_quality_check')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_quality')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_hmmscan')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidate_fasta')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput().setRepo('custom_hmms').setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_promotion_status')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_clustering')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_group_geneclass_data')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_domainscan')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_hmmscan')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_patent_search')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidate_fasta')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('patent_databases')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_geneclass')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_group_geneclass_data')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_patent_check')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_patent_search')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_group_promo_data')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_quality_check')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_completeness')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_neighbors')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_candidates')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_geneclass')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_patent_check'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_promotionfilter')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_group_promo_data')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_group_promo_clstr')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_promotionfilter')
              .setProject('Trait-Discovery'),
          ),
          new Input().setPfs(
            new PFSInput()
              .setRepo('pachy_trait_promotion_status')
              .setProject('Trait-Discovery'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_promoclstr_filter')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_group_promo_clstr')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_promotion')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_promoclstr_filter')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('pachy_trait_atgs')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput()
            .setRepo('pachy_trait_promotion')
            .setProject('Trait-Discovery'),
        ),
      ),
    ),
];

const multiProjectPipelineA = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('Node_2')
        .setProject(new Project().setName('Multi-Project-Pipeline-A')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('Node_1')
              .setProject('Multi-Project-Pipeline-B'),
          ),
        )
        .setOutputBranch('master'),
    ),
];
const multiProjectPipelineB = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('Node_2')
        .setProject(new Project().setName('Multi-Project-Pipeline-B')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('Node_1')
              .setProject('Multi-Project-Pipeline-A'),
          ),
        )
        .setOutputBranch('master'),
    ),
];

const pipelinesProject = [
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('service-pipeline')
        .setProject(new Project().setName('Pipelines-Project')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setPfs(
            new PFSInput()
              .setRepo('service-pipeline-input')
              .setProject('Pipelines-Project'),
          ),
        )
        .setOutputBranch('master')
        .setService(new Service().setExternalPort(30000)),
    ),
  new PipelineInfo()
    .setPipeline(
      new Pipeline()
        .setName('spout-pipeline')
        .setProject(new Project().setName('Pipelines-Project')),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setOutputBranch('master')
        .setSpout(new Spout()),
    ),
];

const getLoadPipelines = (count: number) => {
  return [...new Array(count).keys()].map((i) => {
    return new PipelineInfo()
      .setPipeline(
        new Pipeline()
          .setName(`load-pipeline-${i}`)
          .setProject(new Project().setName('Load-Project')),
      )
      .setDetails(
        new PipelineInfo.Details().setInput(
          new Input().setPfs(new PFSInput().setRepo(`load-repo-${i}`)),
        ),
      );
  });
};

const pipelines: {[projectId: string]: PipelineInfo[]} = {
  'Solar-Panel-Data-Sorting': tutorial,
  'Data-Cleaning-Process': customerTeam,
  'Solar-Power-Data-Logger-Team-Collab': cron,
  'Solar-Price-Prediction-Modal': solarPricePredictionModal,
  'Egress-Examples': egress,
  'Empty-Project': [],
  'Trait-Discovery': traitDiscovery,
  'OpenCV-Tutorial': [],
  'Load-Project': getLoadPipelines(DAGS),
  default: defaultPipelines,
  'Multi-Project-Pipeline-A': multiProjectPipelineA,
  'Multi-Project-Pipeline-B': multiProjectPipelineB,
  'Pipelines-Project': pipelinesProject,
};

export default pipelines;

import {
  CronInput,
  Input,
  PFSInput,
  Pipeline,
  PipelineInfo,
  PipelineState,
} from '@pachyderm/proto/pb/pps/pps_pb';

const tutorial = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('montage'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('edges')),
        new Input().setPfs(new PFSInput().setRepo('images')),
      ]),
    )
    .setDescription('Not my favorite pipeline')
    .setState(PipelineState.PIPELINE_FAILURE)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('edges'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('images')))
    .setDescription('Very cool edges description')
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),
];

const customerTeam = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('likelihoods'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('samples')),
        new Input().setPfs(new PFSInput().setRepo('reference')),
      ]),
    )
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('models'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('training')))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('joint_call'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('reference')),
        new Input().setPfs(new PFSInput().setRepo('likelihoods')),
      ]),
    )
    .setState(PipelineState.PIPELINE_FAILURE)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('split'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('raw_data')))
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('model'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('split')),
        new Input().setPfs(new PFSInput().setRepo('parameters')),
      ]),
    )
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('test'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('split')),
        new Input().setPfs(new PFSInput().setRepo('model')),
      ]),
    )
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('select'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('test')),
        new Input().setPfs(new PFSInput().setRepo('model')),
      ]),
    )
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('detect'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('model')),
        new Input().setPfs(new PFSInput().setRepo('images')),
      ]),
    )
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),
];

const cron = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('processor'))
    .setInput(new Input().setCron(new CronInput().setRepo('cron')))
    .setOutputBranch('master')
    .setCacheSize('64M'),
];

const pipelines: {[projectId: string]: PipelineInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': tutorial,
  default: [...tutorial, ...customerTeam],
};

export default pipelines;

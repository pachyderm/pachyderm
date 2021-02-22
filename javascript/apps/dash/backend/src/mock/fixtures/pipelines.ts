import {
  Input,
  PFSInput,
  Pipeline,
  PipelineInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

const pipelines: {[projectId: string]: PipelineInfo[]} = {
  tutorial: [
    new PipelineInfo()
      .setPipeline(new Pipeline().setName('montage'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('edges')),
          new Input().setPfs(new PFSInput().setRepo('images')),
        ]),
      ),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('edges'))
      .setInput(new Input().setPfs(new PFSInput().setRepo('images'))),
  ],
  customerTeam: [
    new PipelineInfo()
      .setPipeline(new Pipeline().setName('likelihoods'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('samples')),
          new Input().setPfs(new PFSInput().setRepo('reference')),
        ]),
      ),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('models'))
      .setInput(new Input().setPfs(new PFSInput().setRepo('training'))),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('joint_call'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('reference')),
          new Input().setPfs(new PFSInput().setRepo('likelihoods')),
        ]),
      ),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('split'))
      .setInput(new Input().setPfs(new PFSInput().setRepo('raw_data'))),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('model'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('split')),
          new Input().setPfs(new PFSInput().setRepo('parameters')),
        ]),
      ),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('test'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('split')),
          new Input().setPfs(new PFSInput().setRepo('model')),
        ]),
      ),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('select'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('test')),
          new Input().setPfs(new PFSInput().setRepo('model')),
        ]),
      ),

    new PipelineInfo()
      .setPipeline(new Pipeline().setName('detect'))
      .setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('model')),
          new Input().setPfs(new PFSInput().setRepo('images')),
        ]),
      ),
  ],
};

export default pipelines;

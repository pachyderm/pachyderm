import {
  Input,
  PFSInput,
  Pipeline,
  PipelineInfo,
} from '@pachyderm/proto/pb/client/pps/pps_pb';

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
};

export default pipelines;

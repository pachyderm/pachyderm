import {IAPIServer} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {PipelineInfos} from '@pachyderm/proto/pb/pps/pps_pb';

import pipelines from 'mock/fixtures/pipelines';

const pps: Pick<IAPIServer, 'listPipeline'> = {
  listPipeline: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');

    const reply = new PipelineInfos();

    // "tutorial" in this case represents the default/catch-all project in core pach
    reply.setPipelineInfoList(
      projectId ? pipelines[projectId.toString()] : pipelines['1'],
    );

    callback(null, reply);
  },
};

export default pps;

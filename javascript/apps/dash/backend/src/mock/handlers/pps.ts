import {Status} from '@grpc/grpc-js/build/src/constants';
import {IAPIServer} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {PipelineInfos} from '@pachyderm/proto/pb/pps/pps_pb';

import pipelines from '@dash-backend/mock/fixtures/pipelines';

const pps: Pick<IAPIServer, 'listPipeline'> = {
  listPipeline: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const [authToken] = call.metadata.get('authn-token');

    if (authToken && authToken === 'expired') {
      callback(
        {
          code: Status.UNAUTHENTICATED,
          details:
            'provided auth token is corrupted or has expired (try logging in again)',
        },
        null,
      );
    }

    const reply = new PipelineInfos();

    // "tutorial" in this case represents the default/catch-all project in core pach
    reply.setPipelineInfoList(
      projectId ? pipelines[projectId.toString()] : pipelines['1'],
    );

    callback(null, reply);
  },
};

export default pps;

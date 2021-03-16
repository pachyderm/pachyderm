import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListPipelineRequest,
  PipelineInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

import {ServiceArgs} from '@dash-backend/lib/types';

const pps = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
  log,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listPipeline: () => {
      log.info('listPipeline request');

      return new Promise<PipelineInfo.AsObject[]>((resolve, reject) => {
        client.listPipeline(
          new ListPipelineRequest(),
          credentialMetadata,
          (error, res) => {
            if (error) {
              log.error({error: error.message}, 'listPipeline request failed');
              return reject(error);
            }

            log.info('listPipeline request succeeded');
            return resolve(res.toObject().pipelineInfoList);
          },
        );
      });
    },
  };
};

export default pps;

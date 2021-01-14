import {APIClient} from '@pachyderm/proto/pb/client/pps/pps_grpc_pb';
import {
  ListPipelineRequest,
  PipelineInfo,
} from '@pachyderm/proto/pb/client/pps/pps_pb';

import createCredentials from './createCredentials';

const pps = (address: string, authToken: string) => {
  const client = new APIClient(address, createCredentials(authToken));

  return {
    listPipeline: () => {
      return new Promise<PipelineInfo.AsObject[]>((resolve, reject) => {
        client.listPipeline(new ListPipelineRequest(), (err, res) => {
          if (err) return reject(err);

          return resolve(res.toObject().pipelineInfoList);
        });
      });
    },
  };
};

export default pps;

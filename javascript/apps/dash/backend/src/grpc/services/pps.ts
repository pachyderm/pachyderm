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
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listPipeline: () => {
      return new Promise<PipelineInfo.AsObject[]>((resolve, reject) => {
        client.listPipeline(
          new ListPipelineRequest(),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }

            return resolve(res.toObject().pipelineInfoList);
          },
        );
      });
    },
  };
};

export default pps;

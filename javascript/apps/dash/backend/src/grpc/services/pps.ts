import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListPipelineRequest,
  PipelineInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

const pps = (
  pachdAddress: string,
  channelCredentials: ChannelCredentials,
  credentialMetadata: Metadata,
) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listPipeline: () => {
      return new Promise<PipelineInfo.AsObject[]>((resolve, reject) => {
        client.listPipeline(
          new ListPipelineRequest(),
          credentialMetadata,
          (err, res) => {
            if (err) return reject(err);

            return resolve(res.toObject().pipelineInfoList);
          },
        );
      });
    },
  };
};

export default pps;

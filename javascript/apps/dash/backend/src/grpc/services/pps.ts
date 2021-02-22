import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListPipelineRequest,
  PipelineInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

const pps = (pachdAddress: string, channelCredentials: ChannelCredentials) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listPipeline: (projectId = '') => {
      return new Promise<PipelineInfo.AsObject[]>((resolve, reject) => {
        const metadata = new Metadata();
        metadata.set('project-id', projectId);

        client.listPipeline(new ListPipelineRequest(), metadata, (err, res) => {
          if (err) return reject(err);

          return resolve(res.toObject().pipelineInfoList);
        });
      });
    },
  };
};

export default pps;

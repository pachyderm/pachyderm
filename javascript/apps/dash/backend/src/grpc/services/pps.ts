import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListJobRequest,
  ListPipelineRequest,
  PipelineInfo,
  JobInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

import {ServiceArgs} from '@dash-backend/lib/types';

const pps = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listPipeline: (jq = '') => {
      return new Promise<PipelineInfo.AsObject[]>((resolve, reject) => {
        client.listPipeline(
          new ListPipelineRequest().setJqfilter(jq),
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
    listJobs: () => {
      const listJobRequest = new ListJobRequest();
      const stream = client.listJob(listJobRequest, credentialMetadata);

      return new Promise<JobInfo.AsObject[]>((resolve, reject) => {
        const jobs: JobInfo.AsObject[] = [];

        stream.on('data', (chunk: JobInfo) => {
          jobs.push(chunk.toObject());
        });
        stream.on('error', (err) => reject(err));
        stream.on('end', () => resolve(jobs));
      });
    },
  };
};

export default pps;

import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListPipelineJobRequest,
  ListPipelineRequest,
  PipelineInfo,
  PipelineJobInfo,
  InspectPipelineJobRequest,
  InspectPipelineRequest,
} from '@pachyderm/proto/pb/pps/pps_pb';

import {ServiceArgs} from '@dash-backend/lib/types';

import {pipelineJobFromObject, pipelineFromObject} from '../builders/pps';

import {DEFAULT_JOBS_LIMIT} from './constants/pps';

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
    // TODO: update name
    listJobs: ({limit = DEFAULT_JOBS_LIMIT, jq = ''} = {}) => {
      const listJobRequest = new ListPipelineJobRequest().setJqfilter(jq);
      const stream = client.listPipelineJob(listJobRequest, credentialMetadata);

      return new Promise<PipelineJobInfo.AsObject[]>((resolve, reject) => {
        const jobs: PipelineJobInfo.AsObject[] = [];

        stream.on('data', (chunk: PipelineJobInfo) => {
          jobs.push(chunk.toObject());
          if (limit && jobs.length >= limit) {
            stream.destroy();
          }
        });
        stream.on('error', (err) => reject(err));
        stream.on('close', () => resolve(jobs));
        stream.on('end', () => resolve(jobs));
      });
    },

    // TODO: update name
    inspectJob: (id: string) => {
      return new Promise<PipelineJobInfo.AsObject>((resolve, reject) => {
        client.inspectPipelineJob(
          new InspectPipelineJobRequest().setPipelineJob(
            pipelineJobFromObject({id}),
          ),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          },
        );
      });
    },

    inspectPipeline: (id: string) => {
      return new Promise<PipelineInfo.AsObject>((resolve, reject) => {
        client.inspectPipeline(
          new InspectPipelineRequest().setPipeline(
            pipelineFromObject({name: id}),
          ),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          },
        );
      });
    },
  };
};

export default pps;

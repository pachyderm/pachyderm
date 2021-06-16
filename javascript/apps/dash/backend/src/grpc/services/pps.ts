import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListJobRequest,
  ListPipelineRequest,
  PipelineInfo,
  JobInfo,
  InspectJobRequest,
  InspectPipelineRequest,
} from '@pachyderm/proto/pb/pps/pps_pb';

import streamToObjectArray from '@dash-backend/grpc/utils/streamToObjectArray';
import {ServiceArgs} from '@dash-backend/lib/types';

import {jobFromObject, pipelineFromObject} from '../builders/pps';

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

    listJobs: ({limit = DEFAULT_JOBS_LIMIT, jq = ''} = {}) => {
      const listJobRequest = new ListJobRequest().setJqfilter(jq);
      const stream = client.listJob(listJobRequest, credentialMetadata);

      return streamToObjectArray<JobInfo, JobInfo.AsObject>(stream, limit);
    },

    inspectJob: (id: string) => {
      return new Promise<JobInfo.AsObject>((resolve, reject) => {
        client.inspectJob(
          new InspectJobRequest().setJob(jobFromObject({id})),
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

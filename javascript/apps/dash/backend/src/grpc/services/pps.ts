import {ClientReadableStream} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListJobRequest,
  ListPipelineRequest,
  PipelineInfo,
  JobInfo,
  InspectJobRequest,
  InspectPipelineRequest,
  Jobset,
  InspectJobsetRequest,
  LogMessage,
} from '@pachyderm/proto/pb/pps/pps_pb';

import streamToObjectArray from '@dash-backend/grpc/utils/streamToObjectArray';
import {ServiceArgs} from '@dash-backend/lib/types';
import {JobsetQueryArgs, JobQueryArgs} from '@graphqlTypes';

import {
  jobFromObject,
  pipelineFromObject,
  GetLogsRequestObject,
  getLogsRequestFromObject,
} from '../builders/pps';

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

    inspectJobset: ({id}: JobsetQueryArgs) => {
      const inspectJobsetRequest = new InspectJobsetRequest()
        .setWait(false)
        .setJobset(new Jobset().setId(id));

      const stream = client.inspectJobset(
        inspectJobsetRequest,
        credentialMetadata,
      );

      return streamToObjectArray<JobInfo, JobInfo.AsObject>(stream);
    },

    inspectJob: ({id, pipelineName}: JobQueryArgs) => {
      return new Promise<JobInfo.AsObject>((resolve, reject) => {
        client.inspectJob(
          new InspectJobRequest().setJob(
            jobFromObject({id}).setPipeline(
              pipelineFromObject({name: pipelineName}),
            ),
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

    getLogs: (request: GetLogsRequestObject) => {
      const getLogsRequest = getLogsRequestFromObject(request);
      const stream = client.getLogs(getLogsRequest, credentialMetadata);

      return streamToObjectArray<LogMessage, LogMessage.AsObject>(stream);
    },

    getLogsStream: (request: GetLogsRequestObject) => {
      return new Promise<ClientReadableStream<LogMessage>>(
        (resolve, reject) => {
          try {
            const getLogsRequest = getLogsRequestFromObject(request);
            const stream = client.getLogs(getLogsRequest, credentialMetadata);

            return resolve(stream);
          } catch (error) {
            return reject(error);
          }
        },
      );
    },
  };
};

export default pps;

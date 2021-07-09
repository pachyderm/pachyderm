import {ClientReadableStream} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {
  ListJobRequest,
  ListPipelineRequest,
  PipelineInfo,
  JobInfo,
  InspectJobRequest,
  InspectPipelineRequest,
  JobSet,
  InspectJobSetRequest,
  LogMessage,
  Pipeline,
  ListJobSetRequest,
  JobSetInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

import {
  jobFromObject,
  pipelineFromObject,
  GetLogsRequestObject,
  getLogsRequestFromObject,
} from '../builders/pps';
import {JobSetQueryArgs, JobQueryArgs, ServiceArgs} from '../lib/types';
import {DEFAULT_JOBS_LIMIT} from '../services/constants/pps';
import streamToObjectArray from '../utils/streamToObjectArray';

interface ListArgs {
  limit?: number | null;
}
interface ListJobArgs extends ListArgs {
  pipelineId?: string | null;
}

const pps = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    listPipeline: (jq = '') => {
      const listPipelineRequest = new ListPipelineRequest()
        .setJqfilter(jq)
        .setDetails(true);
      const stream = client.listPipeline(
        listPipelineRequest,
        credentialMetadata,
      );

      return streamToObjectArray<PipelineInfo, PipelineInfo.AsObject>(stream);
    },

    listJobs: ({limit, pipelineId}: ListJobArgs = {}) => {
      const listJobRequest = new ListJobRequest().setDetails(true);

      if (pipelineId) {
        listJobRequest.setPipeline(new Pipeline().setName(pipelineId));
      }

      const stream = client.listJob(listJobRequest, credentialMetadata);

      return streamToObjectArray<JobInfo, JobInfo.AsObject>(
        stream,
        limit || DEFAULT_JOBS_LIMIT,
      );
    },

    inspectPipeline: (id: string) => {
      return new Promise<PipelineInfo.AsObject>((resolve, reject) => {
        client.inspectPipeline(
          new InspectPipelineRequest()
            .setPipeline(pipelineFromObject({name: id}))
            .setDetails(true),
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

    inspectJobSet: ({id}: JobSetQueryArgs) => {
      const inspectJobSetRequest = new InspectJobSetRequest()
        .setWait(false)
        .setJobSet(new JobSet().setId(id))
        .setDetails(true);

      const stream = client.inspectJobSet(
        inspectJobSetRequest,
        credentialMetadata,
      );

      return streamToObjectArray<JobInfo, JobInfo.AsObject>(stream);
    },

    listJobSets: ({limit}: ListArgs = {}) => {
      const listJobSetRequest = new ListJobSetRequest().setDetails(true);

      const stream = client.listJobSet(listJobSetRequest, credentialMetadata);

      return streamToObjectArray<JobSetInfo, JobSetInfo.AsObject>(
        stream,
        limit || DEFAULT_JOBS_LIMIT,
      );
    },

    inspectJob: ({id, pipelineName}: JobQueryArgs) => {
      return new Promise<JobInfo.AsObject>((resolve, reject) => {
        client.inspectJob(
          new InspectJobRequest()
            .setJob(
              jobFromObject({id}).setPipeline(
                pipelineFromObject({name: pipelineName}),
              ),
            )
            .setDetails(true),
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

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
  CreatePipelineRequest,
  Transform,
  Input,
} from '@pachyderm/proto/pb/pps/pps_pb';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {commitFromObject} from 'builders/pfs';
import {durationFromObject} from 'builders/protobuf';

import {
  jobFromObject,
  pipelineFromObject,
  GetLogsRequestObject,
  getLogsRequestFromObject,
  egressFromObject,
  inputFromObject,
  parallelismSpecFromObject,
  chunkSpecFromObject,
  resourceSpecFromObject,
  schedulingSpecFromObject,
  serviceFromObject,
  spoutFromObject,
  transformFromObject,
} from '../builders/pps';
import {JobSetQueryArgs, JobQueryArgs, ServiceArgs} from '../lib/types';
import {DEFAULT_JOBS_LIMIT} from '../services/constants/pps';
import streamToObjectArray from '../utils/streamToObjectArray';

export interface ListArgs {
  limit?: number | null;
}
export interface ListJobArgs extends ListArgs {
  pipelineId?: string | null;
}

interface CreatePipelineRequestOptions
  extends Omit<
    CreatePipelineRequest.AsObject,
    | 'autoscaling'
    | 'pipeline'
    | 'description'
    | 'transform'
    | 'podPatch'
    | 'podSpec'
    | 'outputBranch'
    | 'reprocess'
    | 'reprocessSpec'
    | 'tfJob'
    | 'extension'
    | 'update'
    | 'salt'
    | 's3Out'
    | 'datumTries'
  > {
  autoscaling?: CreatePipelineRequest.AsObject['autoscaling'];
  name: string;
  transform: Transform.AsObject;
  description?: CreatePipelineRequest.AsObject['description'];
  input: Input.AsObject;
  podPatch?: CreatePipelineRequest.AsObject['podPatch'];
  podSpec?: CreatePipelineRequest.AsObject['podSpec'];
  outputBranch?: CreatePipelineRequest.AsObject['outputBranch'];
  reprocess?: CreatePipelineRequest.AsObject['reprocess'];
  reprocessSpec?: CreatePipelineRequest.AsObject['reprocessSpec'];
  update?: CreatePipelineRequest.AsObject['update'];
  salt?: CreatePipelineRequest.AsObject['salt'];
  s3Out?: CreatePipelineRequest.AsObject['s3Out'];
  datumTries?: CreatePipelineRequest.AsObject['datumTries'];
}

const pps = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    createPipeline: (options: CreatePipelineRequestOptions) => {
      const request = new CreatePipelineRequest();
      if (options.autoscaling) request.setAutoscaling(options.autoscaling);
      if (options.datumSetSpec)
        request.setDatumSetSpec(chunkSpecFromObject(options.datumSetSpec));
      if (options.datumTimeout)
        request.setDatumTimeout(durationFromObject(options.datumTimeout));
      if (options.datumTries) request.setDatumTries(options.datumTries);
      if (options.description) request.setDescription(options.description);
      if (options.egress) request.setEgress(egressFromObject(options.egress));
      if (options.input) request.setInput(inputFromObject(options.input));
      if (options.jobTimeout)
        request.setJobTimeout(durationFromObject(options.jobTimeout));
      // if (options.metadata) request.setMetadata(options.metadata);
      if (options.outputBranch) request.setOutputBranch(options.outputBranch);
      if (options.parallelismSpec)
        request.setParallelismSpec(
          parallelismSpecFromObject(options.parallelismSpec),
        );
      request.setPipeline(pipelineFromObject({name: options.name}));
      if (options.podPatch) request.setPodPatch(options.podPatch);
      if (options.podSpec) request.setPodSpec(options.podSpec);
      if (options.s3Out) request.setS3Out(options.s3Out);
      if (options.reprocess) request.setReprocess(options.reprocess);
      if (options.reprocessSpec)
        request.setReprocessSpec(options.reprocessSpec);
      if (options.resourceLimits)
        request.setResourceLimits(
          resourceSpecFromObject(options.resourceLimits),
        );
      if (options.resourceRequests)
        request.setResourceRequests(
          resourceSpecFromObject(options.resourceRequests),
        );
      if (options.schedulingSpec)
        request.setSchedulingSpec(
          schedulingSpecFromObject(options.schedulingSpec),
        );
      if (options.service)
        request.setService(serviceFromObject(options.service));
      if (options.sidecarResourceLimits)
        request.setSidecarResourceLimits(
          resourceSpecFromObject(options.sidecarResourceLimits),
        );
      if (options.spout) request.setSpout(spoutFromObject(options.spout));
      if (options.update) request.setUpdate(options.update);
      request.setTransform(transformFromObject(options.transform));
      if (options.salt) request.setSalt(options.salt);
      if (options.specCommit)
        request.setSpecCommit(commitFromObject(options.specCommit));

      return new Promise<Empty.AsObject>((resolve, reject) => {
        client.createPipeline(request, (error) => {
          if (error) return reject(error);
          return resolve({});
        });
      });
    },
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

    deleteAll: () => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        client.deleteAll(new Empty(), (error) => {
          if (error) {
            return reject(error);
          }
          return resolve({});
        });
      });
    },
  };
};

export default pps;

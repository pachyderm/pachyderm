import {ClientReadableStream} from '@grpc/grpc-js';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {commitFromObject, CommitObject} from '../builders/pfs';
import {
  jobFromObject,
  pipelineFromObject,
  egressFromObject,
  inputFromObject,
  parallelismSpecFromObject,
  chunkSpecFromObject,
  resourceSpecFromObject,
  schedulingSpecFromObject,
  serviceFromObject,
  spoutFromObject,
  transformFromObject,
  InputObject,
  PipelineObject,
  TransformObject,
  SchedulingSpecObject,
  SpoutObject,
  ServiceObject,
  TFJobObject,
  ParallelismSpecObject,
  EgressObject,
  ResourceSpecObject,
  DatumSetSpecObject,
  getLogsRequestFromArgs,
} from '../builders/pps';
import {durationFromObject, DurationObject} from '../builders/protobuf';
import {
  JobSetQueryArgs,
  JobQueryArgs,
  ServiceArgs,
  GetLogsRequestArgs,
  InspectDatumRequestArgs,
  ListDatumsRequestArgs,
} from '../lib/types';
import {APIClient} from '../proto/pps/pps_grpc_pb';
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
  DeletePipelineRequest,
  InspectDatumRequest,
  ListDatumRequest,
  DatumInfo,
  Datum,
  Job,
} from '../proto/pps/pps_pb';
import {DEFAULT_JOBS_LIMIT} from '../services/constants/pps';
import streamToObjectArray from '../utils/streamToObjectArray';

import {RPC_DEADLINE_MS} from './constants/rpc';

export interface ListArgs {
  limit?: number | null;
}
export interface ListJobArgs extends ListArgs {
  pipelineId?: string | null;
}

export interface ListJobSetArgs extends ListArgs {
  details?: boolean;
}

export interface PipelineArgs {
  name: string;
}

export interface DeletePipelineArgs {
  pipeline: PipelineArgs;
  keepRepo?: boolean;
  force?: boolean;
}

export interface CreatePipelineRequestOptions
  extends Omit<
    CreatePipelineRequest.AsObject,
    | 'pipeline'
    | 'autoscaling'
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
    | 'input'
    | 'schedulingSpec'
    | 'datumTimeout'
    | 'jobTimeout'
    | 'spout'
    | 'service'
    | 'parallismSpec'
    | 'egress'
    | 'resourceRequests'
    | 'resourceLimits'
    | 'sidecarResourceLimits'
    | 'datumSetSpec'
    | 'specCommit'
  > {
  schedulingSpec?: SchedulingSpecObject;
  pipeline: PipelineObject;
  autoscaling?: CreatePipelineRequest.AsObject['autoscaling'];
  transform: TransformObject;
  description?: CreatePipelineRequest.AsObject['description'];
  input: InputObject;
  podPatch?: CreatePipelineRequest.AsObject['podPatch'];
  podSpec?: CreatePipelineRequest.AsObject['podSpec'];
  outputBranch?: CreatePipelineRequest.AsObject['outputBranch'];
  reprocess?: CreatePipelineRequest.AsObject['reprocess'];
  reprocessSpec?: CreatePipelineRequest.AsObject['reprocessSpec'];
  update?: CreatePipelineRequest.AsObject['update'];
  salt?: CreatePipelineRequest.AsObject['salt'];
  s3Out?: CreatePipelineRequest.AsObject['s3Out'];
  tfjob?: TFJobObject;
  datumTries?: CreatePipelineRequest.AsObject['datumTries'];
  datumTimeout?: DurationObject;
  jobTimeout?: DurationObject;
  spout?: SpoutObject;
  service?: ServiceObject;
  parallelismSpec?: ParallelismSpecObject;
  egress?: EgressObject;
  resourceRequests?: ResourceSpecObject;
  resourceLimits?: ResourceSpecObject;
  sidecarResourceLimits?: ResourceSpecObject;
  datumSetSpec?: DatumSetSpecObject;
  specCommit?: CommitObject;
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
      request.setPipeline(pipelineFromObject(options.pipeline));
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
      if (options.datumTimeout)
        request.setDatumTimeout(durationFromObject(options.datumTimeout));
      if (options.jobTimeout)
        request.setJobTimeout(durationFromObject(options.jobTimeout));
      if (options.spout) request.setSpout(spoutFromObject(options.spout));

      return new Promise<Empty.AsObject>((resolve, reject) => {
        client.createPipeline(request, credentialMetadata, (error) => {
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
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
      );

      return streamToObjectArray<PipelineInfo, PipelineInfo.AsObject>(stream);
    },

    deletePipeline: ({
      pipeline,
      keepRepo = false,
      force = false,
    }: DeletePipelineArgs) => {
      const deletePipelineRequest = new DeletePipelineRequest()
        .setPipeline(new Pipeline().setName(pipeline.name))
        .setForce(force)
        .setKeepRepo(keepRepo);
      return new Promise<Empty.AsObject>((resolve, reject) => {
        client.deletePipeline(
          deletePipelineRequest,
          credentialMetadata,
          (error) => {
            if (error) {
              reject(error);
            }
            return resolve({});
          },
        );
      });
    },

    listJobs: ({limit, pipelineId}: ListJobArgs = {}) => {
      const listJobRequest = new ListJobRequest().setDetails(true);

      if (pipelineId) {
        listJobRequest.setPipeline(new Pipeline().setName(pipelineId));
      }

      const stream = client.listJob(listJobRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

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

    inspectJobSet: ({id, details = true}: JobSetQueryArgs) => {
      const inspectJobSetRequest = new InspectJobSetRequest()
        .setWait(false)
        .setJobSet(new JobSet().setId(id))
        .setDetails(details);

      const stream = client.inspectJobSet(
        inspectJobSetRequest,
        credentialMetadata,
        {
          deadline: Date.now() + RPC_DEADLINE_MS,
        },
      );

      return streamToObjectArray<JobInfo, JobInfo.AsObject>(stream);
    },

    listJobSets: ({limit, details = true}: ListJobSetArgs = {}) => {
      const listJobSetRequest = new ListJobSetRequest().setDetails(details);

      const stream = client.listJobSet(listJobSetRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<JobSetInfo, JobSetInfo.AsObject>(
        stream,
        limit || DEFAULT_JOBS_LIMIT,
      );
    },

    inspectJob: ({id, pipelineName, wait = false}: JobQueryArgs) => {
      return new Promise<JobInfo.AsObject>((resolve, reject) => {
        client.inspectJob(
          new InspectJobRequest()
            .setWait(wait)
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

    getLogs: ({
      pipelineName,
      jobId,
      since,
      follow = false,
    }: GetLogsRequestArgs) => {
      const getLogsRequest = getLogsRequestFromArgs({
        pipelineName,
        jobId,
        since,
        follow,
      });
      const stream = client.getLogs(getLogsRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<LogMessage, LogMessage.AsObject>(stream);
    },

    getLogsStream: ({
      pipelineName,
      jobId,
      since,
      follow = false,
    }: GetLogsRequestArgs) => {
      return new Promise<ClientReadableStream<LogMessage>>(
        (resolve, reject) => {
          try {
            const getLogsRequest = getLogsRequestFromArgs({
              pipelineName,
              jobId,
              since,
              follow,
            });
            const stream = client.getLogs(getLogsRequest, credentialMetadata, {
              deadline: Date.now() + RPC_DEADLINE_MS,
            });

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

    inspectDatum: ({id, jobId, pipelineName}: InspectDatumRequestArgs) => {
      return new Promise<DatumInfo>((resolve, reject) => {
        client.inspectDatum(
          new InspectDatumRequest().setDatum(
            new Datum()
              .setId(id)
              .setJob(
                new Job()
                  .setId(jobId)
                  .setPipeline(new Pipeline().setName(pipelineName)),
              ),
          ),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res);
          },
        );
      });
    },

    listDatums: ({jobId, pipelineName, filter}: ListDatumsRequestArgs) => {
      const request = new ListDatumRequest().setJob(
        new Job()
          .setId(jobId)
          .setPipeline(new Pipeline().setName(pipelineName)),
      );

      if (filter) {
        request.setFilter(new ListDatumRequest.Filter().setStateList(filter));
      }

      const stream = client.listDatum(request, credentialMetadata);

      return streamToObjectArray<DatumInfo, DatumInfo.AsObject>(stream);
    },
  };
};

export default pps;

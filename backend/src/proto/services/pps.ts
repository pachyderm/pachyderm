import {type Nullable} from '@dash-backend/lib/types';
import {ClientReadableStream} from '@grpc/grpc-js';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {grpcApiConstructorArgs} from '@dash-backend/proto/utils/createGrpcApiClient';

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
import {
  durationFromObject,
  DurationObject,
  timestampFromObject,
  TimestampObject,
} from '../builders/protobuf';
import {
  JobSetQueryArgs,
  JobQueryArgs,
  ServiceArgs,
  GetLogsRequestArgs,
  InspectDatumRequestArgs,
  ListDatumsRequestArgs,
} from '../lib/types';
import {Project} from '../proto/pfs/pfs_pb';
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
  DeletePipelinesRequest,
  CreatePipelineV2Request,
  CreatePipelineV2Response,
  GetClusterDefaultsResponse,
  GetClusterDefaultsRequest,
  SetClusterDefaultsRequest,
  SetClusterDefaultsResponse,
  RerunPipelineRequest,
} from '../proto/pps/pps_pb';
import streamToObjectArray from '../utils/streamToObjectArray';

import {RPC_DEADLINE_MS} from './constants/rpc';

export interface ListJobArgs {
  pipelineId?: string | null;
  jqFilter?: string;
  projectId: string;
  history?: number;
  number?: number;
  reverse?: boolean;
  cursor?: TimestampObject;
  details?: boolean;
}

export interface ListJobSetArgs {
  details?: boolean;
  projectIds: string[];
  number?: number;
  reverse?: boolean;
  cursor?: TimestampObject;
}

export interface PipelineArgs {
  name: string;
}

export interface DeletePipelineArgs {
  pipeline: PipelineArgs;
  keepRepo?: boolean;
  force?: boolean;
  projectId: string;
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
    | 'tolerationsList'
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

let client: APIClient;

const ppsServiceRpcHandler = ({
  credentialMetadata,
}: Pick<ServiceArgs, 'credentialMetadata'>) => {
  client = client ?? new APIClient(...grpcApiConstructorArgs());

  return {
    createPipelineV2: (
      args?: Nullable<Partial<CreatePipelineV2Request.AsObject>>,
    ) => {
      const req = new CreatePipelineV2Request();
      args?.dryRun && req.setDryRun(args.dryRun);
      args?.reprocess && req.setReprocess(args.reprocess);
      args?.update && req.setUpdate(args.update);
      args?.createPipelineRequestJson &&
        req.setCreatePipelineRequestJson(args.createPipelineRequestJson);

      return new Promise<CreatePipelineV2Response.AsObject>(
        (resolve, reject) => {
          client.createPipelineV2(req, credentialMetadata, (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          });
        },
      );
    },
    getClusterDefaults: () => {
      const req = new GetClusterDefaultsRequest();
      return new Promise<GetClusterDefaultsResponse.AsObject>(
        (resolve, reject) => {
          client.getClusterDefaults(req, credentialMetadata, (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          });
        },
      );
    },
    /**
     * @param clusterDefaultsJson
     * create_pipeline_request is the top level key
     */
    setClusterDefaults: (
      args?: Nullable<Partial<SetClusterDefaultsRequest.AsObject>>,
    ) => {
      const req = new SetClusterDefaultsRequest();
      args?.dryRun && req.setDryRun(args.dryRun);
      args?.reprocess && req.setReprocess(args.reprocess);
      args?.regenerate && req.setRegenerate(args.regenerate);
      args?.clusterDefaultsJson &&
        req.setClusterDefaultsJson(args.clusterDefaultsJson);

      return new Promise<SetClusterDefaultsResponse.AsObject>(
        (resolve, reject) => {
          client.setClusterDefaults(req, credentialMetadata, (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          });
        },
      );
    },
    // TODO: createPipeline should support projects
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
    listPipeline: ({
      projectIds,
      details = true,
      jq = '',
    }: {
      projectIds: string[];
      details?: boolean;
      jq?: string;
    }) => {
      const listPipelineRequest = new ListPipelineRequest()
        .setJqfilter(jq)
        .setDetails(details)
        .setProjectsList(
          projectIds.map((projectName) => new Project().setName(projectName)),
        );
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
      projectId,
      pipeline,
      keepRepo = false,
      force = false,
    }: DeletePipelineArgs) => {
      const deletePipelineRequest = new DeletePipelineRequest()
        .setPipeline(
          new Pipeline()
            .setName(pipeline.name)
            .setProject(new Project().setName(projectId)),
        )
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
    deletePipelines: ({projectIds}: {projectIds: string[]}) => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        const deletePipelinesRequest = new DeletePipelinesRequest();

        const projects = projectIds.map((id) => new Project().setName(id));
        deletePipelinesRequest.setProjectsList(projects);

        client.deletePipelines(
          deletePipelinesRequest,
          credentialMetadata,
          (error) => {
            if (error) {
              return reject(error);
            }
            return resolve({});
          },
        );
      });
    },

    // TODO: should implement multi project list ?
    // SEE PROTO: History -1 means return jobs from all historical versions.
    listJobs: ({
      projectId,
      pipelineId,
      jqFilter,
      history = -1,
      cursor,
      reverse,
      number,
      details = true,
    }: ListJobArgs) => {
      const listJobRequest = new ListJobRequest().setDetails(details);

      listJobRequest.setProjectsList([new Project().setName(projectId)]);

      listJobRequest.setHistory(history);

      if (pipelineId && projectId) {
        listJobRequest.setPipeline(
          new Pipeline()
            .setName(pipelineId)
            .setProject(new Project().setName(projectId)),
        );
      } else if (pipelineId) {
        listJobRequest.setPipeline(new Pipeline().setName(pipelineId));
      }

      if (jqFilter) {
        listJobRequest.setJqfilter(jqFilter);
      }

      if (projectId) {
        listJobRequest.setProjectsList([new Project().setName(projectId)]);
      }

      if (cursor) {
        listJobRequest.setPaginationmarker(timestampFromObject(cursor));
      }

      if (number) {
        listJobRequest.setNumber(number);
      }

      if (reverse) {
        listJobRequest.setReverse(reverse);
      }

      const stream = client.listJob(listJobRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<JobInfo, JobInfo.AsObject>(stream);
    },

    inspectPipeline: ({
      projectId,
      pipelineId,
    }: {
      projectId: string;
      pipelineId: string;
    }) => {
      return new Promise<PipelineInfo.AsObject>((resolve, reject) => {
        const pipelineRequest = new InspectPipelineRequest()
          .setPipeline(
            new Pipeline()
              .setName(pipelineId)
              .setProject(new Project().setName(projectId)),
          )
          .setDetails(true);

        client.inspectPipeline(
          pipelineRequest,
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

    listJobSetServerDerivedFromListJobs: async ({
      projectId,
      limit = 10,
    }: {
      projectId: string;
      limit?: number;
    }) => {
      const listJobRequest = new ListJobRequest();

      listJobRequest.setDetails(false);
      listJobRequest.setProjectsList([new Project().setName(projectId)]);
      listJobRequest.setHistory(-1);

      const stream = client.listJob(listJobRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      /**
       * This function converts a stream of JobInfo data into a Map where the key is the JobSetId and the value is an array of jobs.
       *
       * There is a limit parameter that restricts the Map to only contain the first 'limit' number of unique JobSets. However, the
       * function continues scanning through the remaining jobs in the stream in order to add jobs to the existing job sets, if they
       * belong there. This is because the job sets in the stream might not be sequential.
       */
      const streamToObjectArray = (stream: ClientReadableStream<JobInfo>) => {
        return new Promise<Map<string, JobInfo.AsObject[]>>(
          (resolve, reject) => {
            const jobsMap = new Map<string, JobInfo.AsObject[]>();

            stream.on('data', (chunk: JobInfo) => {
              const job = chunk.toObject();

              const addNewJobSets = jobsMap.size < limit;

              if (job?.job?.id) {
                const jobId = job.job.id;

                if (!addNewJobSets && !jobsMap.has(jobId)) {
                  return;
                }

                const jobsInJobSet = jobsMap.get(jobId) || [];
                jobsInJobSet.push(job);
                jobsMap.set(jobId, jobsInJobSet);
              }
            });
            stream.on('end', () => resolve(jobsMap));
            stream.on('close', () => resolve(jobsMap));
            stream.on('error', (err) => reject(err));
          },
        );
      };

      return streamToObjectArray(stream);
    },

    listJobSets: (
      {projectIds, details = true, cursor, number, reverse}: ListJobSetArgs = {
        projectIds: [],
      },
    ) => {
      const listJobSetRequest = new ListJobSetRequest().setDetails(details);

      listJobSetRequest.setProjectsList(
        projectIds.map((projectName) => new Project().setName(projectName)),
      );

      if (cursor) {
        listJobSetRequest.setPaginationmarker(timestampFromObject(cursor));
      }

      if (number) {
        listJobSetRequest.setNumber(number);
      }

      if (reverse) {
        listJobSetRequest.setReverse(reverse);
      }

      const stream = client.listJobSet(listJobSetRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<JobSetInfo, JobSetInfo.AsObject>(stream);
    },

    inspectJob: ({projectId, id, pipelineName, wait = false}: JobQueryArgs) => {
      return new Promise<JobInfo.AsObject>((resolve, reject) => {
        client.inspectJob(
          new InspectJobRequest()
            .setWait(wait)
            .setJob(
              jobFromObject({id}).setPipeline(
                pipelineFromObject({
                  name: pipelineName,
                  project: {name: projectId},
                }),
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
      projectId,
      pipelineName,
      jobId,
      datumId,
      since,
      follow = false,
      master = false,
      limit,
    }: GetLogsRequestArgs) => {
      const getLogsRequest = getLogsRequestFromArgs({
        projectId,
        pipelineName,
        jobId,
        datumId,
        since,
        follow,
        master,
      });

      const stream = client.getLogs(getLogsRequest, credentialMetadata, {
        deadline: Date.now() + RPC_DEADLINE_MS,
      });

      return streamToObjectArray<LogMessage, LogMessage.AsObject>(
        stream,
        limit,
      );
    },

    getLogsStream: ({
      projectId,
      pipelineName,
      jobId,
      datumId,
      since,
      follow = false,
    }: GetLogsRequestArgs) => {
      return new Promise<ClientReadableStream<LogMessage>>(
        (resolve, reject) => {
          try {
            const getLogsRequest = getLogsRequestFromArgs({
              projectId,
              pipelineName,
              jobId,
              datumId,
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

    inspectDatum: ({
      projectId,
      id,
      jobId,
      pipelineName,
    }: InspectDatumRequestArgs) => {
      return new Promise<DatumInfo>((resolve, reject) => {
        client.inspectDatum(
          new InspectDatumRequest().setDatum(
            new Datum()
              .setId(id)
              .setJob(
                new Job()
                  .setId(jobId)
                  .setPipeline(
                    new Pipeline()
                      .setName(pipelineName)
                      .setProject(new Project().setName(projectId)),
                  ),
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

    listDatums: ({
      projectId,
      jobId,
      pipelineName,
      filter,
      number,
      cursor,
    }: ListDatumsRequestArgs) => {
      const request = new ListDatumRequest().setJob(
        new Job()
          .setId(jobId)
          .setPipeline(
            new Pipeline()
              .setName(pipelineName)
              .setProject(new Project().setName(projectId)),
          ),
      );
      if (filter) {
        request.setFilter(new ListDatumRequest.Filter().setStateList(filter));
      }
      if (number) {
        request.setNumber(number);
      }
      if (cursor) {
        request.setPaginationmarker(cursor);
      }
      const stream = client.listDatum(request, credentialMetadata);

      return streamToObjectArray<DatumInfo, DatumInfo.AsObject>(stream);
    },
    rerunPipeline: ({
      projectId,
      pipelineId,
      reprocess = false,
    }: {
      projectId: string;
      pipelineId: string;
      reprocess?: boolean;
    }) => {
      const req = new RerunPipelineRequest();

      req.setPipeline(
        new Pipeline()
          .setName(pipelineId)
          .setProject(new Project().setName(projectId)),
      );
      reprocess && req.setReprocess(reprocess);

      return new Promise<void>((resolve, reject) => {
        client.rerunPipeline(req, credentialMetadata, (error) => {
          if (error) {
            return reject(error);
          }
          return resolve();
        });
      });
    },
  };
};

export default ppsServiceRpcHandler;

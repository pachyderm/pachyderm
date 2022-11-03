import {Status} from '@grpc/grpc-js/build/src/constants';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {
  PpsIAPIServer,
  LogMessage,
  JobSetInfo,
  Pipeline,
  PipelineInfo,
  JobSet,
  RepoInfo,
  Repo,
  Branch,
} from '@dash-backend/proto';
import {pipelineInfoFromObject} from '@dash-backend/proto/builders/pps';
import {timestampFromObject} from '@dash-backend/proto/builders/protobuf';
import {createServiceError} from '@dash-backend/testHelpers';

import runJQFilter from '../utils/runJQFilter';

import MockState from './MockState';

const DEFAULT_SINCE_TIME = 86400;

const pps = () => {
  return {
    getService: (): Pick<
      PpsIAPIServer,
      | 'listPipeline'
      | 'listJob'
      | 'inspectJob'
      | 'inspectPipeline'
      | 'inspectJobSet'
      | 'listJobSet'
      | 'getLogs'
      | 'createPipeline'
      | 'deletePipeline'
      | 'inspectDatum'
      | 'listDatum'
    > => {
      return {
        listPipeline: async (call) => {
          const [projectId] = call.metadata.get('project-id');
          const [authToken] = call.metadata.get('authn-token');
          if (authToken && authToken === 'expired') {
            call.emit(
              'error',
              createServiceError({
                code: Status.INTERNAL,
                details: 'token expiration is in the past',
              }),
            );
          }
          let replyPipelines = projectId
            ? MockState.state.pipelines[projectId.toString()]
            : MockState.state.pipelines['1'];

          if (call.request.getJqfilter()) {
            replyPipelines = await runJQFilter({
              jqFilter: `.pipelineInfoList[] | ${call.request.getJqfilter()}`,
              object: {
                pipelineInfoList: replyPipelines.map((p) => p.toObject()),
              },
              objectMapper: pipelineInfoFromObject,
            });
          }

          replyPipelines.forEach((pipeline) => call.write(pipeline));
          call.end();
        },
        listJob: async (call) => {
          const [projectId] = call.metadata.get('project-id');
          let replyJobs = projectId
            ? MockState.state.jobs[projectId.toString()]
            : MockState.state.jobs['1'];

          const pipeline = call.request.getPipeline();
          if (pipeline) {
            replyJobs = replyJobs.filter(
              (job) =>
                job.getJob()?.getPipeline()?.getName() === pipeline.getName(),
            );
          }

          replyJobs.forEach((job) => call.write(job));
          call.end();
        },
        inspectJob: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const replyJobs = projectId
            ? MockState.state.jobs[projectId.toString()]
            : MockState.state.jobs['1'];

          const foundJob = replyJobs.find(
            (job) =>
              job.getJob()?.getId() === call.request.getJob()?.getId() &&
              job.getJob()?.getPipeline()?.getName() ===
                call.request.getJob()?.getPipeline()?.getName(),
          );
          if (foundJob) {
            callback(null, foundJob);
          } else {
            callback({code: Status.NOT_FOUND, details: 'job not found'});
          }
        },
        inspectPipeline: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const projectPipelines = projectId
            ? MockState.state.pipelines[projectId.toString()]
            : MockState.state.pipelines['1'];
          const foundPipeline = projectPipelines.find((pipeline) => {
            return (
              pipeline.getPipeline()?.getName() ===
              call.request.getPipeline()?.getName()
            );
          });

          if (foundPipeline) {
            callback(null, foundPipeline);
          } else {
            callback({code: Status.NOT_FOUND, details: 'pipeline not found'});
          }
        },
        inspectJobSet: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const projectJobSets =
            MockState.state.jobSets[projectId.toString()] ||
            MockState.state.jobSets['default'];

          const foundJobSet =
            projectJobSets[call.request.getJobSet()?.getId() || ''];

          if (foundJobSet) {
            foundJobSet.forEach((job) => call.write(job));
          }
          call.end();
        },
        listJobSet: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const projectJobSets =
            MockState.state.jobSets[projectId.toString()] ||
            MockState.state.jobSets['default'];

          Object.keys(projectJobSets).forEach((jobId) =>
            call.write(
              new JobSetInfo()
                .setJobSet(new JobSet().setId(jobId))
                .setJobsList(projectJobSets[jobId]),
            ),
          );
          call.end();
        },
        getLogs: async (call) => {
          if (!call.request.getPipeline() && !call.request.getJob()) {
            MockState.state.workspaceLogs
              .slice(-call.request.getTail() || 0)
              .forEach((log) => {
                call.write(log);
              });
          } else {
            let filteredLogs: LogMessage[] = [];
            const [projectId] = call.metadata.get('project-id');
            const projectLogs = projectId
              ? MockState.state.pipelineAndJobLogs[projectId.toString()]
              : MockState.state.pipelineAndJobLogs['1'];

            const pipelineName = call.request.getPipeline()?.getName();
            if (pipelineName) {
              filteredLogs = projectLogs.filter(
                (log) => pipelineName && log.getPipelineName() === pipelineName,
              );
            } else {
              const jobId = call.request.getJob()?.getId();
              const pipelineJobName = call.request
                .getJob()
                ?.getPipeline()
                ?.getName();

              filteredLogs = projectLogs.filter(
                (log) =>
                  jobId &&
                  log.getJobId() === jobId &&
                  pipelineJobName &&
                  log.getPipelineName() === pipelineJobName,
              );
            }

            filteredLogs.slice(-call.request.getTail() || 0).forEach((log) => {
              const now = Math.floor(Date.now() / 1000);
              const timeGap =
                (call.request.getSince()?.getSeconds() || DEFAULT_SINCE_TIME) +
                (log.getTs()?.getSeconds() || 0);
              if (timeGap >= now) {
                call.write(log);
              }
            });
          }
          if (!call.request.getFollow()) {
            call.end();
          }
        },
        createPipeline: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const pipelineName = call.request.getPipeline()?.getName();
          const update = call.request.getUpdate();
          const transform = call.request.getTransform();
          const input = call.request.getInput();
          const description = call.request.getDescription();
          const projectPipelines = projectId
            ? MockState.state.pipelines[projectId.toString()]
            : MockState.state.pipelines['1'];

          if (pipelineName) {
            const existingPipeline = projectPipelines.find(
              (pipeline) => pipeline.getPipeline()?.getName() === pipelineName,
            );
            if (existingPipeline) {
              if (!update) {
                callback({
                  code: Status.ALREADY_EXISTS,
                  details: `pipeline ${pipelineName} already exists`,
                });
              } else {
                const details = existingPipeline.getDetails();
                details?.setTransform(transform);
                details?.setInput(input).setDescription(description);
                existingPipeline.setDetails(details);
              }
            } else {
              const newPipeline = new PipelineInfo()
                .setPipeline(new Pipeline().setName(pipelineName))
                .setDetails(
                  new PipelineInfo.Details()
                    .setTransform(transform)
                    .setInput(input)
                    .setDescription(description),
                );
              projectPipelines.push(newPipeline);

              const projectRepos = projectId
                ? MockState.state.repos[projectId.toString()]
                : MockState.state.repos['1'];
              const newRepo = new RepoInfo()
                .setRepo(new Repo().setName(pipelineName).setType('user'))
                .setDetails(new RepoInfo.Details().setSizeBytes(0))
                .setBranchesList([new Branch().setName('master')])
                .setCreated(
                  timestampFromObject({
                    seconds: Math.floor(Date.now() / 1000),
                    nanos: 0,
                  }),
                )
                .setDescription(`Output repo for ${pipelineName}`);
              projectRepos.push(newRepo);
            }
          }
          callback(null, new Empty());
        },
        deletePipeline: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const pipelineName = call.request.getPipeline()?.getName();

          const projectPipelines = projectId
            ? MockState.state.pipelines[projectId.toString()]
            : MockState.state.pipelines['1'];

          MockState.state.pipelines[projectId.toString()] =
            projectPipelines.filter((pipelineInfo) => {
              return pipelineName !== pipelineInfo.getPipeline()?.getName();
            });

          callback(null, new Empty());
        },
        listDatum: async (call) => {
          const [projectId] = call.metadata.get('project-id');
          const pipelineName =
            call.request.getJob()?.getPipeline()?.getName() || '';
          const jobId = call.request.getJob()?.getId() || '';
          const projectDatums =
            MockState.state.datums[projectId.toString()] ||
            MockState.state.datums['default'];

          const datums = projectDatums[pipelineName][jobId];

          datums.forEach((datum) => call.write(datum));
          call.end();
        },
        inspectDatum: (call, callback) => {
          const [projectId] = call.metadata.get('project-id');
          const pipelineName =
            call.request.getDatum()?.getJob()?.getPipeline()?.getName() || '';
          const jobId = call.request.getDatum()?.getJob()?.getId() || '';
          const datumId = call.request.getDatum()?.getId();
          const projectDatums =
            MockState.state.datums[projectId.toString()] ||
            MockState.state.datums['default'];

          const datums = projectDatums[pipelineName][jobId];

          const matchingDatum = datums.find(
            (datumInfo) => datumInfo.getDatum()?.getId() === datumId,
          );

          if (matchingDatum) {
            callback(null, matchingDatum);
          } else {
            callback({code: Status.NOT_FOUND, details: 'datum not found'});
          }
        },
      };
    },
  };
};

export default pps();

import {Status} from '@grpc/grpc-js/build/src/constants';
import {
  PpsIAPIServer,
  LogMessage,
  JobSetInfo,
  Pipeline,
  PipelineInfo,
} from '@pachyderm/node-pachyderm';
import {pipelineInfoFromObject} from '@pachyderm/node-pachyderm/dist/builders/pps';
import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import jobs from '@dash-backend/mock/fixtures/jobs';
import pipelines from '@dash-backend/mock/fixtures/pipelines';
import {createServiceError} from '@dash-backend/testHelpers';

import jobSets from '../fixtures/jobSets';
import {pipelineAndJobLogs, workspaceLogs} from '../fixtures/logs';
import runJQFilter from '../utils/runJQFilter';

const DEFAULT_SINCE_TIME = 86400;

const pps = () => {
  let state = {pipelines, jobs, jobSets, pipelineAndJobLogs, workspaceLogs};
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
            ? state.pipelines[projectId.toString()]
            : state.pipelines['1'];

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
            ? state.jobs[projectId.toString()]
            : state.jobs['1'];

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
            ? state.jobs[projectId.toString()]
            : state.jobs['1'];

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
            ? state.pipelines[projectId.toString()]
            : state.pipelines['1'];
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
            state.jobSets[projectId.toString()] || state.jobSets['default'];

          const foundJobSet =
            projectJobSets[call.request.getJobSet()?.getId() || ''];

          if (!foundJobSet) {
            call.emit(
              'error',
              createServiceError({
                code: Status.UNKNOWN,
                details: 'no commits found',
              }),
            );
          } else {
            foundJobSet.forEach((job) => call.write(job));
          }
          call.end();
        },
        listJobSet: (call) => {
          const [projectId] = call.metadata.get('project-id');
          const projectJobSets =
            state.jobSets[projectId.toString()] || state.jobSets['default'];

          Object.keys(projectJobSets).forEach((jobId) =>
            call.write(new JobSetInfo().setJobsList(projectJobSets[jobId])),
          );
          call.end();
        },
        getLogs: async (call) => {
          if (!call.request.getPipeline() && !call.request.getJob()) {
            state.workspaceLogs
              .slice(-call.request.getTail() || 0)
              .forEach((log) => {
                call.write(log);
              });
          } else {
            let filteredLogs: LogMessage[] = [];
            const [projectId] = call.metadata.get('project-id');
            const projectLogs = projectId
              ? state.pipelineAndJobLogs[projectId.toString()]
              : state.pipelineAndJobLogs['1'];

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
            ? state.pipelines[projectId.toString()]
            : state.pipelines['1'];

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
            }
          }
          callback(null, new Empty());
        },
      };
    },
    resetState: () => {
      state = {
        pipelines,
        jobs,
        jobSets,
        pipelineAndJobLogs,
        workspaceLogs,
      };
    },
  };
};

export default pps();

import {Status} from '@grpc/grpc-js/build/src/constants';
import {IAPIServer} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {JobSetInfo, LogMessage} from '@pachyderm/proto/pb/pps/pps_pb';

import jobs from '@dash-backend/mock/fixtures/jobs';
import pipelines from '@dash-backend/mock/fixtures/pipelines';
import {createServiceError} from '@dash-backend/testHelpers';

import {pipelineInfoFromObject} from '../../grpc/builders/pps';
import jobSets from '../fixtures/jobSets';
import {pipelineAndJobLogs, workspaceLogs} from '../fixtures/logs';
import runJQFilter from '../utils/runJQFilter';

const DEFAULT_SINCE_TIME = 86400;

const pps: Pick<
  IAPIServer,
  | 'listPipeline'
  | 'listJob'
  | 'inspectJob'
  | 'inspectPipeline'
  | 'inspectJobSet'
  | 'listJobSet'
  | 'getLogs'
> = {
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
      ? pipelines[projectId.toString()]
      : pipelines['1'];

    if (call.request.getJqfilter()) {
      replyPipelines = await runJQFilter({
        jqFilter: `.pipelineInfoList[] | ${call.request.getJqfilter()}`,
        object: {pipelineInfoList: replyPipelines.map((p) => p.toObject())},
        objectMapper: pipelineInfoFromObject,
      });
    }

    replyPipelines.forEach((pipeline) => call.write(pipeline));
    call.end();
  },
  listJob: async (call) => {
    const [projectId] = call.metadata.get('project-id');
    let replyJobs = projectId ? jobs[projectId.toString()] : jobs['1'];

    const pipeline = call.request.getPipeline();
    if (pipeline) {
      replyJobs = replyJobs.filter(
        (job) => job.getJob()?.getPipeline()?.getName() === pipeline.getName(),
      );
    }

    replyJobs.forEach((job) => call.write(job));
    call.end();
  },
  inspectJob: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const replyJobs = projectId ? jobs[projectId.toString()] : jobs['1'];

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
      ? pipelines[projectId.toString()]
      : pipelines['1'];
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
    const projectJobSets = jobSets[projectId.toString()] || jobSets['default'];

    const foundJobSet = projectJobSets[call.request.getJobSet()?.getId() || ''];

    if (!foundJobSet) {
      call.emit(
        'error',
        createServiceError({code: Status.UNKNOWN, details: 'no commits found'}),
      );
    } else {
      foundJobSet.forEach((job) => call.write(job));
    }
    call.end();
  },
  listJobSet: (call) => {
    const [projectId] = call.metadata.get('project-id');
    const projectJobSets = jobSets[projectId.toString()] || jobSets['default'];

    Object.keys(projectJobSets).forEach((jobId) =>
      call.write(new JobSetInfo().setJobsList(projectJobSets[jobId])),
    );
    call.end();
  },
  getLogs: async (call) => {
    if (!call.request.getPipeline() && !call.request.getJob()) {
      workspaceLogs.slice(-call.request.getTail() || 0).forEach((log) => {
        call.write(log);
      });
    } else {
      let filteredLogs: LogMessage[] = [];
      const [projectId] = call.metadata.get('project-id');
      const projectLogs = projectId
        ? pipelineAndJobLogs[projectId.toString()]
        : pipelineAndJobLogs['1'];

      const pipelineName = call.request.getPipeline()?.getName();
      if (pipelineName) {
        filteredLogs = projectLogs.filter(
          (log) => pipelineName && log.getPipelineName() === pipelineName,
        );
      } else {
        const jobId = call.request.getJob()?.getId();
        const pipelineJobName = call.request.getJob()?.getPipeline()?.getName();

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
};

export default pps;

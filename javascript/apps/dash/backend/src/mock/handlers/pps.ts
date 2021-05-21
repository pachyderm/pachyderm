import {Status} from '@grpc/grpc-js/build/src/constants';
import {IAPIServer} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {PipelineInfos} from '@pachyderm/proto/pb/pps/pps_pb';

import pipelineJobs from '@dash-backend/mock/fixtures/pipelineJobs';
import pipelines from '@dash-backend/mock/fixtures/pipelines';

import {
  pipelineJobInfoFromObject,
  pipelineInfoFromObject,
} from '../../grpc/builders/pps';
import runJQFilter from '../utils/runJQFilter';

const pps: Pick<
  IAPIServer,
  'listPipeline' | 'listPipelineJob' | 'inspectPipelineJob' | 'inspectPipeline'
> = {
  listPipeline: async (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const [authToken] = call.metadata.get('authn-token');

    if (authToken && authToken === 'expired') {
      callback(
        {
          code: Status.INTERNAL,
          details: 'token expiration is in the past',
        },
        null,
      );
    }

    const reply = new PipelineInfos();
    // "tutorial" in this case represents the default/catch-all project in core pach
    reply.setPipelineInfoList(
      projectId ? pipelines[projectId.toString()] : pipelines['1'],
    );

    if (call.request.getJqfilter()) {
      return callback(
        null,
        reply.setPipelineInfoList(
          await runJQFilter({
            jqFilter: `.pipelineInfoList[] | ${call.request.getJqfilter()}`,
            object: reply.toObject(),
            objectMapper: pipelineInfoFromObject,
          }),
        ),
      );
    } else {
      callback(null, reply);
    }
  },
  listPipelineJob: async (call) => {
    const [projectId] = call.metadata.get('project-id');
    let replyJobs = projectId
      ? pipelineJobs[projectId.toString()]
      : pipelineJobs['1'];

    if (call.request.getJqfilter()) {
      replyJobs = await runJQFilter({
        jqFilter: `.pipelineJobInfoList[] | ${call.request.getJqfilter()}`,
        object: {pipelineJobInfoList: replyJobs.map((rj) => rj.toObject())},
        objectMapper: pipelineJobInfoFromObject,
      });
    }

    replyJobs.forEach((job) => call.write(job));
    call.end();
  },
  inspectPipelineJob: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const replyJobs = projectId
      ? pipelineJobs[projectId.toString()]
      : pipelineJobs['1'];
    const foundJob = replyJobs.find(
      (job) =>
        job.getPipelineJob()?.getId() ===
        call.request.getPipelineJob()?.getId(),
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
};

export default pps;

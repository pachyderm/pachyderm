import {RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {
  PipelineJobInfo,
  PipelineJobState,
  PipelineInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';
import fromPairs from 'lodash/fromPairs';

import {
  toGQLPipelineJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import {PipelineJob, Pipeline, PipelineType} from '@graphqlTypes';

const derivePipelineType = (pipelineInfo: PipelineInfo.AsObject) => {
  if (pipelineInfo.service) {
    return PipelineType.SERVICE;
  }

  if (pipelineInfo.spout) {
    return PipelineType.SPOUT;
  }

  return PipelineType.STANDARD;
};

export const pipelineInfoToGQLPipeline = (
  pipelineInfo: PipelineInfo.AsObject,
): Pipeline => {
  const jobStates = fromPairs(pipelineInfo.jobCountsMap);

  return {
    // pipelines don't always have an ID, most of the time it uses
    // name as the global identifier
    id: pipelineInfo.pipeline?.name || '',
    name: pipelineInfo.pipeline?.name || '',
    description: pipelineInfo.description || '',
    version: pipelineInfo.version,
    createdAt: pipelineInfo.createdAt?.seconds || 0,
    state: toGQLPipelineState(pipelineInfo.state),
    stopped: pipelineInfo.stopped,
    recentError: pipelineInfo.recentError,
    numOfJobsStarting: jobStates[PipelineJobState.JOB_STARTING] || 0,
    numOfJobsRunning: jobStates[PipelineJobState.JOB_RUNNING] || 0,
    numOfJobsFailing: jobStates[PipelineJobState.JOB_FAILURE] || 0,
    numOfJobsSucceeding: jobStates[PipelineJobState.JOB_SUCCESS] || 0,
    numOfJobsKilled: jobStates[PipelineJobState.JOB_KILLED] || 0,
    numOfJobsEgressing: jobStates[PipelineJobState.JOB_EGRESSING] || 0,
    lastJobState: toGQLPipelineJobState(pipelineInfo.lastJobState),
    type: derivePipelineType(pipelineInfo),
    transform: pipelineInfo.transform,
    inputString: JSON.stringify(pipelineInfo.input, null, 2),
    cacheSize: pipelineInfo.cacheSize,
    datumTimeoutS: pipelineInfo.datumTimeout?.seconds,
    datumTries: pipelineInfo.datumTries,
    jobTimeoutS: pipelineInfo.jobTimeout?.seconds,
    enableStats: pipelineInfo.enableStats,
    outputBranch: pipelineInfo.outputBranch,
    s3OutputRepo:
      pipelineInfo.s3Out && pipelineInfo.pipeline
        ? `s3//${pipelineInfo.pipeline.name}`
        : undefined,
    egress: Boolean(pipelineInfo.egress),
    schedulingSpec: pipelineInfo.schedulingSpec
      ? {
          nodeSelectorMap: pipelineInfo.schedulingSpec.nodeSelectorMap.map(
            ([key, value]) => ({
              key,
              value,
            }),
          ),
          priorityClassName: pipelineInfo.schedulingSpec.priorityClassName,
        }
      : undefined,
  };
};

export const jobInfoToGQLJob = (
  jobInfo: PipelineJobInfo.AsObject,
): PipelineJob => {
  return {
    id: jobInfo.pipelineJob?.id || '',
    state: toGQLPipelineJobState(jobInfo.state),
    createdAt: jobInfo.started?.seconds || 0,
    finishedAt: jobInfo.finished?.seconds,
    pipelineName: jobInfo.pipeline?.name || '',
    transform: jobInfo.transform,
    inputString: jobInfo.input
      ? JSON.stringify(jobInfo.input, null, 2)
      : undefined,
    inputBranch: jobInfo.input?.pfs?.branch,
  };
};

export const repoInfoToGQLRepo = (repoInfo: RepoInfo.AsObject) => {
  return {
    createdAt: repoInfo.created?.seconds || 0,
    description: repoInfo.description,
    name: repoInfo.repo?.name || '',
    sizeInBytes: repoInfo.sizeBytes,
    id: repoInfo?.repo?.name || '',
  };
};

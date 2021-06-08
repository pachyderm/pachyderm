import {RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {JobInfo, JobState, PipelineInfo} from '@pachyderm/proto/pb/pps/pps_pb';
import fromPairs from 'lodash/fromPairs';

import {
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import {Job, Pipeline, PipelineType} from '@graphqlTypes';

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
    numOfJobsStarting: jobStates[JobState.JOB_STARTING] || 0,
    numOfJobsRunning: jobStates[JobState.JOB_RUNNING] || 0,
    numOfJobsFailing: jobStates[JobState.JOB_FAILURE] || 0,
    numOfJobsSucceeding: jobStates[JobState.JOB_SUCCESS] || 0,
    numOfJobsKilled: jobStates[JobState.JOB_KILLED] || 0,
    numOfJobsEgressing: jobStates[JobState.JOB_EGRESSING] || 0,
    lastJobState: toGQLJobState(pipelineInfo.lastJobState),
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

export const jobInfoToGQLJob = (jobInfo: JobInfo.AsObject): Job => {
  return {
    id: jobInfo.job?.id || '',
    state: toGQLJobState(jobInfo.state),
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

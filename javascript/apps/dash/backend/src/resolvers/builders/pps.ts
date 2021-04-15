import {JobInfo, JobState, PipelineInfo} from '@pachyderm/proto/pb/pps/pps_pb';
import fromPairs from 'lodash/fromPairs';

import {PipelineType} from '@graphqlTypes';

const derivePipelineType = (pipelineInfo: PipelineInfo.AsObject) => {
  if (pipelineInfo.service) {
    return PipelineType.Service;
  }

  if (pipelineInfo.spout) {
    return PipelineType.Spout;
  }

  return PipelineType.Standard;
};

export const pipelineInfoToGQLPipeline = (
  pipelineInfo: PipelineInfo.AsObject,
) => {
  const jobStates = fromPairs(pipelineInfo.jobCountsMap);

  return {
    // pipelines don't always have an ID, most of the time it uses
    // name as the global identifier
    id: pipelineInfo.id || pipelineInfo.pipeline?.name || '',
    name: pipelineInfo.pipeline?.name || '',
    description: pipelineInfo.description || '',
    version: pipelineInfo.version,
    createdAt: pipelineInfo.createdAt?.seconds || 0,
    state: pipelineInfo.state || 0,
    stopped: pipelineInfo.stopped,
    recentError: pipelineInfo.recentError,
    numOfJobsStarting: jobStates[JobState.JOB_STARTING] || 0,
    numOfJobsRunning: jobStates[JobState.JOB_RUNNING] || 0,
    numOfJobsFailing: jobStates[JobState.JOB_FAILURE] || 0,
    numOfJobsSucceeding: jobStates[JobState.JOB_SUCCESS] || 0,
    numOfJobsKilled: jobStates[JobState.JOB_KILLED] || 0,
    numOfJobsEgressing: jobStates[JobState.JOB_EGRESSING] || 0,
    lastJobState: pipelineInfo.lastJobState || 0,
    type: derivePipelineType(pipelineInfo),
    transform: pipelineInfo.transform
      ? {
          cmdList: pipelineInfo.transform.cmdList,
          image: pipelineInfo.transform.image,
        }
      : undefined,
    inputString: JSON.stringify(pipelineInfo.input, null, 2),
    cacheSize: pipelineInfo.cacheSize,
    datumTimeoutS: pipelineInfo.datumTimeout?.seconds,
    datumTries: pipelineInfo.datumTries,
    jobTimeoutS: pipelineInfo.jobTimeout?.seconds,
    enableStats: pipelineInfo.enableStats,
    outputBranch: pipelineInfo.outputBranch,
  };
};

export const jobInfoToGQLJob = (jobInfo: JobInfo.AsObject) => {
  return {
    id: jobInfo.job?.id || '',
    state: jobInfo.state,
    createdAt: jobInfo.started?.seconds || 0,
  };
};

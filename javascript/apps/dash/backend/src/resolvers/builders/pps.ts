import {Branch, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {
  JobInfo,
  JobState,
  PipelineInfo,
  LogMessage,
  JobSetInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';
import fromPairs from 'lodash/fromPairs';

import formatBytes from '@dash-backend/lib/formatBytes';
import {
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import {
  Job,
  Pipeline,
  PipelineType,
  JobState as GQLJobState,
  Repo,
  Branch as GQLBranch,
  JobSet,
} from '@graphqlTypes';

const derivePipelineType = (pipelineInfo: PipelineInfo.AsObject) => {
  if (pipelineInfo.details?.service) {
    return PipelineType.SERVICE;
  }

  if (pipelineInfo.details?.spout) {
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
    description: pipelineInfo.details?.description || '',
    version: pipelineInfo.version,
    createdAt: pipelineInfo.details?.createdAt?.seconds || 0,
    state: toGQLPipelineState(pipelineInfo.state),
    stopped: pipelineInfo.stopped,
    recentError: pipelineInfo.details?.recentError,
    numOfJobsCreated: jobStates[JobState.JOB_CREATED] || 0,
    numOfJobsStarting: jobStates[JobState.JOB_STARTING] || 0,
    numOfJobsRunning: jobStates[JobState.JOB_RUNNING] || 0,
    numOfJobsFailing: jobStates[JobState.JOB_FAILURE] || 0,
    numOfJobsSucceeding: jobStates[JobState.JOB_SUCCESS] || 0,
    numOfJobsKilled: jobStates[JobState.JOB_KILLED] || 0,
    numOfJobsEgressing: jobStates[JobState.JOB_EGRESSING] || 0,
    lastJobState: toGQLJobState(pipelineInfo.lastJobState),
    type: derivePipelineType(pipelineInfo),
    transform: pipelineInfo.details?.transform,
    inputString: pipelineInfo.details?.input
      ? JSON.stringify(pipelineInfo.details?.input, null, 2)
      : '',
    datumTimeoutS: pipelineInfo.details?.datumTimeout?.seconds,
    datumTries: pipelineInfo.details?.datumTries || 0,
    jobTimeoutS: pipelineInfo.details?.jobTimeout?.seconds,
    outputBranch: pipelineInfo.details?.outputBranch || '',
    s3OutputRepo:
      pipelineInfo.details?.s3Out && pipelineInfo.pipeline
        ? `s3//${pipelineInfo.pipeline.name}`
        : undefined,
    egress: Boolean(pipelineInfo.details?.egress),
    schedulingSpec: pipelineInfo.details?.schedulingSpec
      ? {
          nodeSelectorMap:
            pipelineInfo.details?.schedulingSpec.nodeSelectorMap.map(
              ([key, value]) => ({
                key,
                value,
              }),
            ),
          priorityClassName:
            pipelineInfo.details?.schedulingSpec.priorityClassName,
        }
      : undefined,
  };
};

export const jobInfoToGQLJob = (jobInfo: JobInfo.AsObject): Job => {
  return {
    id: jobInfo.job?.id || '',
    state: toGQLJobState(jobInfo.state),
    createdAt: jobInfo.created?.seconds,
    startedAt: jobInfo.started?.seconds,
    finishedAt: jobInfo.finished?.seconds,
    pipelineName: jobInfo.job?.pipeline?.name || '',
    transform: jobInfo.details?.transform,
    inputString: jobInfo.details?.input
      ? JSON.stringify(jobInfo.details?.input, null, 2)
      : undefined,
    inputBranch: jobInfo.details?.input?.pfs?.branch,
  };
};

const branchInfoToGQLBranch = (branch: Branch.AsObject): GQLBranch => {
  return {
    id: branch.name,
    name: branch.name,
  };
};

export const repoInfoToGQLRepo = (repoInfo: RepoInfo.AsObject): Repo => {
  return {
    createdAt: repoInfo.created?.seconds || 0,
    description: repoInfo.description,
    name: repoInfo.repo?.name || '',
    sizeBytes: repoInfo.details?.sizeBytes || 0,
    id: repoInfo?.repo?.name || '',
    branches: repoInfo.branchesList.map(branchInfoToGQLBranch),
    // derived in field level resolver
    commits: [],
    sizeDisplay: formatBytes(repoInfo.details?.sizeBytes || 0),
  };
};

const getAggregateJobState = (jobs: Job[]) => {
  for (let i = 0; i < jobs.length; i++) {
    if (jobs[i].state !== GQLJobState.JOB_SUCCESS) {
      return jobs[i].state;
    }
  }

  return GQLJobState.JOB_SUCCESS;
};

export const jobInfosToGQLJobSet = (jobInfos: JobInfo.AsObject[]): JobSet => {
  const jobs = jobInfos
    .map(jobInfoToGQLJob)
    .sort((a, b) => (a.startedAt || 0) - (b.startedAt || 0));

  return {
    id: jobs[0].id,
    // grab the oldest jobs createdAt date
    createdAt: jobs[0].createdAt,
    state: getAggregateJobState(jobs),
    jobs,
  };
};

export const jobSetsToGQLJobSets = (
  jobSet: JobSetInfo.AsObject[],
): JobSet[] => {
  return jobSet.map((jobSet) => jobInfosToGQLJobSet(jobSet.jobsList));
};

export const logMessageToGQLLog = (logMessage: LogMessage.AsObject) => {
  return {
    message: logMessage.message,
    timestamp: logMessage.ts,
    user: logMessage.user,
  };
};

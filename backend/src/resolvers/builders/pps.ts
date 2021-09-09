import {
  Branch,
  RepoInfo,
  PipelineInfo,
  LogMessage,
  JobInfo,
  JobState,
  JobSetInfo,
} from '@pachyderm/node-pachyderm';
import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';
import fromPairs from 'lodash/fromPairs';
import isEmpty from 'lodash/isEmpty';
import sortBy from 'lodash/sortBy';

import formatBytes from '@dash-backend/lib/formatBytes';
import {
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import omitByDeep from '@dash-backend/lib/omitByDeep';
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

const deriveJSONJobDetails = (jobInfo: JobInfo.AsObject) => {
  const spec = {
    pipelineVersion: jobInfo.pipelineVersion,
    dataTotal: jobInfo.dataTotal,
    dataFailed: jobInfo.dataFailed,
    stats: jobInfo.stats,
    salt: jobInfo.details?.salt,
    datumTries: jobInfo.details?.datumTries,
  };

  const simplifiedSpec = omitByDeep(
    spec,
    (val, _) => !val || (typeof val === 'object' && isEmpty(val)),
  );

  return JSON.stringify(simplifiedSpec, null, 2);
};

const deriveJSONPipelineSpec = (pipelineInfo: PipelineInfo.AsObject) => {
  const spec = {
    metadata: pipelineInfo.details?.metadata,
    transform: pipelineInfo.details?.transform,
    parallelismSpec: pipelineInfo.details?.parallelismSpec,
    resourceRequests: pipelineInfo.details?.resourceRequests,
    resourceLimits: pipelineInfo.details?.resourceLimits,
    sidecarResourceLimits: pipelineInfo.details?.sidecarResourceLimits,
    input: pipelineInfo.details?.input,
    autoscaling: pipelineInfo.details?.autoscaling,
    reprocessSpec: pipelineInfo.details?.reprocessSpec,
    schedulingSpec: pipelineInfo.details?.schedulingSpec,
    podSpec: pipelineInfo.details?.podSpec,
    podPatch: pipelineInfo.details?.podPatch,
  };

  const simplifiedSpec = omitByDeep(
    spec,
    (val, _) => !val || (typeof val === 'object' && isEmpty(val)),
  );

  return JSON.stringify(simplifiedSpec, null, 2);
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
    datumTimeoutS: pipelineInfo.details?.datumTimeout?.seconds,
    datumTries: pipelineInfo.details?.datumTries || 0,
    jobTimeoutS: pipelineInfo.details?.jobTimeout?.seconds,
    outputBranch: pipelineInfo.details?.outputBranch || '',
    s3OutputRepo:
      pipelineInfo.details?.s3Out && pipelineInfo.pipeline
        ? `s3//${pipelineInfo.pipeline.name}`
        : undefined,
    egress: Boolean(pipelineInfo.details?.egress),
    jsonSpec: deriveJSONPipelineSpec(pipelineInfo),
    reason: pipelineInfo.reason,
  };
};

export const jobInfoToGQLJob = (jobInfo: JobInfo.AsObject): Job => {
  return {
    id: jobInfo.job?.id || '',
    state: toGQLJobState(jobInfo.state),
    reason: jobInfo.reason,
    createdAt: jobInfo.created?.seconds,
    startedAt: jobInfo.started?.seconds,
    finishedAt: jobInfo.finished?.seconds,
    pipelineName: jobInfo.job?.pipeline?.name || '',
    transform: jobInfo.details?.transform,
    inputString: jobInfo.details?.input
      ? JSON.stringify(jobInfo.details?.input, null, 2)
      : undefined,
    inputBranch: jobInfo.details?.input?.pfs?.branch,
    outputBranch: jobInfo.outputCommit?.branch?.name,
    jsonDetails: deriveJSONJobDetails(jobInfo),
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

const timestampToNanos = (timestamp: Timestamp.AsObject) => {
  return timestamp.seconds * 1e9 + timestamp.nanos;
};

export const jobInfosToGQLJobSet = (jobInfos: JobInfo.AsObject[]): JobSet => {
  const sortedJobInfos = sortBy(jobInfos, [
    (jobInfo) => (jobInfo.started ? timestampToNanos(jobInfo.started) : 0),
    (jobInfo) => jobInfo.job?.pipeline?.name || '',
  ]);

  const jobs = sortedJobInfos.map(jobInfoToGQLJob);

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

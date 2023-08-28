import isEmpty from 'lodash/isEmpty';

import formatBytes from '@dash-backend/lib/formatBytes';
import {
  toGQLDatumState,
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import hasRepoReadPermissions from '@dash-backend/lib/hasRepoReadPermissions';
import {
  gqlPipelineStateToNodeState,
  gqlJobStateToNodeState,
} from '@dash-backend/lib/nodeStateMappers';
import omitByDeep from '@dash-backend/lib/omitByDeep';
import removeGeneratedSuffixes from '@dash-backend/lib/removeGeneratedSuffixes';
import sortJobInfos from '@dash-backend/lib/sortJobInfos';
import {
  Branch,
  DatumInfo,
  JobInfo,
  JobSetInfo,
  LogMessage,
  PipelineInfo,
  RepoInfo,
} from '@dash-backend/proto';
import {
  Branch as GQLBranch,
  Datum,
  Job,
  JobSet,
  JobState as GQLJobState,
  Pipeline,
  PipelineType,
  Repo,
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
    removeGeneratedSuffixes(spec),
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
    service: pipelineInfo.details?.service,
    autoscaling: pipelineInfo.details?.autoscaling,
    reprocessSpec: pipelineInfo.details?.reprocessSpec,
    schedulingSpec: pipelineInfo.details?.schedulingSpec,
    podSpec: pipelineInfo.details?.podSpec,
    podPatch: pipelineInfo.details?.podPatch,
    egress: pipelineInfo.details?.egress,
  };

  const simplifiedSpec = omitByDeep(
    removeGeneratedSuffixes(spec),
    (val, _) => !val || (typeof val === 'object' && isEmpty(val)),
  );

  return JSON.stringify(simplifiedSpec, null, 2);
};

export const pipelineInfoToGQLPipeline = (
  pipelineInfo: PipelineInfo.AsObject,
): Pipeline => {
  return {
    id: `${pipelineInfo.pipeline?.project?.name || ''}_${
      pipelineInfo.pipeline?.name || ''
    }`,
    name: pipelineInfo.pipeline?.name || '',
    description: pipelineInfo.details?.description || '',
    version: pipelineInfo.version,
    createdAt: pipelineInfo.details?.createdAt?.seconds || 0,
    state: toGQLPipelineState(pipelineInfo.state),
    nodeState: gqlPipelineStateToNodeState(
      toGQLPipelineState(pipelineInfo.state),
    ),
    stopped: pipelineInfo.stopped,
    recentError: pipelineInfo.details?.recentError,
    lastJobState: toGQLJobState(pipelineInfo.lastJobState),
    lastJobNodeState: gqlJobStateToNodeState(
      toGQLJobState(pipelineInfo.lastJobState),
    ),
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
    nodeState: gqlJobStateToNodeState(toGQLJobState(jobInfo.state)),
    reason: jobInfo.reason,
    createdAt: jobInfo.created?.seconds,
    startedAt: jobInfo.started?.seconds,
    finishedAt: jobInfo.finished?.seconds,
    restarts: jobInfo.restart,
    pipelineName: jobInfo.job?.pipeline?.name || '',
    pipelineVersion: jobInfo.pipelineVersion,
    transform: jobInfo.details?.transform,
    inputString: jobInfo.details?.input
      ? JSON.stringify(removeGeneratedSuffixes(jobInfo.details?.input), null, 2)
      : undefined,
    transformString: jobInfo.details?.transform
      ? JSON.stringify(
          removeGeneratedSuffixes(jobInfo.details?.transform),
          null,
          2,
        )
      : undefined,
    inputBranch: jobInfo.details?.input?.pfs?.branch,
    outputBranch: jobInfo.outputCommit?.branch?.name,
    outputCommit: jobInfo.outputCommit?.id,
    jsonDetails: deriveJSONJobDetails(jobInfo),
    uploadBytesDisplay: formatBytes(jobInfo.stats?.uploadBytes || 0),
    downloadBytesDisplay: formatBytes(jobInfo.stats?.downloadBytes || 0),
    dataFailed: jobInfo.dataFailed,
    dataProcessed: jobInfo.dataProcessed,
    dataRecovered: jobInfo.dataRecovered,
    dataSkipped: jobInfo.dataSkipped,
    dataTotal: jobInfo.dataTotal,
  };
};

export const branchInfoToGQLBranch = (branch: Branch.AsObject): GQLBranch => {
  return {
    name: branch.name,
    repo: (branch.repo && {name: branch.repo?.name}) || undefined,
  };
};

export const repoInfoToGQLRepo = (repoInfo: RepoInfo.AsObject): Repo => {
  // repoInfo.details.sizeBytes is deprecated in favor of repoInfo.sizeBytesUpperBound
  return {
    createdAt: repoInfo.created?.seconds || 0,
    description: repoInfo.description,
    name: repoInfo.repo?.name || '',
    sizeBytes: repoInfo.sizeBytesUpperBound,
    id: repoInfo?.repo?.name || '',
    access: hasRepoReadPermissions(repoInfo?.authInfo?.permissionsList),
    branches: repoInfo.branchesList.map(branchInfoToGQLBranch),
    sizeDisplay: formatBytes(repoInfo.sizeBytesUpperBound),
    projectId: repoInfo.repo?.project?.name || '',
    authInfo: {rolesList: repoInfo.authInfo?.rolesList},
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

export const jobInfosToGQLJobSet = (
  jobInfos: JobInfo.AsObject[],
  id: string,
): JobSet => {
  const sortedJobInfos = sortJobInfos(jobInfos);

  const jobs = sortedJobInfos.map(jobInfoToGQLJob);

  const createdAt =
    jobs.length > 0 ? jobs.find((job) => job.createdAt)?.createdAt : null;
  const startedAt =
    jobs.length > 0 ? jobs.find((job) => job.startedAt)?.startedAt : null;
  const finishTimes = jobs
    .map((j) => Number(j.finishedAt))
    .filter((finishedTime) => !isNaN(finishedTime));
  const finishedAt = finishTimes.length > 0 ? Math.max(...finishTimes) : null;
  const inProgress = Boolean(jobs.find((j) => !j.finishedAt));

  return {
    id,
    createdAt,
    startedAt,
    finishedAt,
    inProgress,
    state: getAggregateJobState(jobs),
    jobs,
  };
};

export const jobsMapToGQLJobSets = (
  jobsMap: Map<string, JobInfo.AsObject[]>,
): JobSet[] => {
  const data = [];
  for (const [jobId, jobs] of jobsMap) {
    data.push(jobInfosToGQLJobSet(jobs, jobId));
  }

  return data;
};

export const jobSetsToGQLJobSets = (
  jobSet: JobSetInfo.AsObject[],
): JobSet[] => {
  return jobSet.map((jobSet) =>
    jobInfosToGQLJobSet(jobSet.jobsList, jobSet.jobSet?.id || ''),
  );
};

export const logMessageToGQLLog = (logMessage: LogMessage.AsObject) => {
  return {
    message: logMessage.message,
    timestamp: logMessage.ts,
    user: logMessage.user,
  };
};

export const datumInfoToGQLDatum = (
  datumInfo: DatumInfo.AsObject,
  jobId: string,
): Datum => {
  return {
    id: datumInfo.datum?.id || '',
    requestedJobId: jobId,
    jobId: datumInfo.datum?.job?.id,

    state: toGQLDatumState(datumInfo.state),

    downloadTimestamp: datumInfo.stats?.downloadTime,
    uploadTimestamp: datumInfo.stats?.uploadTime,
    processTimestamp: datumInfo.stats?.processTime,

    downloadBytes: datumInfo.stats?.downloadBytes,
    uploadBytes: datumInfo.stats?.uploadBytes,
  };
};

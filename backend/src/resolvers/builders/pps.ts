import isEmpty from 'lodash/isEmpty';

import formatBytes from '@dash-backend/lib/formatBytes';
import {
  toGQLJobState,
  toGQLPipelineState,
  toGQLDatumState,
} from '@dash-backend/lib/gqlEnumMappers';
import omitByDeep from '@dash-backend/lib/omitByDeep';
import removeGeneratedSuffixes from '@dash-backend/lib/removeGeneratedSuffixes';
import sortJobInfos from '@dash-backend/lib/sortJobInfos';
import {
  Branch,
  RepoInfo,
  PipelineInfo,
  LogMessage,
  JobInfo,
  JobSetInfo,
  DatumInfo,
} from '@dash-backend/proto';
import {
  Job,
  Pipeline,
  PipelineType,
  JobState as GQLJobState,
  Repo,
  Branch as GQLBranch,
  JobSet,
  Datum,
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
  return {
    createdAt: repoInfo.created?.seconds || 0,
    description: repoInfo.description,
    name: repoInfo.repo?.name || '',
    sizeBytes: repoInfo.details?.sizeBytes || 0,
    id: repoInfo?.repo?.name || '',
    branches: repoInfo.branchesList.map(branchInfoToGQLBranch),
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

export const jobInfosToGQLJobSet = (
  jobInfos: JobInfo.AsObject[],
  id: string,
): JobSet => {
  const sortedJobInfos = sortJobInfos(jobInfos);

  const jobs = sortedJobInfos.map(jobInfoToGQLJob);

  return {
    id,
    // grab the oldest jobs createdAt date
    createdAt:
      jobs.length > 0 ? jobs.find((job) => job.createdAt)?.createdAt : null,
    state: getAggregateJobState(jobs),
    jobs,
  };
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

export const datumInfoToGQLDatum = (datumInfo: DatumInfo.AsObject): Datum => {
  return {
    id: datumInfo.datum?.id || '',
    state: toGQLDatumState(datumInfo.state),
    downloadBytes: datumInfo.stats?.downloadBytes,
    uploadTime: datumInfo.stats?.uploadTime?.seconds,
    processTime: datumInfo.stats?.processTime?.seconds,
    downloadTime: datumInfo.stats?.downloadTime?.seconds,
  };
};

import {
  GetLogsRequest,
  ListDatumRequest,
  LogMessage,
} from '@dash-frontend/api/pps';

const clusterQueryKeys = {
  account: ['account'],
  version: ['version'],
  enterpriseState: ['enterpriseState'],
  clusterDefaults: ['clusterDefaults'],
  inspectCluster: ['inspectCluster'],
  allPipelines: ['allPipelines'],
  projects: ['projects'],
  jobSet: ({jobSetId}: {jobSetId?: string}) => ['jobSet', jobSetId],
  roleBinding: <ArgsType>({args}: {args: ArgsType}) => [args],
  authorize: <ArgsType>({args}: {args: ArgsType}) => [args],
};

const projectQueryKeys = {
  project: ({projectId}: {projectId?: string}) => [projectId],
  projectStatus: ({projectId}: {projectId?: string}) => [
    projectId,
    'projectStatus',
  ],
  projectDefaults: ({projectId}: {projectId: string}) => [
    projectId,
    'projectDefaults',
  ],
  pipelines: ({projectId}: {projectId?: string}) => [projectId, 'pipelines'],
  repos: ({projectId}: {projectId?: string}) => [projectId, 'repos'],
  jobs: <ArgsType>({projectId, args}: {projectId?: string; args: ArgsType}) => [
    projectId,
    'jobs',
    args,
  ],
  jobSets: ({projectId, args}: {projectId?: string; args: {limit: number}}) => [
    projectId,
    'jobSets',
    args,
  ],
};

const pipelineQueryKeys = {
  pipeline: ({
    projectId,
    pipelineId,
  }: {
    projectId?: string;
    pipelineId?: string;
  }) => [projectId, 'pipeline', pipelineId],
  job: ({
    projectId,
    pipelineId,
    jobId,
  }: {
    projectId?: string;
    pipelineId?: string;
    jobId?: string;
  }) => [...queryKeys.pipeline({projectId, pipelineId}), 'job', jobId],
  jobsByPipeline: ({
    projectId,
    pipelineId,
    args,
  }: {
    projectId?: string;
    pipelineId?: string;
    args: {limit?: number};
  }) => [
    ...queryKeys.pipeline({projectId, pipelineId}),
    'jobsByPipeline',
    args,
  ],
  datum: ({
    projectId,
    pipelineId,
    jobId,
    datumId,
  }: {
    projectId?: string;
    pipelineId?: string;
    jobId?: string;
    datumId?: string;
  }) => [...queryKeys.job({projectId, pipelineId, jobId}), 'datum', datumId],
  datums: ({
    projectId,
    pipelineId,
    jobId,
    args,
  }: {
    projectId?: string;
    pipelineId?: string;
    jobId?: string;
    args: ListDatumRequest;
  }) => [...queryKeys.job({projectId, pipelineId, jobId}), 'datums', {args}],
};

const repoQueryKeys = {
  repo: ({projectId, repoId}: {projectId?: string; repoId?: string}) => [
    projectId,
    'repo',
    repoId,
  ],
  commits: <ArgsType>({
    projectId,
    repoId,
    args,
  }: {
    projectId?: string;
    repoId?: string;
    args: ArgsType;
  }) => [...queryKeys.repo({projectId, repoId}), 'commits', args],
  findCommits: <ArgsType>({
    projectId,
    repoId,
    args,
  }: {
    projectId?: string;
    repoId?: string;
    args: ArgsType;
  }) => [...queryKeys.repo({projectId, repoId}), 'findCommits', args],
  commitSearch: ({
    projectId,
    repoId,
    commitId,
  }: {
    projectId?: string;
    repoId?: string;
    commitId?: string;
  }) => [...queryKeys.repo({projectId, repoId}), 'commitSearch', commitId],
  commitDiff: <ArgsType>({
    projectId,
    repoId,
    commitId,
    args,
  }: {
    projectId?: string;
    repoId?: string;
    commitId?: string;
    args: ArgsType;
  }) => [
    ...queryKeys.repo({projectId, repoId}),
    'commit',
    commitId,
    'fileDiff',
    args,
  ],
  branches: ({projectId, repoId}: {projectId?: string; repoId?: string}) => [
    ...queryKeys.repo({projectId, repoId}),
    'branches',
  ],
  files: <ArgsType>({
    projectId,
    repoId,
    branchName,
    commitId,
    path,
    args,
  }: {
    projectId?: string;
    repoId?: string;
    branchName?: string;
    commitId?: string;
    path?: string;
    args: ArgsType;
  }) => [
    ...queryKeys.repo({projectId, repoId}),
    'branch',
    branchName,
    'commit',
    commitId,
    'path',
    path,
    args,
  ],
};

const logQueryKeys = {
  log: ({
    pipelineName,
    projectName,
    datumId,
    jobId,
    req,
    args,
  }: {
    projectName: string;
    pipelineName: string;
    datumId: string;
    jobId: string;
    req: GetLogsRequest;
    args: {
      cursor?: LogMessage;
      limit?: number;
      since?: string;
    };
  }) => [projectName, 'log', pipelineName, jobId, datumId, req, args],
};

const queryKeys = {
  ...clusterQueryKeys,
  ...projectQueryKeys,
  ...pipelineQueryKeys,
  ...repoQueryKeys,
  ...logQueryKeys,
};

export default queryKeys;

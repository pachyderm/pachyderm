/* eslint-disable @typescript-eslint/ban-types */
/* eslint-disable import/no-duplicates */
/* eslint-disable @typescript-eslint/no-explicit-any */

import {GraphQLResolveInfo} from 'graphql';

import {Context} from '@dash-backend/lib/types';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends {[key: string]: unknown}> = {[K in keyof T]: T[K]};
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]?: Maybe<T[SubKey]>;
};
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]: Maybe<T[SubKey]>;
};
export type RequireFields<T, K extends keyof T> = Omit<T, K> & {
  [P in K]-?: NonNullable<T[P]>;
};
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
};

export type Account = {
  __typename?: 'Account';
  email: Scalars['String'];
  id: Scalars['ID'];
  name?: Maybe<Scalars['String']>;
};

export type AdminInfo = {
  __typename?: 'AdminInfo';
  clusterId?: Maybe<Scalars['String']>;
};

export type AuthConfig = {
  __typename?: 'AuthConfig';
  authEndpoint: Scalars['String'];
  clientId: Scalars['String'];
  pachdClientId: Scalars['String'];
};

export type Branch = {
  __typename?: 'Branch';
  name: Scalars['ID'];
  repo?: Maybe<RepoInfo>;
};

export type BranchInfo = {
  __typename?: 'BranchInfo';
  branch?: Maybe<Branch>;
  head?: Maybe<Commit>;
};

export type BranchInput = {
  name: Scalars['ID'];
  repo: RepoInput;
};

export type BranchQueryArgs = {
  branch: BranchInput;
  projectId: Scalars['String'];
};

export type BranchesQueryArgs = {
  projectId: Scalars['String'];
  repoName: Scalars['String'];
};

export type Commit = {
  __typename?: 'Commit';
  branch?: Maybe<Branch>;
  description?: Maybe<Scalars['String']>;
  diff?: Maybe<Diff>;
  finished: Scalars['Int'];
  hasLinkedJob: Scalars['Boolean'];
  id: Scalars['ID'];
  originKind?: Maybe<OriginKind>;
  repoName: Scalars['String'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
  started: Scalars['Int'];
};

export type CommitInput = {
  branch?: InputMaybe<BranchInput>;
  id: Scalars['ID'];
};

export type CommitQueryArgs = {
  branchName?: InputMaybe<Scalars['String']>;
  id?: InputMaybe<Scalars['ID']>;
  projectId: Scalars['String'];
  repoName: Scalars['String'];
  withDiff?: InputMaybe<Scalars['Boolean']>;
};

export type CommitSearchQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
  repoName: Scalars['String'];
};

export type CommitsQueryArgs = {
  branchName?: InputMaybe<Scalars['String']>;
  commitIdCursor?: InputMaybe<Scalars['String']>;
  cursor?: InputMaybe<TimestampInput>;
  number?: InputMaybe<Scalars['Int']>;
  originKind?: InputMaybe<OriginKind>;
  pipelineName?: InputMaybe<Scalars['String']>;
  projectId: Scalars['String'];
  repoName: Scalars['String'];
};

export type CreateBranchArgs = {
  branch?: InputMaybe<BranchInput>;
  head?: InputMaybe<CommitInput>;
  newCommitSet?: InputMaybe<Scalars['Boolean']>;
  projectId: Scalars['String'];
  provenance?: InputMaybe<Array<BranchInput>>;
};

export type CreatePipelineArgs = {
  crossList?: InputMaybe<Array<Pfs>>;
  description?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  pfs?: InputMaybe<Pfs>;
  projectId: Scalars['String'];
  transform: TransformInput;
  update?: InputMaybe<Scalars['Boolean']>;
};

export type CreateProjectArgs = {
  description?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
};

export type CreateRepoArgs = {
  description?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  projectId: Scalars['String'];
  update?: InputMaybe<Scalars['Boolean']>;
};

export type CronInput = {
  __typename?: 'CronInput';
  name: Scalars['String'];
  repo: Repo;
};

export type DagQueryArgs = {
  jobSetId?: InputMaybe<Scalars['ID']>;
  projectId: Scalars['ID'];
};

export type Datum = {
  __typename?: 'Datum';
  downloadBytes?: Maybe<Scalars['Float']>;
  downloadTimestamp?: Maybe<Timestamp>;
  id: Scalars['ID'];
  jobId?: Maybe<Scalars['ID']>;
  processTimestamp?: Maybe<Timestamp>;
  requestedJobId: Scalars['ID'];
  state: DatumState;
  uploadBytes?: Maybe<Scalars['Float']>;
  uploadTimestamp?: Maybe<Timestamp>;
};

export enum DatumFilter {
  FAILED = 'FAILED',
  RECOVERED = 'RECOVERED',
  SKIPPED = 'SKIPPED',
  SUCCESS = 'SUCCESS',
}

export type DatumQueryArgs = {
  id: Scalars['ID'];
  jobId: Scalars['ID'];
  pipelineId: Scalars['ID'];
  projectId: Scalars['String'];
};

export enum DatumState {
  FAILED = 'FAILED',
  RECOVERED = 'RECOVERED',
  SKIPPED = 'SKIPPED',
  STARTING = 'STARTING',
  SUCCESS = 'SUCCESS',
  UNKNOWN = 'UNKNOWN',
}

export type DatumsQueryArgs = {
  cursor?: InputMaybe<Scalars['String']>;
  filter?: InputMaybe<Array<DatumFilter>>;
  jobId: Scalars['ID'];
  limit?: InputMaybe<Scalars['Int']>;
  pipelineId: Scalars['ID'];
  projectId: Scalars['String'];
};

export type DeleteFileArgs = {
  branch: Scalars['String'];
  filePath: Scalars['String'];
  force?: InputMaybe<Scalars['Boolean']>;
  projectId: Scalars['String'];
  repo: Scalars['String'];
};

export type DeletePipelineArgs = {
  name: Scalars['String'];
  projectId: Scalars['String'];
};

export type DeleteProjectAndResourcesArgs = {
  name: Scalars['String'];
};

export type DeleteRepoArgs = {
  force?: InputMaybe<Scalars['Boolean']>;
  projectId: Scalars['String'];
  repo: RepoInput;
};

export type Diff = {
  __typename?: 'Diff';
  filesAdded: DiffCount;
  filesDeleted: DiffCount;
  filesUpdated: DiffCount;
  size: Scalars['Float'];
  sizeDisplay: Scalars['String'];
};

export type DiffCount = {
  __typename?: 'DiffCount';
  count: Scalars['Int'];
  sizeDelta: Scalars['Int'];
};

export type EnterpriseInfo = {
  __typename?: 'EnterpriseInfo';
  expiration: Scalars['Int'];
  state: EnterpriseState;
};

export enum EnterpriseState {
  ACTIVE = 'ACTIVE',
  EXPIRED = 'EXPIRED',
  HEARTBEAT_FAILED = 'HEARTBEAT_FAILED',
  NONE = 'NONE',
}

export type File = {
  __typename?: 'File';
  commitAction?: Maybe<FileCommitState>;
  commitId: Scalars['String'];
  committed?: Maybe<Timestamp>;
  download?: Maybe<Scalars['String']>;
  downloadDisabled?: Maybe<Scalars['Boolean']>;
  hash: Scalars['String'];
  path: Scalars['String'];
  repoName: Scalars['String'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
  type: FileType;
};

export enum FileCommitState {
  ADDED = 'ADDED',
  DELETED = 'DELETED',
  UPDATED = 'UPDATED',
}

export type FileFromUrl = {
  path: Scalars['String'];
  url: Scalars['String'];
};

export type FileQueryArgs = {
  branchName: Scalars['String'];
  commitId?: InputMaybe<Scalars['String']>;
  cursorPath?: InputMaybe<Scalars['String']>;
  limit?: InputMaybe<Scalars['Int']>;
  path?: InputMaybe<Scalars['String']>;
  projectId: Scalars['String'];
  repoName: Scalars['String'];
  reverse?: InputMaybe<Scalars['Boolean']>;
};

export type FileQueryResponse = {
  __typename?: 'FileQueryResponse';
  diff?: Maybe<Diff>;
  files: Array<File>;
};

export enum FileType {
  DIR = 'DIR',
  FILE = 'FILE',
  RESERVED = 'RESERVED',
}

export type FinishCommitArgs = {
  commit: OpenCommitInput;
  projectId: Scalars['String'];
};

export type GitInput = {
  __typename?: 'GitInput';
  name: Scalars['String'];
  url: Scalars['String'];
};

export type Input = {
  __typename?: 'Input';
  cronInput?: Maybe<CronInput>;
  crossedWith: Array<Scalars['String']>;
  gitInput?: Maybe<GitInput>;
  groupedWith: Array<Scalars['String']>;
  id: Scalars['ID'];
  joinedWith: Array<Scalars['String']>;
  pfsInput?: Maybe<PfsInput>;
  type: InputType;
  unionedWith: Array<Scalars['String']>;
};

export type InputPipeline = {
  __typename?: 'InputPipeline';
  id: Scalars['ID'];
};

export enum InputType {
  CRON = 'CRON',
  GIT = 'GIT',
  PFS = 'PFS',
}

export type Job = {
  __typename?: 'Job';
  createdAt?: Maybe<Scalars['Int']>;
  dataFailed: Scalars['Int'];
  dataProcessed: Scalars['Int'];
  dataRecovered: Scalars['Int'];
  dataSkipped: Scalars['Int'];
  dataTotal: Scalars['Int'];
  downloadBytesDisplay: Scalars['String'];
  finishedAt?: Maybe<Scalars['Int']>;
  id: Scalars['ID'];
  inputBranch?: Maybe<Scalars['String']>;
  inputString?: Maybe<Scalars['String']>;
  jsonDetails: Scalars['String'];
  nodeState: NodeState;
  outputBranch?: Maybe<Scalars['String']>;
  outputCommit?: Maybe<Scalars['String']>;
  pipelineName: Scalars['String'];
  reason?: Maybe<Scalars['String']>;
  restarts: Scalars['Int'];
  startedAt?: Maybe<Scalars['Int']>;
  state: JobState;
  transform?: Maybe<Transform>;
  transformString?: Maybe<Scalars['String']>;
  uploadBytesDisplay: Scalars['String'];
};

export type JobQueryArgs = {
  id: Scalars['ID'];
  pipelineName: Scalars['String'];
  projectId: Scalars['String'];
};

export type JobSet = {
  __typename?: 'JobSet';
  createdAt?: Maybe<Scalars['Int']>;
  finishedAt?: Maybe<Scalars['Int']>;
  id: Scalars['ID'];
  inProgress: Scalars['Boolean'];
  jobs: Array<Job>;
  startedAt?: Maybe<Scalars['Int']>;
  state: JobState;
};

export type JobSetQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
};

export type JobSetsQueryArgs = {
  cursor?: InputMaybe<TimestampInput>;
  limit?: InputMaybe<Scalars['Int']>;
  projectId: Scalars['String'];
  reverse?: InputMaybe<Scalars['Boolean']>;
};

export enum JobState {
  JOB_CREATED = 'JOB_CREATED',
  JOB_EGRESSING = 'JOB_EGRESSING',
  JOB_FAILURE = 'JOB_FAILURE',
  JOB_FINISHING = 'JOB_FINISHING',
  JOB_KILLED = 'JOB_KILLED',
  JOB_RUNNING = 'JOB_RUNNING',
  JOB_STARTING = 'JOB_STARTING',
  JOB_STATE_UNKNOWN = 'JOB_STATE_UNKNOWN',
  JOB_SUCCESS = 'JOB_SUCCESS',
  JOB_UNRUNNABLE = 'JOB_UNRUNNABLE',
}

export type JobsByPipelineQueryArgs = {
  limit?: InputMaybe<Scalars['Int']>;
  pipelineIds: Array<InputMaybe<Scalars['String']>>;
  projectId: Scalars['ID'];
};

export type JobsQueryArgs = {
  cursor?: InputMaybe<TimestampInput>;
  jobSetIds?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  limit?: InputMaybe<Scalars['Int']>;
  pipelineId?: InputMaybe<Scalars['String']>;
  pipelineIds?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  projectId: Scalars['ID'];
  reverse?: InputMaybe<Scalars['Boolean']>;
};

export type Log = {
  __typename?: 'Log';
  message: Scalars['String'];
  timestamp?: Maybe<Timestamp>;
  user: Scalars['Boolean'];
};

export type LogsArgs = {
  datumId?: InputMaybe<Scalars['String']>;
  jobId?: InputMaybe<Scalars['String']>;
  pipelineName: Scalars['String'];
  projectId: Scalars['String'];
  reverse?: InputMaybe<Scalars['Boolean']>;
  start?: InputMaybe<Scalars['Int']>;
};

export type Mutation = {
  __typename?: 'Mutation';
  createBranch: Branch;
  createPipeline: Pipeline;
  createProject: Project;
  createRepo: Repo;
  deleteFile: Scalars['ID'];
  deletePipeline?: Maybe<Scalars['Boolean']>;
  deleteProjectAndResources: Scalars['Boolean'];
  deleteRepo?: Maybe<Scalars['Boolean']>;
  exchangeCode: Tokens;
  finishCommit: Scalars['Boolean'];
  putFilesFromURLs: Array<Scalars['String']>;
  startCommit: OpenCommit;
  updateProject: Project;
};

export type MutationCreateBranchArgs = {
  args: CreateBranchArgs;
};

export type MutationCreatePipelineArgs = {
  args: CreatePipelineArgs;
};

export type MutationCreateProjectArgs = {
  args: CreateProjectArgs;
};

export type MutationCreateRepoArgs = {
  args: CreateRepoArgs;
};

export type MutationDeleteFileArgs = {
  args: DeleteFileArgs;
};

export type MutationDeletePipelineArgs = {
  args: DeletePipelineArgs;
};

export type MutationDeleteProjectAndResourcesArgs = {
  args: DeleteProjectAndResourcesArgs;
};

export type MutationDeleteRepoArgs = {
  args: DeleteRepoArgs;
};

export type MutationExchangeCodeArgs = {
  code: Scalars['String'];
};

export type MutationFinishCommitArgs = {
  args: FinishCommitArgs;
};

export type MutationPutFilesFromUrLsArgs = {
  args: PutFilesFromUrLsArgs;
};

export type MutationStartCommitArgs = {
  args: StartCommitArgs;
};

export type MutationUpdateProjectArgs = {
  args: UpdateProjectArgs;
};

export type NodeSelector = {
  __typename?: 'NodeSelector';
  key: Scalars['String'];
  value: Scalars['String'];
};

export enum NodeState {
  BUSY = 'BUSY',
  ERROR = 'ERROR',
  IDLE = 'IDLE',
  PAUSED = 'PAUSED',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
}

export enum NodeType {
  EGRESS = 'EGRESS',
  INPUT_REPO = 'INPUT_REPO',
  OUTPUT_REPO = 'OUTPUT_REPO',
  PIPELINE = 'PIPELINE',
}

export type OpenCommit = {
  __typename?: 'OpenCommit';
  branch: Branch;
  id: Scalars['ID'];
};

export type OpenCommitInput = {
  branch: BranchInput;
  id: Scalars['ID'];
};

export enum OriginKind {
  ALIAS = 'ALIAS',
  AUTO = 'AUTO',
  FSCK = 'FSCK',
  ORIGIN_KIND_UNKNOWN = 'ORIGIN_KIND_UNKNOWN',
  USER = 'USER',
}

export type Pfs = {
  branch?: InputMaybe<Scalars['String']>;
  glob?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  repo: RepoInput;
};

export type PfsInput = {
  __typename?: 'PFSInput';
  name: Scalars['String'];
  repo: Repo;
};

export type Pach = {
  __typename?: 'Pach';
  id: Scalars['ID'];
};

export type PageableCommit = {
  __typename?: 'PageableCommit';
  cursor?: Maybe<Timestamp>;
  items: Array<Commit>;
  parentCommit?: Maybe<Scalars['String']>;
};

export type PageableDatum = {
  __typename?: 'PageableDatum';
  cursor?: Maybe<Scalars['String']>;
  hasNextPage?: Maybe<Scalars['Boolean']>;
  items: Array<Datum>;
};

export type PageableFile = {
  __typename?: 'PageableFile';
  cursor?: Maybe<Scalars['String']>;
  diff?: Maybe<Diff>;
  files: Array<File>;
  hasNextPage?: Maybe<Scalars['Boolean']>;
};

export type PageableJob = {
  __typename?: 'PageableJob';
  cursor?: Maybe<Timestamp>;
  hasNextPage?: Maybe<Scalars['Boolean']>;
  items: Array<Job>;
};

export type PageableJobSet = {
  __typename?: 'PageableJobSet';
  cursor?: Maybe<Timestamp>;
  hasNextPage?: Maybe<Scalars['Boolean']>;
  items: Array<JobSet>;
};

export type Pipeline = {
  __typename?: 'Pipeline';
  createdAt: Scalars['Int'];
  datumTimeoutS?: Maybe<Scalars['Int']>;
  datumTries: Scalars['Int'];
  description?: Maybe<Scalars['String']>;
  egress: Scalars['Boolean'];
  id: Scalars['ID'];
  jobTimeoutS?: Maybe<Scalars['Int']>;
  jsonSpec: Scalars['String'];
  lastJobNodeState?: Maybe<NodeState>;
  lastJobState?: Maybe<JobState>;
  name: Scalars['String'];
  nodeState: NodeState;
  outputBranch: Scalars['String'];
  reason?: Maybe<Scalars['String']>;
  recentError?: Maybe<Scalars['String']>;
  s3OutputRepo?: Maybe<Scalars['String']>;
  state: PipelineState;
  stopped: Scalars['Boolean'];
  type: PipelineType;
  version: Scalars['Int'];
};

export type PipelineQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
};

export enum PipelineState {
  PIPELINE_CRASHING = 'PIPELINE_CRASHING',
  PIPELINE_FAILURE = 'PIPELINE_FAILURE',
  PIPELINE_PAUSED = 'PIPELINE_PAUSED',
  PIPELINE_RESTARTING = 'PIPELINE_RESTARTING',
  PIPELINE_RUNNING = 'PIPELINE_RUNNING',
  PIPELINE_STANDBY = 'PIPELINE_STANDBY',
  PIPELINE_STARTING = 'PIPELINE_STARTING',
  PIPELINE_STATE_UNKNOWN = 'PIPELINE_STATE_UNKNOWN',
}

export enum PipelineType {
  SERVICE = 'SERVICE',
  SPOUT = 'SPOUT',
  STANDARD = 'STANDARD',
}

export type PipelinesQueryArgs = {
  jobSetId?: InputMaybe<Scalars['ID']>;
  projectIds: Array<Scalars['String']>;
};

export type Project = {
  __typename?: 'Project';
  description?: Maybe<Scalars['String']>;
  id: Scalars['String'];
  status: ProjectStatus;
};

export type ProjectDetails = {
  __typename?: 'ProjectDetails';
  jobSets: Array<JobSet>;
  pipelineCount: Scalars['Int'];
  repoCount: Scalars['Int'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
};

export type ProjectDetailsQueryArgs = {
  jobSetsLimit?: InputMaybe<Scalars['Int']>;
  projectId: Scalars['String'];
};

export enum ProjectStatus {
  HEALTHY = 'HEALTHY',
  UNHEALTHY = 'UNHEALTHY',
}

export type ProjectWithoutStatus = {
  __typename?: 'ProjectWithoutStatus';
  description?: Maybe<Scalars['String']>;
  id: Scalars['String'];
  status?: Maybe<ProjectStatus>;
};

export type PutFilesFromUrLsArgs = {
  branch: Scalars['String'];
  files: Array<FileFromUrl>;
  projectId: Scalars['String'];
  repo: Scalars['String'];
};

export type Query = {
  __typename?: 'Query';
  account: Account;
  adminInfo: AdminInfo;
  authConfig: AuthConfig;
  branch: Branch;
  branches: Array<Maybe<Branch>>;
  commit?: Maybe<Commit>;
  commitSearch?: Maybe<Commit>;
  commits: PageableCommit;
  dag: Array<Vertex>;
  datum: Datum;
  datumSearch?: Maybe<Datum>;
  datums: PageableDatum;
  enterpriseInfo: EnterpriseInfo;
  files: PageableFile;
  job: Job;
  jobSet: JobSet;
  jobSets: PageableJobSet;
  jobs: PageableJob;
  jobsByPipeline: Array<Job>;
  loggedIn: Scalars['Boolean'];
  logs: Array<Maybe<Log>>;
  pipeline: Pipeline;
  pipelines: Array<Maybe<Pipeline>>;
  project: Project;
  projectDetails: ProjectDetails;
  projects: Array<Project>;
  repo: Repo;
  repos: Array<Maybe<Repo>>;
  searchResults: SearchResults;
  workspaceLogs: Array<Maybe<Log>>;
};

export type QueryBranchArgs = {
  args: BranchQueryArgs;
};

export type QueryBranchesArgs = {
  args: BranchesQueryArgs;
};

export type QueryCommitArgs = {
  args: CommitQueryArgs;
};

export type QueryCommitSearchArgs = {
  args: CommitSearchQueryArgs;
};

export type QueryCommitsArgs = {
  args: CommitsQueryArgs;
};

export type QueryDagArgs = {
  args: DagQueryArgs;
};

export type QueryDatumArgs = {
  args: DatumQueryArgs;
};

export type QueryDatumSearchArgs = {
  args: DatumQueryArgs;
};

export type QueryDatumsArgs = {
  args: DatumsQueryArgs;
};

export type QueryFilesArgs = {
  args: FileQueryArgs;
};

export type QueryJobArgs = {
  args: JobQueryArgs;
};

export type QueryJobSetArgs = {
  args: JobSetQueryArgs;
};

export type QueryJobSetsArgs = {
  args: JobSetsQueryArgs;
};

export type QueryJobsArgs = {
  args: JobsQueryArgs;
};

export type QueryJobsByPipelineArgs = {
  args: JobsByPipelineQueryArgs;
};

export type QueryLogsArgs = {
  args: LogsArgs;
};

export type QueryPipelineArgs = {
  args: PipelineQueryArgs;
};

export type QueryPipelinesArgs = {
  args: PipelinesQueryArgs;
};

export type QueryProjectArgs = {
  id: Scalars['ID'];
};

export type QueryProjectDetailsArgs = {
  args: ProjectDetailsQueryArgs;
};

export type QueryRepoArgs = {
  args: RepoQueryArgs;
};

export type QueryReposArgs = {
  args: ReposQueryArgs;
};

export type QuerySearchResultsArgs = {
  args: SearchResultQueryArgs;
};

export type QueryWorkspaceLogsArgs = {
  args: WorkspaceLogsArgs;
};

export type Repo = {
  __typename?: 'Repo';
  access: Scalars['Boolean'];
  branches: Array<Branch>;
  createdAt: Scalars['Int'];
  description: Scalars['String'];
  id: Scalars['ID'];
  lastCommit?: Maybe<Commit>;
  linkedPipeline?: Maybe<Pipeline>;
  name: Scalars['ID'];
  projectId: Scalars['String'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
};

export type RepoInfo = {
  __typename?: 'RepoInfo';
  name?: Maybe<Scalars['String']>;
  type?: Maybe<Scalars['String']>;
};

export type RepoInput = {
  name: Scalars['ID'];
  type?: InputMaybe<Scalars['String']>;
};

export type RepoQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
};

export type ReposQueryArgs = {
  jobSetId?: InputMaybe<Scalars['ID']>;
  projectId: Scalars['String'];
};

export type SchedulingSpec = {
  __typename?: 'SchedulingSpec';
  nodeSelectorMap: Array<NodeSelector>;
  priorityClassName: Scalars['String'];
};

export type SearchResultQueryArgs = {
  globalIdFilter?: InputMaybe<Scalars['ID']>;
  limit?: InputMaybe<Scalars['Int']>;
  projectId: Scalars['String'];
  query: Scalars['String'];
};

export type SearchResults = {
  __typename?: 'SearchResults';
  jobSet?: Maybe<JobSet>;
  pipelines: Array<Pipeline>;
  repos: Array<Repo>;
};

export type StartCommitArgs = {
  branchName: Scalars['String'];
  projectId: Scalars['String'];
  repoName: Scalars['String'];
};

export type Subscription = {
  __typename?: 'Subscription';
  dags: Array<Vertex>;
  logs: Log;
  workspaceLogs: Log;
};

export type SubscriptionDagsArgs = {
  args: DagQueryArgs;
};

export type SubscriptionLogsArgs = {
  args: LogsArgs;
};

export type SubscriptionWorkspaceLogsArgs = {
  args: WorkspaceLogsArgs;
};

export type Timestamp = {
  __typename?: 'Timestamp';
  nanos: Scalars['Int'];
  seconds: Scalars['Int'];
};

export type TimestampInput = {
  nanos: Scalars['Int'];
  seconds: Scalars['Int'];
};

export type Tokens = {
  __typename?: 'Tokens';
  idToken: Scalars['String'];
  pachToken: Scalars['String'];
};

export type Transform = {
  __typename?: 'Transform';
  cmdList: Array<Scalars['String']>;
  debug: Scalars['Boolean'];
  image: Scalars['String'];
};

export type TransformInput = {
  cmdList: Array<Scalars['String']>;
  image: Scalars['String'];
  stdinList?: InputMaybe<Array<Scalars['String']>>;
};

export type UpdateProjectArgs = {
  description: Scalars['String'];
  name: Scalars['String'];
};

export type Vertex = {
  __typename?: 'Vertex';
  access: Scalars['Boolean'];
  createdAt?: Maybe<Scalars['Int']>;
  id: Scalars['String'];
  jobState?: Maybe<NodeState>;
  name: Scalars['String'];
  parents: Array<Scalars['String']>;
  state?: Maybe<NodeState>;
  type: NodeType;
};

export type WorkspaceLogsArgs = {
  projectId: Scalars['String'];
  start?: InputMaybe<Scalars['Int']>;
};

export type WithIndex<TObject> = TObject & Record<string, any>;
export type ResolversObject<TObject> = WithIndex<TObject>;

export type ResolverTypeWrapper<T> = Promise<T> | T;

export type ResolverWithResolve<TResult, TParent, TContext, TArgs> = {
  resolve: ResolverFn<TResult, TParent, TContext, TArgs>;
};
export type Resolver<TResult, TParent = {}, TContext = {}, TArgs = {}> =
  | ResolverFn<TResult, TParent, TContext, TArgs>
  | ResolverWithResolve<TResult, TParent, TContext, TArgs>;

export type ResolverFn<TResult, TParent, TContext, TArgs> = (
  parent: TParent,
  args: TArgs,
  context: TContext,
  info: GraphQLResolveInfo,
) => Promise<TResult> | TResult;

export type SubscriptionSubscribeFn<TResult, TParent, TContext, TArgs> = (
  parent: TParent,
  args: TArgs,
  context: TContext,
  info: GraphQLResolveInfo,
) => AsyncIterable<TResult> | Promise<AsyncIterable<TResult>>;

export type SubscriptionResolveFn<TResult, TParent, TContext, TArgs> = (
  parent: TParent,
  args: TArgs,
  context: TContext,
  info: GraphQLResolveInfo,
) => TResult | Promise<TResult>;

export interface SubscriptionSubscriberObject<
  TResult,
  TKey extends string,
  TParent,
  TContext,
  TArgs,
> {
  subscribe: SubscriptionSubscribeFn<
    {[key in TKey]: TResult},
    TParent,
    TContext,
    TArgs
  >;
  resolve?: SubscriptionResolveFn<
    TResult,
    {[key in TKey]: TResult},
    TContext,
    TArgs
  >;
}

export interface SubscriptionResolverObject<TResult, TParent, TContext, TArgs> {
  subscribe: SubscriptionSubscribeFn<any, TParent, TContext, TArgs>;
  resolve: SubscriptionResolveFn<TResult, any, TContext, TArgs>;
}

export type SubscriptionObject<
  TResult,
  TKey extends string,
  TParent,
  TContext,
  TArgs,
> =
  | SubscriptionSubscriberObject<TResult, TKey, TParent, TContext, TArgs>
  | SubscriptionResolverObject<TResult, TParent, TContext, TArgs>;

export type SubscriptionResolver<
  TResult,
  TKey extends string,
  TParent = {},
  TContext = {},
  TArgs = {},
> =
  | ((
      ...args: any[]
    ) => SubscriptionObject<TResult, TKey, TParent, TContext, TArgs>)
  | SubscriptionObject<TResult, TKey, TParent, TContext, TArgs>;

export type TypeResolveFn<TTypes, TParent = {}, TContext = {}> = (
  parent: TParent,
  context: TContext,
  info: GraphQLResolveInfo,
) => Maybe<TTypes> | Promise<Maybe<TTypes>>;

export type IsTypeOfResolverFn<T = {}, TContext = {}> = (
  obj: T,
  context: TContext,
  info: GraphQLResolveInfo,
) => boolean | Promise<boolean>;

export type NextResolverFn<T> = () => Promise<T>;

export type DirectiveResolverFn<
  TResult = {},
  TParent = {},
  TContext = {},
  TArgs = {},
> = (
  next: NextResolverFn<TResult>,
  parent: TParent,
  args: TArgs,
  context: TContext,
  info: GraphQLResolveInfo,
) => TResult | Promise<TResult>;

/** Mapping between all available schema types and the resolvers types */
export type ResolversTypes = ResolversObject<{
  Account: ResolverTypeWrapper<Account>;
  AdminInfo: ResolverTypeWrapper<AdminInfo>;
  AuthConfig: ResolverTypeWrapper<AuthConfig>;
  Boolean: ResolverTypeWrapper<Scalars['Boolean']>;
  Branch: ResolverTypeWrapper<Branch>;
  BranchInfo: ResolverTypeWrapper<BranchInfo>;
  BranchInput: BranchInput;
  BranchQueryArgs: BranchQueryArgs;
  BranchesQueryArgs: BranchesQueryArgs;
  Commit: ResolverTypeWrapper<Commit>;
  CommitInput: CommitInput;
  CommitQueryArgs: CommitQueryArgs;
  CommitSearchQueryArgs: CommitSearchQueryArgs;
  CommitsQueryArgs: CommitsQueryArgs;
  CreateBranchArgs: CreateBranchArgs;
  CreatePipelineArgs: CreatePipelineArgs;
  CreateProjectArgs: CreateProjectArgs;
  CreateRepoArgs: CreateRepoArgs;
  CronInput: ResolverTypeWrapper<CronInput>;
  DagQueryArgs: DagQueryArgs;
  Datum: ResolverTypeWrapper<Datum>;
  DatumFilter: DatumFilter;
  DatumQueryArgs: DatumQueryArgs;
  DatumState: DatumState;
  DatumsQueryArgs: DatumsQueryArgs;
  DeleteFileArgs: DeleteFileArgs;
  DeletePipelineArgs: DeletePipelineArgs;
  DeleteProjectAndResourcesArgs: DeleteProjectAndResourcesArgs;
  DeleteRepoArgs: DeleteRepoArgs;
  Diff: ResolverTypeWrapper<Diff>;
  DiffCount: ResolverTypeWrapper<DiffCount>;
  EnterpriseInfo: ResolverTypeWrapper<EnterpriseInfo>;
  EnterpriseState: EnterpriseState;
  File: ResolverTypeWrapper<File>;
  FileCommitState: FileCommitState;
  FileFromURL: FileFromUrl;
  FileQueryArgs: FileQueryArgs;
  FileQueryResponse: ResolverTypeWrapper<FileQueryResponse>;
  FileType: FileType;
  FinishCommitArgs: FinishCommitArgs;
  Float: ResolverTypeWrapper<Scalars['Float']>;
  GitInput: ResolverTypeWrapper<GitInput>;
  ID: ResolverTypeWrapper<Scalars['ID']>;
  Input: ResolverTypeWrapper<Input>;
  InputPipeline: ResolverTypeWrapper<InputPipeline>;
  InputType: InputType;
  Int: ResolverTypeWrapper<Scalars['Int']>;
  Job: ResolverTypeWrapper<Job>;
  JobQueryArgs: JobQueryArgs;
  JobSet: ResolverTypeWrapper<JobSet>;
  JobSetQueryArgs: JobSetQueryArgs;
  JobSetsQueryArgs: JobSetsQueryArgs;
  JobState: JobState;
  JobsByPipelineQueryArgs: JobsByPipelineQueryArgs;
  JobsQueryArgs: JobsQueryArgs;
  Log: ResolverTypeWrapper<Log>;
  LogsArgs: LogsArgs;
  Mutation: ResolverTypeWrapper<{}>;
  NodeSelector: ResolverTypeWrapper<NodeSelector>;
  NodeState: NodeState;
  NodeType: NodeType;
  OpenCommit: ResolverTypeWrapper<OpenCommit>;
  OpenCommitInput: OpenCommitInput;
  OriginKind: OriginKind;
  PFS: Pfs;
  PFSInput: ResolverTypeWrapper<PfsInput>;
  Pach: ResolverTypeWrapper<Pach>;
  PageableCommit: ResolverTypeWrapper<PageableCommit>;
  PageableDatum: ResolverTypeWrapper<PageableDatum>;
  PageableFile: ResolverTypeWrapper<PageableFile>;
  PageableJob: ResolverTypeWrapper<PageableJob>;
  PageableJobSet: ResolverTypeWrapper<PageableJobSet>;
  Pipeline: ResolverTypeWrapper<Pipeline>;
  PipelineQueryArgs: PipelineQueryArgs;
  PipelineState: PipelineState;
  PipelineType: PipelineType;
  PipelinesQueryArgs: PipelinesQueryArgs;
  Project: ResolverTypeWrapper<Project>;
  ProjectDetails: ResolverTypeWrapper<ProjectDetails>;
  ProjectDetailsQueryArgs: ProjectDetailsQueryArgs;
  ProjectStatus: ProjectStatus;
  ProjectWithoutStatus: ResolverTypeWrapper<ProjectWithoutStatus>;
  PutFilesFromURLsArgs: PutFilesFromUrLsArgs;
  Query: ResolverTypeWrapper<{}>;
  Repo: ResolverTypeWrapper<Repo>;
  RepoInfo: ResolverTypeWrapper<RepoInfo>;
  RepoInput: RepoInput;
  RepoQueryArgs: RepoQueryArgs;
  ReposQueryArgs: ReposQueryArgs;
  SchedulingSpec: ResolverTypeWrapper<SchedulingSpec>;
  SearchResultQueryArgs: SearchResultQueryArgs;
  SearchResults: ResolverTypeWrapper<SearchResults>;
  StartCommitArgs: StartCommitArgs;
  String: ResolverTypeWrapper<Scalars['String']>;
  Subscription: ResolverTypeWrapper<{}>;
  Timestamp: ResolverTypeWrapper<Timestamp>;
  TimestampInput: TimestampInput;
  Tokens: ResolverTypeWrapper<Tokens>;
  Transform: ResolverTypeWrapper<Transform>;
  TransformInput: TransformInput;
  UpdateProjectArgs: UpdateProjectArgs;
  Vertex: ResolverTypeWrapper<Vertex>;
  WorkspaceLogsArgs: WorkspaceLogsArgs;
}>;

/** Mapping between all available schema types and the resolvers parents */
export type ResolversParentTypes = ResolversObject<{
  Account: Account;
  AdminInfo: AdminInfo;
  AuthConfig: AuthConfig;
  Boolean: Scalars['Boolean'];
  Branch: Branch;
  BranchInfo: BranchInfo;
  BranchInput: BranchInput;
  BranchQueryArgs: BranchQueryArgs;
  BranchesQueryArgs: BranchesQueryArgs;
  Commit: Commit;
  CommitInput: CommitInput;
  CommitQueryArgs: CommitQueryArgs;
  CommitSearchQueryArgs: CommitSearchQueryArgs;
  CommitsQueryArgs: CommitsQueryArgs;
  CreateBranchArgs: CreateBranchArgs;
  CreatePipelineArgs: CreatePipelineArgs;
  CreateProjectArgs: CreateProjectArgs;
  CreateRepoArgs: CreateRepoArgs;
  CronInput: CronInput;
  DagQueryArgs: DagQueryArgs;
  Datum: Datum;
  DatumQueryArgs: DatumQueryArgs;
  DatumsQueryArgs: DatumsQueryArgs;
  DeleteFileArgs: DeleteFileArgs;
  DeletePipelineArgs: DeletePipelineArgs;
  DeleteProjectAndResourcesArgs: DeleteProjectAndResourcesArgs;
  DeleteRepoArgs: DeleteRepoArgs;
  Diff: Diff;
  DiffCount: DiffCount;
  EnterpriseInfo: EnterpriseInfo;
  File: File;
  FileFromURL: FileFromUrl;
  FileQueryArgs: FileQueryArgs;
  FileQueryResponse: FileQueryResponse;
  FinishCommitArgs: FinishCommitArgs;
  Float: Scalars['Float'];
  GitInput: GitInput;
  ID: Scalars['ID'];
  Input: Input;
  InputPipeline: InputPipeline;
  Int: Scalars['Int'];
  Job: Job;
  JobQueryArgs: JobQueryArgs;
  JobSet: JobSet;
  JobSetQueryArgs: JobSetQueryArgs;
  JobSetsQueryArgs: JobSetsQueryArgs;
  JobsByPipelineQueryArgs: JobsByPipelineQueryArgs;
  JobsQueryArgs: JobsQueryArgs;
  Log: Log;
  LogsArgs: LogsArgs;
  Mutation: {};
  NodeSelector: NodeSelector;
  OpenCommit: OpenCommit;
  OpenCommitInput: OpenCommitInput;
  PFS: Pfs;
  PFSInput: PfsInput;
  Pach: Pach;
  PageableCommit: PageableCommit;
  PageableDatum: PageableDatum;
  PageableFile: PageableFile;
  PageableJob: PageableJob;
  PageableJobSet: PageableJobSet;
  Pipeline: Pipeline;
  PipelineQueryArgs: PipelineQueryArgs;
  PipelinesQueryArgs: PipelinesQueryArgs;
  Project: Project;
  ProjectDetails: ProjectDetails;
  ProjectDetailsQueryArgs: ProjectDetailsQueryArgs;
  ProjectWithoutStatus: ProjectWithoutStatus;
  PutFilesFromURLsArgs: PutFilesFromUrLsArgs;
  Query: {};
  Repo: Repo;
  RepoInfo: RepoInfo;
  RepoInput: RepoInput;
  RepoQueryArgs: RepoQueryArgs;
  ReposQueryArgs: ReposQueryArgs;
  SchedulingSpec: SchedulingSpec;
  SearchResultQueryArgs: SearchResultQueryArgs;
  SearchResults: SearchResults;
  StartCommitArgs: StartCommitArgs;
  String: Scalars['String'];
  Subscription: {};
  Timestamp: Timestamp;
  TimestampInput: TimestampInput;
  Tokens: Tokens;
  Transform: Transform;
  TransformInput: TransformInput;
  UpdateProjectArgs: UpdateProjectArgs;
  Vertex: Vertex;
  WorkspaceLogsArgs: WorkspaceLogsArgs;
}>;

export type AccountResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Account'] = ResolversParentTypes['Account'],
> = ResolversObject<{
  email?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type AdminInfoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['AdminInfo'] = ResolversParentTypes['AdminInfo'],
> = ResolversObject<{
  clusterId?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type AuthConfigResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['AuthConfig'] = ResolversParentTypes['AuthConfig'],
> = ResolversObject<{
  authEndpoint?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  clientId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  pachdClientId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type BranchResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Branch'] = ResolversParentTypes['Branch'],
> = ResolversObject<{
  name?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  repo?: Resolver<Maybe<ResolversTypes['RepoInfo']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type BranchInfoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['BranchInfo'] = ResolversParentTypes['BranchInfo'],
> = ResolversObject<{
  branch?: Resolver<Maybe<ResolversTypes['Branch']>, ParentType, ContextType>;
  head?: Resolver<Maybe<ResolversTypes['Commit']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type CommitResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Commit'] = ResolversParentTypes['Commit'],
> = ResolversObject<{
  branch?: Resolver<Maybe<ResolversTypes['Branch']>, ParentType, ContextType>;
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  diff?: Resolver<Maybe<ResolversTypes['Diff']>, ParentType, ContextType>;
  finished?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  hasLinkedJob?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  originKind?: Resolver<
    Maybe<ResolversTypes['OriginKind']>,
    ParentType,
    ContextType
  >;
  repoName?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  started?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type CronInputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['CronInput'] = ResolversParentTypes['CronInput'],
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  repo?: Resolver<ResolversTypes['Repo'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type DatumResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Datum'] = ResolversParentTypes['Datum'],
> = ResolversObject<{
  downloadBytes?: Resolver<
    Maybe<ResolversTypes['Float']>,
    ParentType,
    ContextType
  >;
  downloadTimestamp?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  jobId?: Resolver<Maybe<ResolversTypes['ID']>, ParentType, ContextType>;
  processTimestamp?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  requestedJobId?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  state?: Resolver<ResolversTypes['DatumState'], ParentType, ContextType>;
  uploadBytes?: Resolver<
    Maybe<ResolversTypes['Float']>,
    ParentType,
    ContextType
  >;
  uploadTimestamp?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type DiffResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Diff'] = ResolversParentTypes['Diff'],
> = ResolversObject<{
  filesAdded?: Resolver<ResolversTypes['DiffCount'], ParentType, ContextType>;
  filesDeleted?: Resolver<ResolversTypes['DiffCount'], ParentType, ContextType>;
  filesUpdated?: Resolver<ResolversTypes['DiffCount'], ParentType, ContextType>;
  size?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type DiffCountResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['DiffCount'] = ResolversParentTypes['DiffCount'],
> = ResolversObject<{
  count?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  sizeDelta?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type EnterpriseInfoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['EnterpriseInfo'] = ResolversParentTypes['EnterpriseInfo'],
> = ResolversObject<{
  expiration?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  state?: Resolver<ResolversTypes['EnterpriseState'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type FileResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['File'] = ResolversParentTypes['File'],
> = ResolversObject<{
  commitAction?: Resolver<
    Maybe<ResolversTypes['FileCommitState']>,
    ParentType,
    ContextType
  >;
  commitId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  committed?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  download?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  downloadDisabled?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType
  >;
  hash?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  path?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  repoName?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  type?: Resolver<ResolversTypes['FileType'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type FileQueryResponseResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['FileQueryResponse'] = ResolversParentTypes['FileQueryResponse'],
> = ResolversObject<{
  diff?: Resolver<Maybe<ResolversTypes['Diff']>, ParentType, ContextType>;
  files?: Resolver<Array<ResolversTypes['File']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type GitInputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['GitInput'] = ResolversParentTypes['GitInput'],
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  url?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type InputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Input'] = ResolversParentTypes['Input'],
> = ResolversObject<{
  cronInput?: Resolver<
    Maybe<ResolversTypes['CronInput']>,
    ParentType,
    ContextType
  >;
  crossedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  gitInput?: Resolver<
    Maybe<ResolversTypes['GitInput']>,
    ParentType,
    ContextType
  >;
  groupedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  joinedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  pfsInput?: Resolver<
    Maybe<ResolversTypes['PFSInput']>,
    ParentType,
    ContextType
  >;
  type?: Resolver<ResolversTypes['InputType'], ParentType, ContextType>;
  unionedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type InputPipelineResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['InputPipeline'] = ResolversParentTypes['InputPipeline'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type JobResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Job'] = ResolversParentTypes['Job'],
> = ResolversObject<{
  createdAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  dataFailed?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataProcessed?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataRecovered?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataSkipped?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataTotal?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  downloadBytesDisplay?: Resolver<
    ResolversTypes['String'],
    ParentType,
    ContextType
  >;
  finishedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  inputBranch?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  inputString?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  jsonDetails?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  nodeState?: Resolver<ResolversTypes['NodeState'], ParentType, ContextType>;
  outputBranch?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  outputCommit?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  pipelineName?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  reason?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  restarts?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  startedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  transform?: Resolver<
    Maybe<ResolversTypes['Transform']>,
    ParentType,
    ContextType
  >;
  transformString?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  uploadBytesDisplay?: Resolver<
    ResolversTypes['String'],
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type JobSetResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['JobSet'] = ResolversParentTypes['JobSet'],
> = ResolversObject<{
  createdAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  finishedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  inProgress?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  jobs?: Resolver<Array<ResolversTypes['Job']>, ParentType, ContextType>;
  startedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type LogResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Log'] = ResolversParentTypes['Log'],
> = ResolversObject<{
  message?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  timestamp?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  user?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type MutationResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Mutation'] = ResolversParentTypes['Mutation'],
> = ResolversObject<{
  createBranch?: Resolver<
    ResolversTypes['Branch'],
    ParentType,
    ContextType,
    RequireFields<MutationCreateBranchArgs, 'args'>
  >;
  createPipeline?: Resolver<
    ResolversTypes['Pipeline'],
    ParentType,
    ContextType,
    RequireFields<MutationCreatePipelineArgs, 'args'>
  >;
  createProject?: Resolver<
    ResolversTypes['Project'],
    ParentType,
    ContextType,
    RequireFields<MutationCreateProjectArgs, 'args'>
  >;
  createRepo?: Resolver<
    ResolversTypes['Repo'],
    ParentType,
    ContextType,
    RequireFields<MutationCreateRepoArgs, 'args'>
  >;
  deleteFile?: Resolver<
    ResolversTypes['ID'],
    ParentType,
    ContextType,
    RequireFields<MutationDeleteFileArgs, 'args'>
  >;
  deletePipeline?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType,
    RequireFields<MutationDeletePipelineArgs, 'args'>
  >;
  deleteProjectAndResources?: Resolver<
    ResolversTypes['Boolean'],
    ParentType,
    ContextType,
    RequireFields<MutationDeleteProjectAndResourcesArgs, 'args'>
  >;
  deleteRepo?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType,
    RequireFields<MutationDeleteRepoArgs, 'args'>
  >;
  exchangeCode?: Resolver<
    ResolversTypes['Tokens'],
    ParentType,
    ContextType,
    RequireFields<MutationExchangeCodeArgs, 'code'>
  >;
  finishCommit?: Resolver<
    ResolversTypes['Boolean'],
    ParentType,
    ContextType,
    RequireFields<MutationFinishCommitArgs, 'args'>
  >;
  putFilesFromURLs?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType,
    RequireFields<MutationPutFilesFromUrLsArgs, 'args'>
  >;
  startCommit?: Resolver<
    ResolversTypes['OpenCommit'],
    ParentType,
    ContextType,
    RequireFields<MutationStartCommitArgs, 'args'>
  >;
  updateProject?: Resolver<
    ResolversTypes['Project'],
    ParentType,
    ContextType,
    RequireFields<MutationUpdateProjectArgs, 'args'>
  >;
}>;

export type NodeSelectorResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['NodeSelector'] = ResolversParentTypes['NodeSelector'],
> = ResolversObject<{
  key?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  value?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type OpenCommitResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['OpenCommit'] = ResolversParentTypes['OpenCommit'],
> = ResolversObject<{
  branch?: Resolver<ResolversTypes['Branch'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PfsInputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PFSInput'] = ResolversParentTypes['PFSInput'],
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  repo?: Resolver<ResolversTypes['Repo'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PachResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Pach'] = ResolversParentTypes['Pach'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PageableCommitResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PageableCommit'] = ResolversParentTypes['PageableCommit'],
> = ResolversObject<{
  cursor?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  items?: Resolver<Array<ResolversTypes['Commit']>, ParentType, ContextType>;
  parentCommit?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PageableDatumResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PageableDatum'] = ResolversParentTypes['PageableDatum'],
> = ResolversObject<{
  cursor?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  hasNextPage?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType
  >;
  items?: Resolver<Array<ResolversTypes['Datum']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PageableFileResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PageableFile'] = ResolversParentTypes['PageableFile'],
> = ResolversObject<{
  cursor?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  diff?: Resolver<Maybe<ResolversTypes['Diff']>, ParentType, ContextType>;
  files?: Resolver<Array<ResolversTypes['File']>, ParentType, ContextType>;
  hasNextPage?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PageableJobResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PageableJob'] = ResolversParentTypes['PageableJob'],
> = ResolversObject<{
  cursor?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  hasNextPage?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType
  >;
  items?: Resolver<Array<ResolversTypes['Job']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PageableJobSetResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PageableJobSet'] = ResolversParentTypes['PageableJobSet'],
> = ResolversObject<{
  cursor?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  hasNextPage?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType
  >;
  items?: Resolver<Array<ResolversTypes['JobSet']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PipelineResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Pipeline'] = ResolversParentTypes['Pipeline'],
> = ResolversObject<{
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  datumTimeoutS?: Resolver<
    Maybe<ResolversTypes['Int']>,
    ParentType,
    ContextType
  >;
  datumTries?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  egress?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  jobTimeoutS?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  jsonSpec?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  lastJobNodeState?: Resolver<
    Maybe<ResolversTypes['NodeState']>,
    ParentType,
    ContextType
  >;
  lastJobState?: Resolver<
    Maybe<ResolversTypes['JobState']>,
    ParentType,
    ContextType
  >;
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  nodeState?: Resolver<ResolversTypes['NodeState'], ParentType, ContextType>;
  outputBranch?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  reason?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  recentError?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  s3OutputRepo?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  state?: Resolver<ResolversTypes['PipelineState'], ParentType, ContextType>;
  stopped?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  type?: Resolver<ResolversTypes['PipelineType'], ParentType, ContextType>;
  version?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type ProjectResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Project'] = ResolversParentTypes['Project'],
> = ResolversObject<{
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  id?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  status?: Resolver<ResolversTypes['ProjectStatus'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type ProjectDetailsResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['ProjectDetails'] = ResolversParentTypes['ProjectDetails'],
> = ResolversObject<{
  jobSets?: Resolver<Array<ResolversTypes['JobSet']>, ParentType, ContextType>;
  pipelineCount?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  repoCount?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type ProjectWithoutStatusResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['ProjectWithoutStatus'] = ResolversParentTypes['ProjectWithoutStatus'],
> = ResolversObject<{
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  id?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  status?: Resolver<
    Maybe<ResolversTypes['ProjectStatus']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type QueryResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Query'] = ResolversParentTypes['Query'],
> = ResolversObject<{
  account?: Resolver<ResolversTypes['Account'], ParentType, ContextType>;
  adminInfo?: Resolver<ResolversTypes['AdminInfo'], ParentType, ContextType>;
  authConfig?: Resolver<ResolversTypes['AuthConfig'], ParentType, ContextType>;
  branch?: Resolver<
    ResolversTypes['Branch'],
    ParentType,
    ContextType,
    RequireFields<QueryBranchArgs, 'args'>
  >;
  branches?: Resolver<
    Array<Maybe<ResolversTypes['Branch']>>,
    ParentType,
    ContextType,
    RequireFields<QueryBranchesArgs, 'args'>
  >;
  commit?: Resolver<
    Maybe<ResolversTypes['Commit']>,
    ParentType,
    ContextType,
    RequireFields<QueryCommitArgs, 'args'>
  >;
  commitSearch?: Resolver<
    Maybe<ResolversTypes['Commit']>,
    ParentType,
    ContextType,
    RequireFields<QueryCommitSearchArgs, 'args'>
  >;
  commits?: Resolver<
    ResolversTypes['PageableCommit'],
    ParentType,
    ContextType,
    RequireFields<QueryCommitsArgs, 'args'>
  >;
  dag?: Resolver<
    Array<ResolversTypes['Vertex']>,
    ParentType,
    ContextType,
    RequireFields<QueryDagArgs, 'args'>
  >;
  datum?: Resolver<
    ResolversTypes['Datum'],
    ParentType,
    ContextType,
    RequireFields<QueryDatumArgs, 'args'>
  >;
  datumSearch?: Resolver<
    Maybe<ResolversTypes['Datum']>,
    ParentType,
    ContextType,
    RequireFields<QueryDatumSearchArgs, 'args'>
  >;
  datums?: Resolver<
    ResolversTypes['PageableDatum'],
    ParentType,
    ContextType,
    RequireFields<QueryDatumsArgs, 'args'>
  >;
  enterpriseInfo?: Resolver<
    ResolversTypes['EnterpriseInfo'],
    ParentType,
    ContextType
  >;
  files?: Resolver<
    ResolversTypes['PageableFile'],
    ParentType,
    ContextType,
    RequireFields<QueryFilesArgs, 'args'>
  >;
  job?: Resolver<
    ResolversTypes['Job'],
    ParentType,
    ContextType,
    RequireFields<QueryJobArgs, 'args'>
  >;
  jobSet?: Resolver<
    ResolversTypes['JobSet'],
    ParentType,
    ContextType,
    RequireFields<QueryJobSetArgs, 'args'>
  >;
  jobSets?: Resolver<
    ResolversTypes['PageableJobSet'],
    ParentType,
    ContextType,
    RequireFields<QueryJobSetsArgs, 'args'>
  >;
  jobs?: Resolver<
    ResolversTypes['PageableJob'],
    ParentType,
    ContextType,
    RequireFields<QueryJobsArgs, 'args'>
  >;
  jobsByPipeline?: Resolver<
    Array<ResolversTypes['Job']>,
    ParentType,
    ContextType,
    RequireFields<QueryJobsByPipelineArgs, 'args'>
  >;
  loggedIn?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  logs?: Resolver<
    Array<Maybe<ResolversTypes['Log']>>,
    ParentType,
    ContextType,
    RequireFields<QueryLogsArgs, 'args'>
  >;
  pipeline?: Resolver<
    ResolversTypes['Pipeline'],
    ParentType,
    ContextType,
    RequireFields<QueryPipelineArgs, 'args'>
  >;
  pipelines?: Resolver<
    Array<Maybe<ResolversTypes['Pipeline']>>,
    ParentType,
    ContextType,
    RequireFields<QueryPipelinesArgs, 'args'>
  >;
  project?: Resolver<
    ResolversTypes['Project'],
    ParentType,
    ContextType,
    RequireFields<QueryProjectArgs, 'id'>
  >;
  projectDetails?: Resolver<
    ResolversTypes['ProjectDetails'],
    ParentType,
    ContextType,
    RequireFields<QueryProjectDetailsArgs, 'args'>
  >;
  projects?: Resolver<
    Array<ResolversTypes['Project']>,
    ParentType,
    ContextType
  >;
  repo?: Resolver<
    ResolversTypes['Repo'],
    ParentType,
    ContextType,
    RequireFields<QueryRepoArgs, 'args'>
  >;
  repos?: Resolver<
    Array<Maybe<ResolversTypes['Repo']>>,
    ParentType,
    ContextType,
    RequireFields<QueryReposArgs, 'args'>
  >;
  searchResults?: Resolver<
    ResolversTypes['SearchResults'],
    ParentType,
    ContextType,
    RequireFields<QuerySearchResultsArgs, 'args'>
  >;
  workspaceLogs?: Resolver<
    Array<Maybe<ResolversTypes['Log']>>,
    ParentType,
    ContextType,
    RequireFields<QueryWorkspaceLogsArgs, 'args'>
  >;
}>;

export type RepoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Repo'] = ResolversParentTypes['Repo'],
> = ResolversObject<{
  access?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  branches?: Resolver<Array<ResolversTypes['Branch']>, ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  description?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  lastCommit?: Resolver<
    Maybe<ResolversTypes['Commit']>,
    ParentType,
    ContextType
  >;
  linkedPipeline?: Resolver<
    Maybe<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
  name?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  projectId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type RepoInfoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['RepoInfo'] = ResolversParentTypes['RepoInfo'],
> = ResolversObject<{
  name?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  type?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type SchedulingSpecResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['SchedulingSpec'] = ResolversParentTypes['SchedulingSpec'],
> = ResolversObject<{
  nodeSelectorMap?: Resolver<
    Array<ResolversTypes['NodeSelector']>,
    ParentType,
    ContextType
  >;
  priorityClassName?: Resolver<
    ResolversTypes['String'],
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type SearchResultsResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['SearchResults'] = ResolversParentTypes['SearchResults'],
> = ResolversObject<{
  jobSet?: Resolver<Maybe<ResolversTypes['JobSet']>, ParentType, ContextType>;
  pipelines?: Resolver<
    Array<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
  repos?: Resolver<Array<ResolversTypes['Repo']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type SubscriptionResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Subscription'] = ResolversParentTypes['Subscription'],
> = ResolversObject<{
  dags?: SubscriptionResolver<
    Array<ResolversTypes['Vertex']>,
    'dags',
    ParentType,
    ContextType,
    RequireFields<SubscriptionDagsArgs, 'args'>
  >;
  logs?: SubscriptionResolver<
    ResolversTypes['Log'],
    'logs',
    ParentType,
    ContextType,
    RequireFields<SubscriptionLogsArgs, 'args'>
  >;
  workspaceLogs?: SubscriptionResolver<
    ResolversTypes['Log'],
    'workspaceLogs',
    ParentType,
    ContextType,
    RequireFields<SubscriptionWorkspaceLogsArgs, 'args'>
  >;
}>;

export type TimestampResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Timestamp'] = ResolversParentTypes['Timestamp'],
> = ResolversObject<{
  nanos?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  seconds?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type TokensResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Tokens'] = ResolversParentTypes['Tokens'],
> = ResolversObject<{
  idToken?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  pachToken?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type TransformResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Transform'] = ResolversParentTypes['Transform'],
> = ResolversObject<{
  cmdList?: Resolver<Array<ResolversTypes['String']>, ParentType, ContextType>;
  debug?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  image?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type VertexResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Vertex'] = ResolversParentTypes['Vertex'],
> = ResolversObject<{
  access?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  createdAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  id?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  jobState?: Resolver<
    Maybe<ResolversTypes['NodeState']>,
    ParentType,
    ContextType
  >;
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  parents?: Resolver<Array<ResolversTypes['String']>, ParentType, ContextType>;
  state?: Resolver<Maybe<ResolversTypes['NodeState']>, ParentType, ContextType>;
  type?: Resolver<ResolversTypes['NodeType'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type Resolvers<ContextType = Context> = ResolversObject<{
  Account?: AccountResolvers<ContextType>;
  AdminInfo?: AdminInfoResolvers<ContextType>;
  AuthConfig?: AuthConfigResolvers<ContextType>;
  Branch?: BranchResolvers<ContextType>;
  BranchInfo?: BranchInfoResolvers<ContextType>;
  Commit?: CommitResolvers<ContextType>;
  CronInput?: CronInputResolvers<ContextType>;
  Datum?: DatumResolvers<ContextType>;
  Diff?: DiffResolvers<ContextType>;
  DiffCount?: DiffCountResolvers<ContextType>;
  EnterpriseInfo?: EnterpriseInfoResolvers<ContextType>;
  File?: FileResolvers<ContextType>;
  FileQueryResponse?: FileQueryResponseResolvers<ContextType>;
  GitInput?: GitInputResolvers<ContextType>;
  Input?: InputResolvers<ContextType>;
  InputPipeline?: InputPipelineResolvers<ContextType>;
  Job?: JobResolvers<ContextType>;
  JobSet?: JobSetResolvers<ContextType>;
  Log?: LogResolvers<ContextType>;
  Mutation?: MutationResolvers<ContextType>;
  NodeSelector?: NodeSelectorResolvers<ContextType>;
  OpenCommit?: OpenCommitResolvers<ContextType>;
  PFSInput?: PfsInputResolvers<ContextType>;
  Pach?: PachResolvers<ContextType>;
  PageableCommit?: PageableCommitResolvers<ContextType>;
  PageableDatum?: PageableDatumResolvers<ContextType>;
  PageableFile?: PageableFileResolvers<ContextType>;
  PageableJob?: PageableJobResolvers<ContextType>;
  PageableJobSet?: PageableJobSetResolvers<ContextType>;
  Pipeline?: PipelineResolvers<ContextType>;
  Project?: ProjectResolvers<ContextType>;
  ProjectDetails?: ProjectDetailsResolvers<ContextType>;
  ProjectWithoutStatus?: ProjectWithoutStatusResolvers<ContextType>;
  Query?: QueryResolvers<ContextType>;
  Repo?: RepoResolvers<ContextType>;
  RepoInfo?: RepoInfoResolvers<ContextType>;
  SchedulingSpec?: SchedulingSpecResolvers<ContextType>;
  SearchResults?: SearchResultsResolvers<ContextType>;
  Subscription?: SubscriptionResolvers<ContextType>;
  Timestamp?: TimestampResolvers<ContextType>;
  Tokens?: TokensResolvers<ContextType>;
  Transform?: TransformResolvers<ContextType>;
  Vertex?: VertexResolvers<ContextType>;
}>;

export type BranchFragmentFragment = {
  __typename?: 'Branch';
  name: string;
  repo?: {
    __typename?: 'RepoInfo';
    name?: string | null;
    type?: string | null;
  } | null;
};

export type CommitFragmentFragment = {
  __typename?: 'Commit';
  repoName: string;
  description?: string | null;
  originKind?: OriginKind | null;
  id: string;
  started: number;
  finished: number;
  sizeBytes: number;
  sizeDisplay: string;
  hasLinkedJob: boolean;
  branch?: {__typename?: 'Branch'; name: string} | null;
};

export type DatumFragment = {
  __typename?: 'Datum';
  id: string;
  jobId?: string | null;
  requestedJobId: string;
  state: DatumState;
  downloadBytes?: number | null;
  uploadBytes?: number | null;
  downloadTimestamp?: {
    __typename?: 'Timestamp';
    seconds: number;
    nanos: number;
  } | null;
  uploadTimestamp?: {
    __typename?: 'Timestamp';
    seconds: number;
    nanos: number;
  } | null;
  processTimestamp?: {
    __typename?: 'Timestamp';
    seconds: number;
    nanos: number;
  } | null;
};

export type DiffFragmentFragment = {
  __typename?: 'Diff';
  size: number;
  sizeDisplay: string;
  filesUpdated: {__typename?: 'DiffCount'; count: number; sizeDelta: number};
  filesAdded: {__typename?: 'DiffCount'; count: number; sizeDelta: number};
  filesDeleted: {__typename?: 'DiffCount'; count: number; sizeDelta: number};
};

export type JobOverviewFragment = {
  __typename?: 'Job';
  id: string;
  state: JobState;
  nodeState: NodeState;
  createdAt?: number | null;
  startedAt?: number | null;
  finishedAt?: number | null;
  restarts: number;
  pipelineName: string;
  reason?: string | null;
  dataProcessed: number;
  dataSkipped: number;
  dataFailed: number;
  dataTotal: number;
  dataRecovered: number;
  downloadBytesDisplay: string;
  uploadBytesDisplay: string;
  outputCommit?: string | null;
};

export type JobSetFieldsFragment = {
  __typename?: 'JobSet';
  id: string;
  state: JobState;
  createdAt?: number | null;
  startedAt?: number | null;
  finishedAt?: number | null;
  inProgress: boolean;
  jobs: Array<{
    __typename?: 'Job';
    inputString?: string | null;
    inputBranch?: string | null;
    transformString?: string | null;
    id: string;
    state: JobState;
    nodeState: NodeState;
    createdAt?: number | null;
    startedAt?: number | null;
    finishedAt?: number | null;
    restarts: number;
    pipelineName: string;
    reason?: string | null;
    dataProcessed: number;
    dataSkipped: number;
    dataFailed: number;
    dataTotal: number;
    dataRecovered: number;
    downloadBytesDisplay: string;
    uploadBytesDisplay: string;
    outputCommit?: string | null;
    transform?: {
      __typename?: 'Transform';
      cmdList: Array<string>;
      image: string;
    } | null;
  }>;
};

export type LogFieldsFragment = {
  __typename?: 'Log';
  user: boolean;
  message: string;
  timestamp?: {__typename?: 'Timestamp'; seconds: number; nanos: number} | null;
};

export type RepoFragmentFragment = {
  __typename?: 'Repo';
  createdAt: number;
  description: string;
  id: string;
  name: string;
  sizeDisplay: string;
  sizeBytes: number;
  access: boolean;
  projectId: string;
  branches: Array<{__typename?: 'Branch'; name: string}>;
  linkedPipeline?: {__typename?: 'Pipeline'; id: string; name: string} | null;
};

export type CreateBranchMutationVariables = Exact<{
  args: CreateBranchArgs;
}>;

export type CreateBranchMutation = {
  __typename?: 'Mutation';
  createBranch: {
    __typename?: 'Branch';
    name: string;
    repo?: {__typename?: 'RepoInfo'; name?: string | null} | null;
  };
};

export type CreatePipelineMutationVariables = Exact<{
  args: CreatePipelineArgs;
}>;

export type CreatePipelineMutation = {
  __typename?: 'Mutation';
  createPipeline: {
    __typename?: 'Pipeline';
    id: string;
    name: string;
    state: PipelineState;
    type: PipelineType;
    description?: string | null;
    datumTimeoutS?: number | null;
    datumTries: number;
    jobTimeoutS?: number | null;
    outputBranch: string;
    s3OutputRepo?: string | null;
    egress: boolean;
    jsonSpec: string;
    reason?: string | null;
  };
};

export type CreateProjectMutationVariables = Exact<{
  args: CreateProjectArgs;
}>;

export type CreateProjectMutation = {
  __typename?: 'Mutation';
  createProject: {
    __typename?: 'Project';
    id: string;
    description?: string | null;
    status: ProjectStatus;
  };
};

export type CreateRepoMutationVariables = Exact<{
  args: CreateRepoArgs;
}>;

export type CreateRepoMutation = {
  __typename?: 'Mutation';
  createRepo: {
    __typename?: 'Repo';
    createdAt: number;
    description: string;
    id: string;
    name: string;
    sizeDisplay: string;
  };
};

export type DeleteFileMutationVariables = Exact<{
  args: DeleteFileArgs;
}>;

export type DeleteFileMutation = {__typename?: 'Mutation'; deleteFile: string};

export type DeletePipelineMutationVariables = Exact<{
  args: DeletePipelineArgs;
}>;

export type DeletePipelineMutation = {
  __typename?: 'Mutation';
  deletePipeline?: boolean | null;
};

export type DeleteProjectAndResourcesMutationVariables = Exact<{
  args: DeleteProjectAndResourcesArgs;
}>;

export type DeleteProjectAndResourcesMutation = {
  __typename?: 'Mutation';
  deleteProjectAndResources: boolean;
};

export type DeleteRepoMutationVariables = Exact<{
  args: DeleteRepoArgs;
}>;

export type DeleteRepoMutation = {
  __typename?: 'Mutation';
  deleteRepo?: boolean | null;
};

export type ExchangeCodeMutationVariables = Exact<{
  code: Scalars['String'];
}>;

export type ExchangeCodeMutation = {
  __typename?: 'Mutation';
  exchangeCode: {__typename?: 'Tokens'; pachToken: string; idToken: string};
};

export type FinishCommitMutationVariables = Exact<{
  args: FinishCommitArgs;
}>;

export type FinishCommitMutation = {
  __typename?: 'Mutation';
  finishCommit: boolean;
};

export type PutFilesFromUrLsMutationVariables = Exact<{
  args: PutFilesFromUrLsArgs;
}>;

export type PutFilesFromUrLsMutation = {
  __typename?: 'Mutation';
  putFilesFromURLs: Array<string>;
};

export type StartCommitMutationVariables = Exact<{
  args: StartCommitArgs;
}>;

export type StartCommitMutation = {
  __typename?: 'Mutation';
  startCommit: {
    __typename?: 'OpenCommit';
    id: string;
    branch: {
      __typename?: 'Branch';
      name: string;
      repo?: {
        __typename?: 'RepoInfo';
        name?: string | null;
        type?: string | null;
      } | null;
    };
  };
};

export type UpdateProjectMutationVariables = Exact<{
  args: UpdateProjectArgs;
}>;

export type UpdateProjectMutation = {
  __typename?: 'Mutation';
  updateProject: {
    __typename?: 'Project';
    id: string;
    description?: string | null;
    status: ProjectStatus;
  };
};

export type GetAccountQueryVariables = Exact<{[key: string]: never}>;

export type GetAccountQuery = {
  __typename?: 'Query';
  account: {
    __typename?: 'Account';
    id: string;
    email: string;
    name?: string | null;
  };
};

export type GetAdminInfoQueryVariables = Exact<{[key: string]: never}>;

export type GetAdminInfoQuery = {
  __typename?: 'Query';
  adminInfo: {__typename?: 'AdminInfo'; clusterId?: string | null};
};

export type AuthConfigQueryVariables = Exact<{[key: string]: never}>;

export type AuthConfigQuery = {
  __typename?: 'Query';
  authConfig: {
    __typename?: 'AuthConfig';
    authEndpoint: string;
    clientId: string;
    pachdClientId: string;
  };
};

export type GetBranchesQueryVariables = Exact<{
  args: BranchesQueryArgs;
}>;

export type GetBranchesQuery = {
  __typename?: 'Query';
  branches: Array<{
    __typename?: 'Branch';
    name: string;
    repo?: {
      __typename?: 'RepoInfo';
      name?: string | null;
      type?: string | null;
    } | null;
  } | null>;
};

export type CommitQueryVariables = Exact<{
  args: CommitQueryArgs;
}>;

export type CommitQuery = {
  __typename?: 'Query';
  commit?: {
    __typename?: 'Commit';
    repoName: string;
    description?: string | null;
    originKind?: OriginKind | null;
    id: string;
    started: number;
    finished: number;
    sizeBytes: number;
    sizeDisplay: string;
    hasLinkedJob: boolean;
    diff?: {
      __typename?: 'Diff';
      size: number;
      sizeDisplay: string;
      filesUpdated: {
        __typename?: 'DiffCount';
        count: number;
        sizeDelta: number;
      };
      filesAdded: {__typename?: 'DiffCount'; count: number; sizeDelta: number};
      filesDeleted: {
        __typename?: 'DiffCount';
        count: number;
        sizeDelta: number;
      };
    } | null;
    branch?: {__typename?: 'Branch'; name: string} | null;
  } | null;
};

export type CommitSearchQueryVariables = Exact<{
  args: CommitSearchQueryArgs;
}>;

export type CommitSearchQuery = {
  __typename?: 'Query';
  commitSearch?: {
    __typename?: 'Commit';
    repoName: string;
    description?: string | null;
    originKind?: OriginKind | null;
    id: string;
    started: number;
    finished: number;
    sizeBytes: number;
    sizeDisplay: string;
    hasLinkedJob: boolean;
    branch?: {__typename?: 'Branch'; name: string} | null;
  } | null;
};

export type GetCommitsQueryVariables = Exact<{
  args: CommitsQueryArgs;
}>;

export type GetCommitsQuery = {
  __typename?: 'Query';
  commits: {
    __typename?: 'PageableCommit';
    parentCommit?: string | null;
    items: Array<{
      __typename?: 'Commit';
      repoName: string;
      description?: string | null;
      originKind?: OriginKind | null;
      id: string;
      started: number;
      finished: number;
      sizeBytes: number;
      sizeDisplay: string;
      hasLinkedJob: boolean;
      branch?: {__typename?: 'Branch'; name: string} | null;
    }>;
    cursor?: {__typename?: 'Timestamp'; seconds: number; nanos: number} | null;
  };
};

export type GetDagQueryVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagQuery = {
  __typename?: 'Query';
  dag: Array<{
    __typename?: 'Vertex';
    id: string;
    name: string;
    state?: NodeState | null;
    access: boolean;
    parents: Array<string>;
    type: NodeType;
    jobState?: NodeState | null;
    createdAt?: number | null;
  }>;
};

export type GetDagsSubscriptionVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagsSubscription = {
  __typename?: 'Subscription';
  dags: Array<{
    __typename?: 'Vertex';
    id: string;
    name: string;
    state?: NodeState | null;
    access: boolean;
    parents: Array<string>;
    type: NodeType;
    jobState?: NodeState | null;
    createdAt?: number | null;
  }>;
};

export type DatumQueryVariables = Exact<{
  args: DatumQueryArgs;
}>;

export type DatumQuery = {
  __typename?: 'Query';
  datum: {
    __typename?: 'Datum';
    id: string;
    jobId?: string | null;
    requestedJobId: string;
    state: DatumState;
    downloadBytes?: number | null;
    uploadBytes?: number | null;
    downloadTimestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
    uploadTimestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
    processTimestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
  };
};

export type DatumSearchQueryVariables = Exact<{
  args: DatumQueryArgs;
}>;

export type DatumSearchQuery = {
  __typename?: 'Query';
  datumSearch?: {
    __typename?: 'Datum';
    id: string;
    jobId?: string | null;
    requestedJobId: string;
    state: DatumState;
    downloadBytes?: number | null;
    uploadBytes?: number | null;
    downloadTimestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
    uploadTimestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
    processTimestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
  } | null;
};

export type DatumsQueryVariables = Exact<{
  args: DatumsQueryArgs;
}>;

export type DatumsQuery = {
  __typename?: 'Query';
  datums: {
    __typename?: 'PageableDatum';
    cursor?: string | null;
    hasNextPage?: boolean | null;
    items: Array<{
      __typename?: 'Datum';
      id: string;
      jobId?: string | null;
      requestedJobId: string;
      state: DatumState;
      downloadBytes?: number | null;
      uploadBytes?: number | null;
      downloadTimestamp?: {
        __typename?: 'Timestamp';
        seconds: number;
        nanos: number;
      } | null;
      uploadTimestamp?: {
        __typename?: 'Timestamp';
        seconds: number;
        nanos: number;
      } | null;
      processTimestamp?: {
        __typename?: 'Timestamp';
        seconds: number;
        nanos: number;
      } | null;
    }>;
  };
};

export type GetEnterpriseInfoQueryVariables = Exact<{[key: string]: never}>;

export type GetEnterpriseInfoQuery = {
  __typename?: 'Query';
  enterpriseInfo: {
    __typename?: 'EnterpriseInfo';
    state: EnterpriseState;
    expiration: number;
  };
};

export type GetFilesQueryVariables = Exact<{
  args: FileQueryArgs;
}>;

export type GetFilesQuery = {
  __typename?: 'Query';
  files: {
    __typename?: 'PageableFile';
    cursor?: string | null;
    hasNextPage?: boolean | null;
    diff?: {
      __typename?: 'Diff';
      size: number;
      sizeDisplay: string;
      filesUpdated: {
        __typename?: 'DiffCount';
        count: number;
        sizeDelta: number;
      };
      filesAdded: {__typename?: 'DiffCount'; count: number; sizeDelta: number};
      filesDeleted: {
        __typename?: 'DiffCount';
        count: number;
        sizeDelta: number;
      };
    } | null;
    files: Array<{
      __typename?: 'File';
      commitId: string;
      download?: string | null;
      hash: string;
      path: string;
      repoName: string;
      sizeBytes: number;
      type: FileType;
      sizeDisplay: string;
      downloadDisabled?: boolean | null;
      commitAction?: FileCommitState | null;
      committed?: {
        __typename?: 'Timestamp';
        nanos: number;
        seconds: number;
      } | null;
    }>;
  };
};

export type JobQueryVariables = Exact<{
  args: JobQueryArgs;
}>;

export type JobQuery = {
  __typename?: 'Query';
  job: {
    __typename?: 'Job';
    inputString?: string | null;
    inputBranch?: string | null;
    outputBranch?: string | null;
    outputCommit?: string | null;
    reason?: string | null;
    jsonDetails: string;
    transformString?: string | null;
    id: string;
    state: JobState;
    nodeState: NodeState;
    createdAt?: number | null;
    startedAt?: number | null;
    finishedAt?: number | null;
    restarts: number;
    pipelineName: string;
    dataProcessed: number;
    dataSkipped: number;
    dataFailed: number;
    dataTotal: number;
    dataRecovered: number;
    downloadBytesDisplay: string;
    uploadBytesDisplay: string;
    transform?: {
      __typename?: 'Transform';
      cmdList: Array<string>;
      image: string;
      debug: boolean;
    } | null;
  };
};

export type JobSetsQueryVariables = Exact<{
  args: JobSetsQueryArgs;
}>;

export type JobSetsQuery = {
  __typename?: 'Query';
  jobSets: {
    __typename?: 'PageableJobSet';
    hasNextPage?: boolean | null;
    items: Array<{
      __typename?: 'JobSet';
      id: string;
      state: JobState;
      createdAt?: number | null;
      startedAt?: number | null;
      finishedAt?: number | null;
      inProgress: boolean;
      jobs: Array<{
        __typename?: 'Job';
        inputString?: string | null;
        inputBranch?: string | null;
        transformString?: string | null;
        id: string;
        state: JobState;
        nodeState: NodeState;
        createdAt?: number | null;
        startedAt?: number | null;
        finishedAt?: number | null;
        restarts: number;
        pipelineName: string;
        reason?: string | null;
        dataProcessed: number;
        dataSkipped: number;
        dataFailed: number;
        dataTotal: number;
        dataRecovered: number;
        downloadBytesDisplay: string;
        uploadBytesDisplay: string;
        outputCommit?: string | null;
        transform?: {
          __typename?: 'Transform';
          cmdList: Array<string>;
          image: string;
        } | null;
      }>;
    }>;
    cursor?: {__typename?: 'Timestamp'; seconds: number; nanos: number} | null;
  };
};

export type JobsByPipelineQueryVariables = Exact<{
  args: JobsByPipelineQueryArgs;
}>;

export type JobsByPipelineQuery = {
  __typename?: 'Query';
  jobsByPipeline: Array<{
    __typename?: 'Job';
    inputString?: string | null;
    inputBranch?: string | null;
    outputBranch?: string | null;
    outputCommit?: string | null;
    reason?: string | null;
    jsonDetails: string;
    transformString?: string | null;
    id: string;
    state: JobState;
    nodeState: NodeState;
    createdAt?: number | null;
    startedAt?: number | null;
    finishedAt?: number | null;
    restarts: number;
    pipelineName: string;
    dataProcessed: number;
    dataSkipped: number;
    dataFailed: number;
    dataTotal: number;
    dataRecovered: number;
    downloadBytesDisplay: string;
    uploadBytesDisplay: string;
    transform?: {
      __typename?: 'Transform';
      cmdList: Array<string>;
      image: string;
      debug: boolean;
    } | null;
  }>;
};

export type JobsQueryVariables = Exact<{
  args: JobsQueryArgs;
}>;

export type JobsQuery = {
  __typename?: 'Query';
  jobs: {
    __typename?: 'PageableJob';
    hasNextPage?: boolean | null;
    items: Array<{
      __typename?: 'Job';
      inputString?: string | null;
      inputBranch?: string | null;
      outputBranch?: string | null;
      outputCommit?: string | null;
      reason?: string | null;
      jsonDetails: string;
      transformString?: string | null;
      id: string;
      state: JobState;
      nodeState: NodeState;
      createdAt?: number | null;
      startedAt?: number | null;
      finishedAt?: number | null;
      restarts: number;
      pipelineName: string;
      dataProcessed: number;
      dataSkipped: number;
      dataFailed: number;
      dataTotal: number;
      dataRecovered: number;
      downloadBytesDisplay: string;
      uploadBytesDisplay: string;
      transform?: {
        __typename?: 'Transform';
        cmdList: Array<string>;
        image: string;
        debug: boolean;
      } | null;
    }>;
    cursor?: {__typename?: 'Timestamp'; seconds: number; nanos: number} | null;
  };
};

export type JobSetQueryVariables = Exact<{
  args: JobSetQueryArgs;
}>;

export type JobSetQuery = {
  __typename?: 'Query';
  jobSet: {
    __typename?: 'JobSet';
    id: string;
    state: JobState;
    createdAt?: number | null;
    startedAt?: number | null;
    finishedAt?: number | null;
    inProgress: boolean;
    jobs: Array<{
      __typename?: 'Job';
      inputString?: string | null;
      inputBranch?: string | null;
      transformString?: string | null;
      id: string;
      state: JobState;
      nodeState: NodeState;
      createdAt?: number | null;
      startedAt?: number | null;
      finishedAt?: number | null;
      restarts: number;
      pipelineName: string;
      reason?: string | null;
      dataProcessed: number;
      dataSkipped: number;
      dataFailed: number;
      dataTotal: number;
      dataRecovered: number;
      downloadBytesDisplay: string;
      uploadBytesDisplay: string;
      outputCommit?: string | null;
      transform?: {
        __typename?: 'Transform';
        cmdList: Array<string>;
        image: string;
      } | null;
    }>;
  };
};

export type GetWorkspaceLogsQueryVariables = Exact<{
  args: WorkspaceLogsArgs;
}>;

export type GetWorkspaceLogsQuery = {
  __typename?: 'Query';
  workspaceLogs: Array<{
    __typename?: 'Log';
    user: boolean;
    message: string;
    timestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
  } | null>;
};

export type GetLogsQueryVariables = Exact<{
  args: LogsArgs;
}>;

export type GetLogsQuery = {
  __typename?: 'Query';
  logs: Array<{
    __typename?: 'Log';
    user: boolean;
    message: string;
    timestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
  } | null>;
};

export type GetWorkspaceLogStreamSubscriptionVariables = Exact<{
  args: WorkspaceLogsArgs;
}>;

export type GetWorkspaceLogStreamSubscription = {
  __typename?: 'Subscription';
  workspaceLogs: {
    __typename?: 'Log';
    user: boolean;
    message: string;
    timestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
  };
};

export type GetLogsStreamSubscriptionVariables = Exact<{
  args: LogsArgs;
}>;

export type GetLogsStreamSubscription = {
  __typename?: 'Subscription';
  logs: {
    __typename?: 'Log';
    user: boolean;
    message: string;
    timestamp?: {
      __typename?: 'Timestamp';
      seconds: number;
      nanos: number;
    } | null;
  };
};

export type PipelineQueryVariables = Exact<{
  args: PipelineQueryArgs;
}>;

export type PipelineQuery = {
  __typename?: 'Query';
  pipeline: {
    __typename?: 'Pipeline';
    id: string;
    name: string;
    description?: string | null;
    version: number;
    createdAt: number;
    state: PipelineState;
    nodeState: NodeState;
    stopped: boolean;
    recentError?: string | null;
    lastJobState?: JobState | null;
    lastJobNodeState?: NodeState | null;
    type: PipelineType;
    datumTimeoutS?: number | null;
    datumTries: number;
    jobTimeoutS?: number | null;
    outputBranch: string;
    s3OutputRepo?: string | null;
    egress: boolean;
    jsonSpec: string;
    reason?: string | null;
  };
};

export type PipelinesQueryVariables = Exact<{
  args: PipelinesQueryArgs;
}>;

export type PipelinesQuery = {
  __typename?: 'Query';
  pipelines: Array<{
    __typename?: 'Pipeline';
    id: string;
    name: string;
    description?: string | null;
    version: number;
    createdAt: number;
    state: PipelineState;
    nodeState: NodeState;
    stopped: boolean;
    recentError?: string | null;
    lastJobState?: JobState | null;
    lastJobNodeState?: NodeState | null;
    type: PipelineType;
    datumTimeoutS?: number | null;
    datumTries: number;
    jobTimeoutS?: number | null;
    outputBranch: string;
    s3OutputRepo?: string | null;
    egress: boolean;
    jsonSpec: string;
    reason?: string | null;
  } | null>;
};

export type ProjectDetailsQueryVariables = Exact<{
  args: ProjectDetailsQueryArgs;
}>;

export type ProjectDetailsQuery = {
  __typename?: 'Query';
  projectDetails: {
    __typename?: 'ProjectDetails';
    sizeDisplay: string;
    repoCount: number;
    pipelineCount: number;
    jobSets: Array<{
      __typename?: 'JobSet';
      id: string;
      state: JobState;
      createdAt?: number | null;
      startedAt?: number | null;
      finishedAt?: number | null;
      inProgress: boolean;
      jobs: Array<{
        __typename?: 'Job';
        inputString?: string | null;
        inputBranch?: string | null;
        transformString?: string | null;
        id: string;
        state: JobState;
        nodeState: NodeState;
        createdAt?: number | null;
        startedAt?: number | null;
        finishedAt?: number | null;
        restarts: number;
        pipelineName: string;
        reason?: string | null;
        dataProcessed: number;
        dataSkipped: number;
        dataFailed: number;
        dataTotal: number;
        dataRecovered: number;
        downloadBytesDisplay: string;
        uploadBytesDisplay: string;
        outputCommit?: string | null;
        transform?: {
          __typename?: 'Transform';
          cmdList: Array<string>;
          image: string;
        } | null;
      }>;
    }>;
  };
};

export type ProjectQueryVariables = Exact<{
  id: Scalars['ID'];
}>;

export type ProjectQuery = {
  __typename?: 'Query';
  project: {
    __typename?: 'Project';
    id: string;
    description?: string | null;
    status: ProjectStatus;
  };
};

export type ProjectsQueryVariables = Exact<{[key: string]: never}>;

export type ProjectsQuery = {
  __typename?: 'Query';
  projects: Array<{
    __typename?: 'Project';
    id: string;
    description?: string | null;
    status: ProjectStatus;
  }>;
};

export type RepoQueryVariables = Exact<{
  args: RepoQueryArgs;
}>;

export type RepoQuery = {
  __typename?: 'Query';
  repo: {
    __typename?: 'Repo';
    createdAt: number;
    description: string;
    id: string;
    name: string;
    sizeDisplay: string;
    sizeBytes: number;
    access: boolean;
    projectId: string;
    branches: Array<{__typename?: 'Branch'; name: string}>;
    linkedPipeline?: {__typename?: 'Pipeline'; id: string; name: string} | null;
  };
};

export type RepoWithCommitQueryVariables = Exact<{
  args: RepoQueryArgs;
}>;

export type RepoWithCommitQuery = {
  __typename?: 'Query';
  repo: {
    __typename?: 'Repo';
    createdAt: number;
    description: string;
    id: string;
    name: string;
    sizeDisplay: string;
    sizeBytes: number;
    access: boolean;
    projectId: string;
    lastCommit?: {
      __typename?: 'Commit';
      repoName: string;
      description?: string | null;
      originKind?: OriginKind | null;
      id: string;
      started: number;
      finished: number;
      sizeBytes: number;
      sizeDisplay: string;
      hasLinkedJob: boolean;
      branch?: {__typename?: 'Branch'; name: string} | null;
    } | null;
    branches: Array<{__typename?: 'Branch'; name: string}>;
    linkedPipeline?: {__typename?: 'Pipeline'; id: string; name: string} | null;
  };
};

export type ReposQueryVariables = Exact<{
  args: ReposQueryArgs;
}>;

export type ReposQuery = {
  __typename?: 'Query';
  repos: Array<{
    __typename?: 'Repo';
    createdAt: number;
    description: string;
    id: string;
    name: string;
    sizeDisplay: string;
    sizeBytes: number;
    access: boolean;
    projectId: string;
    branches: Array<{__typename?: 'Branch'; name: string}>;
    linkedPipeline?: {__typename?: 'Pipeline'; id: string; name: string} | null;
  } | null>;
};

export type ReposWithCommitQueryVariables = Exact<{
  args: ReposQueryArgs;
}>;

export type ReposWithCommitQuery = {
  __typename?: 'Query';
  repos: Array<{
    __typename?: 'Repo';
    createdAt: number;
    description: string;
    id: string;
    name: string;
    sizeDisplay: string;
    sizeBytes: number;
    access: boolean;
    projectId: string;
    lastCommit?: {
      __typename?: 'Commit';
      repoName: string;
      description?: string | null;
      originKind?: OriginKind | null;
      id: string;
      started: number;
      finished: number;
      sizeBytes: number;
      sizeDisplay: string;
      hasLinkedJob: boolean;
      branch?: {__typename?: 'Branch'; name: string} | null;
    } | null;
    branches: Array<{__typename?: 'Branch'; name: string}>;
    linkedPipeline?: {__typename?: 'Pipeline'; id: string; name: string} | null;
  } | null>;
};

export type SearchResultsQueryVariables = Exact<{
  args: SearchResultQueryArgs;
}>;

export type SearchResultsQuery = {
  __typename?: 'Query';
  searchResults: {
    __typename?: 'SearchResults';
    pipelines: Array<{__typename?: 'Pipeline'; name: string; id: string}>;
    repos: Array<{__typename?: 'Repo'; name: string; id: string}>;
    jobSet?: {__typename?: 'JobSet'; id: string} | null;
  };
};

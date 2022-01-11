/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable @typescript-eslint/ban-types */
/* eslint-disable import/no-duplicates */
/* eslint-disable @typescript-eslint/no-explicit-any */

import {GraphQLResolveInfo} from 'graphql';

import {Context} from '@dash-backend/lib/types';
export type Maybe<T> = T | null;
export type Exact<T extends {[key: string]: unknown}> = {[K in keyof T]: T[K]};
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]?: Maybe<T[SubKey]>;
};
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]: Maybe<T[SubKey]>;
};
export type RequireFields<T, K extends keyof T> = {
  [X in Exclude<keyof T, K>]?: T[X];
} & {[P in K]-?: NonNullable<T[P]>};
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
  id: Scalars['ID'];
  email: Scalars['String'];
  name?: Maybe<Scalars['String']>;
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
  projectId: Scalars['String'];
  branch: BranchInput;
};

export type Commit = {
  __typename?: 'Commit';
  repoName: Scalars['String'];
  branch?: Maybe<Branch>;
  description?: Maybe<Scalars['String']>;
  originKind?: Maybe<OriginKind>;
  hasLinkedJob: Scalars['Boolean'];
  id: Scalars['ID'];
  started?: Maybe<Scalars['Int']>;
  finished?: Maybe<Scalars['Int']>;
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
};

export type CommitInput = {
  id: Scalars['ID'];
  branch?: Maybe<BranchInput>;
};

export type CommitsQueryArgs = {
  projectId: Scalars['String'];
  repoName: Scalars['String'];
  branchName?: Maybe<Scalars['String']>;
  pipelineName?: Maybe<Scalars['String']>;
  originKind?: Maybe<OriginKind>;
  number?: Maybe<Scalars['Int']>;
};

export type CreateBranchArgs = {
  head?: Maybe<CommitInput>;
  branch?: Maybe<BranchInput>;
  provenance?: Maybe<Array<BranchInput>>;
  newCommitSet?: Maybe<Scalars['Boolean']>;
  projectId: Scalars['String'];
};

export type CreatePipelineArgs = {
  name: Scalars['String'];
  transform: TransformInput;
  pfs?: Maybe<Pfs>;
  crossList?: Maybe<Array<Pfs>>;
  projectId: Scalars['String'];
  description?: Maybe<Scalars['String']>;
  update?: Maybe<Scalars['Boolean']>;
};

export type CreateRepoArgs = {
  name: Scalars['String'];
  description?: Maybe<Scalars['String']>;
  update?: Maybe<Scalars['Boolean']>;
  projectId: Scalars['String'];
};

export type CronInput = {
  __typename?: 'CronInput';
  name: Scalars['String'];
  repo: Repo;
};

export type DagQueryArgs = {
  projectId: Scalars['ID'];
  jobSetId?: Maybe<Scalars['ID']>;
};

export type File = {
  __typename?: 'File';
  committed?: Maybe<Timestamp>;
  commitId: Scalars['String'];
  download?: Maybe<Scalars['String']>;
  hash: Scalars['String'];
  path: Scalars['String'];
  repoName: Scalars['String'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
  downloadDisabled?: Maybe<Scalars['Boolean']>;
  type: FileType;
};

export type FileFromUrl = {
  url: Scalars['String'];
  path: Scalars['String'];
};

export type FileQueryArgs = {
  projectId: Scalars['String'];
  commitId?: Maybe<Scalars['String']>;
  path?: Maybe<Scalars['String']>;
  repoName: Scalars['String'];
  branchName: Scalars['String'];
};

export enum FileType {
  RESERVED = 'RESERVED',
  DIR = 'DIR',
  FILE = 'FILE',
}

export type GitInput = {
  __typename?: 'GitInput';
  name: Scalars['String'];
  url: Scalars['String'];
};

export type Input = {
  __typename?: 'Input';
  id: Scalars['ID'];
  type: InputType;
  joinedWith: Array<Scalars['String']>;
  groupedWith: Array<Scalars['String']>;
  crossedWith: Array<Scalars['String']>;
  unionedWith: Array<Scalars['String']>;
  pfsInput?: Maybe<PfsInput>;
  cronInput?: Maybe<CronInput>;
  gitInput?: Maybe<GitInput>;
};

export type InputPipeline = {
  __typename?: 'InputPipeline';
  id: Scalars['ID'];
};

export enum InputType {
  PFS = 'PFS',
  CRON = 'CRON',
  GIT = 'GIT',
}

export type Job = {
  __typename?: 'Job';
  id: Scalars['ID'];
  createdAt?: Maybe<Scalars['Int']>;
  startedAt?: Maybe<Scalars['Int']>;
  finishedAt?: Maybe<Scalars['Int']>;
  state: JobState;
  pipelineName: Scalars['String'];
  transform?: Maybe<Transform>;
  inputString?: Maybe<Scalars['String']>;
  inputBranch?: Maybe<Scalars['String']>;
  outputBranch?: Maybe<Scalars['String']>;
  reason?: Maybe<Scalars['String']>;
  jsonDetails: Scalars['String'];
  dataProcessed: Scalars['Int'];
  dataSkipped: Scalars['Int'];
  dataFailed: Scalars['Int'];
  dataRecovered: Scalars['Int'];
  dataTotal: Scalars['Int'];
};

export type JobQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
  pipelineName: Scalars['String'];
};

export type JobSet = {
  __typename?: 'JobSet';
  id: Scalars['ID'];
  createdAt?: Maybe<Scalars['Int']>;
  state: JobState;
  jobs: Array<Job>;
};

export type JobSetQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
};

export type JobSetsQueryArgs = {
  projectId: Scalars['String'];
};

export enum JobState {
  JOB_STATE_UNKNOWN = 'JOB_STATE_UNKNOWN',
  JOB_CREATED = 'JOB_CREATED',
  JOB_STARTING = 'JOB_STARTING',
  JOB_RUNNING = 'JOB_RUNNING',
  JOB_FAILURE = 'JOB_FAILURE',
  JOB_SUCCESS = 'JOB_SUCCESS',
  JOB_KILLED = 'JOB_KILLED',
  JOB_EGRESSING = 'JOB_EGRESSING',
  JOB_FINISHING = 'JOB_FINISHING',
}

export type JobsQueryArgs = {
  projectId: Scalars['ID'];
  limit?: Maybe<Scalars['Int']>;
  pipelineId?: Maybe<Scalars['String']>;
};

export type Log = {
  __typename?: 'Log';
  message: Scalars['String'];
  timestamp?: Maybe<Timestamp>;
  user: Scalars['Boolean'];
};

export type LogsArgs = {
  projectId: Scalars['String'];
  pipelineName: Scalars['String'];
  jobId?: Maybe<Scalars['String']>;
  start?: Maybe<Scalars['Int']>;
  reverse?: Maybe<Scalars['Boolean']>;
};

export type Mutation = {
  __typename?: 'Mutation';
  exchangeCode: Tokens;
  createRepo: Repo;
  createPipeline: Pipeline;
  createBranch: Branch;
  putFilesFromURLs: Array<Scalars['String']>;
};

export type MutationExchangeCodeArgs = {
  code: Scalars['String'];
};

export type MutationCreateRepoArgs = {
  args: CreateRepoArgs;
};

export type MutationCreatePipelineArgs = {
  args: CreatePipelineArgs;
};

export type MutationCreateBranchArgs = {
  args: CreateBranchArgs;
};

export type MutationPutFilesFromUrLsArgs = {
  args: PutFilesFromUrLsArgs;
};

export type NodeSelector = {
  __typename?: 'NodeSelector';
  key: Scalars['String'];
  value: Scalars['String'];
};

export enum NodeState {
  SUCCESS = 'SUCCESS',
  IDLE = 'IDLE',
  PAUSED = 'PAUSED',
  BUSY = 'BUSY',
  ERROR = 'ERROR',
}

export enum NodeType {
  PIPELINE = 'PIPELINE',
  OUTPUT_REPO = 'OUTPUT_REPO',
  INPUT_REPO = 'INPUT_REPO',
  EGRESS = 'EGRESS',
}

export enum OriginKind {
  USER = 'USER',
  AUTO = 'AUTO',
  FSCK = 'FSCK',
  ALIAS = 'ALIAS',
  ORIGIN_KIND_UNKNOWN = 'ORIGIN_KIND_UNKNOWN',
}

export type Pfs = {
  name: Scalars['String'];
  repo: RepoInput;
  glob?: Maybe<Scalars['String']>;
  branch?: Maybe<Scalars['String']>;
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

export type Pipeline = {
  __typename?: 'Pipeline';
  id: Scalars['ID'];
  name: Scalars['String'];
  version: Scalars['Int'];
  createdAt: Scalars['Int'];
  state: PipelineState;
  stopped: Scalars['Boolean'];
  recentError?: Maybe<Scalars['String']>;
  numOfJobsCreated: Scalars['Int'];
  numOfJobsStarting: Scalars['Int'];
  numOfJobsRunning: Scalars['Int'];
  numOfJobsFailing: Scalars['Int'];
  numOfJobsSucceeding: Scalars['Int'];
  numOfJobsKilled: Scalars['Int'];
  numOfJobsEgressing: Scalars['Int'];
  lastJobState?: Maybe<JobState>;
  description?: Maybe<Scalars['String']>;
  type: PipelineType;
  datumTimeoutS?: Maybe<Scalars['Int']>;
  datumTries: Scalars['Int'];
  jobTimeoutS?: Maybe<Scalars['Int']>;
  outputBranch: Scalars['String'];
  s3OutputRepo?: Maybe<Scalars['String']>;
  egress: Scalars['Boolean'];
  jsonSpec: Scalars['String'];
  reason?: Maybe<Scalars['String']>;
};

export type PipelineQueryArgs = {
  projectId: Scalars['String'];
  id: Scalars['ID'];
};

export enum PipelineState {
  PIPELINE_STATE_UNKNOWN = 'PIPELINE_STATE_UNKNOWN',
  PIPELINE_STARTING = 'PIPELINE_STARTING',
  PIPELINE_RUNNING = 'PIPELINE_RUNNING',
  PIPELINE_RESTARTING = 'PIPELINE_RESTARTING',
  PIPELINE_FAILURE = 'PIPELINE_FAILURE',
  PIPELINE_PAUSED = 'PIPELINE_PAUSED',
  PIPELINE_STANDBY = 'PIPELINE_STANDBY',
  PIPELINE_CRASHING = 'PIPELINE_CRASHING',
}

export enum PipelineType {
  STANDARD = 'STANDARD',
  SPOUT = 'SPOUT',
  SERVICE = 'SERVICE',
}

export type Project = {
  __typename?: 'Project';
  id: Scalars['ID'];
  name: Scalars['String'];
  status: ProjectStatus;
  description: Scalars['String'];
  createdAt: Scalars['Int'];
};

export type ProjectDetails = {
  __typename?: 'ProjectDetails';
  repoCount: Scalars['Int'];
  pipelineCount: Scalars['Int'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
  jobSets: Array<JobSet>;
};

export type ProjectDetailsQueryArgs = {
  projectId: Scalars['String'];
  jobSetsLimit?: Maybe<Scalars['Int']>;
};

export enum ProjectStatus {
  HEALTHY = 'HEALTHY',
  UNHEALTHY = 'UNHEALTHY',
}

export type PutFilesFromUrLsArgs = {
  files: Array<FileFromUrl>;
  branch: Scalars['String'];
  repo: Scalars['String'];
  projectId: Scalars['String'];
};

export type Query = {
  __typename?: 'Query';
  account: Account;
  authConfig: AuthConfig;
  branch: Branch;
  commits: Array<Commit>;
  dag: Array<Vertex>;
  files: Array<File>;
  job: Job;
  jobSet: JobSet;
  jobSets: Array<JobSet>;
  jobs: Array<Job>;
  loggedIn: Scalars['Boolean'];
  logs: Array<Maybe<Log>>;
  pipeline: Pipeline;
  project: Project;
  projectDetails: ProjectDetails;
  projects: Array<Project>;
  repo: Repo;
  searchResults: SearchResults;
  workspaceLogs: Array<Maybe<Log>>;
};

export type QueryBranchArgs = {
  args: BranchQueryArgs;
};

export type QueryCommitsArgs = {
  args: CommitsQueryArgs;
};

export type QueryDagArgs = {
  args: DagQueryArgs;
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

export type QueryLogsArgs = {
  args: LogsArgs;
};

export type QueryPipelineArgs = {
  args: PipelineQueryArgs;
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

export type QuerySearchResultsArgs = {
  args: SearchResultQueryArgs;
};

export type QueryWorkspaceLogsArgs = {
  args: WorkspaceLogsArgs;
};

export type Repo = {
  __typename?: 'Repo';
  branches: Array<Branch>;
  createdAt: Scalars['Int'];
  description: Scalars['String'];
  id: Scalars['ID'];
  name: Scalars['ID'];
  sizeBytes: Scalars['Float'];
  sizeDisplay: Scalars['String'];
  linkedPipeline?: Maybe<Pipeline>;
};

export type RepoInfo = {
  __typename?: 'RepoInfo';
  name?: Maybe<Scalars['String']>;
  type?: Maybe<Scalars['String']>;
};

export type RepoInput = {
  name: Scalars['ID'];
};

export type RepoQueryArgs = {
  projectId: Scalars['String'];
  id: Scalars['ID'];
};

export type SchedulingSpec = {
  __typename?: 'SchedulingSpec';
  nodeSelectorMap: Array<NodeSelector>;
  priorityClassName: Scalars['String'];
};

export type SearchResultQueryArgs = {
  projectId: Scalars['String'];
  query: Scalars['String'];
  limit?: Maybe<Scalars['Int']>;
};

export type SearchResults = {
  __typename?: 'SearchResults';
  pipelines: Array<Pipeline>;
  repos: Array<Repo>;
  jobSet?: Maybe<JobSet>;
};

export type Subscription = {
  __typename?: 'Subscription';
  dags: Array<Vertex>;
  workspaceLogs: Log;
  logs: Log;
};

export type SubscriptionDagsArgs = {
  args: DagQueryArgs;
};

export type SubscriptionWorkspaceLogsArgs = {
  args: WorkspaceLogsArgs;
};

export type SubscriptionLogsArgs = {
  args: LogsArgs;
};

export type Timestamp = {
  __typename?: 'Timestamp';
  seconds: Scalars['Int'];
  nanos: Scalars['Int'];
};

export type Tokens = {
  __typename?: 'Tokens';
  pachToken: Scalars['String'];
  idToken: Scalars['String'];
};

export type Transform = {
  __typename?: 'Transform';
  cmdList: Array<Scalars['String']>;
  image: Scalars['String'];
};

export type TransformInput = {
  cmdList: Array<Scalars['String']>;
  image: Scalars['String'];
  stdinList?: Maybe<Array<Scalars['String']>>;
};

export type Vertex = {
  __typename?: 'Vertex';
  name: Scalars['String'];
  state?: Maybe<NodeState>;
  access: Scalars['Boolean'];
  parents: Array<Scalars['String']>;
  type: NodeType;
  jobState?: Maybe<JobState>;
  createdAt?: Maybe<Scalars['Int']>;
};

export type WorkspaceLogsArgs = {
  start?: Maybe<Scalars['Int']>;
};

export type WithIndex<TObject> = TObject & Record<string, any>;
export type ResolversObject<TObject> = WithIndex<TObject>;

export type ResolverTypeWrapper<T> = Promise<T> | T;

export type LegacyStitchingResolver<TResult, TParent, TContext, TArgs> = {
  fragment: string;
  resolve: ResolverFn<TResult, TParent, TContext, TArgs>;
};

export type NewStitchingResolver<TResult, TParent, TContext, TArgs> = {
  selectionSet: string;
  resolve: ResolverFn<TResult, TParent, TContext, TArgs>;
};
export type StitchingResolver<TResult, TParent, TContext, TArgs> =
  | LegacyStitchingResolver<TResult, TParent, TContext, TArgs>
  | NewStitchingResolver<TResult, TParent, TContext, TArgs>;
export type Resolver<TResult, TParent = {}, TContext = {}, TArgs = {}> =
  | ResolverFn<TResult, TParent, TContext, TArgs>
  | StitchingResolver<TResult, TParent, TContext, TArgs>;

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
) => AsyncIterator<TResult> | Promise<AsyncIterator<TResult>>;

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
  ID: ResolverTypeWrapper<Scalars['ID']>;
  String: ResolverTypeWrapper<Scalars['String']>;
  AuthConfig: ResolverTypeWrapper<AuthConfig>;
  Branch: ResolverTypeWrapper<Branch>;
  BranchInfo: ResolverTypeWrapper<BranchInfo>;
  BranchInput: BranchInput;
  BranchQueryArgs: BranchQueryArgs;
  Commit: ResolverTypeWrapper<Commit>;
  Boolean: ResolverTypeWrapper<Scalars['Boolean']>;
  Int: ResolverTypeWrapper<Scalars['Int']>;
  Float: ResolverTypeWrapper<Scalars['Float']>;
  CommitInput: CommitInput;
  CommitsQueryArgs: CommitsQueryArgs;
  CreateBranchArgs: CreateBranchArgs;
  CreatePipelineArgs: CreatePipelineArgs;
  CreateRepoArgs: CreateRepoArgs;
  CronInput: ResolverTypeWrapper<CronInput>;
  DagQueryArgs: DagQueryArgs;
  File: ResolverTypeWrapper<File>;
  FileFromURL: FileFromUrl;
  FileQueryArgs: FileQueryArgs;
  FileType: FileType;
  GitInput: ResolverTypeWrapper<GitInput>;
  Input: ResolverTypeWrapper<Input>;
  InputPipeline: ResolverTypeWrapper<InputPipeline>;
  InputType: InputType;
  Job: ResolverTypeWrapper<Job>;
  JobQueryArgs: JobQueryArgs;
  JobSet: ResolverTypeWrapper<JobSet>;
  JobSetQueryArgs: JobSetQueryArgs;
  JobSetsQueryArgs: JobSetsQueryArgs;
  JobState: JobState;
  JobsQueryArgs: JobsQueryArgs;
  Log: ResolverTypeWrapper<Log>;
  LogsArgs: LogsArgs;
  Mutation: ResolverTypeWrapper<{}>;
  NodeSelector: ResolverTypeWrapper<NodeSelector>;
  NodeState: NodeState;
  NodeType: NodeType;
  OriginKind: OriginKind;
  PFS: Pfs;
  PFSInput: ResolverTypeWrapper<PfsInput>;
  Pach: ResolverTypeWrapper<Pach>;
  Pipeline: ResolverTypeWrapper<Pipeline>;
  PipelineQueryArgs: PipelineQueryArgs;
  PipelineState: PipelineState;
  PipelineType: PipelineType;
  Project: ResolverTypeWrapper<Project>;
  ProjectDetails: ResolverTypeWrapper<ProjectDetails>;
  ProjectDetailsQueryArgs: ProjectDetailsQueryArgs;
  ProjectStatus: ProjectStatus;
  PutFilesFromURLsArgs: PutFilesFromUrLsArgs;
  Query: ResolverTypeWrapper<{}>;
  Repo: ResolverTypeWrapper<Repo>;
  RepoInfo: ResolverTypeWrapper<RepoInfo>;
  RepoInput: RepoInput;
  RepoQueryArgs: RepoQueryArgs;
  SchedulingSpec: ResolverTypeWrapper<SchedulingSpec>;
  SearchResultQueryArgs: SearchResultQueryArgs;
  SearchResults: ResolverTypeWrapper<SearchResults>;
  Subscription: ResolverTypeWrapper<{}>;
  Timestamp: ResolverTypeWrapper<Timestamp>;
  Tokens: ResolverTypeWrapper<Tokens>;
  Transform: ResolverTypeWrapper<Transform>;
  TransformInput: TransformInput;
  Vertex: ResolverTypeWrapper<Vertex>;
  WorkspaceLogsArgs: WorkspaceLogsArgs;
}>;

/** Mapping between all available schema types and the resolvers parents */
export type ResolversParentTypes = ResolversObject<{
  Account: Account;
  ID: Scalars['ID'];
  String: Scalars['String'];
  AuthConfig: AuthConfig;
  Branch: Branch;
  BranchInfo: BranchInfo;
  BranchInput: BranchInput;
  BranchQueryArgs: BranchQueryArgs;
  Commit: Commit;
  Boolean: Scalars['Boolean'];
  Int: Scalars['Int'];
  Float: Scalars['Float'];
  CommitInput: CommitInput;
  CommitsQueryArgs: CommitsQueryArgs;
  CreateBranchArgs: CreateBranchArgs;
  CreatePipelineArgs: CreatePipelineArgs;
  CreateRepoArgs: CreateRepoArgs;
  CronInput: CronInput;
  DagQueryArgs: DagQueryArgs;
  File: File;
  FileFromURL: FileFromUrl;
  FileQueryArgs: FileQueryArgs;
  GitInput: GitInput;
  Input: Input;
  InputPipeline: InputPipeline;
  Job: Job;
  JobQueryArgs: JobQueryArgs;
  JobSet: JobSet;
  JobSetQueryArgs: JobSetQueryArgs;
  JobSetsQueryArgs: JobSetsQueryArgs;
  JobsQueryArgs: JobsQueryArgs;
  Log: Log;
  LogsArgs: LogsArgs;
  Mutation: {};
  NodeSelector: NodeSelector;
  PFS: Pfs;
  PFSInput: PfsInput;
  Pach: Pach;
  Pipeline: Pipeline;
  PipelineQueryArgs: PipelineQueryArgs;
  Project: Project;
  ProjectDetails: ProjectDetails;
  ProjectDetailsQueryArgs: ProjectDetailsQueryArgs;
  PutFilesFromURLsArgs: PutFilesFromUrLsArgs;
  Query: {};
  Repo: Repo;
  RepoInfo: RepoInfo;
  RepoInput: RepoInput;
  RepoQueryArgs: RepoQueryArgs;
  SchedulingSpec: SchedulingSpec;
  SearchResultQueryArgs: SearchResultQueryArgs;
  SearchResults: SearchResults;
  Subscription: {};
  Timestamp: Timestamp;
  Tokens: Tokens;
  Transform: Transform;
  TransformInput: TransformInput;
  Vertex: Vertex;
  WorkspaceLogsArgs: WorkspaceLogsArgs;
}>;

export type AccountResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Account'] = ResolversParentTypes['Account'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  email?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  name?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
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
  repoName?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  branch?: Resolver<Maybe<ResolversTypes['Branch']>, ParentType, ContextType>;
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  originKind?: Resolver<
    Maybe<ResolversTypes['OriginKind']>,
    ParentType,
    ContextType
  >;
  hasLinkedJob?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  started?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  finished?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
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

export type FileResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['File'] = ResolversParentTypes['File'],
> = ResolversObject<{
  committed?: Resolver<
    Maybe<ResolversTypes['Timestamp']>,
    ParentType,
    ContextType
  >;
  commitId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  download?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  hash?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  path?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  repoName?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  downloadDisabled?: Resolver<
    Maybe<ResolversTypes['Boolean']>,
    ParentType,
    ContextType
  >;
  type?: Resolver<ResolversTypes['FileType'], ParentType, ContextType>;
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
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  type?: Resolver<ResolversTypes['InputType'], ParentType, ContextType>;
  joinedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  groupedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  crossedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  unionedWith?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  pfsInput?: Resolver<
    Maybe<ResolversTypes['PFSInput']>,
    ParentType,
    ContextType
  >;
  cronInput?: Resolver<
    Maybe<ResolversTypes['CronInput']>,
    ParentType,
    ContextType
  >;
  gitInput?: Resolver<
    Maybe<ResolversTypes['GitInput']>,
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
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  createdAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  startedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  finishedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  pipelineName?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  transform?: Resolver<
    Maybe<ResolversTypes['Transform']>,
    ParentType,
    ContextType
  >;
  inputString?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  inputBranch?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  outputBranch?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  reason?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  jsonDetails?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  dataProcessed?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataSkipped?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataFailed?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataRecovered?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  dataTotal?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type JobSetResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['JobSet'] = ResolversParentTypes['JobSet'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  createdAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  jobs?: Resolver<Array<ResolversTypes['Job']>, ParentType, ContextType>;
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
  exchangeCode?: Resolver<
    ResolversTypes['Tokens'],
    ParentType,
    ContextType,
    RequireFields<MutationExchangeCodeArgs, 'code'>
  >;
  createRepo?: Resolver<
    ResolversTypes['Repo'],
    ParentType,
    ContextType,
    RequireFields<MutationCreateRepoArgs, 'args'>
  >;
  createPipeline?: Resolver<
    ResolversTypes['Pipeline'],
    ParentType,
    ContextType,
    RequireFields<MutationCreatePipelineArgs, 'args'>
  >;
  createBranch?: Resolver<
    ResolversTypes['Branch'],
    ParentType,
    ContextType,
    RequireFields<MutationCreateBranchArgs, 'args'>
  >;
  putFilesFromURLs?: Resolver<
    Array<ResolversTypes['String']>,
    ParentType,
    ContextType,
    RequireFields<MutationPutFilesFromUrLsArgs, 'args'>
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

export type PipelineResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Pipeline'] = ResolversParentTypes['Pipeline'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  version?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  state?: Resolver<ResolversTypes['PipelineState'], ParentType, ContextType>;
  stopped?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  recentError?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  numOfJobsCreated?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  numOfJobsStarting?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  numOfJobsRunning?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  numOfJobsFailing?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  numOfJobsSucceeding?: Resolver<
    ResolversTypes['Int'],
    ParentType,
    ContextType
  >;
  numOfJobsKilled?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  numOfJobsEgressing?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  lastJobState?: Resolver<
    Maybe<ResolversTypes['JobState']>,
    ParentType,
    ContextType
  >;
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  type?: Resolver<ResolversTypes['PipelineType'], ParentType, ContextType>;
  datumTimeoutS?: Resolver<
    Maybe<ResolversTypes['Int']>,
    ParentType,
    ContextType
  >;
  datumTries?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  jobTimeoutS?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  outputBranch?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  s3OutputRepo?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  egress?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  jsonSpec?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  reason?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type ProjectResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Project'] = ResolversParentTypes['Project'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  status?: Resolver<ResolversTypes['ProjectStatus'], ParentType, ContextType>;
  description?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type ProjectDetailsResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['ProjectDetails'] = ResolversParentTypes['ProjectDetails'],
> = ResolversObject<{
  repoCount?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  pipelineCount?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  jobSets?: Resolver<Array<ResolversTypes['JobSet']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type QueryResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Query'] = ResolversParentTypes['Query'],
> = ResolversObject<{
  account?: Resolver<ResolversTypes['Account'], ParentType, ContextType>;
  authConfig?: Resolver<ResolversTypes['AuthConfig'], ParentType, ContextType>;
  branch?: Resolver<
    ResolversTypes['Branch'],
    ParentType,
    ContextType,
    RequireFields<QueryBranchArgs, 'args'>
  >;
  commits?: Resolver<
    Array<ResolversTypes['Commit']>,
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
  files?: Resolver<
    Array<ResolversTypes['File']>,
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
    Array<ResolversTypes['JobSet']>,
    ParentType,
    ContextType,
    RequireFields<QueryJobSetsArgs, 'args'>
  >;
  jobs?: Resolver<
    Array<ResolversTypes['Job']>,
    ParentType,
    ContextType,
    RequireFields<QueryJobsArgs, 'args'>
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
  branches?: Resolver<Array<ResolversTypes['Branch']>, ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  description?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  linkedPipeline?: Resolver<
    Maybe<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
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
  pipelines?: Resolver<
    Array<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
  repos?: Resolver<Array<ResolversTypes['Repo']>, ParentType, ContextType>;
  jobSet?: Resolver<Maybe<ResolversTypes['JobSet']>, ParentType, ContextType>;
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
  workspaceLogs?: SubscriptionResolver<
    ResolversTypes['Log'],
    'workspaceLogs',
    ParentType,
    ContextType,
    RequireFields<SubscriptionWorkspaceLogsArgs, 'args'>
  >;
  logs?: SubscriptionResolver<
    ResolversTypes['Log'],
    'logs',
    ParentType,
    ContextType,
    RequireFields<SubscriptionLogsArgs, 'args'>
  >;
}>;

export type TimestampResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Timestamp'] = ResolversParentTypes['Timestamp'],
> = ResolversObject<{
  seconds?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  nanos?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type TokensResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Tokens'] = ResolversParentTypes['Tokens'],
> = ResolversObject<{
  pachToken?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  idToken?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type TransformResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Transform'] = ResolversParentTypes['Transform'],
> = ResolversObject<{
  cmdList?: Resolver<Array<ResolversTypes['String']>, ParentType, ContextType>;
  image?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type VertexResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Vertex'] = ResolversParentTypes['Vertex'],
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  state?: Resolver<Maybe<ResolversTypes['NodeState']>, ParentType, ContextType>;
  access?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  parents?: Resolver<Array<ResolversTypes['String']>, ParentType, ContextType>;
  type?: Resolver<ResolversTypes['NodeType'], ParentType, ContextType>;
  jobState?: Resolver<
    Maybe<ResolversTypes['JobState']>,
    ParentType,
    ContextType
  >;
  createdAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type Resolvers<ContextType = Context> = ResolversObject<{
  Account?: AccountResolvers<ContextType>;
  AuthConfig?: AuthConfigResolvers<ContextType>;
  Branch?: BranchResolvers<ContextType>;
  BranchInfo?: BranchInfoResolvers<ContextType>;
  Commit?: CommitResolvers<ContextType>;
  CronInput?: CronInputResolvers<ContextType>;
  File?: FileResolvers<ContextType>;
  GitInput?: GitInputResolvers<ContextType>;
  Input?: InputResolvers<ContextType>;
  InputPipeline?: InputPipelineResolvers<ContextType>;
  Job?: JobResolvers<ContextType>;
  JobSet?: JobSetResolvers<ContextType>;
  Log?: LogResolvers<ContextType>;
  Mutation?: MutationResolvers<ContextType>;
  NodeSelector?: NodeSelectorResolvers<ContextType>;
  PFSInput?: PfsInputResolvers<ContextType>;
  Pach?: PachResolvers<ContextType>;
  Pipeline?: PipelineResolvers<ContextType>;
  Project?: ProjectResolvers<ContextType>;
  ProjectDetails?: ProjectDetailsResolvers<ContextType>;
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

/**
 * @deprecated
 * Use "Resolvers" root object instead. If you wish to get "IResolvers", add "typesPrefix: I" to your config.
 */
export type IResolvers<ContextType = Context> = Resolvers<ContextType>;

export type JobOverviewFragment = {__typename?: 'Job'} & Pick<
  Job,
  | 'id'
  | 'state'
  | 'createdAt'
  | 'startedAt'
  | 'finishedAt'
  | 'pipelineName'
  | 'reason'
  | 'dataProcessed'
  | 'dataSkipped'
  | 'dataFailed'
  | 'dataTotal'
  | 'dataRecovered'
>;

export type JobSetFieldsFragment = {__typename?: 'JobSet'} & Pick<
  JobSet,
  'id' | 'state' | 'createdAt'
> & {
    jobs: Array<
      {__typename?: 'Job'} & Pick<Job, 'inputString' | 'inputBranch'> & {
          transform?: Maybe<
            {__typename?: 'Transform'} & Pick<Transform, 'cmdList' | 'image'>
          >;
        } & JobOverviewFragment
    >;
  };

export type LogFieldsFragment = {__typename?: 'Log'} & Pick<
  Log,
  'user' | 'message'
> & {
    timestamp?: Maybe<
      {__typename?: 'Timestamp'} & Pick<Timestamp, 'seconds' | 'nanos'>
    >;
  };

export type CreateBranchMutationVariables = Exact<{
  args: CreateBranchArgs;
}>;

export type CreateBranchMutation = {__typename?: 'Mutation'} & {
  createBranch: {__typename?: 'Branch'} & Pick<Branch, 'name'> & {
      repo?: Maybe<{__typename?: 'RepoInfo'} & Pick<RepoInfo, 'name'>>;
    };
};

export type CreatePipelineMutationVariables = Exact<{
  args: CreatePipelineArgs;
}>;

export type CreatePipelineMutation = {__typename?: 'Mutation'} & {
  createPipeline: {__typename?: 'Pipeline'} & Pick<
    Pipeline,
    | 'id'
    | 'name'
    | 'state'
    | 'type'
    | 'description'
    | 'datumTimeoutS'
    | 'datumTries'
    | 'jobTimeoutS'
    | 'outputBranch'
    | 's3OutputRepo'
    | 'egress'
    | 'jsonSpec'
    | 'reason'
  >;
};

export type CreateRepoMutationVariables = Exact<{
  args: CreateRepoArgs;
}>;

export type CreateRepoMutation = {__typename?: 'Mutation'} & {
  createRepo: {__typename?: 'Repo'} & Pick<
    Repo,
    'createdAt' | 'description' | 'id' | 'name' | 'sizeDisplay'
  >;
};

export type ExchangeCodeMutationVariables = Exact<{
  code: Scalars['String'];
}>;

export type ExchangeCodeMutation = {__typename?: 'Mutation'} & {
  exchangeCode: {__typename?: 'Tokens'} & Pick<Tokens, 'pachToken' | 'idToken'>;
};

export type PutFilesFromUrLsMutationVariables = Exact<{
  args: PutFilesFromUrLsArgs;
}>;

export type PutFilesFromUrLsMutation = {__typename?: 'Mutation'} & Pick<
  Mutation,
  'putFilesFromURLs'
>;

export type GetAccountQueryVariables = Exact<{[key: string]: never}>;

export type GetAccountQuery = {__typename?: 'Query'} & {
  account: {__typename?: 'Account'} & Pick<Account, 'id' | 'email' | 'name'>;
};

export type AuthConfigQueryVariables = Exact<{[key: string]: never}>;

export type AuthConfigQuery = {__typename?: 'Query'} & {
  authConfig: {__typename?: 'AuthConfig'} & Pick<
    AuthConfig,
    'authEndpoint' | 'clientId' | 'pachdClientId'
  >;
};

export type GetCommitsQueryVariables = Exact<{
  args: CommitsQueryArgs;
}>;

export type GetCommitsQuery = {__typename?: 'Query'} & {
  commits: Array<
    {__typename?: 'Commit'} & Pick<
      Commit,
      | 'repoName'
      | 'description'
      | 'originKind'
      | 'id'
      | 'started'
      | 'finished'
      | 'sizeBytes'
      | 'sizeDisplay'
      | 'hasLinkedJob'
    > & {branch?: Maybe<{__typename?: 'Branch'} & Pick<Branch, 'name'>>}
  >;
};

export type GetDagQueryVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagQuery = {__typename?: 'Query'} & {
  dag: Array<
    {__typename?: 'Vertex'} & Pick<
      Vertex,
      | 'name'
      | 'state'
      | 'access'
      | 'parents'
      | 'type'
      | 'jobState'
      | 'createdAt'
    >
  >;
};

export type GetDagsSubscriptionVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagsSubscription = {__typename?: 'Subscription'} & {
  dags: Array<
    {__typename?: 'Vertex'} & Pick<
      Vertex,
      | 'name'
      | 'state'
      | 'access'
      | 'parents'
      | 'type'
      | 'jobState'
      | 'createdAt'
    >
  >;
};

export type GetFilesQueryVariables = Exact<{
  args: FileQueryArgs;
}>;

export type GetFilesQuery = {__typename?: 'Query'} & {
  files: Array<
    {__typename?: 'File'} & Pick<
      File,
      | 'commitId'
      | 'download'
      | 'hash'
      | 'path'
      | 'repoName'
      | 'sizeBytes'
      | 'type'
      | 'sizeDisplay'
      | 'downloadDisabled'
    > & {
        committed?: Maybe<
          {__typename?: 'Timestamp'} & Pick<Timestamp, 'nanos' | 'seconds'>
        >;
      }
  >;
};

export type JobQueryVariables = Exact<{
  args: JobQueryArgs;
}>;

export type JobQuery = {__typename?: 'Query'} & {
  job: {__typename?: 'Job'} & Pick<
    Job,
    'inputString' | 'inputBranch' | 'outputBranch' | 'reason' | 'jsonDetails'
  > & {
      transform?: Maybe<
        {__typename?: 'Transform'} & Pick<Transform, 'cmdList' | 'image'>
      >;
    } & JobOverviewFragment;
};

export type JobSetsQueryVariables = Exact<{
  args: JobSetsQueryArgs;
}>;

export type JobSetsQuery = {__typename?: 'Query'} & {
  jobSets: Array<{__typename?: 'JobSet'} & JobSetFieldsFragment>;
};

export type JobsQueryVariables = Exact<{
  args: JobsQueryArgs;
}>;

export type JobsQuery = {__typename?: 'Query'} & {
  jobs: Array<
    {__typename?: 'Job'} & Pick<Job, 'inputString' | 'inputBranch'> & {
        transform?: Maybe<
          {__typename?: 'Transform'} & Pick<Transform, 'cmdList' | 'image'>
        >;
      } & JobOverviewFragment
  >;
};

export type JobSetQueryVariables = Exact<{
  args: JobSetQueryArgs;
}>;

export type JobSetQuery = {__typename?: 'Query'} & {
  jobSet: {__typename?: 'JobSet'} & JobSetFieldsFragment;
};

export type GetWorkspaceLogsQueryVariables = Exact<{
  args: WorkspaceLogsArgs;
}>;

export type GetWorkspaceLogsQuery = {__typename?: 'Query'} & {
  workspaceLogs: Array<Maybe<{__typename?: 'Log'} & LogFieldsFragment>>;
};

export type GetLogsQueryVariables = Exact<{
  args: LogsArgs;
}>;

export type GetLogsQuery = {__typename?: 'Query'} & {
  logs: Array<Maybe<{__typename?: 'Log'} & LogFieldsFragment>>;
};

export type GetWorkspaceLogStreamSubscriptionVariables = Exact<{
  args: WorkspaceLogsArgs;
}>;

export type GetWorkspaceLogStreamSubscription = {
  __typename?: 'Subscription';
} & {workspaceLogs: {__typename?: 'Log'} & LogFieldsFragment};

export type GetLogsStreamSubscriptionVariables = Exact<{
  args: LogsArgs;
}>;

export type GetLogsStreamSubscription = {__typename?: 'Subscription'} & {
  logs: {__typename?: 'Log'} & LogFieldsFragment;
};

export type PipelineQueryVariables = Exact<{
  args: PipelineQueryArgs;
}>;

export type PipelineQuery = {__typename?: 'Query'} & {
  pipeline: {__typename?: 'Pipeline'} & Pick<
    Pipeline,
    | 'id'
    | 'name'
    | 'state'
    | 'type'
    | 'description'
    | 'datumTimeoutS'
    | 'datumTries'
    | 'jobTimeoutS'
    | 'outputBranch'
    | 's3OutputRepo'
    | 'egress'
    | 'jsonSpec'
    | 'reason'
  >;
};

export type ProjectDetailsQueryVariables = Exact<{
  args: ProjectDetailsQueryArgs;
}>;

export type ProjectDetailsQuery = {__typename?: 'Query'} & {
  projectDetails: {__typename?: 'ProjectDetails'} & Pick<
    ProjectDetails,
    'sizeDisplay' | 'repoCount' | 'pipelineCount'
  > & {jobSets: Array<{__typename?: 'JobSet'} & JobSetFieldsFragment>};
};

export type ProjectQueryVariables = Exact<{
  id: Scalars['ID'];
}>;

export type ProjectQuery = {__typename?: 'Query'} & {
  project: {__typename?: 'Project'} & Pick<
    Project,
    'id' | 'name' | 'description' | 'createdAt' | 'status'
  >;
};

export type ProjectsQueryVariables = Exact<{[key: string]: never}>;

export type ProjectsQuery = {__typename?: 'Query'} & {
  projects: Array<
    {__typename?: 'Project'} & Pick<
      Project,
      'id' | 'name' | 'description' | 'createdAt' | 'status'
    >
  >;
};

export type RepoQueryVariables = Exact<{
  args: RepoQueryArgs;
}>;

export type RepoQuery = {__typename?: 'Query'} & {
  repo: {__typename?: 'Repo'} & Pick<
    Repo,
    'createdAt' | 'description' | 'id' | 'name' | 'sizeDisplay'
  > & {
      branches: Array<{__typename?: 'Branch'} & Pick<Branch, 'name'>>;
      linkedPipeline?: Maybe<
        {__typename?: 'Pipeline'} & Pick<Pipeline, 'id' | 'name'>
      >;
    };
};

export type SearchResultsQueryVariables = Exact<{
  args: SearchResultQueryArgs;
}>;

export type SearchResultsQuery = {__typename?: 'Query'} & {
  searchResults: {__typename?: 'SearchResults'} & {
    pipelines: Array<{__typename?: 'Pipeline'} & Pick<Pipeline, 'name' | 'id'>>;
    repos: Array<{__typename?: 'Repo'} & Pick<Repo, 'name' | 'id'>>;
    jobSet?: Maybe<{__typename?: 'JobSet'} & Pick<JobSet, 'id'>>;
  };
};

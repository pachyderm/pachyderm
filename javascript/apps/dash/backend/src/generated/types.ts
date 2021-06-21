/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable @typescript-eslint/ban-types */
/* eslint-disable import/no-duplicates */
/* eslint-disable @typescript-eslint/no-explicit-any */

import {GraphQLResolveInfo} from 'graphql';

import {Context} from '@dash-backend/lib/types';
export type Maybe<T> = T | null;
export type Exact<T extends {[key: string]: unknown}> = {[K in keyof T]: T[K]};
export type MakeOptional<T, K extends keyof T> = Omit<T, K> &
  {[SubKey in K]?: Maybe<T[SubKey]>};
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> &
  {[SubKey in K]: Maybe<T[SubKey]>};
export type RequireFields<T, K extends keyof T> = {
  [X in Exclude<keyof T, K>]?: T[X];
} &
  {[P in K]-?: NonNullable<T[P]>};
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
  authUrl: Scalars['String'];
  clientId: Scalars['String'];
  pachdClientId: Scalars['String'];
};

export type Branch = {
  __typename?: 'Branch';
  id: Scalars['ID'];
  name: Scalars['ID'];
};

export type Commit = {
  __typename?: 'Commit';
  branch?: Maybe<Branch>;
  description?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  started: Scalars['Int'];
  finished: Scalars['Int'];
  sizeBytes: Scalars['Int'];
  sizeDisplay: Scalars['String'];
};

export type CronInput = {
  __typename?: 'CronInput';
  name: Scalars['String'];
  repo: Repo;
};

export type Dag = {
  __typename?: 'Dag';
  id: Scalars['String'];
  nodes: Array<Node>;
  links: Array<Link>;
};

export enum DagDirection {
  UP = 'UP',
  DOWN = 'DOWN',
  LEFT = 'LEFT',
  RIGHT = 'RIGHT',
}

export type DagQueryArgs = {
  projectId: Scalars['ID'];
  nodeWidth: Scalars['Int'];
  nodeHeight: Scalars['Int'];
  direction: DagDirection;
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

export type FileQueryArgs = {
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
  createdAt: Scalars['Int'];
  finishedAt?: Maybe<Scalars['Int']>;
  state: JobState;
  pipelineName: Scalars['String'];
  transform?: Maybe<Transform>;
  inputString?: Maybe<Scalars['String']>;
  inputBranch?: Maybe<Scalars['String']>;
};

export type JobQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
  pipelineName: Scalars['String'];
};

export enum JobState {
  JOB_CREATED = 'JOB_CREATED',
  JOB_STARTING = 'JOB_STARTING',
  JOB_RUNNING = 'JOB_RUNNING',
  JOB_FAILURE = 'JOB_FAILURE',
  JOB_SUCCESS = 'JOB_SUCCESS',
  JOB_KILLED = 'JOB_KILLED',
  JOB_EGRESSING = 'JOB_EGRESSING',
}

export type JobsQueryArgs = {
  projectId: Scalars['ID'];
  limit?: Maybe<Scalars['Int']>;
  pipelineId?: Maybe<Scalars['String']>;
};

export type Jobset = {
  __typename?: 'Jobset';
  id: Scalars['ID'];
  createdAt: Scalars['Int'];
  state: JobState;
  jobs: Array<Job>;
};

export type JobsetQueryArgs = {
  id: Scalars['ID'];
  projectId: Scalars['String'];
};

export type Link = {
  __typename?: 'Link';
  id: Scalars['ID'];
  source: Scalars['String'];
  sourceState?: Maybe<PipelineState>;
  targetState?: Maybe<PipelineState>;
  target: Scalars['String'];
  state?: Maybe<JobState>;
  bendPoints: Array<PointCoordinates>;
  startPoint: PointCoordinates;
  endPoint: PointCoordinates;
  transferring: Scalars['Boolean'];
};

export type Mutation = {
  __typename?: 'Mutation';
  exchangeCode: Tokens;
};

export type MutationExchangeCodeArgs = {
  code: Scalars['String'];
};

export type Node = {
  __typename?: 'Node';
  id: Scalars['ID'];
  name: Scalars['String'];
  type: NodeType;
  x: Scalars['Float'];
  y: Scalars['Float'];
  state?: Maybe<PipelineState>;
  access: Scalars['Boolean'];
};

export type NodeSelector = {
  __typename?: 'NodeSelector';
  key: Scalars['String'];
  value: Scalars['String'];
};

export enum NodeType {
  PIPELINE = 'PIPELINE',
  OUTPUT_REPO = 'OUTPUT_REPO',
  INPUT_REPO = 'INPUT_REPO',
  EGRESS = 'EGRESS',
}

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
  transform?: Maybe<Transform>;
  inputString: Scalars['String'];
  cacheSize: Scalars['String'];
  datumTimeoutS?: Maybe<Scalars['Int']>;
  datumTries: Scalars['Int'];
  jobTimeoutS?: Maybe<Scalars['Int']>;
  outputBranch: Scalars['String'];
  s3OutputRepo?: Maybe<Scalars['String']>;
  egress: Scalars['Boolean'];
  schedulingSpec?: Maybe<SchedulingSpec>;
};

export type PipelineQueryArgs = {
  projectId: Scalars['String'];
  id: Scalars['ID'];
};

export enum PipelineState {
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

export type PointCoordinates = {
  __typename?: 'PointCoordinates';
  x: Scalars['Float'];
  y: Scalars['Float'];
};

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
  jobs: Array<Job>;
};

export type ProjectDetailsQueryArgs = {
  projectId: Scalars['String'];
  jobsLimit?: Maybe<Scalars['Int']>;
};

export enum ProjectStatus {
  HEALTHY = 'HEALTHY',
  UNHEALTHY = 'UNHEALTHY',
}

export type Query = {
  __typename?: 'Query';
  account: Account;
  authConfig: AuthConfig;
  dag: Dag;
  files: Array<File>;
  job: Job;
  jobs: Array<Job>;
  jobset: Jobset;
  loggedIn: Scalars['Boolean'];
  pipeline: Pipeline;
  project: Project;
  projectDetails: ProjectDetails;
  projects: Array<Project>;
  repo: Repo;
  searchResults: SearchResults;
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

export type QueryJobsArgs = {
  args: JobsQueryArgs;
};

export type QueryJobsetArgs = {
  args: JobsetQueryArgs;
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

export type Repo = {
  __typename?: 'Repo';
  branches: Array<Branch>;
  commits: Array<Commit>;
  createdAt: Scalars['Int'];
  description: Scalars['String'];
  id: Scalars['ID'];
  name: Scalars['ID'];
  sizeBytes: Scalars['Int'];
  sizeDisplay: Scalars['String'];
  linkedPipeline?: Maybe<Pipeline>;
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
  jobset?: Maybe<Jobset>;
};

export type Subscription = {
  __typename?: 'Subscription';
  dags: Array<Dag>;
};

export type SubscriptionDagsArgs = {
  args: DagQueryArgs;
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
  Commit: ResolverTypeWrapper<Commit>;
  Int: ResolverTypeWrapper<Scalars['Int']>;
  CronInput: ResolverTypeWrapper<CronInput>;
  Dag: ResolverTypeWrapper<Dag>;
  DagDirection: DagDirection;
  DagQueryArgs: DagQueryArgs;
  File: ResolverTypeWrapper<File>;
  Float: ResolverTypeWrapper<Scalars['Float']>;
  Boolean: ResolverTypeWrapper<Scalars['Boolean']>;
  FileQueryArgs: FileQueryArgs;
  FileType: FileType;
  GitInput: ResolverTypeWrapper<GitInput>;
  Input: ResolverTypeWrapper<Input>;
  InputPipeline: ResolverTypeWrapper<InputPipeline>;
  InputType: InputType;
  Job: ResolverTypeWrapper<Job>;
  JobQueryArgs: JobQueryArgs;
  JobState: JobState;
  JobsQueryArgs: JobsQueryArgs;
  Jobset: ResolverTypeWrapper<Jobset>;
  JobsetQueryArgs: JobsetQueryArgs;
  Link: ResolverTypeWrapper<Link>;
  Mutation: ResolverTypeWrapper<{}>;
  Node: ResolverTypeWrapper<Node>;
  NodeSelector: ResolverTypeWrapper<NodeSelector>;
  NodeType: NodeType;
  PFSInput: ResolverTypeWrapper<PfsInput>;
  Pach: ResolverTypeWrapper<Pach>;
  Pipeline: ResolverTypeWrapper<Pipeline>;
  PipelineQueryArgs: PipelineQueryArgs;
  PipelineState: PipelineState;
  PipelineType: PipelineType;
  PointCoordinates: ResolverTypeWrapper<PointCoordinates>;
  Project: ResolverTypeWrapper<Project>;
  ProjectDetails: ResolverTypeWrapper<ProjectDetails>;
  ProjectDetailsQueryArgs: ProjectDetailsQueryArgs;
  ProjectStatus: ProjectStatus;
  Query: ResolverTypeWrapper<{}>;
  Repo: ResolverTypeWrapper<Repo>;
  RepoQueryArgs: RepoQueryArgs;
  SchedulingSpec: ResolverTypeWrapper<SchedulingSpec>;
  SearchResultQueryArgs: SearchResultQueryArgs;
  SearchResults: ResolverTypeWrapper<SearchResults>;
  Subscription: ResolverTypeWrapper<{}>;
  Timestamp: ResolverTypeWrapper<Timestamp>;
  Tokens: ResolverTypeWrapper<Tokens>;
  Transform: ResolverTypeWrapper<Transform>;
}>;

/** Mapping between all available schema types and the resolvers parents */
export type ResolversParentTypes = ResolversObject<{
  Account: Account;
  ID: Scalars['ID'];
  String: Scalars['String'];
  AuthConfig: AuthConfig;
  Branch: Branch;
  Commit: Commit;
  Int: Scalars['Int'];
  CronInput: CronInput;
  Dag: Dag;
  DagQueryArgs: DagQueryArgs;
  File: File;
  Float: Scalars['Float'];
  Boolean: Scalars['Boolean'];
  FileQueryArgs: FileQueryArgs;
  GitInput: GitInput;
  Input: Input;
  InputPipeline: InputPipeline;
  Job: Job;
  JobQueryArgs: JobQueryArgs;
  JobsQueryArgs: JobsQueryArgs;
  Jobset: Jobset;
  JobsetQueryArgs: JobsetQueryArgs;
  Link: Link;
  Mutation: {};
  Node: Node;
  NodeSelector: NodeSelector;
  PFSInput: PfsInput;
  Pach: Pach;
  Pipeline: Pipeline;
  PipelineQueryArgs: PipelineQueryArgs;
  PointCoordinates: PointCoordinates;
  Project: Project;
  ProjectDetails: ProjectDetails;
  ProjectDetailsQueryArgs: ProjectDetailsQueryArgs;
  Query: {};
  Repo: Repo;
  RepoQueryArgs: RepoQueryArgs;
  SchedulingSpec: SchedulingSpec;
  SearchResultQueryArgs: SearchResultQueryArgs;
  SearchResults: SearchResults;
  Subscription: {};
  Timestamp: Timestamp;
  Tokens: Tokens;
  Transform: Transform;
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
  authUrl?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  clientId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  pachdClientId?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type BranchResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Branch'] = ResolversParentTypes['Branch'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
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
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  started?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  finished?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
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

export type DagResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Dag'] = ResolversParentTypes['Dag'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  nodes?: Resolver<Array<ResolversTypes['Node']>, ParentType, ContextType>;
  links?: Resolver<Array<ResolversTypes['Link']>, ParentType, ContextType>;
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
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
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
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type JobsetResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Jobset'] = ResolversParentTypes['Jobset'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  jobs?: Resolver<Array<ResolversTypes['Job']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type LinkResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Link'] = ResolversParentTypes['Link'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  source?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  sourceState?: Resolver<
    Maybe<ResolversTypes['PipelineState']>,
    ParentType,
    ContextType
  >;
  targetState?: Resolver<
    Maybe<ResolversTypes['PipelineState']>,
    ParentType,
    ContextType
  >;
  target?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  state?: Resolver<Maybe<ResolversTypes['JobState']>, ParentType, ContextType>;
  bendPoints?: Resolver<
    Array<ResolversTypes['PointCoordinates']>,
    ParentType,
    ContextType
  >;
  startPoint?: Resolver<
    ResolversTypes['PointCoordinates'],
    ParentType,
    ContextType
  >;
  endPoint?: Resolver<
    ResolversTypes['PointCoordinates'],
    ParentType,
    ContextType
  >;
  transferring?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
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
}>;

export type NodeResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Node'] = ResolversParentTypes['Node'],
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  type?: Resolver<ResolversTypes['NodeType'], ParentType, ContextType>;
  x?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  y?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  state?: Resolver<
    Maybe<ResolversTypes['PipelineState']>,
    ParentType,
    ContextType
  >;
  access?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
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
  transform?: Resolver<
    Maybe<ResolversTypes['Transform']>,
    ParentType,
    ContextType
  >;
  inputString?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  cacheSize?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
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
  schedulingSpec?: Resolver<
    Maybe<ResolversTypes['SchedulingSpec']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PointCoordinatesResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PointCoordinates'] = ResolversParentTypes['PointCoordinates'],
> = ResolversObject<{
  x?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
  y?: Resolver<ResolversTypes['Float'], ParentType, ContextType>;
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
  jobs?: Resolver<Array<ResolversTypes['Job']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type QueryResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Query'] = ResolversParentTypes['Query'],
> = ResolversObject<{
  account?: Resolver<ResolversTypes['Account'], ParentType, ContextType>;
  authConfig?: Resolver<ResolversTypes['AuthConfig'], ParentType, ContextType>;
  dag?: Resolver<
    ResolversTypes['Dag'],
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
  jobs?: Resolver<
    Array<ResolversTypes['Job']>,
    ParentType,
    ContextType,
    RequireFields<QueryJobsArgs, 'args'>
  >;
  jobset?: Resolver<
    ResolversTypes['Jobset'],
    ParentType,
    ContextType,
    RequireFields<QueryJobsetArgs, 'args'>
  >;
  loggedIn?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
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
}>;

export type RepoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Repo'] = ResolversParentTypes['Repo'],
> = ResolversObject<{
  branches?: Resolver<Array<ResolversTypes['Branch']>, ParentType, ContextType>;
  commits?: Resolver<Array<ResolversTypes['Commit']>, ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  description?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  sizeBytes?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  sizeDisplay?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  linkedPipeline?: Resolver<
    Maybe<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
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
  jobset?: Resolver<Maybe<ResolversTypes['Jobset']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type SubscriptionResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Subscription'] = ResolversParentTypes['Subscription'],
> = ResolversObject<{
  dags?: SubscriptionResolver<
    Array<ResolversTypes['Dag']>,
    'dags',
    ParentType,
    ContextType,
    RequireFields<SubscriptionDagsArgs, 'args'>
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

export type Resolvers<ContextType = Context> = ResolversObject<{
  Account?: AccountResolvers<ContextType>;
  AuthConfig?: AuthConfigResolvers<ContextType>;
  Branch?: BranchResolvers<ContextType>;
  Commit?: CommitResolvers<ContextType>;
  CronInput?: CronInputResolvers<ContextType>;
  Dag?: DagResolvers<ContextType>;
  File?: FileResolvers<ContextType>;
  GitInput?: GitInputResolvers<ContextType>;
  Input?: InputResolvers<ContextType>;
  InputPipeline?: InputPipelineResolvers<ContextType>;
  Job?: JobResolvers<ContextType>;
  Jobset?: JobsetResolvers<ContextType>;
  Link?: LinkResolvers<ContextType>;
  Mutation?: MutationResolvers<ContextType>;
  Node?: NodeResolvers<ContextType>;
  NodeSelector?: NodeSelectorResolvers<ContextType>;
  PFSInput?: PfsInputResolvers<ContextType>;
  Pach?: PachResolvers<ContextType>;
  Pipeline?: PipelineResolvers<ContextType>;
  PointCoordinates?: PointCoordinatesResolvers<ContextType>;
  Project?: ProjectResolvers<ContextType>;
  ProjectDetails?: ProjectDetailsResolvers<ContextType>;
  Query?: QueryResolvers<ContextType>;
  Repo?: RepoResolvers<ContextType>;
  SchedulingSpec?: SchedulingSpecResolvers<ContextType>;
  SearchResults?: SearchResultsResolvers<ContextType>;
  Subscription?: SubscriptionResolvers<ContextType>;
  Timestamp?: TimestampResolvers<ContextType>;
  Tokens?: TokensResolvers<ContextType>;
  Transform?: TransformResolvers<ContextType>;
}>;

/**
 * @deprecated
 * Use "Resolvers" root object instead. If you wish to get "IResolvers", add "typesPrefix: I" to your config.
 */
export type IResolvers<ContextType = Context> = Resolvers<ContextType>;

export type JobOverviewFragment = {__typename?: 'Job'} & Pick<
  Job,
  'id' | 'state' | 'createdAt' | 'finishedAt' | 'pipelineName'
>;

export type ExchangeCodeMutationVariables = Exact<{
  code: Scalars['String'];
}>;

export type ExchangeCodeMutation = {__typename?: 'Mutation'} & {
  exchangeCode: {__typename?: 'Tokens'} & Pick<Tokens, 'pachToken' | 'idToken'>;
};

export type GetAccountQueryVariables = Exact<{[key: string]: never}>;

export type GetAccountQuery = {__typename?: 'Query'} & {
  account: {__typename?: 'Account'} & Pick<Account, 'id' | 'email' | 'name'>;
};

export type AuthConfigQueryVariables = Exact<{[key: string]: never}>;

export type AuthConfigQuery = {__typename?: 'Query'} & {
  authConfig: {__typename?: 'AuthConfig'} & Pick<
    AuthConfig,
    'authUrl' | 'clientId' | 'pachdClientId'
  >;
};

export type GetDagQueryVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagQuery = {__typename?: 'Query'} & {
  dag: {__typename?: 'Dag'} & Pick<Dag, 'id'> & {
      nodes: Array<
        {__typename?: 'Node'} & Pick<
          Node,
          'id' | 'name' | 'type' | 'access' | 'state' | 'x' | 'y'
        >
      >;
      links: Array<
        {__typename?: 'Link'} & Pick<
          Link,
          'id' | 'source' | 'target' | 'sourceState' | 'targetState' | 'state'
        > & {
            bendPoints: Array<
              {__typename?: 'PointCoordinates'} & Pick<
                PointCoordinates,
                'x' | 'y'
              >
            >;
            startPoint: {__typename?: 'PointCoordinates'} & Pick<
              PointCoordinates,
              'x' | 'y'
            >;
            endPoint: {__typename?: 'PointCoordinates'} & Pick<
              PointCoordinates,
              'x' | 'y'
            >;
          }
      >;
    };
};

export type GetDagsSubscriptionVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagsSubscription = {__typename?: 'Subscription'} & {
  dags: Array<
    {__typename?: 'Dag'} & Pick<Dag, 'id'> & {
        nodes: Array<
          {__typename?: 'Node'} & Pick<
            Node,
            'id' | 'name' | 'type' | 'access' | 'state' | 'x' | 'y'
          >
        >;
        links: Array<
          {__typename?: 'Link'} & Pick<
            Link,
            | 'id'
            | 'source'
            | 'target'
            | 'sourceState'
            | 'targetState'
            | 'state'
            | 'transferring'
          > & {
              bendPoints: Array<
                {__typename?: 'PointCoordinates'} & Pick<
                  PointCoordinates,
                  'x' | 'y'
                >
              >;
              startPoint: {__typename?: 'PointCoordinates'} & Pick<
                PointCoordinates,
                'x' | 'y'
              >;
              endPoint: {__typename?: 'PointCoordinates'} & Pick<
                PointCoordinates,
                'x' | 'y'
              >;
            }
        >;
      }
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
  job: {__typename?: 'Job'} & Pick<Job, 'inputString' | 'inputBranch'> & {
      transform?: Maybe<
        {__typename?: 'Transform'} & Pick<Transform, 'cmdList' | 'image'>
      >;
    } & JobOverviewFragment;
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

export type JobsetQueryVariables = Exact<{
  args: JobsetQueryArgs;
}>;

export type JobsetQuery = {__typename?: 'Query'} & {
  jobset: {__typename?: 'Jobset'} & Pick<
    Jobset,
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
};

export type LoggedInQueryVariables = Exact<{[key: string]: never}>;

export type LoggedInQuery = {__typename?: 'Query'} & Pick<Query, 'loggedIn'>;

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
    | 'inputString'
    | 'cacheSize'
    | 'datumTimeoutS'
    | 'datumTries'
    | 'jobTimeoutS'
    | 'outputBranch'
    | 's3OutputRepo'
    | 'egress'
  > & {
      transform?: Maybe<
        {__typename?: 'Transform'} & Pick<Transform, 'cmdList' | 'image'>
      >;
      schedulingSpec?: Maybe<
        {__typename?: 'SchedulingSpec'} & Pick<
          SchedulingSpec,
          'priorityClassName'
        > & {
            nodeSelectorMap: Array<
              {__typename?: 'NodeSelector'} & Pick<
                NodeSelector,
                'key' | 'value'
              >
            >;
          }
      >;
    };
};

export type ProjectDetailsQueryVariables = Exact<{
  args: ProjectDetailsQueryArgs;
}>;

export type ProjectDetailsQuery = {__typename?: 'Query'} & {
  projectDetails: {__typename?: 'ProjectDetails'} & Pick<
    ProjectDetails,
    'sizeDisplay' | 'repoCount' | 'pipelineCount'
  > & {jobs: Array<{__typename?: 'Job'} & JobOverviewFragment>};
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
      branches: Array<{__typename?: 'Branch'} & Pick<Branch, 'id' | 'name'>>;
      commits: Array<
        {__typename?: 'Commit'} & Pick<
          Commit,
          'description' | 'id' | 'started' | 'finished' | 'sizeDisplay'
        > & {
            branch?: Maybe<
              {__typename?: 'Branch'} & Pick<Branch, 'id' | 'name'>
            >;
          }
      >;
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
    jobset?: Maybe<{__typename?: 'Jobset'} & Pick<Jobset, 'id'>>;
  };
};

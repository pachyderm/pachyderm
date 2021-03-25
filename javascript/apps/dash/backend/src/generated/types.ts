/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable @typescript-eslint/ban-types */
/* eslint-disable import/no-duplicates */
/* eslint-disable @typescript-eslint/no-explicit-any */

import {FileType} from '@pachyderm/proto/pb/pfs/pfs_pb';
import {PipelineState} from '@pachyderm/proto/pb/pps/pps_pb';
import {JobState} from '@pachyderm/proto/pb/pps/pps_pb';
import {ProjectStatus} from '@pachyderm/proto/pb/projects/projects_pb';
import {GraphQLResolveInfo} from 'graphql';

import {Context} from '@dash-backend/lib/types';
export type Maybe<T> = T | null;
export type Exact<T extends {[key: string]: unknown}> = {[K in keyof T]: T[K]};
export type EnumResolverSignature<T, AllowedValues = any> = {
  [key in keyof T]?: AllowedValues;
};
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

export type PfsInput = {
  __typename?: 'PFSInput';
  name: Scalars['String'];
  repo: Repo;
};

export type CronInput = {
  __typename?: 'CronInput';
  name: Scalars['String'];
  repo: Repo;
};

export type GitInput = {
  __typename?: 'GitInput';
  name: Scalars['String'];
  url: Scalars['String'];
};

export enum InputType {
  Pfs = 'PFS',
  Cron = 'CRON',
  Git = 'GIT',
}

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

export {PipelineState};

export {JobState};

export {ProjectStatus};

export type Pipeline = {
  __typename?: 'Pipeline';
  id: Scalars['ID'];
  name: Scalars['String'];
  version: Scalars['Int'];
  createdAt: Scalars['Int'];
  state: PipelineState;
  stopped: Scalars['Boolean'];
  recentError?: Maybe<Scalars['String']>;
  numOfJobsStarting: Scalars['Int'];
  numOfJobsRunning: Scalars['Int'];
  numOfJobsFailing: Scalars['Int'];
  numOfJobsSucceeding: Scalars['Int'];
  numOfJobsKilled: Scalars['Int'];
  numOfJobsEgressing: Scalars['Int'];
  lastJobState?: Maybe<JobState>;
  inputs: Array<Input>;
  description?: Maybe<Scalars['String']>;
};

export type InputPipeline = {
  __typename?: 'InputPipeline';
  id: Scalars['ID'];
};

export type Repo = {
  __typename?: 'Repo';
  name: Scalars['ID'];
  createdAt: Scalars['Int'];
  sizeInBytes: Scalars['Int'];
  description: Scalars['String'];
  isPipelineOutput: Scalars['Boolean'];
};

export type Pach = {
  __typename?: 'Pach';
  id: Scalars['ID'];
};

export type Job = {
  __typename?: 'Job';
  id: Scalars['ID'];
  createdAt: Scalars['Int'];
  state: JobState;
};

export enum NodeType {
  Pipeline = 'PIPELINE',
  Repo = 'REPO',
}

export type Node = {
  __typename?: 'Node';
  name: Scalars['String'];
  type: NodeType;
  state?: Maybe<PipelineState>;
  access: Scalars['Boolean'];
};

export type Link = {
  __typename?: 'Link';
  source: Scalars['Int'];
  target: Scalars['Int'];
  state?: Maybe<JobState>;
};

export type Dag = {
  __typename?: 'Dag';
  nodes: Array<Node>;
  links: Array<Link>;
};

export type DagQueryArgs = {
  projectId: Scalars['ID'];
};

export type JobQueryArgs = {
  projectId: Scalars['ID'];
};

export type Project = {
  __typename?: 'Project';
  id: Scalars['ID'];
  name: Scalars['String'];
  status: ProjectStatus;
  description: Scalars['String'];
  createdAt: Scalars['Int'];
};

export {FileType};

export type Timestamp = {
  __typename?: 'Timestamp';
  seconds: Scalars['Int'];
  nanos: Scalars['Int'];
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
  type: FileType;
};

export type FileQueryArgs = {
  commitId?: Maybe<Scalars['String']>;
  path?: Maybe<Scalars['String']>;
  repoName: Scalars['String'];
};

export type Tokens = {
  __typename?: 'Tokens';
  pachToken: Scalars['String'];
  idToken: Scalars['String'];
};

export type Account = {
  __typename?: 'Account';
  id: Scalars['ID'];
  email: Scalars['String'];
  name?: Maybe<Scalars['String']>;
};

export type SearchResults = {
  __typename?: 'SearchResults';
  pipelines: Array<Maybe<Pipeline>>;
};

export type Query = {
  __typename?: 'Query';
  projects: Array<Project>;
  pipelines: Array<Pipeline>;
  repos: Array<Repo>;
  jobs: Array<Job>;
  dag: Dag;
  dags: Array<Dag>;
  files: Array<File>;
  account: Account;
  searchResults: SearchResults;
};

export type QueryJobsArgs = {
  args: JobQueryArgs;
};

export type QueryDagArgs = {
  args: DagQueryArgs;
};

export type QueryDagsArgs = {
  args: DagQueryArgs;
};

export type QueryFilesArgs = {
  args: FileQueryArgs;
};

export type QuerySearchResultsArgs = {
  query: Scalars['String'];
};

export type Mutation = {
  __typename?: 'Mutation';
  exchangeCode: Tokens;
};

export type MutationExchangeCodeArgs = {
  code: Scalars['String'];
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
  TArgs
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
  TArgs
> =
  | SubscriptionSubscriberObject<TResult, TKey, TParent, TContext, TArgs>
  | SubscriptionResolverObject<TResult, TParent, TContext, TArgs>;

export type SubscriptionResolver<
  TResult,
  TKey extends string,
  TParent = {},
  TContext = {},
  TArgs = {}
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
  TArgs = {}
> = (
  next: NextResolverFn<TResult>,
  parent: TParent,
  args: TArgs,
  context: TContext,
  info: GraphQLResolveInfo,
) => TResult | Promise<TResult>;

/** Mapping between all available schema types and the resolvers types */
export type ResolversTypes = ResolversObject<{
  PFSInput: ResolverTypeWrapper<PfsInput>;
  String: ResolverTypeWrapper<Scalars['String']>;
  CronInput: ResolverTypeWrapper<CronInput>;
  GitInput: ResolverTypeWrapper<GitInput>;
  InputType: InputType;
  Input: ResolverTypeWrapper<Input>;
  ID: ResolverTypeWrapper<Scalars['ID']>;
  PipelineState: PipelineState;
  JobState: JobState;
  ProjectStatus: ProjectStatus;
  Pipeline: ResolverTypeWrapper<Pipeline>;
  Int: ResolverTypeWrapper<Scalars['Int']>;
  Boolean: ResolverTypeWrapper<Scalars['Boolean']>;
  InputPipeline: ResolverTypeWrapper<InputPipeline>;
  Repo: ResolverTypeWrapper<Repo>;
  Pach: ResolverTypeWrapper<Pach>;
  Job: ResolverTypeWrapper<Job>;
  NodeType: NodeType;
  Node: ResolverTypeWrapper<Node>;
  Link: ResolverTypeWrapper<Link>;
  Dag: ResolverTypeWrapper<Dag>;
  DagQueryArgs: DagQueryArgs;
  JobQueryArgs: JobQueryArgs;
  Project: ResolverTypeWrapper<Project>;
  FileType: FileType;
  Timestamp: ResolverTypeWrapper<Timestamp>;
  File: ResolverTypeWrapper<File>;
  Float: ResolverTypeWrapper<Scalars['Float']>;
  FileQueryArgs: FileQueryArgs;
  Tokens: ResolverTypeWrapper<Tokens>;
  Account: ResolverTypeWrapper<Account>;
  SearchResults: ResolverTypeWrapper<SearchResults>;
  Query: ResolverTypeWrapper<{}>;
  Mutation: ResolverTypeWrapper<{}>;
}>;

/** Mapping between all available schema types and the resolvers parents */
export type ResolversParentTypes = ResolversObject<{
  PFSInput: PfsInput;
  String: Scalars['String'];
  CronInput: CronInput;
  GitInput: GitInput;
  Input: Input;
  ID: Scalars['ID'];
  Pipeline: Pipeline;
  Int: Scalars['Int'];
  Boolean: Scalars['Boolean'];
  InputPipeline: InputPipeline;
  Repo: Repo;
  Pach: Pach;
  Job: Job;
  Node: Node;
  Link: Link;
  Dag: Dag;
  DagQueryArgs: DagQueryArgs;
  JobQueryArgs: JobQueryArgs;
  Project: Project;
  Timestamp: Timestamp;
  File: File;
  Float: Scalars['Float'];
  FileQueryArgs: FileQueryArgs;
  Tokens: Tokens;
  Account: Account;
  SearchResults: SearchResults;
  Query: {};
  Mutation: {};
}>;

export type PfsInputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['PFSInput'] = ResolversParentTypes['PFSInput']
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  repo?: Resolver<ResolversTypes['Repo'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type CronInputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['CronInput'] = ResolversParentTypes['CronInput']
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  repo?: Resolver<ResolversTypes['Repo'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type GitInputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['GitInput'] = ResolversParentTypes['GitInput']
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  url?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type InputResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Input'] = ResolversParentTypes['Input']
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

export type PipelineStateResolvers = EnumResolverSignature<
  {
    PIPELINE_STARTING?: any;
    PIPELINE_RUNNING?: any;
    PIPELINE_RESTARTING?: any;
    PIPELINE_FAILURE?: any;
    PIPELINE_PAUSED?: any;
    PIPELINE_STANDBY?: any;
    PIPELINE_CRASHING?: any;
  },
  ResolversTypes['PipelineState']
>;

export type JobStateResolvers = EnumResolverSignature<
  {
    JOB_STARTING?: any;
    JOB_RUNNING?: any;
    JOB_FAILURE?: any;
    JOB_SUCCESS?: any;
    JOB_KILLED?: any;
    JOB_MERGING?: any;
    JOB_EGRESSING?: any;
  },
  ResolversTypes['JobState']
>;

export type ProjectStatusResolvers = EnumResolverSignature<
  {HEALTHY?: any; UNHEALTHY?: any},
  ResolversTypes['ProjectStatus']
>;

export type PipelineResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Pipeline'] = ResolversParentTypes['Pipeline']
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
  inputs?: Resolver<Array<ResolversTypes['Input']>, ParentType, ContextType>;
  description?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type InputPipelineResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['InputPipeline'] = ResolversParentTypes['InputPipeline']
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type RepoResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Repo'] = ResolversParentTypes['Repo']
> = ResolversObject<{
  name?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  sizeInBytes?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  description?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  isPipelineOutput?: Resolver<
    ResolversTypes['Boolean'],
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type PachResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Pach'] = ResolversParentTypes['Pach']
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type JobResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Job'] = ResolversParentTypes['Job']
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type NodeResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Node'] = ResolversParentTypes['Node']
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  type?: Resolver<ResolversTypes['NodeType'], ParentType, ContextType>;
  state?: Resolver<
    Maybe<ResolversTypes['PipelineState']>,
    ParentType,
    ContextType
  >;
  access?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type LinkResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Link'] = ResolversParentTypes['Link']
> = ResolversObject<{
  source?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  target?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  state?: Resolver<Maybe<ResolversTypes['JobState']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type DagResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Dag'] = ResolversParentTypes['Dag']
> = ResolversObject<{
  nodes?: Resolver<Array<ResolversTypes['Node']>, ParentType, ContextType>;
  links?: Resolver<Array<ResolversTypes['Link']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type ProjectResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Project'] = ResolversParentTypes['Project']
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  status?: Resolver<ResolversTypes['ProjectStatus'], ParentType, ContextType>;
  description?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  createdAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type FileTypeResolvers = EnumResolverSignature<
  {RESERVED?: any; DIR?: any; FILE?: any},
  ResolversTypes['FileType']
>;

export type TimestampResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Timestamp'] = ResolversParentTypes['Timestamp']
> = ResolversObject<{
  seconds?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  nanos?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type FileResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['File'] = ResolversParentTypes['File']
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
  type?: Resolver<ResolversTypes['FileType'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type TokensResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Tokens'] = ResolversParentTypes['Tokens']
> = ResolversObject<{
  pachToken?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  idToken?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type AccountResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Account'] = ResolversParentTypes['Account']
> = ResolversObject<{
  id?: Resolver<ResolversTypes['ID'], ParentType, ContextType>;
  email?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  name?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type SearchResultsResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['SearchResults'] = ResolversParentTypes['SearchResults']
> = ResolversObject<{
  pipelines?: Resolver<
    Array<Maybe<ResolversTypes['Pipeline']>>,
    ParentType,
    ContextType
  >;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type QueryResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Query'] = ResolversParentTypes['Query']
> = ResolversObject<{
  projects?: Resolver<
    Array<ResolversTypes['Project']>,
    ParentType,
    ContextType
  >;
  pipelines?: Resolver<
    Array<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
  repos?: Resolver<Array<ResolversTypes['Repo']>, ParentType, ContextType>;
  jobs?: Resolver<
    Array<ResolversTypes['Job']>,
    ParentType,
    ContextType,
    RequireFields<QueryJobsArgs, 'args'>
  >;
  dag?: Resolver<
    ResolversTypes['Dag'],
    ParentType,
    ContextType,
    RequireFields<QueryDagArgs, 'args'>
  >;
  dags?: Resolver<
    Array<ResolversTypes['Dag']>,
    ParentType,
    ContextType,
    RequireFields<QueryDagsArgs, 'args'>
  >;
  files?: Resolver<
    Array<ResolversTypes['File']>,
    ParentType,
    ContextType,
    RequireFields<QueryFilesArgs, 'args'>
  >;
  account?: Resolver<ResolversTypes['Account'], ParentType, ContextType>;
  searchResults?: Resolver<
    ResolversTypes['SearchResults'],
    ParentType,
    ContextType,
    RequireFields<QuerySearchResultsArgs, 'query'>
  >;
}>;

export type MutationResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Mutation'] = ResolversParentTypes['Mutation']
> = ResolversObject<{
  exchangeCode?: Resolver<
    ResolversTypes['Tokens'],
    ParentType,
    ContextType,
    RequireFields<MutationExchangeCodeArgs, 'code'>
  >;
}>;

export type Resolvers<ContextType = Context> = ResolversObject<{
  PFSInput?: PfsInputResolvers<ContextType>;
  CronInput?: CronInputResolvers<ContextType>;
  GitInput?: GitInputResolvers<ContextType>;
  Input?: InputResolvers<ContextType>;
  PipelineState?: PipelineStateResolvers;
  JobState?: JobStateResolvers;
  ProjectStatus?: ProjectStatusResolvers;
  Pipeline?: PipelineResolvers<ContextType>;
  InputPipeline?: InputPipelineResolvers<ContextType>;
  Repo?: RepoResolvers<ContextType>;
  Pach?: PachResolvers<ContextType>;
  Job?: JobResolvers<ContextType>;
  Node?: NodeResolvers<ContextType>;
  Link?: LinkResolvers<ContextType>;
  Dag?: DagResolvers<ContextType>;
  Project?: ProjectResolvers<ContextType>;
  FileType?: FileTypeResolvers;
  Timestamp?: TimestampResolvers<ContextType>;
  File?: FileResolvers<ContextType>;
  Tokens?: TokensResolvers<ContextType>;
  Account?: AccountResolvers<ContextType>;
  SearchResults?: SearchResultsResolvers<ContextType>;
  Query?: QueryResolvers<ContextType>;
  Mutation?: MutationResolvers<ContextType>;
}>;

/**
 * @deprecated
 * Use "Resolvers" root object instead. If you wish to get "IResolvers", add "typesPrefix: I" to your config.
 */
export type IResolvers<ContextType = Context> = Resolvers<ContextType>;

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

export type GetDagQueryVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagQuery = {__typename?: 'Query'} & {
  dag: {__typename?: 'Dag'} & {
    nodes: Array<
      {__typename?: 'Node'} & Pick<Node, 'name' | 'type' | 'access' | 'state'>
    >;
    links: Array<
      {__typename?: 'Link'} & Pick<Link, 'source' | 'target' | 'state'>
    >;
  };
};

export type GetDagsQueryVariables = Exact<{
  args: DagQueryArgs;
}>;

export type GetDagsQuery = {__typename?: 'Query'} & {
  dags: Array<
    {__typename?: 'Dag'} & {
      nodes: Array<
        {__typename?: 'Node'} & Pick<Node, 'name' | 'type' | 'access' | 'state'>
      >;
      links: Array<
        {__typename?: 'Link'} & Pick<Link, 'source' | 'target' | 'state'>
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
    > & {
        committed?: Maybe<
          {__typename?: 'Timestamp'} & Pick<Timestamp, 'nanos' | 'seconds'>
        >;
      }
  >;
};

export type GetJobsQueryVariables = Exact<{
  args: JobQueryArgs;
}>;

export type GetJobsQuery = {__typename?: 'Query'} & {
  jobs: Array<{__typename?: 'Job'} & Pick<Job, 'id' | 'state' | 'createdAt'>>;
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

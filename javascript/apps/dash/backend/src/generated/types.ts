/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable @typescript-eslint/ban-types */
import {GraphQLResolveInfo} from 'graphql';

import {Context} from 'lib/types';
export type Maybe<T> = T | null;
export type Exact<T extends {[key: string]: unknown}> = {[K in keyof T]: T[K]};
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

export enum PipelineState {
  Starting = 'STARTING',
  Running = 'RUNNING',
  Restarting = 'RESTARTING',
  Failure = 'FAILURE',
  Paused = 'PAUSED',
  Standby = 'STANDBY',
  Crashing = 'CRASHING',
}

export enum JobState {
  Starting = 'STARTING',
  Running = 'RUNNING',
  Failure = 'FAILURE',
  Success = 'SUCCESS',
  Killed = 'KILLED',
  Merging = 'MERGING',
  Egressing = 'EGRESSING',
}

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
  numOfJobsMerging: Scalars['Int'];
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
  pipeline: Pipeline;
  parentJobId?: Maybe<Scalars['String']>;
  startedAt: Scalars['Int'];
  finishedAt?: Maybe<Scalars['Int']>;
  state: JobState;
  reason?: Maybe<Scalars['String']>;
  outputRepo: Repo;
  input: Input;
};

export enum NodeType {
  Pipeline = 'PIPELINE',
  Repo = 'REPO',
}

export type Node = {
  __typename?: 'Node';
  name: Scalars['String'];
  type: NodeType;
  error: Scalars['Boolean'];
  access: Scalars['Boolean'];
};

export type Link = {
  __typename?: 'Link';
  source: Scalars['Int'];
  target: Scalars['Int'];
  error?: Maybe<Scalars['Boolean']>;
  active?: Maybe<Scalars['Boolean']>;
};

export type Dag = {
  __typename?: 'Dag';
  nodes: Array<Node>;
  links: Array<Link>;
};

export type DagQueryArgs = {
  projectId: Scalars['ID'];
};

export type Query = {
  __typename?: 'Query';
  pipelines: Array<Pipeline>;
  repos: Array<Repo>;
  jobs: Array<Job>;
  dag: Dag;
  dags: Array<Dag>;
};

export type QueryDagArgs = {
  args: DagQueryArgs;
};

export type QueryDagsArgs = {
  args: DagQueryArgs;
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
  Query: ResolverTypeWrapper<{}>;
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
  Query: {};
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
  numOfJobsMerging?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
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
  pipeline?: Resolver<ResolversTypes['Pipeline'], ParentType, ContextType>;
  parentJobId?: Resolver<
    Maybe<ResolversTypes['String']>,
    ParentType,
    ContextType
  >;
  startedAt?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  finishedAt?: Resolver<Maybe<ResolversTypes['Int']>, ParentType, ContextType>;
  state?: Resolver<ResolversTypes['JobState'], ParentType, ContextType>;
  reason?: Resolver<Maybe<ResolversTypes['String']>, ParentType, ContextType>;
  outputRepo?: Resolver<ResolversTypes['Repo'], ParentType, ContextType>;
  input?: Resolver<ResolversTypes['Input'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type NodeResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Node'] = ResolversParentTypes['Node']
> = ResolversObject<{
  name?: Resolver<ResolversTypes['String'], ParentType, ContextType>;
  type?: Resolver<ResolversTypes['NodeType'], ParentType, ContextType>;
  error?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  access?: Resolver<ResolversTypes['Boolean'], ParentType, ContextType>;
  __isTypeOf?: IsTypeOfResolverFn<ParentType, ContextType>;
}>;

export type LinkResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Link'] = ResolversParentTypes['Link']
> = ResolversObject<{
  source?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  target?: Resolver<ResolversTypes['Int'], ParentType, ContextType>;
  error?: Resolver<Maybe<ResolversTypes['Boolean']>, ParentType, ContextType>;
  active?: Resolver<Maybe<ResolversTypes['Boolean']>, ParentType, ContextType>;
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

export type QueryResolvers<
  ContextType = Context,
  ParentType extends ResolversParentTypes['Query'] = ResolversParentTypes['Query']
> = ResolversObject<{
  pipelines?: Resolver<
    Array<ResolversTypes['Pipeline']>,
    ParentType,
    ContextType
  >;
  repos?: Resolver<Array<ResolversTypes['Repo']>, ParentType, ContextType>;
  jobs?: Resolver<Array<ResolversTypes['Job']>, ParentType, ContextType>;
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
}>;

export type Resolvers<ContextType = Context> = ResolversObject<{
  PFSInput?: PfsInputResolvers<ContextType>;
  CronInput?: CronInputResolvers<ContextType>;
  GitInput?: GitInputResolvers<ContextType>;
  Input?: InputResolvers<ContextType>;
  Pipeline?: PipelineResolvers<ContextType>;
  InputPipeline?: InputPipelineResolvers<ContextType>;
  Repo?: RepoResolvers<ContextType>;
  Pach?: PachResolvers<ContextType>;
  Job?: JobResolvers<ContextType>;
  Node?: NodeResolvers<ContextType>;
  Link?: LinkResolvers<ContextType>;
  Dag?: DagResolvers<ContextType>;
  Query?: QueryResolvers<ContextType>;
}>;

/**
 * @deprecated
 * Use "Resolvers" root object instead. If you wish to get "IResolvers", add "typesPrefix: I" to your config.
 */
export type IResolvers<ContextType = Context> = Resolvers<ContextType>;

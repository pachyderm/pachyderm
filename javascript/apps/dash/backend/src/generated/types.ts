/* eslint-disable @typescript-eslint/naming-convention */
export type Maybe<T> = T | null;
export type Exact<T extends {[key: string]: unknown}> = {[K in keyof T]: T[K]};
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

export type Query = {
  __typename?: 'Query';
  pipelines: Array<Pipeline>;
  repos: Array<Repo>;
  jobs: Array<Job>;
};

export type Maybe<T> = T | null;
export type Exact<T extends { [key: string]: any }> = { [K in keyof T]: T[K] };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
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

export type Hello = {
  __typename?: 'Hello';
  id: Scalars['ID'];
  message: Scalars['String'];
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
  Pfs = 'PFS',
  Cron = 'CRON',
  Git = 'GIT'
}

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

export enum JobState {
  Starting = 'STARTING',
  Running = 'RUNNING',
  Failure = 'FAILURE',
  Success = 'SUCCESS',
  Killed = 'KILLED',
  Merging = 'MERGING',
  Egressing = 'EGRESSING'
}

export type Pach = {
  __typename?: 'Pach';
  id: Scalars['ID'];
};

export type PfsInput = {
  __typename?: 'PFSInput';
  name: Scalars['String'];
  repo: Repo;
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

export enum PipelineState {
  Starting = 'STARTING',
  Running = 'RUNNING',
  Restarting = 'RESTARTING',
  Failure = 'FAILURE',
  Paused = 'PAUSED',
  Standby = 'STANDBY',
  Crashing = 'CRASHING'
}

export type Query = {
  __typename?: 'Query';
  hello: Hello;
  pipelines: Array<Pipeline>;
  repos: Array<Repo>;
  jobs: Array<Job>;
};

export enum OriginKind {
  User = 'USER',
  Auto = 'AUTO',
  Fsck = 'FSCK'
}

export enum FileType {
  Reserved = 'RESERVED',
  File = 'FILE',
  Dir = 'DIR'
}

export type File = {
  __typename?: 'File';
  path: Scalars['String'];
  fileType: FileType,
  sizeInBytes: Scalars['Int'],
  committedAt: Scalars['Int'];
}

export type Commit = {
  __typename?: 'Commit';
  id: Scalars['String'];
  origin: OriginKind;
  description?: Scalars['String'];
  parentCommitId?: Scalars['String'],
  childCommitIds: Array<Scalars['String']>;
  started: Scalars['Int'],
  finished: Scalars['Int'],
  sizeInBytes: Scalars['Int'],
  files: Array<File>
};

export type Branch = {
  __typename?: 'Branch';
  name: Scalars['String'];
  commits: Array<Commit>;
};

export type Repo = {
  __typename?: 'Repo';
  name: Scalars['ID'];
  createdAt: Scalars['Int'];
  sizeInBytes: Scalars['Int'];
  description: Scalars['String'];
  isPipelineOutput: Scalars['Boolean'];
  branches: Array<Branch>;
};


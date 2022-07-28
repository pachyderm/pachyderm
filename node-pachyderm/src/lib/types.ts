import {ChannelCredentials, Metadata} from '@grpc/grpc-js';

import {
  BranchObject,
  CommitObject,
  CommitSetObject,
  RepoObject,
  TriggerObject,
} from '../builders/pfs';
import {
  CommitState,
  CreateBranchRequest,
  CreateRepoRequest,
  DeleteBranchRequest,
  DeleteRepoRequest,
  FinishCommitRequest,
  InspectCommitSetRequest,
  ListBranchRequest,
  ListCommitRequest,
  OriginKind,
  StartCommitRequest,
  SubscribeCommitRequest,
} from '../proto/pfs/pfs_pb';

export interface GRPCPlugin {
  onCall?: (args: {requestName: string}) => void;
  onCompleted?: (args: {requestName: string}) => void;
  onError?: (args: {error: unknown; requestName: string}) => void;
}

export type ServiceHandlerFunction = (...args: never[]) => Promise<unknown>;
export type ServiceDefinition = Record<string, ServiceHandlerFunction>;

export interface ServiceArgs {
  pachdAddress: string;
  channelCredentials: ChannelCredentials;
  credentialMetadata: Metadata;
  plugins?: GRPCPlugin[];
}

export type JobSetQueryArgs = {
  id: string;
  projectId: string;
  details?: boolean;
};

export type JobQueryArgs = {
  id: string;
  projectId: string;
  pipelineName: string;
  wait?: boolean;
};

export type GetLogsRequestArgs = {
  pipelineName?: string;
  jobId?: string;
  since?: number;
  follow?: boolean;
};

export type ListCommitArgs = {
  repo: RepoObject;
  number?: ListCommitRequest.AsObject['number'];
  reverse?: ListCommitRequest.AsObject['reverse'];
  all?: ListCommitRequest.AsObject['all'];
  originKind?: OriginKind;
  from?: CommitObject;
  to?: CommitObject;
};

export type InspectCommitSetArgs = {
  commitSet: CommitSetObject;
  wait?: InspectCommitSetRequest.AsObject['wait'];
};

export type StartCommitRequestArgs = {
  branch: BranchObject;
  description?: StartCommitRequest.AsObject['description'];
  parent?: CommitObject;
};

export type FinishCommitRequestArgs = {
  commit: CommitObject;
  error?: FinishCommitRequest.AsObject['error'];
  force?: FinishCommitRequest.AsObject['force'];
  description?: FinishCommitRequest.AsObject['description'];
};

export type CreateRepoRequestArgs = {
  repo: RepoObject;
  description?: CreateRepoRequest.AsObject['description'];
  update?: CreateRepoRequest.AsObject['update'];
};

export type DeleteRepoRequestArgs = {
  repo: RepoObject;
  force?: DeleteRepoRequest.AsObject['force'];
};

export type InspectCommitRequestArgs = {
  wait: CommitState;
  commit: CommitObject;
};

export type SubscribeCommitRequestArgs = {
  repo: RepoObject;
  branch?: SubscribeCommitRequest.AsObject['branch'];
  state?: CommitState;
  all?: SubscribeCommitRequest.AsObject['all'];
  originKind?: OriginKind;
  from?: CommitObject;
};

export type ListBranchRequestArgs = {
  repo: RepoObject;
  reverse?: ListBranchRequest.AsObject['reverse'];
};

export type CreateBranchArgs = {
  head?: CommitObject;
  branch?: BranchObject;
  provenance: BranchObject[];
  trigger?: TriggerObject;
  newCommitSet: CreateBranchRequest.AsObject['newCommitSet'];
};

export type DeleteBranchRequestArgs = {
  branch: BranchObject;
  force?: DeleteBranchRequest.AsObject['force'];
};

export type InspectDatumRequestArgs = {
  id: string;
  jobId: string;
  pipelineName: string;
};

export type ListDatumsRequestArgs = {
  jobId: string;
  pipelineName: string;
};

export type RenewFileSetRequestArgs = {
  fileSetId: string;
  duration?: number;
};

export type AddFileSetRequestArgs = {
  fileSetId: string;
  commit: CommitObject;
};

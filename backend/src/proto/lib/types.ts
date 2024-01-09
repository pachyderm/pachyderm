import {Metadata} from '@grpc/grpc-js';

import {BranchObject, CommitObject} from '../builders/pfs';
import {FinishCommitRequest, StartCommitRequest} from '../proto/pfs/pfs_pb';

export interface GRPCPlugin {
  onCall?: (args: {requestName: string}) => void;
  onCompleted?: (args: {requestName: string}) => void;
  onError?: (args: {error: unknown; requestName: string}) => void;
}

type ServiceHandlerFunction = (...args: never[]) => Promise<unknown>;
export type ServiceDefinition = Record<string, ServiceHandlerFunction>;

export interface ServiceArgs {
  credentialMetadata: Metadata;
  plugins?: GRPCPlugin[];
}

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

export type RenewFileSetRequestArgs = {
  fileSetId: string;
  duration?: number;
};

export type AddFileSetRequestArgs = {
  fileSetId: string;
  commit: CommitObject;
};

import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import Logger from 'bunyan';

import {Node, JobState, Account} from '@graphqlTypes';

export interface UnauthenticatedContext {
  authToken?: string;
  pachdAddress?: string;
  log: Logger;
  account?: Account;
}

export interface Context extends UnauthenticatedContext {
  account: Account;
}

export type LinkInputData = {
  source: number;
  target: number;
  error?: boolean;
  active?: boolean;
};

export interface Vertex extends Node {
  parents: string[];
  jobState?: JobState;
}

export interface ServiceArgs {
  pachdAddress: string;
  channelCredentials: ChannelCredentials;
  credentialMetadata: Metadata;
  log: Logger;
}
export interface GRPCPlugin {
  onCall?: (args: {requestName: string}) => void;
  onCompleted?: (args: {requestName: string}) => void;
  onError?: (args: {error: unknown; requestName: string}) => void;
}

export type ServiceHandlerFunction = (...args: never[]) => Promise<unknown>;
export type ServiceDefinition = Record<string, ServiceHandlerFunction>;

import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {CronInput, PFSInput} from '@pachyderm/proto/pb/pps/pps_pb';
import {ApolloError} from 'apollo-server-errors';
import Logger from 'bunyan';
import {ElkExtendedEdge, ElkNode} from 'elkjs/lib/elk-api';

import client from '@dash-backend/grpc/client';
import {Node, JobState, Account, Link, NodeType} from '@graphqlTypes';

export type PachClient = ReturnType<typeof client>;

export interface UnauthenticatedContext {
  authToken?: string;
  pachClient: PachClient;
  pachdAddress?: string;
  log: Logger;
  host: string;
  account?: Account;
}

export interface Context extends UnauthenticatedContext {
  account: Account;
}

export interface LinkInputData
  extends ElkExtendedEdge,
    Pick<Link, 'state' | 'targetState' | 'sourceState' | 'transferring'> {}

export interface NodeInputData extends ElkNode, Omit<Node, 'x' | 'y'> {}

export interface RepoVertex extends Omit<Node, 'x' | 'y' | 'id' | 'type'> {
  parents: string[];
  type: NodeType.OUTPUT_REPO | NodeType.INPUT_REPO;
  jobState?: JobState;
}

export interface EgressVertex extends Omit<Node, 'x' | 'y' | 'id' | 'type'> {
  parents: string[];
  type: NodeType.EGRESS;
  jobState?: JobState;
}

export interface PipelineVertex extends Omit<Node, 'x' | 'y' | 'id' | 'type'> {
  parents: (
    | Omit<PFSInput.AsObject, 'commit'>
    | Omit<CronInput.AsObject, 'commit'>
  )[];
  type: NodeType.PIPELINE;
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

export class NotFoundError extends ApolloError {
  constructor(message: string) {
    super(message, 'NOT_FOUND');

    // Note: We must redefine this property, as ApolloError's
    // "name" attribute is read-only by default
    Object.defineProperty(this, 'name', {value: 'NotFoundError'});
  }
}

export type GRPCClient = typeof client;

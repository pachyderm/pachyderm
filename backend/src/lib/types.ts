import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import {ApolloError} from 'apollo-server-errors';
import Logger from 'bunyan';

import {pachydermClient} from '@dash-backend/proto';
import {Account} from '@graphqlTypes';

export type PachClient = ReturnType<typeof pachydermClient>;

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

export const NOT_FOUND_ERROR_CODE = 'NOT_FOUND';
export class NotFoundError extends ApolloError {
  constructor(message: string) {
    super(message, NOT_FOUND_ERROR_CODE);

    // Note: We must redefine this property, as ApolloError's
    // "name" attribute is read-only by default
    Object.defineProperty(this, 'name', {value: 'NotFoundError'});
  }
}

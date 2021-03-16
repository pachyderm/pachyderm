import {ChannelCredentials, Metadata} from '@grpc/grpc-js';
import Logger from 'bunyan';

import {Node, JobState} from '@graphqlTypes';

export interface Context {
  authToken?: string;
  pachdAddress?: string;
  log: Logger;
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

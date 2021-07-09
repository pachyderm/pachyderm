import {ChannelCredentials, Metadata} from '@grpc/grpc-js';

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
}

export type JobSetQueryArgs = {
  id: string;
  projectId: string;
};

export type JobQueryArgs = {
  id: string;
  projectId: string;
  pipelineName: string;
};

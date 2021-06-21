import {ClientReadableStream} from '@grpc/grpc-js';
import {PubSub} from 'apollo-server-express';
import identity from 'lodash/identity';
import noop from 'lodash/noop';

import withCancel from './withCancel';

type withStreamParameters<T, U> = {
  triggerName: string;
  stream: ClientReadableStream<U>;
  timeout?: number; //TODO: Not implemented yet
  onData: (data: U) => T | U;
  onError?: (error: Error) => void;
  onCancel?: () => void;
  onClose?: () => void;
};

const pubsub = new PubSub();

const withStream = <T, U>({
  triggerName,
  stream,
  timeout, //TODO: Not implemented yet
  onData = (data) => identity(data),
  onError = noop,
  onClose = noop,
  onCancel = noop,
}: withStreamParameters<T, U>) => {
  const asyncIterator = pubsub.asyncIterator(triggerName);

  stream.on('data', (chunk: U) => {
    pubsub.publish(triggerName, onData(chunk));
  });
  stream.on('error', (error) => {
    onError(error);
    stream.destroy();
  });
  stream.on('close', () => {
    onClose();
    stream.destroy();
  });

  if (!asyncIterator.return) {
    asyncIterator.return = () =>
      Promise.resolve({value: undefined, done: true});
  }

  const handleCancel = () => {
    onCancel();
    stream.destroy();
  };

  return withCancel(asyncIterator, handleCancel);
};

export default withStream;

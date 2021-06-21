import {PubSub} from 'apollo-server-express';
import noop from 'lodash/noop';

import withCancel from './withCancel';

type withSubscriptionParameters<T> = {
  triggerNames: [string, string];
  resolver: () => Promise<T | undefined>;
  intervalKey: string;
  onCancel?: () => void;
  interval?: number;
};

type IntervalRecord = {
  interval: NodeJS.Timeout;
  count: number;
};

const intervalMap: Record<string, IntervalRecord> = {};
const pubsub = new PubSub();

const withSubscription = <T>({
  triggerNames,
  resolver,
  intervalKey,
  onCancel = noop,
  interval = 3000,
}: withSubscriptionParameters<T>) => {
  const handleCancel = () => {
    intervalMap[intervalKey].count -= 1;
    if (intervalMap[intervalKey].count === 0) {
      clearInterval(intervalMap[intervalKey].interval);
      delete intervalMap[intervalKey];
    }
    onCancel();
  };

  const asyncIterator = pubsub.asyncIterator(triggerNames);
  const iteratorWithCancel = withCancel(asyncIterator, handleCancel);

  // get initial result so first request does not have to wait
  process.nextTick(async () => {
    intervalMap[intervalKey].count += 1;
    const result = await resolver();
    if (result) {
      pubsub.publish(triggerNames[0], result);
    }
  });

  // initialize polling
  if (!intervalMap[intervalKey]) {
    intervalMap[intervalKey] = {
      interval: setInterval(async () => {
        const result = await resolver();
        if (result) pubsub.publish(triggerNames[1], result);
      }, interval),
      count: 0,
    };
  }

  return iteratorWithCancel;
};

export default withSubscription;

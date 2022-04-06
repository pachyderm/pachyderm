import {PubSub} from 'graphql-subscriptions';
import noop from 'lodash/noop';

import {SUBSCRIPTION_INTERVAL} from '@dash-backend/constants/subscription';

import withCancel from './withCancel';

type withSubscriptionParameters<T> = {
  triggerName: string;
  resolver: () => Promise<T | undefined>;
  intervalKey: string;
  onCancel?: () => void;
  interval?: number;
};

type IntervalRecord = {
  interval: NodeJS.Timeout;
};

const intervalMap: Record<string, IntervalRecord> = {};
const pubsub = new PubSub();

const withSubscription = <T>({
  triggerName,
  resolver,
  intervalKey,
  onCancel = noop,
  interval = SUBSCRIPTION_INTERVAL,
}: withSubscriptionParameters<T>) => {
  const handleCancel = () => {
    if (intervalMap[intervalKey]) {
      clearInterval(intervalMap[intervalKey].interval);
      delete intervalMap[intervalKey];
    }
    onCancel();
  };

  const asyncIterator = pubsub.asyncIterator<T>(triggerName);
  const iteratorWithCancel = withCancel<T>(asyncIterator, handleCancel);

  const getData = async () => {
    try {
      const result = await resolver();
      if (result) pubsub.publish(triggerName, result);
    } catch (err) {
      pubsub.publish(triggerName, err);
    }
  };

  const startInterval = () => {
    intervalMap[intervalKey] = {
      interval: setInterval(async () => {
        await getData();
      }, interval),
    };
  };

  // get initial result so first request does not have to wait
  process.nextTick(async () => {
    await getData();
  });

  //initialize polling
  startInterval();

  return iteratorWithCancel;
};

export default withSubscription;

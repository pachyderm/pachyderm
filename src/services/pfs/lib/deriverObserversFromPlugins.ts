import {GRPCPlugin} from '../../../lib/types';

export const deriveObserversFromPlugins = (plugins: GRPCPlugin[]) => {
  return {
    onCallObservers: plugins.flatMap((p) => (p.onCall ? [p.onCall] : [])),
    onCompleteObservers: plugins.flatMap((p) =>
      p.onCompleted ? [p.onCompleted] : [],
    ),
    onErrorObservers: plugins.flatMap((p) => (p.onError ? [p.onError] : [])),
  };
};

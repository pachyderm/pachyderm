import {split, HttpLink} from '@apollo/client';
import {getMainDefinition} from '@apollo/client/utilities';

import {WebSocketLink} from './websocketLink';

export const splitLink = () => {
  const webSocketLink = new WebSocketLink({
    url: `${window.location.protocol.startsWith('https:') ? 'wss' : 'ws'}://${
      window.location.hostname
    }${process.env.REACT_APP_BACKEND_SUBSCRIPTIONS_PREFIX}`,
    connectionParams: () => {
      return {
        'auth-token': window.localStorage.getItem('auth-token') || '',
        'id-token': window.localStorage.getItem('id-token') || '',
      };
    },
    lazyCloseTimeout: 5000,
  });
  const httpLink = new HttpLink({
    uri: process.env.REACT_APP_BACKEND_GRAPHQL_PREFIX,
  });

  return {
    split: split(
      ({query}) => {
        const definition = getMainDefinition(query);
        return (
          definition.kind === 'OperationDefinition' &&
          definition.operation === 'subscription'
        );
      },
      webSocketLink,
      httpLink,
    ),
    restartWebsocket: webSocketLink.restart,
  };
};

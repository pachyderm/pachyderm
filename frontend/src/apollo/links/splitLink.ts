import {split, HttpLink} from '@apollo/client';
import {getMainDefinition} from '@apollo/client/utilities';

import {getSubscriptionsPrefix} from '@dash-frontend/lib/runtimeVariables';

import {WebSocketLink} from './websocketLink';

export const splitLink = () => {
  const isProd = process.env.NODE_ENV === 'production';
  const wsProtocol = window.location.protocol.startsWith('https:')
    ? 'wss'
    : 'ws';
  const wsUrl = isProd
    ? `${window.location.host}/graphql`
    : `${window.location.hostname}${getSubscriptionsPrefix()}`;

  const webSocketLink = new WebSocketLink({
    url: `${wsProtocol}://${wsUrl}`,
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

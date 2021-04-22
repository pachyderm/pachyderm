import {split, HttpLink} from '@apollo/client';
import {WebSocketLink} from '@apollo/client/link/ws';
import {getMainDefinition} from '@apollo/client/utilities';
import {SubscriptionClient} from 'subscriptions-transport-ws/dist/client';

const wsClientCreate = () => {
  return new SubscriptionClient(
    `ws://${window.location.hostname}${process.env.REACT_APP_BACKEND_SUBSCRIPTIONS_PREFIX}`,
    {
      reconnect: true,
      connectionParams: () => {
        return {
          'auth-token': window.localStorage.getItem('auth-token') || '',
          'id-token': window.localStorage.getItem('id-token') || '',
        };
      },
    },
  );
};

export const splitLink = () => {
  const webSocketClient = wsClientCreate();
  const webSocketLink = new WebSocketLink(webSocketClient);
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
    webSocketClient,
  };
};

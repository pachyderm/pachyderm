import {ApolloProvider} from '@apollo/client';
import React, {useEffect} from 'react';
import {useHistory} from 'react-router';

import createApolloClient from 'apollo';

import SubscriptionClientProvider from './SubscriptionClientProvider/SubscriptionClientProvider';

const DashApolloProvider: React.FC = ({children}) => {
  const browserHistory = useHistory();
  const {client, webSocketClient} = createApolloClient(browserHistory);

  // Apollo doesn't close websockets on client.stop()
  useEffect(() => {
    return () => webSocketClient.close();
  }, [webSocketClient]);

  return (
    <ApolloProvider client={client}>
      <SubscriptionClientProvider client={webSocketClient}>
        {children}
      </SubscriptionClientProvider>
    </ApolloProvider>
  );
};

export default DashApolloProvider;

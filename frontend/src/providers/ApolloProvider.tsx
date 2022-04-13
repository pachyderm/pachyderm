import {ApolloProvider} from '@apollo/client';
import React, {useEffect, useRef} from 'react';

import {useErrorLink} from '@dash-frontend/apollo/links/errorLink';
import useLoggedIn from '@dash-frontend/hooks/useLoggedIn';
import createApolloClient from 'apollo';

const DashApolloProvider: React.FC = ({children}) => {
  const {loggedIn} = useLoggedIn();
  const prevLoggedInRef = useRef(loggedIn);
  const errorLink = useErrorLink();
  const {client, restartWebsocket} = createApolloClient(errorLink);

  useEffect(() => {
    if (loggedIn !== prevLoggedInRef.current) {
      prevLoggedInRef.current = loggedIn;
      restartWebsocket();
    }
  }, [loggedIn, restartWebsocket]);

  return <ApolloProvider client={client}>{children}</ApolloProvider>;
};

export default DashApolloProvider;

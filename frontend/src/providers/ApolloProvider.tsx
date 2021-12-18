import {ApolloProvider} from '@apollo/client';
import React, {useEffect, useRef} from 'react';
import {useHistory} from 'react-router';

import useLoggedIn from '@dash-frontend/hooks/useLoggedIn';
import createApolloClient from 'apollo';

const DashApolloProvider: React.FC = ({children}) => {
  const browserHistory = useHistory();
  const {loggedIn} = useLoggedIn();
  const prevLoggedInRef = useRef(loggedIn);
  const {client, restartWebsocket} = createApolloClient(browserHistory);

  useEffect(() => {
    if (loggedIn !== prevLoggedInRef.current) {
      prevLoggedInRef.current = loggedIn;
      restartWebsocket();
    }
  }, [loggedIn, restartWebsocket]);

  return <ApolloProvider client={client}>{children}</ApolloProvider>;
};

export default DashApolloProvider;

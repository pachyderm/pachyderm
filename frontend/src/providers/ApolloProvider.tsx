import {ApolloProvider} from '@apollo/client';
import {GraphQLError} from 'graphql';
import React, {useEffect, useRef, useState} from 'react';

import {useErrorLink} from '@dash-frontend/apollo/links/errorLink';
import useLoggedIn from '@dash-frontend/hooks/useLoggedIn';
import ErrorView from '@dash-frontend/views/ErrorView/GraphQLErrorView';
import createApolloClient from 'apollo';

const DashApolloProvider = ({children}: {children?: React.ReactNode}) => {
  const {loggedIn} = useLoggedIn();
  const prevLoggedInRef = useRef(loggedIn);
  const [apolloError, setApolloError] = useState<GraphQLError>();
  const errorLink = useErrorLink(setApolloError);
  const {client, restartWebsocket} = createApolloClient(errorLink);

  useEffect(() => {
    if (loggedIn !== prevLoggedInRef.current) {
      prevLoggedInRef.current = loggedIn;
      restartWebsocket();
    }
  }, [loggedIn, restartWebsocket]);

  if (apolloError) {
    return (
      <ApolloProvider client={client}>
        <ErrorView graphQLError={apolloError} />
      </ApolloProvider>
    );
  }

  return <ApolloProvider client={client}>{children}</ApolloProvider>;
};

export default DashApolloProvider;

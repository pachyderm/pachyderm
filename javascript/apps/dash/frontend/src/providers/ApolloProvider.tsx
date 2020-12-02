import {ApolloProvider} from '@apollo/client';
import React from 'react';
import {useHistory} from 'react-router';

import createApolloClient from 'apollo';

const DashApolloProvider: React.FC = ({children}) => {
  const browserHistory = useHistory();
  const client = createApolloClient(browserHistory);

  return <ApolloProvider client={client}>{children}</ApolloProvider>;
};

export default DashApolloProvider;

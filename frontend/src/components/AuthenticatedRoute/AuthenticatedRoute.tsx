import React from 'react';

import ErrorView from '@dash-frontend/views/ErrorView';

import useAuthenticatedRoute from './hooks/useAuthenticatedRoute';

const AuthenticatedRoute = <T,>(
  Component: React.ComponentType<T>,
): React.ComponentType<T> => {
  const WrappedComponent: React.FC<T> = (props) => {
    const {error, loggedIn} = useAuthenticatedRoute();

    if (error) {
      return <ErrorView graphQLError={error.graphQLErrors[0]} />;
    }

    if (loggedIn) {
      return <Component {...props} />;
    }

    return null;
  };

  return WrappedComponent;
};

export default AuthenticatedRoute;

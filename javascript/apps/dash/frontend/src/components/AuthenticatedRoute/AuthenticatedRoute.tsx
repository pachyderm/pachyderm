import React from 'react';
import {Redirect} from 'react-router';

import useAuthenticatedRoute from './hooks/useAuthenticatedRoute';

const AuthenticatedRoute = <T extends unknown>(
  Component: React.ComponentType<T>,
): React.ComponentType<T> => {
  const WrappedComponent: React.FC<T> = (props) => {
    const {error, loggedIn, redirectSearchString} = useAuthenticatedRoute();

    // TODO: Think about what to show the user here
    if (error) {
      return (
        <Redirect
          to={{
            pathname: '/error',
            search: redirectSearchString,
          }}
        />
      );
    }

    if (loggedIn) {
      return <Component {...props} />;
    }

    return null;
  };

  return WrappedComponent;
};

export default AuthenticatedRoute;

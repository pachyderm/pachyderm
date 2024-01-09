import React from 'react';

import CommunityEditionBannerProvider from '@dash-frontend/components/CommunityEditionBanner/CommunityEditionBannerProvider';
import AuthenticatedRouteErrorView from '@dash-frontend/views/ErrorView/AuthenticatedRouteErrorView';

import useAuthenticatedRoute from './hooks/useAuthenticatedRoute';

const AuthenticatedRoute = <T,>(Component: React.ComponentType<T>) => {
  const WrappedComponent: React.FC<T & JSX.IntrinsicAttributes> = (props) => {
    const {error, loggedIn} = useAuthenticatedRoute();

    if (error) {
      return <AuthenticatedRouteErrorView error={error} />;
    }

    if (loggedIn) {
      return (
        <>
          <Component {...props} />
          <CommunityEditionBannerProvider />
        </>
      );
    }

    return null;
  };

  return WrappedComponent;
};

export default AuthenticatedRoute;
